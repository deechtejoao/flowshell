package app

import (
	"context"
	"fmt"
	"regexp"
	"runtime"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
)

// GEN:NodeAction
type LinesAction struct {
	IncludeCarriageReturns bool
}

func NewLinesNode() *Node {
	return &Node{
		Name: "Lines",

		InputPorts: []NodePort{{
			Name: "Text",
			Type: FlowType{Kind: FSKindBytes},
		}},
		OutputPorts: []NodePort{{
			Name: "Lines",
			Type: FlowType{Kind: FSKindList, ContainedType: &FlowType{Kind: FSKindBytes}},
		}},

		Action: &LinesAction{
			IncludeCarriageReturns: runtime.GOOS == "windows",
		},
	}
}

var _ NodeAction = &LinesAction{}

func (c *LinesAction) UpdateAndValidate(n *Node) {
	n.Valid = true

	if _, ok := n.GetInputWire(0); !ok {
		n.Valid = false
	}
}

func (l *LinesAction) UI(n *Node) {
	clay.CLAY(clay.IDI("LinesUI", n.ID), clay.EL{
		Layout: clay.LAY{Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.CLAY(clay.IDI("LinesInputCol", n.ID), clay.EL{ // inputs
			Layout: clay.LAY{
				LayoutDirection: clay.TopToBottom,
				Sizing:          GROWH,
				ChildAlignment:  clay.ChildAlignment{Y: clay.AlignYCenter},
			},
		}, func() {
			UIInputPort(n, 0)
		})
		clay.CLAY(clay.IDI("LinesOutputCol", n.ID), clay.EL{ // outputs
			Layout: clay.LAY{
				LayoutDirection: clay.TopToBottom,
				Sizing:          GROWH,
				ChildAlignment:  clay.ChildAlignment{X: clay.AlignXRight, Y: clay.AlignYCenter},
			},
		}, func() {
			UIOutputPort(n, 0)
		})
	})

	clay.CLAY(clay.IDI("LinesOptionsRow", n.ID), clay.EL{
		Layout: clay.LAY{ChildGap: S1, ChildAlignment: clay.ChildAlignment{Y: clay.AlignYCenter}},
	}, func() {
		UIButton(clay.IDI("LinesCRToggle", n.ID), UIButtonConfig{
			OnClick: func(_ clay.ElementID, _ clay.PointerData, _ any) {
				l.IncludeCarriageReturns = !l.IncludeCarriageReturns
			},
		}, func() {
			clay.CLAY(clay.IDI("LinesCRIndicator", n.ID), clay.EL{
				Layout:          clay.LAY{Sizing: WH(16, 16), ChildAlignment: ALLCENTER},
				Border:          clay.B{Width: BA, Color: White},
				BackgroundColor: util.Tern(l.IncludeCarriageReturns, White, clay.Color{}),
			})
		})
		clay.TEXT("Keep \\r", clay.TextElementConfig{TextColor: White})
	})
}

var LFSplit = regexp.MustCompile(`\n`)
var CRLFSplit = regexp.MustCompile(`\r?\n`)

func (l *LinesAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult)

	go func() {
		var res NodeActionResult
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				res = NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
			}
			done <- res
		}()

		select {
		case <-ctx.Done():
			res.Err = ctx.Err()
			return
		default:
		}

		text, ok, err := n.GetInputValue(0)
		if !ok {
			panic(fmt.Errorf("node %s: no text input, should have been caught by validation", n))
		}
		if err != nil {
			res.Err = err
			return
		}
		linesStrs := util.Tern(l.IncludeCarriageReturns, CRLFSplit, LFSplit).Split(string(text.BytesValue), -1)
		lines := util.Map(linesStrs, func(line string) FlowValue { return NewStringValue(line) })

		res = NodeActionResult{
			Outputs: []FlowValue{{
				Type:      &FlowType{Kind: FSKindList, ContainedType: &FlowType{Kind: FSKindBytes}},
				ListValue: lines,
			}},
		}
	}()
	return done
}

func (l *LinesAction) Run(n *Node) <-chan NodeActionResult {
	return l.RunContext(context.Background(), n)
}

func (n *LinesAction) Serialize(s *Serializer) bool {
	SBool(s, &n.IncludeCarriageReturns)
	return s.Ok()
}
