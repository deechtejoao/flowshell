package tests

import (
	"context"
	"testing"
	"time"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
	"github.com/stretchr/testify/assert"
)

func TestAggregateNode(t *testing.T) {
	t.Run("Min Int64", func(t *testing.T) {
		node := nodes.NewAggregateNode("Min")
		action := node.Action.(*nodes.AggregateAction)

		vals := []core.FlowValue{
			core.NewInt64Value(10, 0),
			core.NewInt64Value(5, 0),
			core.NewInt64Value(20, 0),
		}
		list := core.NewListValue(core.FlowType{Kind: core.FSKindInt64}, vals)

		setupGraph(node, list)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Len(t, res.Outputs, 1)
		assert.Equal(t, int64(5), res.Outputs[0].Int64Value)
	})

	t.Run("Max Float64", func(t *testing.T) {
		node := nodes.NewAggregateNode("Max")
		action := node.Action.(*nodes.AggregateAction)

		vals := []core.FlowValue{
			core.NewFloat64Value(10.5, 2),
			core.NewFloat64Value(5.5, 2),
			core.NewFloat64Value(20.5, 2),
		}
		list := core.NewListValue(core.FlowType{Kind: core.FSKindFloat64}, vals)

		setupGraph(node, list)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Len(t, res.Outputs, 1)
		assert.Equal(t, 20.5, res.Outputs[0].Float64Value)
	})

	t.Run("Mean Int64", func(t *testing.T) {
		node := nodes.NewAggregateNode("Mean")
		action := node.Action.(*nodes.AggregateAction)

		vals := []core.FlowValue{
			core.NewInt64Value(10, 0),
			core.NewInt64Value(20, 0),
		}
		list := core.NewListValue(core.FlowType{Kind: core.FSKindInt64}, vals)

		setupGraph(node, list)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Len(t, res.Outputs, 1)
		// (10 + 20) / 2 = 15
		assert.Equal(t, int64(15), res.Outputs[0].Int64Value)
	})
}

func TestConcatTablesNode(t *testing.T) {
	node := nodes.NewConcatTablesNode()
	action := node.Action.(*nodes.ConcatTablesAction)

	// Create two identical tables
	fields := []core.FlowField{{Name: "col1", Type: &core.FlowType{Kind: core.FSKindBytes}}}
	recType := &core.FlowType{Kind: core.FSKindRecord, Fields: fields}
	tableType := &core.FlowType{Kind: core.FSKindTable, ContainedType: recType}

	row1 := []core.FlowValueField{{Name: "col1", Value: core.NewStringValue("A")}}
	row2 := []core.FlowValueField{{Name: "col1", Value: core.NewStringValue("B")}}

	t1 := core.FlowValue{Type: tableType, TableValue: [][]core.FlowValueField{row1}}
	t2 := core.FlowValue{Type: tableType, TableValue: [][]core.FlowValueField{row2}}

	// Setup graph
	// ConcatTables dynamically adds inputs. Default has 1. We need 2.
	node.InputPorts = append(node.InputPorts, core.NodePort{Name: "Table 2", Type: core.NewAnyTableType()})

	// Manually setup wiring since setupGraph helper might not handle dynamic ports easily (logic is simple though)
	g := core.NewGraph()
	g.AddNode(node)

	inputNode1 := createInputNode(t1)
	inputNode2 := createInputNode(t2)
	g.AddNode(inputNode1)
	g.AddNode(inputNode2)
	g.AddWire(inputNode1, 0, node, 0)
	g.AddWire(inputNode2, 0, node, 1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := action.RunContext(ctx, node)
	res := <-done

	assert.NoError(t, res.Err)
	assert.Len(t, res.Outputs, 1)
	assert.Len(t, res.Outputs[0].TableValue, 2)
	assert.Equal(t, "A", string(res.Outputs[0].TableValue[0][0].Value.BytesValue))
	assert.Equal(t, "B", string(res.Outputs[0].TableValue[1][0].Value.BytesValue))
}

func TestTransposeNode(t *testing.T) {
	node := nodes.NewTransposeNode()
	action := node.Action.(*nodes.TransposeAction)

	// 2x3 Table
	// c1 c2 c3
	// a  b  c
	// d  e  f
	fields := []core.FlowField{
		{Name: "c1", Type: &core.FlowType{Kind: core.FSKindBytes}},
		{Name: "c2", Type: &core.FlowType{Kind: core.FSKindBytes}},
		{Name: "c3", Type: &core.FlowType{Kind: core.FSKindBytes}},
	}
	recType := &core.FlowType{Kind: core.FSKindRecord, Fields: fields}
	tableType := &core.FlowType{Kind: core.FSKindTable, ContainedType: recType}

	row1 := []core.FlowValueField{
		{Name: "c1", Value: core.NewStringValue("a")},
		{Name: "c2", Value: core.NewStringValue("b")},
		{Name: "c3", Value: core.NewStringValue("c")},
	}
	row2 := []core.FlowValueField{
		{Name: "c1", Value: core.NewStringValue("d")},
		{Name: "c2", Value: core.NewStringValue("e")},
		{Name: "c3", Value: core.NewStringValue("f")},
	}
	tableVal := core.FlowValue{Type: tableType, TableValue: [][]core.FlowValueField{row1, row2}}

	setupGraph(node, tableVal)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := action.RunContext(ctx, node)
	res := <-done

	assert.NoError(t, res.Err)
	assert.Len(t, res.Outputs, 1)

	outTable := res.Outputs[0].TableValue
	// Should be 3x2
	// Row 0 Row 1
	// a     d
	// b     e
	// c     f

	assert.Len(t, outTable, 3)
	assert.Len(t, outTable[0], 2)

	assert.Equal(t, "a", string(outTable[0][0].Value.BytesValue))
	assert.Equal(t, "d", string(outTable[0][1].Value.BytesValue))

	assert.Equal(t, "b", string(outTable[1][0].Value.BytesValue))
	assert.Equal(t, "f", string(outTable[2][1].Value.BytesValue))
}

func TestFilterEmptyNode(t *testing.T) {
	node := nodes.NewFilterEmptyNode()
	action := node.Action.(*nodes.FilterEmptyAction)
	action.Column = "col1"

	fields := []core.FlowField{{Name: "col1", Type: &core.FlowType{Kind: core.FSKindBytes}}}
	recType := &core.FlowType{Kind: core.FSKindRecord, Fields: fields}
	tableType := &core.FlowType{Kind: core.FSKindTable, ContainedType: recType}

	row1 := []core.FlowValueField{{Name: "col1", Value: core.NewStringValue("kept")}}
	row2 := []core.FlowValueField{{Name: "col1", Value: core.NewStringValue("")}}

	tableVal := core.FlowValue{Type: tableType, TableValue: [][]core.FlowValueField{row1, row2}}

	setupGraph(node, tableVal)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := action.RunContext(ctx, node)
	res := <-done

	assert.NoError(t, res.Err)
	assert.Len(t, res.Outputs[0].TableValue, 1)
	assert.Equal(t, "kept", string(res.Outputs[0].TableValue[0][0].Value.BytesValue))
}

func TestMapNode(t *testing.T) {
	// 1. Create Subgraph: Input -> JoinText(" Sub") -> Output
	subB := core.NewGraphBuilder()
	input := subB.Add(nodes.NewGraphInputNode())
	output := subB.Add(nodes.NewGraphOutputNode())

	// We want to append " Sub" to the input string.
	// JoinText joins a list.
	// Or maybe just use a simpler node like CaseConvert as a proxy for "operation".
	// Let's use CaseConvert to Upper.

	convert := subB.Add(nodes.NewCaseConvertNode())
	convert.Node.Action.(*nodes.CaseConvertAction).Mode = nodes.CaseUpper

	input.To(convert).To(output)

	subGraph := subB.Graph

	// 2. Create Map Node
	node := nodes.NewMapNode()
	action := node.Action.(*nodes.MapAction)
	action.CachedGraph = subGraph // Inject subgraph

	// 3. Create Input List ["a", "b", "c"]
	vals := []core.FlowValue{
		core.NewStringValue("a"),
		core.NewStringValue("b"),
		core.NewStringValue("c"),
	}
	list := core.NewListValue(core.FlowType{Kind: core.FSKindBytes}, vals)

	setupGraph(node, list)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 4. Run
	done := action.RunContext(ctx, node)
	res := <-done

	// 5. Verify
	assert.NoError(t, res.Err)
	assert.Len(t, res.Outputs, 1)
	outList := res.Outputs[0].ListValue
	assert.Len(t, outList, 3)
	assert.Equal(t, "A", string(outList[0].BytesValue))
	assert.Equal(t, "B", string(outList[1].BytesValue))
	assert.Equal(t, "C", string(outList[2].BytesValue))
}
