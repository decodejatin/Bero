package h3

import (
	"testing"
)

func TestCompactCells_AllChildren(t *testing.T) {
	// Create 9 children (our grid produces 3x3 = 9 per parent, threshold is 7)
	// All children of a resolution-9 cell at (100, 100) = parent at res-8 (33, 33)
	parentRes8 := encodeCell(8, 33, 33)
	var children []Cell
	for dr := 0; dr < 3; dr++ {
		for dc := 0; dc < 3; dc++ {
			child := encodeCell(9, 33*3+dr, 33*3+dc)
			children = append(children, child)
		}
	}

	compacted := CompactCells(children)

	// Should compact 9 children into 1 parent
	found := false
	for _, c := range compacted {
		if c == parentRes8 {
			found = true
		}
	}

	if !found {
		t.Errorf("expected parent %s in compacted result, got %v", parentRes8, compacted)
	}
	t.Logf("Compacted %d children → %d cells", len(children), len(compacted))
}

func TestCompactCells_PartialChildren(t *testing.T) {
	// Only 3 children — should NOT compact
	var partial []Cell
	for dc := 0; dc < 3; dc++ {
		partial = append(partial, encodeCell(9, 99, 200+dc))
	}

	compacted := CompactCells(partial)

	if len(compacted) != 3 {
		t.Errorf("partial children should not compact: got %d cells", len(compacted))
	}
}

func TestCompactCells_Empty(t *testing.T) {
	if CompactCells(nil) != nil {
		t.Error("empty should return nil")
	}
}

func TestUncompactCells(t *testing.T) {
	parent := encodeCell(8, 10, 10)
	uncompacted := UncompactCells([]Cell{parent}, 9)

	if len(uncompacted) == 0 {
		t.Fatal("should produce children")
	}

	// Each parent should produce 9 children (3x3 grid)
	if len(uncompacted) != 9 {
		t.Errorf("expected 9 children, got %d", len(uncompacted))
	}

	t.Logf("Uncompacted 1 parent → %d children at res 9", len(uncompacted))
}

func TestUncompactCells_SameResolution(t *testing.T) {
	cells := []Cell{encodeCell(9, 10, 20), encodeCell(9, 10, 21)}
	result := UncompactCells(cells, 9)
	if len(result) != 2 {
		t.Errorf("same-res should return as-is: got %d", len(result))
	}
}

func TestCompactedSet_Contains(t *testing.T) {
	// Build a set from 9 children that compact to 1 parent
	var children []Cell
	for dr := 0; dr < 3; dr++ {
		for dc := 0; dc < 3; dc++ {
			children = append(children, encodeCell(9, 30+dr, 30+dc))
		}
	}

	set := NewCompactedSet(children)

	t.Logf("CompactedSet: %d stored cells from %d original", set.Len(), len(children))

	// All original children should be found (via parent walk)
	for _, child := range children {
		if !set.Contains(child) {
			t.Errorf("compacted set should contain child %s", child)
		}
	}

	// A cell in a different area should not be found
	other := encodeCell(9, 500, 500)
	if set.Contains(other) {
		t.Error("should not contain unrelated cell")
	}
}

func TestCompactedSet_MemorySavings(t *testing.T) {
	// Simulate a "city" with many cells at resolution 9
	var cells []Cell
	for row := 0; row < 30; row++ {
		for col := 0; col < 30; col++ {
			cells = append(cells, encodeCell(9, row, col))
		}
	}

	set := NewCompactedSet(cells)
	savings := 100.0 * (1.0 - float64(set.Len())/float64(len(cells)))

	t.Logf("Original: %d cells, Compacted: %d cells (%.1f%% savings)",
		len(cells), set.Len(), savings)

	if set.Len() >= len(cells) {
		t.Error("compacted set should be smaller than original")
	}
}

func TestCellToParent(t *testing.T) {
	child := encodeCell(9, 99, 99)
	parent := CellToParent(child)

	if parent == "" {
		t.Fatal("parent should not be empty")
	}

	parentRes := CellResolution(parent)
	if parentRes != 8 {
		t.Errorf("parent should be res 8, got %d", parentRes)
	}

	t.Logf("Child %s → Parent %s", child, parent)
}

func TestCellToParent_Resolution0(t *testing.T) {
	cell := encodeCell(0, 0, 0)
	parent := CellToParent(cell)
	if parent != "" {
		t.Errorf("res-0 cell should have no parent, got %s", parent)
	}
}
