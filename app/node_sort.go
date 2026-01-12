package app

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
)

// GEN:NodeAction
type SortAction struct {
	Reverse bool
}

func NewSortNode() *Node {
	return &Node{
		Name: "Sort",

		InputPorts: []NodePort{{
			Name: "List",
			Type: NewListType(FlowType{Kind: FSKindBytes}),
		}},
		OutputPorts: []NodePort{{
			Name: "Sorted",
			Type: NewListType(FlowType{Kind: FSKindBytes}),
		}},

		Action: &SortAction{},
	}
}

var _ NodeAction = &SortAction{}

func (c *SortAction) Serialize(s *Serializer) bool {
	SBool(s, &c.Reverse)
	return s.Ok()
}

func (c *SortAction) UpdateAndValidate(n *Node) {
	n.Valid = true

	if wire, ok := n.GetInputWire(0); ok {
		n.InputPorts[0].Type = wire.Type()
		n.OutputPorts[0].Type = wire.Type()

		if wire.Type().Kind != FSKindList {
			n.Valid = false
		}
	} else {
		n.Valid = false
	}
}

func (c *SortAction) UI(n *Node) {
	clay.CLAY_AUTO_ID(clay.EL{
		Layout: clay.LAY{
			Sizing:         GROWH,
			ChildAlignment: YCENTER,
		},
	}, func() {
		UIInputPort(n, 0)
		UISpacer(clay.AUTO_ID, GROWH)

		// Reverse Checkbox
		clay.CLAY_AUTO_ID(clay.EL{
			Layout: clay.LAY{ChildGap: S1, ChildAlignment: YCENTER},
		}, func() {
			UIButton(clay.AUTO_ID, UIButtonConfig{
				OnClick: func(_ clay.ElementID, _ clay.PointerData, _ any) {
					c.Reverse = !c.Reverse
				},
			}, func() {
				UIImage(clay.AUTO_ID, util.Tern(c.Reverse, ImgToggleDown, ImgToggleRight), clay.EL{}) // Using toggle icons for checkbox for now
			})
			clay.TEXT("Reverse", clay.TextElementConfig{TextColor: White})
		})

		UISpacer(clay.AUTO_ID, GROWH)
		UIOutputPort(n, 0)
	})
}

func (c *SortAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)

	go func() {
		var res NodeActionResult
		defer func() {
			if r := recover(); r != nil {
				res = NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
			}
			done <- res
			close(done)
		}()

		select {
		case <-ctx.Done():
			res.Err = ctx.Err()
			return
		default:
		}

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

		// Copy list to avoid modifying input
		sorted := make([]FlowValue, len(input.ListValue))
		copy(sorted, input.ListValue)

		// Recover from panic caused by cancellation
		defer func() {
			if r := recover(); r != nil {
				if r == ctx.Err() {
					res.Err = ctx.Err()
					// done <- res will happen in the outer defer
				} else {
					panic(r) // Re-panic if it's not our cancellation
				}
			}
		}()

		slices.SortFunc(sorted, func(a, b FlowValue) int {
			// Check for cancellation
			if ctx.Err() != nil {
				panic(ctx.Err())
			}

			res := 0
			switch a.Type.Kind {
			case FSKindBytes:
				res = cmp.Compare(string(a.BytesValue), string(b.BytesValue))
			case FSKindInt64:
				res = cmp.Compare(a.Int64Value, b.Int64Value)
			case FSKindFloat64:
				res = cmp.Compare(a.Float64Value, b.Float64Value)
			}

			if c.Reverse {
				return -res
			}
			return res
		})

		res.Outputs = []FlowValue{NewListValue(*input.Type.ContainedType, sorted)}
	}()

	return done
}

func (c *SortAction) Run(n *Node) <-chan NodeActionResult {
	return c.RunContext(context.Background(), n)
}
