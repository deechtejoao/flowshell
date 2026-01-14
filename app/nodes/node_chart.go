package nodes

import (
	"fmt"
	"math"
	"strconv"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/bvisness/flowshell/app/core"
)

type ChartType int

const (
	ChartTypeLine ChartType = iota
	ChartTypeBar
	ChartTypeScatter
)

type ChartRenderData struct {
	Type   ChartType
	Points []rl.Vector2
	MinX   float64
	MaxX   float64
	MinY   float64
	MaxY   float64
	Error  string
}

// Helper to extract data
func ExtractChartData(n *core.Node, xCol, yCol string, chartType ChartType) *ChartRenderData {
	renderData := &ChartRenderData{
		Type: chartType,
		MinX: math.MaxFloat64,
		MaxX: -math.MaxFloat64,
		MinY: math.MaxFloat64,
		MaxY: -math.MaxFloat64,
	}

	res, _ := n.GetResult()
	if res.Err != nil {
		renderData.Error = fmt.Sprintf("Error: %v", res.Err)
		return renderData
	}

	if len(res.Outputs) == 0 {
		return renderData
	}

	val := res.Outputs[0]
	if val.Type.Kind != core.FSKindTable {
		return renderData
	}

	xIdx := -1
	yIdx := -1

	if val.Type.ContainedType != nil {
		for i, f := range val.Type.ContainedType.Fields {
			if xCol != "" && f.Name == xCol {
				xIdx = i
			}
			if yCol != "" && f.Name == yCol {
				yIdx = i
			}
		}
	}

	if yIdx == -1 {
		if val.Type.ContainedType != nil && len(val.Type.ContainedType.Fields) > 0 {
			yIdx = 0
		}
	}

	maxPoints := 1000
	step := 1
	if len(val.TableValue) > maxPoints {
		step = len(val.TableValue) / maxPoints
	}

	for i := 0; i < len(val.TableValue); i += step {
		row := val.TableValue[i]
		var x, y float64
		var err error

		if xIdx != -1 && xIdx < len(row) {
			x, err = toFloat(row[xIdx].Value)
		} else {
			x = float64(i)
		}

		if yIdx != -1 && yIdx < len(row) {
			valY, errY := toFloat(row[yIdx].Value)
			if errY == nil {
				y = valY
			} else {
				continue
			}
		} else {
			continue
		}

		if err == nil {
			renderData.Points = append(renderData.Points, rl.Vector2{X: float32(x), Y: float32(y)})
			if x < renderData.MinX {
				renderData.MinX = x
			}
			if x > renderData.MaxX {
				renderData.MaxX = x
			}
			if y < renderData.MinY {
				renderData.MinY = y
			}
			if y > renderData.MaxY {
				renderData.MaxY = y
			}
		}
	}

	// Pad ranges
	xRange := renderData.MaxX - renderData.MinX
	yRange := renderData.MaxY - renderData.MinY
	if xRange == 0 {
		renderData.MinX -= 0.5
		renderData.MaxX += 0.5
	}
	if yRange == 0 {
		renderData.MinY -= 0.5
		renderData.MaxY += 0.5
	}

	// For Bar charts, Y baseline should often include 0
	if chartType == ChartTypeBar {
		if renderData.MinY > 0 {
			renderData.MinY = 0
		}
		if renderData.MaxY < 0 {
			renderData.MaxY = 0
		}
	}

	return renderData
}

func toFloat(v interface{}) (float64, error) {
	if fv, ok := v.(core.FlowValue); ok {
		switch fv.Type.Kind {
		case core.FSKindInt64:
			return float64(fv.Int64Value), nil
		case core.FSKindFloat64:
			return fv.Float64Value, nil
		case core.FSKindBytes:
			return strconv.ParseFloat(string(fv.BytesValue), 64)
		}
	}
	return 0, fmt.Errorf("not a number")
}

// --- Line Chart ---

// GEN:NodeAction
type LineChartAction struct {
	XColumn string
	YColumn string
	// UI State
}

func NewLineChartNode() *core.Node {
	return &core.Node{
		Name:        "Line Chart",
		InputPorts:  []core.NodePort{{Name: "Data", Type: core.FlowType{Kind: core.FSKindTable}}},
		OutputPorts: []core.NodePort{{Name: "Data", Type: core.FlowType{Kind: core.FSKindTable}}},
		Action:      &LineChartAction{},
	}
}

var _ core.NodeAction = &LineChartAction{}

func (c *LineChartAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &c.XColumn)
	core.SStr(s, &c.YColumn)
	return s.Ok()
}

func (c *LineChartAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
	if w, ok := n.GetInputWire(0); ok {
		n.OutputPorts[0].Type = w.Type()
	} else {
		n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindTable}
	}
}

func (c *LineChartAction) UI(n *core.Node) {
	renderData := ExtractChartData(n, c.XColumn, c.YColumn, ChartTypeLine)

	clay.CLAY(clay.IDI("LineChartNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: clay.Sizing{Width: clay.SizingFixed(400)}, ChildGap: core.S2},
	}, func() {
		clay.CLAY(clay.IDI("Settings", n.ID), clay.EL{Layout: clay.LAY{Sizing: core.GROWH, ChildGap: core.S2, ChildAlignment: core.YCENTER}}, func() {
			clay.TEXT("X:", clay.TextElementConfig{TextColor: core.LightGray, FontID: core.InterRegular})
			core.UITextBox(clay.IDI("XCol", n.ID), &c.XColumn, core.UITextBoxConfig{El: clay.EL{Layout: clay.LAY{Sizing: clay.Sizing{Width: clay.SizingFixed(80)}}}, OnSubmit: func(s string) { c.XColumn = s }})
			clay.TEXT("Y:", clay.TextElementConfig{TextColor: core.LightGray, FontID: core.InterRegular})
			core.UITextBox(clay.IDI("YCol", n.ID), &c.YColumn, core.UITextBoxConfig{El: clay.EL{Layout: clay.LAY{Sizing: clay.Sizing{Width: clay.SizingFixed(80)}}}, OnSubmit: func(s string) { c.YColumn = s }})
		})
		clay.CLAY(clay.IDI("ChartArea", n.ID), clay.EL{
			Layout:          clay.LAY{Sizing: clay.Sizing{Width: clay.SizingGrow(0, 400), Height: clay.SizingFixed(200)}},
			BackgroundColor: core.Black,
			Border:          clay.B{Color: core.Gray, Width: clay.BorderWidth{Left: 1, Right: 1, Top: 1, Bottom: 1}},
			Custom:          clay.CustomElementConfig{CustomData: renderData},
		}, func() {})
	})
	core.UIOutputPort(n, 0)
}

func (c *LineChartAction) Run(n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult)
	go func() {
		defer close(done)
		val, _, err := n.GetInputValue(0)
		if err != nil {
			done <- core.NodeActionResult{Err: err}
			return
		}
		done <- core.NodeActionResult{Outputs: []core.FlowValue{val}}
	}()
	return done
}

// --- Bar Chart ---

// GEN:NodeAction
type BarChartAction struct {
	XColumn string
	YColumn string
}

func NewBarChartNode() *core.Node {
	return &core.Node{
		Name:        "Bar Chart",
		InputPorts:  []core.NodePort{{Name: "Data", Type: core.FlowType{Kind: core.FSKindTable}}},
		OutputPorts: []core.NodePort{{Name: "Data", Type: core.FlowType{Kind: core.FSKindTable}}},
		Action:      &BarChartAction{},
	}
}

var _ core.NodeAction = &BarChartAction{}

func (c *BarChartAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &c.XColumn)
	core.SStr(s, &c.YColumn)
	return s.Ok()
}

func (c *BarChartAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
	if w, ok := n.GetInputWire(0); ok {
		n.OutputPorts[0].Type = w.Type()
	} else {
		n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindTable}
	}
}

func (c *BarChartAction) UI(n *core.Node) {
	renderData := ExtractChartData(n, c.XColumn, c.YColumn, ChartTypeBar)

	clay.CLAY(clay.IDI("BarChartNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: clay.Sizing{Width: clay.SizingFixed(400)}, ChildGap: core.S2},
	}, func() {
		clay.CLAY(clay.IDI("Settings", n.ID), clay.EL{Layout: clay.LAY{Sizing: core.GROWH, ChildGap: core.S2, ChildAlignment: core.YCENTER}}, func() {
			clay.TEXT("X:", clay.TextElementConfig{TextColor: core.LightGray, FontID: core.InterRegular})
			core.UITextBox(clay.IDI("XCol", n.ID), &c.XColumn, core.UITextBoxConfig{El: clay.EL{Layout: clay.LAY{Sizing: clay.Sizing{Width: clay.SizingFixed(80)}}}, OnSubmit: func(s string) { c.XColumn = s }})
			clay.TEXT("Y:", clay.TextElementConfig{TextColor: core.LightGray, FontID: core.InterRegular})
			core.UITextBox(clay.IDI("YCol", n.ID), &c.YColumn, core.UITextBoxConfig{El: clay.EL{Layout: clay.LAY{Sizing: clay.Sizing{Width: clay.SizingFixed(80)}}}, OnSubmit: func(s string) { c.YColumn = s }})
		})
		clay.CLAY(clay.IDI("ChartArea", n.ID), clay.EL{
			Layout:          clay.LAY{Sizing: clay.Sizing{Width: clay.SizingGrow(0, 400), Height: clay.SizingFixed(200)}},
			BackgroundColor: core.Black,
			Border:          clay.B{Color: core.Gray, Width: clay.BorderWidth{Left: 1, Right: 1, Top: 1, Bottom: 1}},
			Custom:          clay.CustomElementConfig{CustomData: renderData},
		}, func() {})
	})
	core.UIOutputPort(n, 0)
}

func (c *BarChartAction) Run(n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult)
	go func() {
		defer close(done)
		val, _, err := n.GetInputValue(0)
		if err != nil {
			done <- core.NodeActionResult{Err: err}
			return
		}
		done <- core.NodeActionResult{Outputs: []core.FlowValue{val}}
	}()
	return done
}

// --- Scatter Plot ---

// GEN:NodeAction
type ScatterPlotAction struct {
	XColumn string
	YColumn string
}

func NewScatterPlotNode() *core.Node {
	return &core.Node{
		Name:        "Scatter Plot",
		InputPorts:  []core.NodePort{{Name: "Data", Type: core.FlowType{Kind: core.FSKindTable}}},
		OutputPorts: []core.NodePort{{Name: "Data", Type: core.FlowType{Kind: core.FSKindTable}}},
		Action:      &ScatterPlotAction{},
	}
}

var _ core.NodeAction = &ScatterPlotAction{}

func (c *ScatterPlotAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &c.XColumn)
	core.SStr(s, &c.YColumn)
	return s.Ok()
}

func (c *ScatterPlotAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
	if w, ok := n.GetInputWire(0); ok {
		n.OutputPorts[0].Type = w.Type()
	} else {
		n.OutputPorts[0].Type = core.FlowType{Kind: core.FSKindTable}
	}
}

func (c *ScatterPlotAction) UI(n *core.Node) {
	renderData := ExtractChartData(n, c.XColumn, c.YColumn, ChartTypeScatter)

	clay.CLAY(clay.IDI("ScatterPlotNode", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: clay.Sizing{Width: clay.SizingFixed(400)}, ChildGap: core.S2},
	}, func() {
		clay.CLAY(clay.IDI("Settings", n.ID), clay.EL{Layout: clay.LAY{Sizing: core.GROWH, ChildGap: core.S2, ChildAlignment: core.YCENTER}}, func() {
			clay.TEXT("X:", clay.TextElementConfig{TextColor: core.LightGray, FontID: core.InterRegular})
			core.UITextBox(clay.IDI("XCol", n.ID), &c.XColumn, core.UITextBoxConfig{El: clay.EL{Layout: clay.LAY{Sizing: clay.Sizing{Width: clay.SizingFixed(80)}}}, OnSubmit: func(s string) { c.XColumn = s }})
			clay.TEXT("Y:", clay.TextElementConfig{TextColor: core.LightGray, FontID: core.InterRegular})
			core.UITextBox(clay.IDI("YCol", n.ID), &c.YColumn, core.UITextBoxConfig{El: clay.EL{Layout: clay.LAY{Sizing: clay.Sizing{Width: clay.SizingFixed(80)}}}, OnSubmit: func(s string) { c.YColumn = s }})
		})
		clay.CLAY(clay.IDI("ChartArea", n.ID), clay.EL{
			Layout:          clay.LAY{Sizing: clay.Sizing{Width: clay.SizingGrow(0, 400), Height: clay.SizingFixed(200)}},
			BackgroundColor: core.Black,
			Border:          clay.B{Color: core.Gray, Width: clay.BorderWidth{Left: 1, Right: 1, Top: 1, Bottom: 1}},
			Custom:          clay.CustomElementConfig{CustomData: renderData},
		}, func() {})
	})
	core.UIOutputPort(n, 0)
}

func (c *ScatterPlotAction) Run(n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult)
	go func() {
		defer close(done)
		val, _, err := n.GetInputValue(0)
		if err != nil {
			done <- core.NodeActionResult{Err: err}
			return
		}
		done <- core.NodeActionResult{Outputs: []core.FlowValue{val}}
	}()
	return done
}

// GEN:CustomRender
func RenderChart(bbox clay.BoundingBox, data *ChartRenderData) {
	width := bbox.Width
	height := bbox.Height

	if data.Error != "" {
		rl.DrawText(data.Error, int32(bbox.X)+10, int32(bbox.Y)+10, 20, rl.Red)
		return
	}

	if len(data.Points) < 1 {
		rl.DrawText("No Data", int32(bbox.X)+10, int32(bbox.Y)+10, 20, rl.Gray)
		return
	}

	xRange := data.MaxX - data.MinX
	yRange := data.MaxY - data.MinY

	mapPt := func(pt rl.Vector2) rl.Vector2 {
		x := (float64(pt.X) - data.MinX) / xRange
		y := (float64(pt.Y) - data.MinY) / yRange
		return rl.Vector2{
			X: bbox.X + float32(x)*width,
			Y: bbox.Y + height - float32(y)*height,
		}
	}

	// Draw Grid/Axes
	rl.DrawRectangleLinesEx(rl.Rectangle{X: bbox.X, Y: bbox.Y, Width: width, Height: height}, 1, rl.DarkGray)

	// Draw Zero Line Y if visible
	if data.MinY < 0 && data.MaxY > 0 {
		zeroY := bbox.Y + height - float32((0-data.MinY)/yRange)*height
		rl.DrawLineEx(rl.Vector2{X: bbox.X, Y: zeroY}, rl.Vector2{X: bbox.X + width, Y: zeroY}, 1, rl.DarkGray)
	}

	switch data.Type {
	case ChartTypeLine:
		for i := 0; i < len(data.Points)-1; i++ {
			p1 := mapPt(data.Points[i])
			p2 := mapPt(data.Points[i+1])
			rl.DrawLineEx(p1, p2, 2, rl.Green)
		}
	case ChartTypeBar:
		barWidth := width / float32(len(data.Points)) * 0.8
		for _, pt := range data.Points {
			p := mapPt(pt)
			base := mapPt(rl.Vector2{X: pt.X, Y: 0})
			// Draw bar from base to p
			// We need to handle rect

			// Actually mapPt X is center? No, X is value.
			// Bar charts usually treat X as category.
			// Here we just plot X/Y.

			// Let's draw a vertical line or thin rect
			h := base.Y - p.Y
			// If h is negative (value < 0), p.Y > base.Y

			r := rl.Rectangle{
				X:      p.X - barWidth/2,
				Y:      util.Tern(h > 0, p.Y, base.Y),
				Width:  barWidth,
				Height: float32(math.Abs(float64(h))),
			}
			rl.DrawRectangleRec(r, rl.SkyBlue)
		}
	case ChartTypeScatter:
		for _, pt := range data.Points {
			p := mapPt(pt)
			rl.DrawCircleV(p, 3, rl.Yellow)
		}
	}

	rl.DrawText(fmt.Sprintf("%.2f", data.MaxY), int32(bbox.X)+5, int32(bbox.Y)+5, 10, rl.Gray)
	rl.DrawText(fmt.Sprintf("%.2f", data.MinY), int32(bbox.X)+5, int32(bbox.Y+height)-15, 10, rl.Gray)
}