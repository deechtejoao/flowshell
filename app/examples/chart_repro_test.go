package examples

import (
	"testing"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
)

func TestGenerateChartDebugFlow(t *testing.T) {
	b := core.NewGraphBuilder()

	// 1. Literal Table Data
	val := core.FlowValue{
		Type: &core.FlowType{
			Kind: core.FSKindTable,
			ContainedType: &core.FlowType{
				Kind: core.FSKindRecord,
				Fields: []core.FlowField{
					{Name: "Category", Type: &core.FlowType{Kind: core.FSKindBytes}},
					{Name: "Value", Type: &core.FlowType{Kind: core.FSKindInt64}},
				},
			},
		},
		TableValue: [][]core.FlowValueField{
			{
				{Name: "Category", Value: core.NewStringValue("A")},
				{Name: "Value", Value: core.NewInt64Value(10, 0)},
			},
			{
				{Name: "Category", Value: core.NewStringValue("B")},
				{Name: "Value", Value: core.NewInt64Value(20, 0)},
			},
			{
				{Name: "Category", Value: core.NewStringValue("C")},
				{Name: "Value", Value: core.NewInt64Value(15, 0)},
			},
		},
	}

	tableNode := b.Add(nodes.NewValueNode(val)).SetPosition(100, 100)

	// 2. Bar Chart
	chartNode := b.Add(nodes.NewBarChartNode()).SetPosition(400, 100)
	if act, ok := chartNode.Node.Action.(*nodes.BarChartAction); ok {
		act.XColumn = "Category"
		act.YColumn = "Value"
	}

	// Connect Table(Value) -> Chart(Data)
	tableNode.Connect("Value", chartNode, "Data")

	// Save
	if err := core.SaveGraph("../../examples/debug_chart.flow", b.Graph); err != nil {
		t.Fatalf("Failed to save flow: %v", err)
	}
}
