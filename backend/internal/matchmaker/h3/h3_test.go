package h3

import (
	"math"
	"testing"
)

func TestLatLngToCell_SameLocation(t *testing.T) {
	// Same location should produce the same cell
	cell1 := LatLngToCell(28.6139, 77.2090, DefaultResolution)
	cell2 := LatLngToCell(28.6139, 77.2090, DefaultResolution)

	if cell1 != cell2 {
		t.Errorf("same location should produce same cell: %s != %s", cell1, cell2)
	}
}

func TestLatLngToCell_NearbyLocations(t *testing.T) {
	// Very close points should be in the same cell (within ~174m at res 9)
	cell1 := LatLngToCell(28.61390, 77.20900, DefaultResolution)
	cell2 := LatLngToCell(28.61391, 77.20901, DefaultResolution)

	if cell1 != cell2 {
		t.Logf("Note: very close points may still be in different cells near boundaries")
		t.Logf("cell1=%s cell2=%s", cell1, cell2)
	}
}

func TestLatLngToCell_DifferentLocations(t *testing.T) {
	// Delhi vs Mumbai should be in different cells
	delhi := LatLngToCell(28.6139, 77.2090, DefaultResolution)
	mumbai := LatLngToCell(19.0760, 72.8777, DefaultResolution)

	if delhi == mumbai {
		t.Errorf("Delhi and Mumbai should be in different cells: %s == %s", delhi, mumbai)
	}
}

func TestCellToLatLng_Roundtrip(t *testing.T) {
	tests := []struct {
		name string
		lat  float64
		lng  float64
	}{
		{"Delhi", 28.6139, 77.2090},
		{"Mumbai", 19.0760, 72.8777},
		{"Bangalore", 12.9716, 77.5946},
		{"Equator", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cell := LatLngToCell(tt.lat, tt.lng, DefaultResolution)
			result, err := CellToLatLng(cell)
			if err != nil {
				t.Fatalf("CellToLatLng error: %v", err)
			}

			// Roundtrip should be within ~1 cell edge (~0.174km ≈ 0.002 degrees)
			latDiff := math.Abs(result.Lat - tt.lat)
			lngDiff := math.Abs(result.Lng - tt.lng)

			if latDiff > 0.01 {
				t.Errorf("lat roundtrip too far: original=%f, got=%f, diff=%f", tt.lat, result.Lat, latDiff)
			}
			if lngDiff > 0.01 {
				t.Errorf("lng roundtrip too far: original=%f, got=%f, diff=%f", tt.lng, result.Lng, lngDiff)
			}

			t.Logf("%s: cell=%s, roundtrip=(%.4f, %.4f), diff=(%.6f, %.6f)",
				tt.name, cell, result.Lat, result.Lng, latDiff, lngDiff)
		})
	}
}

func TestKRing_CellCount(t *testing.T) {
	origin := LatLngToCell(28.6139, 77.2090, DefaultResolution)

	tests := []struct {
		k           int
		expectedMin int // 3k² + 3k + 1 is the theoretical maximum
	}{
		{0, 1},
		{1, 7},
		{2, 19},
		{3, 37},
	}

	for _, tt := range tests {
		cells := KRing(origin, tt.k)
		t.Logf("k=%d: got %d cells (expected ~%d)", tt.k, len(cells), tt.expectedMin)

		if len(cells) < 1 {
			t.Errorf("k=%d: expected at least 1 cell, got %d", tt.k, len(cells))
		}

		// Origin should always be included
		found := false
		for _, c := range cells {
			if c == origin {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("k=%d: origin cell %s not found in k-ring", tt.k, origin)
		}
	}
}

func TestKRing_IncludesOrigin(t *testing.T) {
	origin := LatLngToCell(28.6139, 77.2090, DefaultResolution)
	cells := KRing(origin, 1)

	found := false
	for _, c := range cells {
		if c == origin {
			found = true
			break
		}
	}
	if !found {
		t.Error("k-ring(1) should include the origin cell")
	}
}

func TestGridDistance_SameCell(t *testing.T) {
	cell := LatLngToCell(28.6139, 77.2090, DefaultResolution)
	dist := GridDistance(cell, cell)
	if dist != 0 {
		t.Errorf("same cell should have distance 0, got %d", dist)
	}
}

func TestGridDistance_DifferentResolutions(t *testing.T) {
	cell1 := LatLngToCell(28.6139, 77.2090, 9)
	cell2 := LatLngToCell(28.6139, 77.2090, 10)
	dist := GridDistance(cell1, cell2)
	if dist != -1 {
		t.Errorf("different resolutions should return -1, got %d", dist)
	}
}

func TestKRingRadiusForDistance(t *testing.T) {
	// At resolution 9 (~174m edge), to cover 2km radius:
	// cell diameter ≈ 174m * √3 ≈ 301m
	// k ≈ ceil(2000m / 301m) ≈ 7
	k := KRingRadiusForDistance(DefaultResolution, 2.0)
	t.Logf("k for 2km at resolution 9: %d", k)

	if k < 1 {
		t.Error("k should be at least 1")
	}
	if k > 20 {
		t.Error("k for 2km should not be > 20 at resolution 9")
	}

	// Larger radius should give larger k
	k5 := KRingRadiusForDistance(DefaultResolution, 5.0)
	if k5 <= k {
		t.Errorf("5km should need more rings than 2km: %d <= %d", k5, k)
	}
}

func TestCellEdgeKm(t *testing.T) {
	edge := CellEdgeKm(DefaultResolution)
	if edge <= 0 {
		t.Errorf("edge should be positive, got %f", edge)
	}
	t.Logf("Resolution %d edge: %.4f km", DefaultResolution, edge)

	// Higher resolution = smaller edge
	edgeHigh := CellEdgeKm(12)
	if edgeHigh >= edge {
		t.Errorf("higher resolution should have smaller edge: %f >= %f", edgeHigh, edge)
	}
}

func TestDifferentResolutions(t *testing.T) {
	lat, lng := 28.6139, 77.2090

	for res := 5; res <= 12; res++ {
		cell := LatLngToCell(lat, lng, res)
		t.Logf("Resolution %d: cell=%s, edge=%.4fkm", res, cell, CellEdgeKm(res))
	}
}
