package app

import (
	"os"
	"testing"
)

func TestSaveLoadGraph(t *testing.T) {
	// Setup some nodes and wires
	g := NewGraph()
	n1 := &Node{ID: 1, Name: "Node 1", Pos: V2{X: 10, Y: 10}, Action: &TrimSpacesAction{}}
	n2 := &Node{ID: 2, Name: "Node 2", Pos: V2{X: 100, Y: 100}, Action: &TrimSpacesAction{}}
	// Manually adding to ensure IDs are preserved for the test
	g.Nodes = append(g.Nodes, n1, n2)
	g.Wires = []*Wire{
		{StartNode: n1, StartPort: 0, EndNode: n2, EndPort: 0},
	}

	tmpFile := "test_graph.flow"
	defer os.Remove(tmpFile)

	// Save
	err := SaveGraph(tmpFile, g)
	if err != nil {
		t.Fatalf("SaveGraph failed: %v", err)
	}

	// Load
	loadedG, err := LoadGraph(tmpFile)
	if err != nil {
		t.Fatalf("LoadGraph failed: %v", err)
	}

	// Verify
	if len(loadedG.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(loadedG.Nodes))
	}
	if len(loadedG.Wires) != 1 {
		t.Errorf("Expected 1 wire, got %d", len(loadedG.Wires))
	}

	// Helper to find node by ID
	findNode := func(id int) *Node {
		for _, n := range loadedG.Nodes {
			if n.ID == id {
				return n
			}
		}
		return nil
	}

	ln1 := findNode(1)
	ln2 := findNode(2)

	if ln1 == nil || ln2 == nil {
		t.Fatalf("Nodes not found")
	}

	if loadedG.Wires[0].StartNode.ID != 1 || loadedG.Wires[0].EndNode.ID != 2 {
		t.Errorf("Wire connections mismatch")
	}
}

func TestSaveLoadGraphComplex(t *testing.T) {
	// Setup
	g := NewGraph()
	n1 := &Node{
		ID: 1, Name: "Lines", Pos: V2{X: 10, Y: 10},
		Action: &LinesAction{IncludeCarriageReturns: true},
	}
	n2 := &Node{
		ID: 2, Name: "Extract", Pos: V2{X: 100, Y: 100},
		Action: &ExtractColumnAction{Column: "MyCol"},
	}
	g.Nodes = append(g.Nodes, n1, n2)
	// No wires needed for this test, just testing node state serialization

	tmpFile := "test_graph_complex.flow"
	defer os.Remove(tmpFile)

	// Save
	if err := SaveGraph(tmpFile, g); err != nil {
		t.Fatalf("SaveGraph failed: %v", err)
	}

	// Load
	loadedG, err := LoadGraph(tmpFile)
	if err != nil {
		t.Fatalf("LoadGraph failed: %v", err)
	}

	// Verify
	if len(loadedG.Nodes) != 2 {
		t.Fatalf("Expected 2 nodes, got %d", len(loadedG.Nodes))
	}

	// Find nodes by ID (order might not be guaranteed, though slice usually preserves it)
	var linesNode, extractNode *Node
	for _, n := range loadedG.Nodes {
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
