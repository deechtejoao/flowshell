package app

import (
	"context"
	"fmt"
	"os"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type MapAction struct {
	SubflowPath string
	CachedGraph *Graph
}

func NewMapNode() *Node {
	return &Node{
		Name: "Map",
		InputPorts: []NodePort{
			{Name: "List", Type: FlowType{Kind: FSKindList}},
		},
		OutputPorts: []NodePort{
			{Name: "Result", Type: FlowType{Kind: FSKindList}},
		},
		Action: &MapAction{},
	}
}

var _ NodeAction = &MapAction{}

func (a *MapAction) UpdateAndValidate(n *Node) {
	n.Valid = a.SubflowPath != "" && n.InputIsWired(0)
}

func (a *MapAction) UI(n *Node) {
	UIInputPort(n, 0)
	clay.TEXT("Map", clay.TextElementConfig{TextColor: White})
	clay.TEXT("Subflow:", clay.TextElementConfig{TextColor: Gray})
	UITextBox(clay.IDI("SubflowPath", n.ID), &a.SubflowPath, UITextBoxConfig{
		El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
	})
	UIOutputPort(n, 0)
}

func (a *MapAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *MapAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		defer close(done)

		// Get Input List
		val, ok, err := n.GetInputValue(0)
		if err != nil || !ok {
			done <- NodeActionResult{Err: fmt.Errorf("failed to get input: %v", err)}
			return
		}
		if val.Type.Kind != FSKindList {
			done <- NodeActionResult{Err: fmt.Errorf("input must be a list")}
			return
		}
		list := val.ListValue

		// Load Graph
		if a.CachedGraph == nil {
			data, err := os.ReadFile(a.SubflowPath)
			if err != nil {
				done <- NodeActionResult{Err: fmt.Errorf("failed to read subflow: %v", err)}
				return
			}
			g, err := DeserializeGraph(data)
			if err != nil {
				done <- NodeActionResult{Err: fmt.Errorf("failed to parse subflow: %v", err)}
				return
			}
			a.CachedGraph = g
		}

		// Find IO Nodes
		var inputNode, outputNode *Node
		for _, node := range a.CachedGraph.Nodes {
			if _, ok := node.Action.(*GraphInputAction); ok {
				inputNode = node
			}
			if _, ok := node.Action.(*GraphOutputAction); ok {
				outputNode = node
			}
		}

		if inputNode == nil || outputNode == nil {
			done <- NodeActionResult{Err: fmt.Errorf("subflow missing GraphInput or GraphOutput node")}
			return
		}

		// Process
		results := make([]FlowValue, 0, len(list))
		for _, item := range list {
			// Check Context
			if ctx.Err() != nil {
				done <- NodeActionResult{Err: ctx.Err()}
				return
			}

			// Set Input
			inputAction := inputNode.Action.(*GraphInputAction)
			inputAction.Value = item

			// Run Subgraph
			// We force rerun inputs to propagate the new value
			outDoneCh := outputNode.Run(ctx, true)
			
			select {
			case <-outDoneCh:
				res, ok := outputNode.GetResult()
				if !ok {
					done <- NodeActionResult{Err: fmt.Errorf("subflow output node produced no result")}
					return
				}
				if res.Err != nil {
					done <- NodeActionResult{Err: fmt.Errorf("subflow error: %v", res.Err)}
					return
				}
				if len(res.Outputs) > 0 {
					results = append(results, res.Outputs[0])
				} else {
					results = append(results, FlowValue{}) // nil/void?
				}
			case <-ctx.Done():
				done <- NodeActionResult{Err: ctx.Err()}
				return
			}
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{{Type: &FlowType{Kind: FSKindList}, ListValue: results}},
		}
	}()
	return done
}

func (a *MapAction) Tag() string {
	return "Map"
}

func (a *MapAction) Serialize(s *Serializer) bool {
	SStr(s, &a.SubflowPath)
	return s.Ok()
}
