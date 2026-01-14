package nodes

import (
	"context"
	"fmt"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type GraphInputAction struct {
	Value core.FlowValue
}

func NewGraphInputNode() *core.Node {
	return &core.Node{
		Name: "core.Graph Input",
		OutputPorts: []core.NodePort{
			{Name: "In", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		Action: &GraphInputAction{},
	}
}

var _ core.NodeAction = &GraphInputAction{}

func (a *GraphInputAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true
}

func (a *GraphInputAction) UI(n *core.Node) {
	clay.TEXT("core.Graph Input", clay.TextElementConfig{TextColor: core.White})
	core.UIOutputPort(n, 0)
}

func (a *GraphInputAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *GraphInputAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult, 1)
	done <- core.NodeActionResult{
		Outputs: []core.FlowValue{a.Value},
	}
	close(done)
	return done
}

func (a *GraphInputAction) Serialize(s *core.Serializer) bool {
	return s.Ok()
}

// GEN:NodeAction
type GraphOutputAction struct {
}

func NewGraphOutputNode() *core.Node {
	return &core.Node{
		Name: "core.Graph Output",
		InputPorts: []core.NodePort{
			{Name: "Out", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		OutputPorts: []core.NodePort{
			{Name: "Result", Type: core.FlowType{Kind: core.FSKindAny}},
		},
		Action: &GraphOutputAction{},
	}
}

var _ core.NodeAction = &GraphOutputAction{}

func (a *GraphOutputAction) UpdateAndValidate(n *core.Node) {
	n.Valid = n.InputIsWired(0)
}

func (a *GraphOutputAction) UI(n *core.Node) {
	core.UIInputPort(n, 0)
	clay.TEXT("core.Graph Output", clay.TextElementConfig{TextColor: core.White})
	core.UIOutputPort(n, 0)
}

func (a *GraphOutputAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *GraphOutputAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult, 1)
	go func() {
		val, ok, err := n.GetInputValue(0)
		if err != nil {
			done <- core.NodeActionResult{Err: err}
		} else if !ok {
			done <- core.NodeActionResult{Err: fmt.Errorf("missing input")}
		} else {
			done <- core.NodeActionResult{Outputs: []core.FlowValue{val}}
		}
		close(done)
	}()
	return done
}

func (a *GraphOutputAction) Serialize(s *core.Serializer) bool {
	return s.Ok()
}
