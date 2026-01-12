package app

import (
	"context"
	"testing"
	"time"
)

func runAction(t *testing.T, node *Node) NodeActionResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if actionCtx, ok := node.Action.(NodeActionWithContext); ok {
		done := actionCtx.RunContext(ctx, node)
		select {
		case res := <-done:
			return res
		case <-ctx.Done():
			t.Fatal("Node execution timed out")
			return NodeActionResult{}
		}
	} else {
		done := node.Action.Run(node)
		select {
		case res := <-done:
			return res
		case <-ctx.Done():
			t.Fatal("Node execution timed out")
			return NodeActionResult{}
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
			node := NewSortNode()
			action := node.Action.(*SortAction)
			action.Reverse = tt.reverse

			// Prepare input
			var inputValues []FlowValue
			for _, s := range tt.input {
				inputValues = append(inputValues, NewStringValue(s))
			}
			inputList := NewListValue(FlowType{Kind: FSKindBytes}, inputValues)

			// Mock input port
			// Since we can't easily wire nodes in unit tests without a full graph,
			// we might need to mock GetInputValue.
			// However, GetInputValue relies on n.InputPorts and Wires.
			// Or we can manually set the input value if the node supports it.
			// But Node struct doesn't store input values directly, it looks them up via wires.
			//
			// HACK: We can create a ValueNode and wire it up, OR we can modify the node to pre-set input.
			// OR we can rely on the fact that GetInputValue checks for wires.
			//
			// Let's use the HACK of setting up global Wires temporarily.
			// This is ugly but effectively tests the Action logic which calls GetInputValue.

			valueNode := NewValueNode(inputList)

			// Setup globals
			oldWires := wires
			oldNodes := nodes
			defer func() {
				wires = oldWires
				nodes = oldNodes
			}()
			wires = nil
			nodes = []*Node{valueNode, node}
			wires = []*Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

			// Pre-calculate input node result
			valueNode.ResultAvailable = true
			valueNode.Result = NodeActionResult{
				Outputs: []FlowValue{inputList},
			}

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
			if outVal.Type.Kind != FSKindList {
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
	node := NewConvertNode()
	action := node.Action.(*ConvertAction)
	action.TargetKind = FSKindBytes

	val := NewInt64Value(123, 0)
	valueNode := NewValueNode(val)

	// Setup globals
	oldWires := wires
	oldNodes := nodes
	defer func() {
		wires = oldWires
		nodes = oldNodes
	}()
	wires = nil
	nodes = []*Node{valueNode, node}
	wires = []*Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.ResultAvailable = true
	valueNode.Result = NodeActionResult{
		Outputs: []FlowValue{val},
	}

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
	if outVal.Type.Kind != FSKindBytes {
		t.Errorf("Expected output kind Bytes, got %v", outVal.Type.Kind)
	}
	if string(outVal.BytesValue) != "123" {
		t.Errorf("Expected '123', got '%s'", string(outVal.BytesValue))
	}
}

func TestConvertAction_IgnoreErrors(t *testing.T) {
	// Test String (Invalid) -> Int64 with IgnoreErrors
	node := NewConvertNode()
	action := node.Action.(*ConvertAction)
	action.TargetKind = FSKindInt64
	action.targetDropdown.SelectByValue(FSKindInt64)
	action.IgnoreErrors = true

	val := NewStringValue("invalid")
	valueNode := NewValueNode(val)

	// Setup globals
	oldWires := wires
	oldNodes := nodes
	defer func() {
		wires = oldWires
		nodes = oldNodes
	}()
	wires = nil
	nodes = []*Node{valueNode, node}
	wires = []*Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.ResultAvailable = true
	valueNode.Result = NodeActionResult{
		Outputs: []FlowValue{val},
	}

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
	if outVal.Type.Kind != FSKindInt64 {
		t.Errorf("Expected output kind Int64, got %v", outVal.Type.Kind)
	}
	if outVal.Int64Value != 0 {
		t.Errorf("Expected 0, got %d", outVal.Int64Value)
	}
}

func TestExtractColumnAction(t *testing.T) {
	node := NewExtractColumnNode()
	action := node.Action.(*ExtractColumnAction)
	action.Column = "col1"

	// Create table input
	row1 := []FlowValueField{
		{Name: "col1", Value: NewStringValue("a")},
		{Name: "col2", Value: NewInt64Value(1, 0)},
	}
	row2 := []FlowValueField{
		{Name: "col1", Value: NewStringValue("b")},
		{Name: "col2", Value: NewInt64Value(2, 0)},
	}
	tableType := FlowType{
		Kind: FSKindTable,
		ContainedType: &FlowType{
			Kind: FSKindRecord,
			Fields: []FlowField{
				{Name: "col1", Type: &FlowType{Kind: FSKindBytes}},
				{Name: "col2", Type: &FlowType{Kind: FSKindInt64}},
			},
		},
	}
	val := FlowValue{
		Type:       &tableType,
		TableValue: [][]FlowValueField{row1, row2},
	}

	valueNode := NewValueNode(val)

	// Setup globals
	oldWires := wires
	oldNodes := nodes
	defer func() {
		wires = oldWires
		nodes = oldNodes
	}()
	wires = nil
	nodes = []*Node{valueNode, node}
	wires = []*Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.ResultAvailable = true
	valueNode.Result = NodeActionResult{
		Outputs: []FlowValue{val},
	}

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
	if outVal.Type.Kind != FSKindList {
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
	node := NewSelectColumnsNode()
	action := node.Action.(*SelectColumnsAction)
	action.SelectedColumns = []string{"col1"}

	// Create table input
	row1 := []FlowValueField{
		{Name: "col1", Value: NewStringValue("a")},
		{Name: "col2", Value: NewInt64Value(1, 0)},
	}
	row2 := []FlowValueField{
		{Name: "col1", Value: NewStringValue("b")},
		{Name: "col2", Value: NewInt64Value(2, 0)},
	}
	tableType := FlowType{
		Kind: FSKindTable,
		ContainedType: &FlowType{
			Kind: FSKindRecord,
			Fields: []FlowField{
				{Name: "col1", Type: &FlowType{Kind: FSKindBytes}},
				{Name: "col2", Type: &FlowType{Kind: FSKindInt64}},
			},
		},
	}
	val := FlowValue{
		Type:       &tableType,
		TableValue: [][]FlowValueField{row1, row2},
	}

	valueNode := NewValueNode(val)

	// Setup globals
	oldWires := wires
	oldNodes := nodes
	defer func() {
		wires = oldWires
		nodes = oldNodes
	}()
	wires = nil
	nodes = []*Node{valueNode, node}
	wires = []*Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.ResultAvailable = true
	valueNode.Result = NodeActionResult{
		Outputs: []FlowValue{val},
	}

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
	if outVal.Type.Kind != FSKindTable {
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
	node := NewSelectColumnsNode()
	action := node.Action.(*SelectColumnsAction)
	action.SelectedColumns = []string{"col1"}

	// Table with no ContainedType (e.g. dynamic or empty)
	tableType := FlowType{
		Kind:          FSKindTable,
		ContainedType: nil,
	}
	val := FlowValue{
		Type:       &tableType,
		TableValue: [][]FlowValueField{},
	}

	valueNode := NewValueNode(val)

	// Setup globals
	oldWires := wires
	oldNodes := nodes
	defer func() {
		wires = oldWires
		nodes = oldNodes
	}()
	wires = nil
	nodes = []*Node{valueNode, node}
	wires = []*Wire{{StartNode: valueNode, EndNode: node, StartPort: 0, EndPort: 0}}

	// Pre-calculate input node result
	valueNode.ResultAvailable = true
	valueNode.Result = NodeActionResult{
		Outputs: []FlowValue{val},
	}

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
