package app

import (
	"context"
	"fmt"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type GraphInputAction struct {
	Value FlowValue
}

func NewGraphInputNode() *Node {
	return &Node{
		Name: "Graph Input",
		OutputPorts: []NodePort{
			{Name: "In", Type: FlowType{Kind: FSKindAny}},
		},
		Action: &GraphInputAction{},
	}
}

var _ NodeAction = &GraphInputAction{}

func (a *GraphInputAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *GraphInputAction) UI(n *Node) {
	clay.TEXT("Graph Input", clay.TextElementConfig{TextColor: White})
	UIOutputPort(n, 0)
}

func (a *GraphInputAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *GraphInputAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	done <- NodeActionResult{
		Outputs: []FlowValue{a.Value},
	}
	close(done)
	return done
}

func (a *GraphInputAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

// GEN:NodeAction
type GraphOutputAction struct {
}

func NewGraphOutputNode() *Node {
	return &Node{
		Name: "Graph Output",
		InputPorts: []NodePort{
			{Name: "Out", Type: FlowType{Kind: FSKindAny}},
		},
		OutputPorts: []NodePort{
			{Name: "Result", Type: FlowType{Kind: FSKindAny}},
		},
		Action: &GraphOutputAction{},
	}
}

var _ NodeAction = &GraphOutputAction{}

func (a *GraphOutputAction) UpdateAndValidate(n *Node) {
	n.Valid = n.InputIsWired(0)
}

func (a *GraphOutputAction) UI(n *Node) {
	UIInputPort(n, 0)
	clay.TEXT("Graph Output", clay.TextElementConfig{TextColor: White})
	UIOutputPort(n, 0)
}

func (a *GraphOutputAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *GraphOutputAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		val, ok, err := n.GetInputValue(0)
		if err != nil {
			done <- NodeActionResult{Err: err}
		} else if !ok {
			done <- NodeActionResult{Err: fmt.Errorf("missing input")}
		} else {
			done <- NodeActionResult{Outputs: []FlowValue{val}}
		}
		close(done)
	}()
	return done
}

func (a *GraphOutputAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

