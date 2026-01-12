package app

import (
	"testing"
)

func TestCopyPaste(t *testing.T) {
	// Setup graph
	currentGraph = NewGraph()
	n1 := &Node{ID: 1, Name: "Node 1", Pos: V2{X: 10, Y: 10}, Action: &TrimSpacesAction{}}
	n2 := &Node{ID: 2, Name: "Node 2", Pos: V2{X: 100, Y: 100}, Action: &TrimSpacesAction{}}
	// Manually add
	currentGraph.AddNode(n1)
	currentGraph.AddNode(n2)
	currentGraph.Wires = []*Wire{
		{StartNode: n1, StartPort: 0, EndNode: n2, EndPort: 0},
	}

	// Select both
	SelectNode(n1.ID, false)
	SelectNode(n2.ID, true)

	// Copy
	data := CopyToData()

	if len(data.Nodes) != 2 {
		t.Errorf("Expected 2 nodes in clipboard, got %d", len(data.Nodes))
	}
	if len(data.Wires) != 1 {
		t.Errorf("Expected 1 wire in clipboard, got %d", len(data.Wires))
	}

	// Paste
	PasteFromData(data)

	// Verify graph
	if len(currentGraph.Nodes) != 4 {
		t.Errorf("Expected 4 nodes in graph, got %d", len(currentGraph.Nodes))
	}
	if len(currentGraph.Wires) != 2 {
		t.Errorf("Expected 2 wires in graph, got %d", len(currentGraph.Wires))
	}

	// Verify new nodes are selected
	if len(selectedNodes) != 2 {
		t.Errorf("Expected 2 selected nodes, got %d", len(selectedNodes))
	}

	// Verify selection contains new nodes (ID 3 and 4)
	if !IsNodeSelected(3) {
		t.Errorf("Node 3 should be selected")
	}
	if !IsNodeSelected(4) {
		t.Errorf("Node 4 should be selected")
	}
}
