package app

import (
	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type ValueAction struct {
	Value FlowValue
}

func NewValueNode(v FlowValue) *Node {
	return &Node{
		ID:   NewNodeID(),
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
	clay.CLAY_AUTO_ID(clay.EL{
		Layout: clay.LAY{
			Sizing:         GROWH,
			ChildAlignment: YCENTER,
		},
	}, func() {
		// Just show the value
		UIFlowValue(c.Value)
		UISpacer(clay.AUTO_ID, GROWH)
		UIOutputPort(n, 0)
	})
}

func (c *ValueAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)
	go func() {
		done <- NodeActionResult{
			Outputs: []FlowValue{c.Value},
		}
	}()
	return done
}
