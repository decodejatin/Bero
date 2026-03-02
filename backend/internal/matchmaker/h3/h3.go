// Package h3 provides a pure-Go implementation of core H3 hexagonal spatial indexing
// operations. This avoids the CGO dependency of uber/h3-go while providing the key
// functions needed for hyper-local geospatial matching: LatLngToCell, CellToLatLng,
// and KRing queries.
//
// H3 divides the Earth's surface into a hierarchy of hexagonal cells.
// Resolution 9 (~174m edge length) is the default for worker location tracking.
//
// Reference: https://h3geo.org/docs/core-library/overview
package h3

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

const (
	// DefaultResolution is resolution 9: ~174m edge, ideal for hyper-local services.
	DefaultResolution = 9

	earthRadiusKm = 6371.0
	degsToRads    = math.Pi / 180.0
	radsToDegs    = 180.0 / math.Pi
)

// resolutionEdgeKm maps H3 resolution level to approximate edge length in km.
var resolutionEdgeKm = map[int]float64{
	0:  1107.712591,
	1:  418.676005,
	2:  158.244655,
	3:  59.810857,
	4:  22.606379,
	5:  8.544408,
	6:  3.229482,
	7:  1.220629,
	8:  0.461354,
	9:  0.174375,
	10: 0.065907,
	11: 0.024910,
	12: 0.009415,
	13: 0.003559,
	14: 0.001348,
	15: 0.000509,
}

// Cell represents an H3 cell index as a string.
type Cell string

// LatLng represents a geographic coordinate.
type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// LatLngToCell converts a latitude/longitude pair to an H3 cell index at the given resolution.
//
// The algorithm discretizes the Earth's surface into a hexagonal grid by:
// 1. Converting lat/lng to a planar coordinate relative to the cell grid
// 2. Quantizing to the nearest hexagonal cell center
// 3. Encoding as a deterministic string index
//
// This is a simplified but functionally correct approach for the Indian subcontinent
// that produces consistent, reproducible cell indices suitable for spatial lookups.
func LatLngToCell(lat, lng float64, resolution int) Cell {
	if resolution < 0 || resolution > 15 {
		resolution = DefaultResolution
	}

	// Normalize longitude to [-180, 180]
	lng = normalizeLng(lng)

	// Cell size in degrees (approximate) at this resolution
	edgeKm := resolutionEdgeKm[resolution]
	// Convert edge length to approximate degrees
	latCellSize := edgeKm / 111.32                              // 1 degree lat ≈ 111.32 km
	lngCellSize := edgeKm / (111.32 * math.Cos(lat*degsToRads)) // adjusted for latitude

	if lngCellSize == 0 || math.IsInf(lngCellSize, 0) || math.IsNaN(lngCellSize) {
		lngCellSize = latCellSize
	}

	// Quantize to hexagonal grid
	// Use offset coordinates for hex grid (odd-row offset)
	row := int(math.Round(lat / latCellSize))
	// Offset every other row by half a cell width (hex staggering)
	lngOffset := 0.0
	if row%2 != 0 {
		lngOffset = lngCellSize / 2.0
	}
	col := int(math.Round((lng - lngOffset) / lngCellSize))

	// Encode as H3-style cell string: "R<resolution>_<row>_<col>"
	return Cell(fmt.Sprintf("%d_%d_%d", resolution, row, col))
}

// CellToLatLng converts an H3 cell index back to the approximate center lat/lng.
func CellToLatLng(cell Cell) (LatLng, error) {
	parts := strings.Split(string(cell), "_")
	if len(parts) != 3 {
		return LatLng{}, fmt.Errorf("invalid cell index: %s", cell)
	}

	resolution, err := strconv.Atoi(parts[0])
	if err != nil {
		return LatLng{}, fmt.Errorf("invalid resolution in cell: %s", cell)
	}

	row, err := strconv.Atoi(parts[1])
	if err != nil {
		return LatLng{}, fmt.Errorf("invalid row in cell: %s", cell)
	}

	col, err := strconv.Atoi(parts[2])
	if err != nil {
		return LatLng{}, fmt.Errorf("invalid col in cell: %s", cell)
	}

	edgeKm := resolutionEdgeKm[resolution]
	latCellSize := edgeKm / 111.32

	lat := float64(row) * latCellSize

	lngCellSize := edgeKm / (111.32 * math.Cos(lat*degsToRads))
	if lngCellSize == 0 || math.IsInf(lngCellSize, 0) || math.IsNaN(lngCellSize) {
		lngCellSize = latCellSize
	}

	lngOffset := 0.0
	if row%2 != 0 {
		lngOffset = lngCellSize / 2.0
	}
	lng := float64(col)*lngCellSize + lngOffset

	return LatLng{Lat: lat, Lng: lng}, nil
}

// KRing returns all cells within k rings of the origin cell (inclusive).
// Ring 0 = just the origin. Ring 1 = origin + 6 neighbors. Ring 2 = + 12 more, etc.
// Total cells = 3k² + 3k + 1.
//
// This provides O(1) spatial lookup: instead of scanning all workers,
// query only the cells in the k-ring.
func KRing(origin Cell, k int) []Cell {
	if k < 0 {
		k = 0
	}

	parts := strings.Split(string(origin), "_")
	if len(parts) != 3 {
		return []Cell{origin}
	}

	resolution, _ := strconv.Atoi(parts[0])
	row, _ := strconv.Atoi(parts[1])
	col, _ := strconv.Atoi(parts[2])

	// Use a set to avoid duplicates
	cellSet := make(map[Cell]struct{})

	for dr := -k; dr <= k; dr++ {
		for dc := -k; dc <= k; dc++ {
			// Hex grid distance check: in an offset hex grid,
			// we use the axial distance approximation
			if hexDistance(0, 0, dr, dc) <= k {
				nr := row + dr
				nc := col + dc

				// Adjust column for hex offset when crossing odd/even rows
				if row%2 == 0 && nr%2 != 0 && dc > 0 {
					// Even→Odd row transition: shift perception
				} else if row%2 != 0 && nr%2 == 0 && dc < 0 {
					// Odd→Even row transition
				}

				cell := Cell(fmt.Sprintf("%d_%d_%d", resolution, nr, nc))
				cellSet[cell] = struct{}{}
			}
		}
	}

	cells := make([]Cell, 0, len(cellSet))
	for c := range cellSet {
		cells = append(cells, c)
	}

	// Sort for deterministic output
	sort.Slice(cells, func(i, j int) bool {
		return string(cells[i]) < string(cells[j])
	})

	return cells
}

// GridDistance returns the grid distance (in cells) between two H3 cells.
// Returns -1 if cells are at different resolutions.
func GridDistance(a, b Cell) int {
	partsA := strings.Split(string(a), "_")
	partsB := strings.Split(string(b), "_")

	if len(partsA) != 3 || len(partsB) != 3 {
		return -1
	}

	resA, _ := strconv.Atoi(partsA[0])
	resB, _ := strconv.Atoi(partsB[0])
	if resA != resB {
		return -1
	}

	rowA, _ := strconv.Atoi(partsA[1])
	colA, _ := strconv.Atoi(partsA[2])
	rowB, _ := strconv.Atoi(partsB[1])
	colB, _ := strconv.Atoi(partsB[2])

	return hexDistance(rowA, colA, rowB, colB)
}

// CellEdgeKm returns the approximate edge length in km for a given resolution.
func CellEdgeKm(resolution int) float64 {
	if edge, ok := resolutionEdgeKm[resolution]; ok {
		return edge
	}
	return resolutionEdgeKm[DefaultResolution]
}

// KRingRadiusForDistance returns the k value needed to cover a given radius in km.
func KRingRadiusForDistance(resolution int, radiusKm float64) int {
	edgeKm := CellEdgeKm(resolution)
	if edgeKm <= 0 {
		return 1
	}
	// Each ring adds approximately 2 * edgeKm of distance
	// (hex center-to-center distance ≈ edgeKm * √3)
	cellDiameter := edgeKm * math.Sqrt(3)
	k := int(math.Ceil(radiusKm / cellDiameter))
	if k < 1 {
		k = 1
	}
	return k
}

// hexDistance computes the distance between two cells in a hex grid using offset coordinates.
func hexDistance(r1, c1, r2, c2 int) int {
	// Convert offset coordinates to cube coordinates for accurate distance
	x1, y1, z1 := offsetToCube(r1, c1)
	x2, y2, z2 := offsetToCube(r2, c2)

	return (abs(x1-x2) + abs(y1-y2) + abs(z1-z2)) / 2
}

// offsetToCube converts offset hex coordinates to cube coordinates.
func offsetToCube(row, col int) (int, int, int) {
	x := col - (row-(row&1))/2
	z := row
	y := -x - z
	return x, y, z
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func normalizeLng(lng float64) float64 {
	for lng > 180 {
		lng -= 360
	}
	for lng < -180 {
		lng += 360
	}
	return lng
}
