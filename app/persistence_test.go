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
