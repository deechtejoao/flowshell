package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
	"github.com/stretchr/testify/assert"
)

func saveGraph(t *testing.T, g *core.Graph, filename string) {
	data, err := core.SerializeGraph(g)
	assert.NoError(t, err)

	// Create examples dir if not exists
	err = os.MkdirAll("../../examples", 0755)
	assert.NoError(t, err)

	path := filepath.Join("../../examples", filename)
	err = os.WriteFile(path, data, 0644)
	assert.NoError(t, err)
}

func TestGenerateCSVProcessingFlow(t *testing.T) {
	b := core.NewGraphBuilder()

	// 1. Load CSV
	load := b.Add(nodes.NewLoadFileNode("corpus/flute1.csv")).SetPosition(100, 100)

	// 2. Filter Empty (assuming "Avg build (us)" column)
	filter := b.Add(nodes.NewFilterEmptyNode()).SetPosition(400, 100)
	filter.Node.Action.(*nodes.FilterEmptyAction).Column = "Avg build (us)"

	// 3. Select Columns
	selectCols := b.Add(nodes.NewSelectColumnsNode()).SetPosition(700, 100)
	selectCols.Node.Action.(*nodes.SelectColumnsAction).SelectedColumns = []string{"Avg build (us)", "Avg frame (us)"}

	// 4. Save
	save := b.Add(nodes.NewSaveFileNode()).SetPosition(1000, 100)
	save.Node.Action.(*nodes.SaveFileAction).Path = "cleaned_data.csv"
	save.Node.Action.(*nodes.SaveFileAction).Format = "csv"

	// Connect: Load -> Filter -> Select -> Save
	load.To(filter).To(selectCols).To(save)

	saveGraph(t, b.Graph, "csv_pipeline.flow")
}

func TestGenerateLogAnalysisFlow(t *testing.T) {
	b := core.NewGraphBuilder()

	// 1. Load Log
	load := b.Add(nodes.NewLoadFileNode("server.log")).SetPosition(100, 200)

	// 2. Split Lines
	lines := b.Add(nodes.NewLinesNode()).SetPosition(300, 200)

	// 3. Regex Match "ERROR"
	match := b.Add(nodes.NewRegexMatchNode()).SetPosition(500, 200)

	pattern := b.Add(nodes.NewValueNode(core.NewStringValue("ERROR"))).SetPosition(500, 350)
	pattern.Connect("Value", match, "Pattern")

	// 4. Aggregate Sum (Count matches)
	count := b.Add(nodes.NewAggregateNode("Sum")).SetPosition(700, 200)

	// Connect: Load -> Lines -> Match -> Count
	load.To(lines).To(match).To(count)

	saveGraph(t, b.Graph, "log_analysis.flow")
}

func TestGenerateHTTPDashboard(t *testing.T) {
	b := core.NewGraphBuilder()

	// 1. HTTP Request
	req := b.Add(nodes.NewHTTPRequestNode()).SetPosition(100, 300)

	urlNode := b.Add(nodes.NewValueNode(core.NewStringValue("https://api.weather.gov/gridpoints/TOP/31,80/forecast"))).SetPosition(100, 450)
	urlNode.Connect("Value", req, "URL")

	// 2. JSON Query
	query := b.Add(nodes.NewJsonQueryNode()).SetPosition(400, 300)
	query.Node.Action.(*nodes.JsonQueryAction).Query = "properties.periods.0.temperature"

	// 3. Chart
	chart := b.Add(nodes.NewBarChartNode()).SetPosition(700, 300)

	// Connect
	req.To(query).To(chart)

	saveGraph(t, b.Graph, "http_dashboard.flow")
}
