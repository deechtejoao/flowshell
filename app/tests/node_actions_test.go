package tests

import (
	"context"
	"testing"
	"time"

	"github.com/bvisness/flowshell/app"
)

func runAction(t *testing.T, node *app.Node) app.NodeActionResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if actionCtx, ok := node.Action.(app.NodeActionWithContext); ok {
		done := actionCtx.RunContext(ctx, node)
		select {
		case res := <-done:
			return res
		case <-ctx.Done():
			t.Fatal("Node execution timed out")
			return app.NodeActionResult{}
		}
	} else {
		done := node.Action.Run(node)
		select {
		case res := <-done:
			return res
		case <-ctx.Done():
			t.Fatal("Node execution timed out")
			return app.NodeActionResult{}
		}
	}
}

func TestSortAction(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		reverse  bool
		expected []string
	}{
		{
			name:     "Sort Strings Ascending",
			input:    []string{"c", "a", "b"},
			reverse:  false,
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "Sort Strings Descending",
			input:    []string{"c", "a", "b"},
			reverse:  true,
			expected: []string{"c", "b", "a"},
		},
		{
			name:     "Already Sorted",
			input:    []string{"1", "2", "3"},
			reverse:  false,
			expected: []string{"1", "2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := app.NewSortNode()
			action := node.Action.(*app.SortAction)
			action.Reverse = tt.reverse

			// Prepare input
			var inputValues []app.FlowValue
			for _, s := range tt.input {
				inputValues = append(inputValues, app.NewStringValue(s))
			}
			inputList := app.NewListValue(app.FlowType{Kind: app.FSKindBytes}, inputValues)

			valueNode := app.NewValueNode(inputList)

			// Setup graph
			g := app.NewGraph()
			g.AddNode(valueNode)
			g.AddNode(node)
			g.Wires = []*app.Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

			// Pre-calculate input node result
			valueNode.SetResult(app.NodeActionResult{
				Outputs: []app.FlowValue{inputList},
			})

			// Run
			action.UpdateAndValidate(node)
			res := runAction(t, node)

			if res.Err != nil {
				t.Fatalf("Node execution failed: %v", res.Err)
			}

			if len(res.Outputs) != 1 {
				t.Fatalf("Expected 1 output, got %d", len(res.Outputs))
			}

			outVal := res.Outputs[0]
			if outVal.Type.Kind != app.FSKindList {
				t.Fatalf("Expected List output, got %v", outVal.Type.Kind)
			}

			if len(outVal.ListValue) != len(tt.expected) {
				t.Fatalf("Expected %d items, got %d", len(tt.expected), len(outVal.ListValue))
			}

			for i, expected := range tt.expected {
				got := string(outVal.ListValue[i].BytesValue)
				if got != expected {
					t.Errorf("Item %d: expected %s, got %s", i, expected, got)
				}
			}
		})
	}
}

func TestConvertAction_Scalar(t *testing.T) {
	// Test Int -> String
	node := app.NewConvertNode()
	action := node.Action.(*app.ConvertAction)
	action.TargetKind = app.FSKindBytes

	val := app.NewInt64Value(123, 0)
	valueNode := app.NewValueNode(val)

	// Setup graph
	g := app.NewGraph()
	g.AddNode(valueNode)
	g.AddNode(node)
	g.Wires = []*app.Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.SetResult(app.NodeActionResult{
		Outputs: []app.FlowValue{val},
	})

	// Run
	action.UpdateAndValidate(node) // Important for Convert node to set output type
	res := runAction(t, node)

	if res.Err != nil {
		t.Fatalf("Node execution failed: %v", res.Err)
	}

	if len(res.Outputs) != 1 {
		t.Fatalf("Expected 1 output, got %d", len(res.Outputs))
	}

	outVal := res.Outputs[0]
	if outVal.Type.Kind != app.FSKindBytes {
		t.Errorf("Expected output kind Bytes, got %v", outVal.Type.Kind)
	}
	if string(outVal.BytesValue) != "123" {
		t.Errorf("Expected '123', got '%s'", string(outVal.BytesValue))
	}
}

func TestConvertAction_IgnoreErrors(t *testing.T) {
	// Test String (Invalid) -> Int64 with IgnoreErrors
	node := app.NewConvertNode()
	action := node.Action.(*app.ConvertAction)
	action.TargetKind = app.FSKindInt64
	action.TargetDropdown.SelectByValue(app.FSKindInt64)
	action.IgnoreErrors = true

	val := app.NewStringValue("invalid")
	valueNode := app.NewValueNode(val)

	// Setup graph
	g := app.NewGraph()
	g.AddNode(valueNode)
	g.AddNode(node)
	g.Wires = []*app.Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.SetResult(app.NodeActionResult{
		Outputs: []app.FlowValue{val},
	})

	// Run
	action.UpdateAndValidate(node)
	res := runAction(t, node)

	if res.Err != nil {
		t.Fatalf("Expected no error, got: %v", res.Err)
	}

	if len(res.Outputs) != 1 {
		t.Fatalf("Expected 1 output, got %d", len(res.Outputs))
	}

	outVal := res.Outputs[0]
	if outVal.Type.Kind != app.FSKindInt64 {
		t.Errorf("Expected output kind Int64, got %v", outVal.Type.Kind)
	}
	if outVal.Int64Value != 0 {
		t.Errorf("Expected 0, got %d", outVal.Int64Value)
	}
}

func TestExtractColumnAction(t *testing.T) {
	node := app.NewExtractColumnNode()
	action := node.Action.(*app.ExtractColumnAction)
	action.Column = "col1"

	// Create table input
	row1 := []app.FlowValueField{
		{Name: "col1", Value: app.NewStringValue("a")},
		{Name: "col2", Value: app.NewInt64Value(1, 0)},
	}
	row2 := []app.FlowValueField{
		{Name: "col1", Value: app.NewStringValue("b")},
		{Name: "col2", Value: app.NewInt64Value(2, 0)},
	}
	tableType := app.FlowType{
		Kind: app.FSKindTable,
		ContainedType: &app.FlowType{
			Kind: app.FSKindRecord,
			Fields: []app.FlowField{
				{Name: "col1", Type: &app.FlowType{Kind: app.FSKindBytes}},
				{Name: "col2", Type: &app.FlowType{Kind: app.FSKindInt64}},
			},
		},
	}
	val := app.FlowValue{
		Type:       &tableType,
		TableValue: [][]app.FlowValueField{row1, row2},
	}

	valueNode := app.NewValueNode(val)

	// Setup graph
	g := app.NewGraph()
	g.AddNode(valueNode)
	g.AddNode(node)
	g.Wires = []*app.Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.SetResult(app.NodeActionResult{
		Outputs: []app.FlowValue{val},
	})

	// Run
	action.UpdateAndValidate(node)
	res := runAction(t, node)

	if res.Err != nil {
		t.Fatalf("Node execution failed: %v", res.Err)
	}

	if len(res.Outputs) != 1 {
		t.Fatalf("Expected 1 output, got %d", len(res.Outputs))
	}

	outVal := res.Outputs[0]
	if outVal.Type.Kind != app.FSKindList {
		t.Errorf("Expected output kind List, got %v", outVal.Type.Kind)
	}

	if len(outVal.ListValue) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(outVal.ListValue))
	}

	if string(outVal.ListValue[0].BytesValue) != "a" {
		t.Errorf("Expected 'a', got '%s'", string(outVal.ListValue[0].BytesValue))
	}
	if string(outVal.ListValue[1].BytesValue) != "b" {
		t.Errorf("Expected 'b', got '%s'", string(outVal.ListValue[1].BytesValue))
	}
}

func TestSelectColumnsAction(t *testing.T) {
	node := app.NewSelectColumnsNode()
	action := node.Action.(*app.SelectColumnsAction)
	action.SelectedColumns = []string{"col1"}

	// Create table input
	row1 := []app.FlowValueField{
		{Name: "col1", Value: app.NewStringValue("a")},
		{Name: "col2", Value: app.NewInt64Value(1, 0)},
	}
	row2 := []app.FlowValueField{
		{Name: "col1", Value: app.NewStringValue("b")},
		{Name: "col2", Value: app.NewInt64Value(2, 0)},
	}
	tableType := app.FlowType{
		Kind: app.FSKindTable,
		ContainedType: &app.FlowType{
			Kind: app.FSKindRecord,
			Fields: []app.FlowField{
				{Name: "col1", Type: &app.FlowType{Kind: app.FSKindBytes}},
				{Name: "col2", Type: &app.FlowType{Kind: app.FSKindInt64}},
			},
		},
	}
	val := app.FlowValue{
		Type:       &tableType,
		TableValue: [][]app.FlowValueField{row1, row2},
	}

	valueNode := app.NewValueNode(val)

	// Setup graph
	g := app.NewGraph()
	g.AddNode(valueNode)
	g.AddNode(node)
	g.Wires = []*app.Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.SetResult(app.NodeActionResult{
		Outputs: []app.FlowValue{val},
	})

	// Run
	action.UpdateAndValidate(node)
	res := runAction(t, node)

	if res.Err != nil {
		t.Fatalf("Node execution failed: %v", res.Err)
	}

	if len(res.Outputs) != 1 {
		t.Fatalf("Expected 1 output, got %d", len(res.Outputs))
	}

	outVal := res.Outputs[0]
	if outVal.Type.Kind != app.FSKindTable {
		t.Errorf("Expected output kind Table, got %v", outVal.Type.Kind)
	}

	if len(outVal.TableValue) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(outVal.TableValue))
	}

	// Check fields
	if len(outVal.Type.ContainedType.Fields) != 1 {
		t.Fatalf("Expected 1 column, got %d", len(outVal.Type.ContainedType.Fields))
	}
	if outVal.Type.ContainedType.Fields[0].Name != "col1" {
		t.Errorf("Expected column 'col1', got '%s'", outVal.Type.ContainedType.Fields[0].Name)
	}

	// Check values
	if string(outVal.TableValue[0][0].Value.BytesValue) != "a" {
		t.Errorf("Row 0: Expected 'a', got '%s'", string(outVal.TableValue[0][0].Value.BytesValue))
	}
	if string(outVal.TableValue[1][0].Value.BytesValue) != "b" {
		t.Errorf("Row 1: Expected 'b', got '%s'", string(outVal.TableValue[1][0].Value.BytesValue))
	}
}

func TestSelectColumnsAction_NilContainedType(t *testing.T) {
	node := app.NewSelectColumnsNode()
	action := node.Action.(*app.SelectColumnsAction)
	action.SelectedColumns = []string{"col1"}

	// Table with no ContainedType (e.g. dynamic or empty)
	tableType := app.FlowType{
		Kind:          app.FSKindTable,
		ContainedType: nil,
	}
	val := app.FlowValue{
		Type:       &tableType,
		TableValue: [][]app.FlowValueField{},
	}

	valueNode := app.NewValueNode(val)

	// Setup graph
	g := app.NewGraph()
	g.AddNode(valueNode)
	g.AddNode(node)
	g.Wires = []*app.Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.SetResult(app.NodeActionResult{
		Outputs: []app.FlowValue{val},
	})

	// Run
	// Should not panic
	action.UpdateAndValidate(node)
	res := runAction(t, node)

	if res.Err != nil {
		t.Fatalf("Node execution failed: %v", res.Err)
	}

	outVal := res.Outputs[0]
	if outVal.Type.ContainedType != nil && len(outVal.Type.ContainedType.Fields) > 0 {
		t.Errorf("Expected empty fields for nil ContainedType input")
	}
}
