package nodes

import (
	"context"
	"fmt"
	"os"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type MapAction struct {
	SubflowPath string
	CachedGraph *core.Graph
}

func NewMapNode() *core.Node {
	return &core.Node{
		Name: "Map",
		InputPorts: []core.NodePort{
			{Name: "List", Type: core.FlowType{Kind: core.FSKindList}},
		},
		OutputPorts: []core.NodePort{
			{Name: "Result", Type: core.FlowType{Kind: core.FSKindList}},
		},
		Action: &MapAction{},
	}
}

var _ core.NodeAction = &MapAction{}

func (a *MapAction) UpdateAndValidate(n *core.Node) {
	n.Valid = a.SubflowPath != "" && n.InputIsWired(0)
}

func (a *MapAction) UI(n *core.Node) {
	core.UIInputPort(n, 0)
	clay.TEXT("Map", clay.TextElementConfig{TextColor: core.White})
	clay.TEXT("Subflow:", clay.TextElementConfig{TextColor: core.Gray})
	core.UITextBox(clay.IDI("SubflowPath", n.ID), &a.SubflowPath, core.UITextBoxConfig{
		El: clay.EL{Layout: clay.LAY{Sizing: core.GROWH}},
	})
	core.UIOutputPort(n, 0)
}

func (a *MapAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *MapAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult, 1)
	go func() {
		defer close(done)

		// Get Input List
		val, ok, err := n.GetInputValue(0)
		if err != nil || !ok {
			done <- core.NodeActionResult{Err: fmt.Errorf("failed to get input: %v", err)}
			return
		}
		if val.Type.Kind != core.FSKindList {
			done <- core.NodeActionResult{Err: fmt.Errorf("input must be a list")}
			return
		}
		list := val.ListValue

		// Load core.Graph
		if a.CachedGraph == nil {
			data, err := os.ReadFile(a.SubflowPath)
			if err != nil {
				done <- core.NodeActionResult{Err: fmt.Errorf("failed to read subflow: %v", err)}
				return
			}
			g, err := core.DeserializeGraph(data)
			if err != nil {
				done <- core.NodeActionResult{Err: fmt.Errorf("failed to parse subflow: %v", err)}
				return
			}
			a.CachedGraph = g
		}

		// Find IO Nodes
		var inputNode, outputNode *core.Node
		for _, node := range a.CachedGraph.Nodes {
			if _, ok := node.Action.(*GraphInputAction); ok {
				inputNode = node
			}
			if _, ok := node.Action.(*GraphOutputAction); ok {
				outputNode = node
			}
		}

		if inputNode == nil || outputNode == nil {
			done <- core.NodeActionResult{Err: fmt.Errorf("subflow missing GraphInput or GraphOutput node")}
			return
		}

		// Process
		results := make([]core.FlowValue, 0, len(list))
		for _, item := range list {
			// Check Context
			if ctx.Err() != nil {
				done <- core.NodeActionResult{Err: ctx.Err()}
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
					done <- core.NodeActionResult{Err: fmt.Errorf("subflow output node produced no result")}
					return
				}
				if res.Err != nil {
					done <- core.NodeActionResult{Err: fmt.Errorf("subflow error: %v", res.Err)}
					return
				}
				if len(res.Outputs) > 0 {
					results = append(results, res.Outputs[0])
				} else {
					results = append(results, core.FlowValue{}) // nil/void?
				}
			case <-ctx.Done():
				done <- core.NodeActionResult{Err: ctx.Err()}
				return
			}
		}

		done <- core.NodeActionResult{
			Outputs: []core.FlowValue{{Type: &core.FlowType{Kind: core.FSKindList}, ListValue: results}},
		}
	}()
	return done
}

func (a *MapAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &a.SubflowPath)
	return s.Ok()
}