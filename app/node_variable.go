package app

import (
	"context"
	"fmt"
	"os"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type GetVariableAction struct {
	VariableName string
}

func NewGetVariableNode() *Node {
	return &Node{
		Name: "Get Variable",
		OutputPorts: []NodePort{
			{Name: "Value", Type: FlowType{Kind: FSKindBytes}},
		},
		Action: &GetVariableAction{},
	}
}

var _ NodeAction = &GetVariableAction{}

func (a *GetVariableAction) UpdateAndValidate(n *Node) {
	if a.VariableName == "" {
		n.Valid = false
	} else {
		n.Valid = true
	}
}

func (a *GetVariableAction) UI(n *Node) {
	clay.CLAY(clay.IDI("GetVariable", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.CLAY(clay.IDI("Row1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER, ChildGap: S2},
		}, func() {
			clay.TEXT("Name:", clay.TextElementConfig{TextColor: White})
			UITextBox(clay.IDI("VarName", n.ID), &a.VariableName, UITextBoxConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
			})
			UIOutputPort(n, 0)
		})
	})
}

func (a *GetVariableAction) Run(n *Node) <-chan NodeActionResult {
	return a.RunContext(context.Background(), n)
}

func (a *GetVariableAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		defer close(done)

		// Check environment variables first (higher priority? or lower? Let's say Graph Variables override Env?)
		// Actually, let's say Env overrides Graph so you can inject secrets at runtime.
		// Wait, "Variables / Secrets" usually implies managing them within the app.
		// Let's check:
		// 1. System Env Vars
		// 2. Graph Variables

		val, found := os.LookupEnv(a.VariableName)
		if !found {
			if n.Graph != nil {
				n.Graph.VarMutex.RLock()
				v, ok := n.Graph.Variables[a.VariableName]
				n.Graph.VarMutex.RUnlock()
				if ok {
					val = v
					found = true
				}
			}
		}

		if !found {
			done <- NodeActionResult{Err: fmt.Errorf("variable '%s' not found in environment or graph", a.VariableName)}
			return
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{NewStringValue(val)},
		}
	}()
	return done
}

func (a *GetVariableAction) Serialize(s *Serializer) bool {
	SStr(s, &a.VariableName)
	return s.Ok()
}

