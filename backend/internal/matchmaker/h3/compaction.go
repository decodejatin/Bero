package h3

import "sort"

// CompactCells compresses a set of H3 cells by merging children into parents.
// When all 7 children of a parent cell are present, they are replaced by the parent.
// This dramatically reduces memory usage for dense areas (e.g., entire cities).
//
// Example: 7 resolution-9 cells covering a hex → 1 resolution-8 cell
func CompactCells(cells []Cell) []Cell {
	if len(cells) == 0 {
		return nil
	}

	// Group cells by their parent
	parentToChildren := make(map[Cell][]Cell)
	for _, cell := range cells {
		parent := CellToParent(cell)
		if parent == "" {
			continue
		}
		parentToChildren[parent] = append(parentToChildren[parent], cell)
	}

	var result []Cell
	compactedParents := make(map[Cell]bool)

	for parent, children := range parentToChildren {
		if len(children) >= 7 {
			// All children present — compact to parent
			result = append(result, parent)
			compactedParents[parent] = true
		}
	}

	// Add remaining un-compacted cells
	for _, cell := range cells {
		parent := CellToParent(cell)
		if !compactedParents[parent] {
			result = append(result, cell)
		}
	}

	// Sort for deterministic output
	sort.Slice(result, func(i, j int) bool {
		return string(result[i]) < string(result[j])
	})

	return result
}

// UncompactCells expands compacted cells back to the target resolution.
// Parent cells are expanded into all their children at the target resolution.
func UncompactCells(cells []Cell, targetRes int) []Cell {
	if len(cells) == 0 {
		return nil
	}

	var result []Cell
	seen := make(map[Cell]bool)

	for _, cell := range cells {
		res := CellResolution(cell)
		if res == targetRes {
			if !seen[cell] {
				result = append(result, cell)
				seen[cell] = true
			}
		} else if res < targetRes {
			// Expand to children
			children := CellToChildren(cell, targetRes)
			for _, child := range children {
				if !seen[child] {
					result = append(result, child)
					seen[child] = true
				}
			}
		}
		//  If res > targetRes, skip (can't expand to coarser resolution)
	}

	sort.Slice(result, func(i, j int) bool {
		return string(result[i]) < string(result[j])
	})

	return result
}

// CompactedSet is a memory-efficient representation of a large set of H3 cells.
// Stores cells at mixed resolutions after compaction.
type CompactedSet struct {
	cells map[Cell]struct{}
}

// NewCompactedSet creates a new compacted set from raw cells.
func NewCompactedSet(cells []Cell) *CompactedSet {
	compacted := CompactCells(cells)
	set := &CompactedSet{
		cells: make(map[Cell]struct{}, len(compacted)),
	}
	for _, c := range compacted {
		set.cells[c] = struct{}{}
	}
	return set
}

// Contains checks if a cell (at any resolution) is within the compacted set.
// It checks the cell itself and walks up to parent cells.
func (cs *CompactedSet) Contains(cell Cell) bool {
	// Check exact match
	if _, ok := cs.cells[cell]; ok {
		return true
	}

	// Walk up the resolution hierarchy
	current := cell
	for {
		parent := CellToParent(current)
		if parent == "" || parent == current {
			break
		}
		if _, ok := cs.cells[parent]; ok {
			return true
		}
		current = parent
	}

	return false
}

// Len returns the number of stored cells (after compaction).
func (cs *CompactedSet) Len() int {
	return len(cs.cells)
}

// Cells returns all stored cells.
func (cs *CompactedSet) Cells() []Cell {
	result := make([]Cell, 0, len(cs.cells))
	for c := range cs.cells {
		result = append(result, c)
	}
	return result
}

// CellToParent returns the parent cell at resolution-1.
func CellToParent(cell Cell) Cell {
	res := CellResolution(cell)
	if res <= 0 {
		return ""
	}

	// Parse the cell and reconstruct at res-1
	parsedRes, row, col := parseCell(cell)
	if parsedRes < 0 {
		return ""
	}

	// Parent cell coordinates: each parent covers ~7 children
	// Using integer division to map child to parent
	parentRow := row / 3
	parentCol := col / 3
	parentRes := parsedRes - 1

	return encodeCell(parentRes, parentRow, parentCol)
}

// CellToChildren returns all children of a cell at the target resolution.
func CellToChildren(cell Cell, targetRes int) []Cell {
	cellRes := CellResolution(cell)
	if cellRes >= targetRes {
		return []Cell{cell}
	}

	// Start with the cell and expand one level at a time
	current := []Cell{cell}
	for r := cellRes + 1; r <= targetRes; r++ {
		var next []Cell
		for _, c := range current {
			_, row, col := parseCell(c)
			// Each cell has ~7 children (hex grid approximation: 3x3 minus 2 corners)
			for dr := 0; dr < 3; dr++ {
				for dc := 0; dc < 3; dc++ {
					childRow := row*3 + dr
					childCol := col*3 + dc
					child := encodeCell(r, childRow, childCol)
					next = append(next, child)
				}
			}
		}
		current = next
	}

	return current
}
