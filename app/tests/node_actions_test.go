package tests

import (
	"context"
	"testing"
	"time"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
)

func runAction(t *testing.T, node *core.Node) core.NodeActionResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if actionCtx, ok := node.Action.(core.NodeActionWithContext); ok {
		done := actionCtx.RunContext(ctx, node)
		select {
		case res := <-done:
			return res
		case <-ctx.Done():
			t.Fatal("Node execution timed out")
			return core.NodeActionResult{}
		}
	} else {
		done := node.Action.Run(node)
		select {
		case res := <-done:
			return res
		case <-ctx.Done():
			t.Fatal("Node execution timed out")
			return core.NodeActionResult{}
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
			node := nodes.NewSortNode()
			action := node.Action.(*nodes.SortAction)
			action.Reverse = tt.reverse

			// Prepare input
			var inputValues []core.FlowValue
			for _, s := range tt.input {
				inputValues = append(inputValues, core.NewStringValue(s))
			}
			inputList := core.NewListValue(core.FlowType{Kind: core.FSKindBytes}, inputValues)

			valueNode := nodes.NewValueNode(inputList)

			// Setup graph
			g := core.NewGraph()
			g.AddNode(valueNode)
			g.AddNode(node)
			g.Wires = []*core.Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

			// Pre-calculate input node result
			valueNode.SetResult(core.NodeActionResult{
				Outputs: []core.FlowValue{inputList},
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
			if outVal.Type.Kind != core.FSKindList {
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
	node := nodes.NewConvertNode()
	action := node.Action.(*nodes.ConvertAction)
	action.TargetKind = core.FSKindBytes

	val := core.NewInt64Value(123, 0)
	valueNode := nodes.NewValueNode(val)

	// Setup graph
	g := core.NewGraph()
	g.AddNode(valueNode)
	g.AddNode(node)
	g.Wires = []*core.Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.SetResult(core.NodeActionResult{
		Outputs: []core.FlowValue{val},
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
	if outVal.Type.Kind != core.FSKindBytes {
		t.Errorf("Expected output kind Bytes, got %v", outVal.Type.Kind)
	}
	if string(outVal.BytesValue) != "123" {
		t.Errorf("Expected '123', got '%s'", string(outVal.BytesValue))
	}
}

func TestConvertAction_IgnoreErrors(t *testing.T) {
	// Test String (Invalid) -> Int64 with IgnoreErrors
	node := nodes.NewConvertNode()
	action := node.Action.(*nodes.ConvertAction)
	action.TargetKind = core.FSKindInt64
	action.TargetDropdown.SelectByValue(core.FSKindInt64)
	action.IgnoreErrors = true

	val := core.NewStringValue("invalid")
	valueNode := nodes.NewValueNode(val)

	// Setup graph
	g := core.NewGraph()
	g.AddNode(valueNode)
	g.AddNode(node)
	g.Wires = []*core.Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.SetResult(core.NodeActionResult{
		Outputs: []core.FlowValue{val},
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
	if outVal.Type.Kind != core.FSKindInt64 {
		t.Errorf("Expected output kind Int64, got %v", outVal.Type.Kind)
	}
	if outVal.Int64Value != 0 {
		t.Errorf("Expected 0, got %d", outVal.Int64Value)
	}
}

func TestExtractColumnAction(t *testing.T) {
	node := nodes.NewExtractColumnNode()
	action := node.Action.(*nodes.ExtractColumnAction)
	action.Column = "col1"

	// Create table input
	row1 := []core.FlowValueField{
		{Name: "col1", Value: core.NewStringValue("a")},
		{Name: "col2", Value: core.NewInt64Value(1, 0)},
	}
	row2 := []core.FlowValueField{
		{Name: "col1", Value: core.NewStringValue("b")},
		{Name: "col2", Value: core.NewInt64Value(2, 0)},
	}
	tableType := core.FlowType{
		Kind: core.FSKindTable,
		ContainedType: &core.FlowType{
			Kind: core.FSKindRecord,
			Fields: []core.FlowField{
				{Name: "col1", Type: &core.FlowType{Kind: core.FSKindBytes}},
				{Name: "col2", Type: &core.FlowType{Kind: core.FSKindInt64}},
			},
		},
	}
	val := core.FlowValue{
		Type:       &tableType,
		TableValue: [][]core.FlowValueField{row1, row2},
	}

	valueNode := nodes.NewValueNode(val)

	// Setup graph
	g := core.NewGraph()
	g.AddNode(valueNode)
	g.AddNode(node)
	g.Wires = []*core.Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.SetResult(core.NodeActionResult{
		Outputs: []core.FlowValue{val},
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
	if outVal.Type.Kind != core.FSKindList {
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
	node := nodes.NewSelectColumnsNode()
	action := node.Action.(*nodes.SelectColumnsAction)
	action.SelectedColumns = []string{"col1"}

	// Create table input
	row1 := []core.FlowValueField{
		{Name: "col1", Value: core.NewStringValue("a")},
		{Name: "col2", Value: core.NewInt64Value(1, 0)},
	}
	row2 := []core.FlowValueField{
		{Name: "col1", Value: core.NewStringValue("b")},
		{Name: "col2", Value: core.NewInt64Value(2, 0)},
	}
	tableType := core.FlowType{
		Kind: core.FSKindTable,
		ContainedType: &core.FlowType{
			Kind: core.FSKindRecord,
			Fields: []core.FlowField{
				{Name: "col1", Type: &core.FlowType{Kind: core.FSKindBytes}},
				{Name: "col2", Type: &core.FlowType{Kind: core.FSKindInt64}},
			},
		},
	}
	val := core.FlowValue{
		Type:       &tableType,
		TableValue: [][]core.FlowValueField{row1, row2},
	}

	valueNode := nodes.NewValueNode(val)

	// Setup graph
	g := core.NewGraph()
	g.AddNode(valueNode)
	g.AddNode(node)
	g.Wires = []*core.Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.SetResult(core.NodeActionResult{
		Outputs: []core.FlowValue{val},
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
	if outVal.Type.Kind != core.FSKindTable {
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
	node := nodes.NewSelectColumnsNode()
	action := node.Action.(*nodes.SelectColumnsAction)
	action.SelectedColumns = []string{"col1"}

	// Table with no ContainedType (e.g. dynamic or empty)
	tableType := core.FlowType{
		Kind:          core.FSKindTable,
		ContainedType: nil,
	}
	val := core.FlowValue{
		Type:       &tableType,
		TableValue: [][]core.FlowValueField{},
	}

	valueNode := nodes.NewValueNode(val)

	// Setup graph
	g := core.NewGraph()
	g.AddNode(valueNode)
	g.AddNode(node)
	g.Wires = []*core.Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.SetResult(core.NodeActionResult{
		Outputs: []core.FlowValue{val},
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
