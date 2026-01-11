package app

import (
	"errors"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type FilterEmptyAction struct{}

func NewFilterEmptyNode() *Node {
	return &Node{
		ID:   NewNodeID(),
		Name: "Filter Empty",

		InputPorts: []NodePort{{
			Name: "List",
			Type: NewListType(FlowType{Kind: FSKindBytes}),
		}},
		OutputPorts: []NodePort{{
			Name: "Filtered",
			Type: NewListType(FlowType{Kind: FSKindBytes}),
		}},

		Action: &FilterEmptyAction{},
	}
}

var _ NodeAction = &FilterEmptyAction{}

func (c *FilterEmptyAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

func (c *FilterEmptyAction) UpdateAndValidate(n *Node) {
	n.Valid = true

	// Check if input is wired and valid
	if wire, ok := n.GetInputWire(0); ok {
		n.InputPorts[0].Type = wire.Type()
		n.OutputPorts[0].Type = wire.Type()

		if wire.Type().Kind != FSKindList {
			n.Valid = false // Only support lists
		}
	} else {
		n.Valid = false
	}
}

func (c *FilterEmptyAction) UI(n *Node) {
	clay.CLAY_AUTO_ID(clay.EL{
		Layout: clay.LAY{
			Sizing:         GROWH,
			ChildAlignment: YCENTER,
		},
	}, func() {
		UIInputPort(n, 0)
		UISpacer(clay.AUTO_ID, GROWH)
		UIOutputPort(n, 0)
	})
}

func (c *FilterEmptyAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)

	go func() {
		var res NodeActionResult
		defer func() { done <- res }()

		input, ok, err := n.GetInputValue(0)
		if !ok {
			res.Err = errors.New("input required")
			return
		}
		if err != nil {
			res.Err = err
			return
		}

		if input.Type.Kind != FSKindList {
			res.Err = errors.New("input must be a list")
			return
		}

		var filtered []FlowValue
		for _, item := range input.ListValue {
			keep := true
			switch item.Type.Kind {
			case FSKindBytes:
				if len(item.BytesValue) == 0 {
					keep = false
				}
				// Add other types if needed (e.g. empty lists?)
			}

			if keep {
				filtered = append(filtered, item)
			}
		}

		res.Outputs = []FlowValue{NewListValue(*input.Type.ContainedType, filtered)}
	}()

	return done
}
