package tests

import (
	"context"
	"testing"
	"time"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
	"github.com/stretchr/testify/assert"
)

func TestChartNode_Extraction(t *testing.T) {
	// Setup Table Data
	fields := []core.FlowField{
		{Name: "x", Type: &core.FlowType{Kind: core.FSKindInt64}},
		{Name: "y", Type: &core.FlowType{Kind: core.FSKindInt64}},
	}
	tableType := &core.FlowType{Kind: core.FSKindTable, ContainedType: &core.FlowType{Kind: core.FSKindRecord, Fields: fields}}

	rows := [][]core.FlowValueField{
		{
			{Name: "x", Value: core.NewInt64Value(1, 0)},
			{Name: "y", Value: core.NewInt64Value(10, 0)},
		},
		{
			{Name: "x", Value: core.NewInt64Value(2, 0)},
			{Name: "y", Value: core.NewInt64Value(20, 0)},
		},
	}
	tableData := core.FlowValue{Type: tableType, TableValue: rows}

	// Just reuse LineChartNode for data extraction test
	node := nodes.NewLineChartNode()
	// Fake a result on the node so ExtractChartData can read it (it reads from node result)
	// But ExtractChartData reads from `n.GetResult()`. So we must set it.

	// Issue: ExtractChartData expects the node to have run and produced a result.
	// But in a real app, the UI calls ExtractChartData *after* run? Or does it run itself?
	// The node.Run() just passes data through.

	node.SetResult(core.NodeActionResult{
		Outputs: []core.FlowValue{tableData},
	})

	data := nodes.ExtractChartData(node, "x", "y", nodes.ChartTypeLine)

	assert.Empty(t, data.Error)
	assert.Len(t, data.Points, 2)
	assert.Equal(t, float32(1), data.Points[0].X)
	assert.Equal(t, float32(10), data.Points[0].Y)
	assert.Equal(t, float32(2), data.Points[1].X)
	assert.Equal(t, float32(20), data.Points[1].Y)

	assert.Equal(t, 1.0, data.MinX)
	assert.Equal(t, 2.0, data.MaxX)
	assert.Equal(t, 10.0, data.MinY)
	assert.Equal(t, 20.0, data.MaxY)
}

func TestGraphIONodes(t *testing.T) {
	t.Run("GraphInput", func(t *testing.T) {
		node := nodes.NewGraphInputNode()
		action := node.Action.(*nodes.GraphInputAction)
		action.Value = core.NewStringValue("injected")

		// It acts like a Value node, outputs what's in Action.Value
		done := action.Run(node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Equal(t, "injected", string(res.Outputs[0].BytesValue))
	})

	t.Run("GraphOutput", func(t *testing.T) {
		node := nodes.NewGraphOutputNode()
		action := node.Action.(*nodes.GraphOutputAction)

		setupGraph(node, core.NewStringValue("result"))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Equal(t, "result", string(res.Outputs[0].BytesValue))
	})
}

func TestPromptNode(t *testing.T) {
	node := nodes.NewPromptUserNode()
	action := node.Action.(*nodes.PromptUserAction)

	// Just check instantiation properties for now
	assert.Equal(t, "Input Required", action.Title)
}

func TestAllNodesInstantiation(t *testing.T) {
	creators := []func() *core.Node{
		nodes.NewAddColumnNode,
		func() *core.Node { return nodes.NewAggregateNode("Sum") },
		nodes.NewConcatTablesNode,
		nodes.NewTransposeNode,
		nodes.NewFilterEmptyNode,
		nodes.NewLinesNode,
		nodes.NewRegexMatchNode,
		nodes.NewRegexFindAllNode,
		nodes.NewRegexReplaceNode,
		nodes.NewRegexSplitNode,
		nodes.NewJoinTextNode,
		nodes.NewSplitTextNode,
		nodes.NewCaseConvertNode,
		nodes.NewFormatStringNode,
		nodes.NewTrimSpacesNode,
		nodes.NewMinifyHTMLNode,
		nodes.NewParseTimeNode,
		nodes.NewGetVariableNode,
		nodes.NewIfElseNode,
		nodes.NewMakeDirNode,
		nodes.NewCopyFileNode,
		nodes.NewMoveFileNode,
		nodes.NewDeleteFileNode,
		func() *core.Node { return nodes.NewListFilesNode(".") },
		nodes.NewSaveFileNode,
		nodes.NewHTTPRequestNode,
		nodes.NewLineChartNode,
		nodes.NewBarChartNode,
		nodes.NewScatterPlotNode,
		nodes.NewGraphInputNode,
		nodes.NewGraphOutputNode,
		nodes.NewPromptUserNode,
		nodes.NewGetMousePositionNode,
		nodes.NewWaitForClickNode,
		// Newly Added
		nodes.NewConvertNode,
		nodes.NewJsonQueryNode,
		nodes.NewXmlQueryNode,
		nodes.NewExtractColumnNode,
		nodes.NewFormulaNode,
		func() *core.Node { return nodes.NewLoadFileNode("test.txt") },
		nodes.NewMapNode,
		nodes.NewSelectColumnsNode,
		nodes.NewSortNode,
		nodes.NewGateNode,
		nodes.NewMergeNode,
		func() *core.Node { return nodes.NewRunProcessNode("echo hello") },
	}

	for _, creator := range creators {
		n := creator()
		assert.NotNil(t, n)
		assert.NotNil(t, n.Action)
	}
}
