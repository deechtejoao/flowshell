package tests

import (
	"testing"

	"github.com/bvisness/flowshell/app"
)

func TestCopyPaste(t *testing.T) {
	// Setup graph
	app.CurrentGraph = app.NewGraph()
	n1 := &app.Node{ID: 1, Name: "Node 1", Pos: app.V2{X: 10, Y: 10}, Action: &app.TrimSpacesAction{}}
	n2 := &app.Node{ID: 2, Name: "Node 2", Pos: app.V2{X: 100, Y: 100}, Action: &app.TrimSpacesAction{}}
	// Manually add
	app.CurrentGraph.AddNode(n1)
	app.CurrentGraph.AddNode(n2)
	app.CurrentGraph.Wires = []*app.Wire{
		{StartNode: n1, StartPort: 0, EndNode: n2, EndPort: 0},
	}

	// Select both
	app.SelectNode(n1.ID, false)
	app.SelectNode(n2.ID, true)

	// Copy
	data := app.CopyToData()

	if len(data.Nodes) != 2 {
		t.Errorf("Expected 2 nodes in clipboard, got %d", len(data.Nodes))
	}
	if len(data.Wires) != 1 {
		t.Errorf("Expected 1 wire in clipboard, got %d", len(data.Wires))
	}

	// Paste
	app.PasteFromData(data)

	// Verify graph
	if len(app.CurrentGraph.Nodes) != 4 {
		t.Errorf("Expected 4 nodes in graph, got %d", len(app.CurrentGraph.Nodes))
	}
	if len(app.CurrentGraph.Wires) != 2 {
		t.Errorf("Expected 2 wires in graph, got %d", len(app.CurrentGraph.Wires))
	}

	// Verify new nodes are selected
	if len(app.SelectedNodes) != 2 {
		t.Errorf("Expected 2 selected nodes, got %d", len(app.SelectedNodes))
	}

	// Verify selection contains new nodes (ID 3 and 4)
	if !app.IsNodeSelected(3) {
		t.Errorf("Node 3 should be selected")
	}
	if !app.IsNodeSelected(4) {
		t.Errorf("Node 4 should be selected")
	}
}
