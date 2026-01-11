package app

import (
	"os"
	"testing"
)

func TestSaveLoadGraph(t *testing.T) {
	// Setup some nodes and wires
	nodes = []*Node{
		{ID: 1, Name: "Node 1", Pos: V2{X: 10, Y: 10}, Action: &TrimSpacesAction{}}, // Using a simple action
		{ID: 2, Name: "Node 2", Pos: V2{X: 100, Y: 100}, Action: &TrimSpacesAction{}},
	}
	wires = []*Wire{
		{StartNode: nodes[0], StartPort: 0, EndNode: nodes[1], EndPort: 0},
	}

	tmpFile := "test_graph.flow"
	defer os.Remove(tmpFile)

	// Save
	err := SaveGraph(tmpFile)
	if err != nil {
		t.Fatalf("SaveGraph failed: %v", err)
	}

	// Clear state
	nodes = nil
	wires = nil

	// Load
	err = LoadGraph(tmpFile)
	if err != nil {
		t.Fatalf("LoadGraph failed: %v", err)
	}

	// Verify
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(nodes))
	}
	if len(wires) != 1 {
		t.Errorf("Expected 1 wire, got %d", len(wires))
	}

	if nodes[0].ID != 1 || nodes[1].ID != 2 {
		t.Errorf("Node IDs mismatch")
	}

	if wires[0].StartNode.ID != 1 || wires[0].EndNode.ID != 2 {
		t.Errorf("Wire connections mismatch")
	}
}

func TestSaveLoadGraphComplex(t *testing.T) {
	// Setup
	nodes = []*Node{
		{
			ID: 1, Name: "Lines", Pos: V2{X: 10, Y: 10},
			Action: &LinesAction{IncludeCarriageReturns: true},
		},
		{
			ID: 2, Name: "Extract", Pos: V2{X: 100, Y: 100},
			Action: &ExtractColumnAction{Column: "MyCol"},
		},
	}
	// No wires needed for this test, just testing node state serialization

	tmpFile := "test_graph_complex.flow"
	defer os.Remove(tmpFile)

	// Save
	if err := SaveGraph(tmpFile); err != nil {
		t.Fatalf("SaveGraph failed: %v", err)
	}

	// Clear
	nodes = nil

	// Load
	if err := LoadGraph(tmpFile); err != nil {
		t.Fatalf("LoadGraph failed: %v", err)
	}

	// Verify
	if len(nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(nodes))
	}

	// Find nodes by ID (order might not be guaranteed, though slice usually preserves it)
	var linesNode, extractNode *Node
	for _, n := range nodes {
		if n.ID == 1 {
			linesNode = n
		}
		if n.ID == 2 {
			extractNode = n
		}
	}

	if linesNode == nil {
		t.Fatal("Lines node not found")
	}
	if extractNode == nil {
		t.Fatal("Extract node not found")
	}

	// Check LinesAction state
	linesAction, ok := linesNode.Action.(*LinesAction)
	if !ok {
		t.Fatalf("Expected LinesAction, got %T", linesNode.Action)
	}
	if !linesAction.IncludeCarriageReturns {
		t.Error("IncludeCarriageReturns not preserved")
	}

	// Check ExtractColumnAction state
	extractAction, ok := extractNode.Action.(*ExtractColumnAction)
	if !ok {
		t.Fatalf("Expected ExtractColumnAction, got %T", extractNode.Action)
	}
	if extractAction.Column != "MyCol" {
		t.Errorf("Expected column 'MyCol', got '%s'", extractAction.Column)
	}
}
