package app

import (
	"context"
	"fmt"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type ValueAction struct {
	Value FlowValue
}

func NewValueNode(v FlowValue) *Node {
	return &Node{
		Name: "Value",

		OutputPorts: []NodePort{{
			Name: "Value",
			Type: *v.Type,
		}},

		Action: &ValueAction{Value: v},
	}
}

var _ NodeAction = &ValueAction{}

func (c *ValueAction) Serialize(s *Serializer) bool {
	SThing(s, &c.Value)
	return s.Ok()
}

func (c *ValueAction) UpdateAndValidate(n *Node) {
	n.Valid = true
	// Ensure output port type matches value type (in case it changed or for safety)
	if n.OutputPorts[0].Type.Kind == FSKindAny {
		n.OutputPorts[0].Type = *c.Value.Type
	}
}

func (c *ValueAction) UI(n *Node) {
	clay.CLAY(clay.IDI("ValueUI", n.ID), clay.EL{
		Layout: clay.LAY{
			Sizing:         GROWH,
			ChildAlignment: YCENTER,
		},
	}, func() {
		// Just show the value
		UIFlowValue(c.Value)
		UISpacer(clay.IDI("ValueSpacer", n.ID), GROWH)
		UIOutputPort(n, 0)
	})
}

func (c *ValueAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				done <- NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
			}
		}()

		select {
		case <-ctx.Done():
			done <- NodeActionResult{Err: ctx.Err()}
			return
		default:
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{c.Value},
		}
	}()
	return done
}

func (c *ValueAction) Run(n *Node) <-chan NodeActionResult {
	return c.RunContext(context.Background(), n)
}
