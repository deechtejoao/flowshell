package nodes

import (
	"context"
	"fmt"
	"regexp"
	"runtime"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type LinesAction struct {
	IncludeCarriageReturns bool
}

func NewLinesNode() *core.Node {
	return &core.Node{
		Name: "Lines",

		InputPorts: []core.NodePort{{
			Name: "Text",
			Type: core.FlowType{Kind: core.FSKindBytes},
		}},
		OutputPorts: []core.NodePort{{
			Name: "Lines",
			Type: core.FlowType{Kind: core.FSKindList, ContainedType: &core.FlowType{Kind: core.FSKindBytes}},
		}},

		Action: &LinesAction{
			IncludeCarriageReturns: runtime.GOOS == "windows",
		},
	}
}

var _ core.NodeAction = &LinesAction{}

func (c *LinesAction) UpdateAndValidate(n *core.Node) {
	n.Valid = true

	if _, ok := n.GetInputWire(0); !ok {
		n.Valid = false
	}
}

func (l *LinesAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("LinesUI", n.ID), clay.EL{
		Layout: clay.LAY{Sizing: core.GROWH, ChildGap: core.S2},
	}, func() {
		clay.CLAY(clay.IDI("LinesInputCol", n.ID), clay.EL{ // inputs
			Layout: clay.LAY{
				LayoutDirection: clay.TopToBottom,
				Sizing:          core.GROWH,
				ChildAlignment:  clay.ChildAlignment{Y: clay.AlignYCenter},
			},
		}, func() {
			core.UIInputPort(n, 0)
		})
		clay.CLAY(clay.IDI("LinesOutputCol", n.ID), clay.EL{ // outputs
			Layout: clay.LAY{
				LayoutDirection: clay.TopToBottom,
				Sizing:          core.GROWH,
				ChildAlignment:  clay.ChildAlignment{X: clay.AlignXRight, Y: clay.AlignYCenter},
			},
		}, func() {
			core.UIOutputPort(n, 0)
		})
	})

	clay.CLAY(clay.IDI("LinesOptionsRow", n.ID), clay.EL{
		Layout: clay.LAY{ChildGap: core.S1, ChildAlignment: clay.ChildAlignment{Y: clay.AlignYCenter}},
	}, func() {
		core.UIButton(clay.IDI("LinesCRToggle", n.ID), core.UIButtonConfig{
			OnClick: func(_ clay.ElementID, _ clay.PointerData, _ any) {
				l.IncludeCarriageReturns = !l.IncludeCarriageReturns
			},
		}, func() {
			clay.CLAY(clay.IDI("LinesCRIndicator", n.ID), clay.EL{
				Layout:          clay.LAY{Sizing: core.WH(16, 16), ChildAlignment: core.ALLCENTER},
				Border:          clay.B{Width: core.BA, Color: core.White},
				BackgroundColor: util.Tern(l.IncludeCarriageReturns, core.White, clay.Color{}),
			})
		})
		clay.TEXT("Keep \\r", clay.TextElementConfig{TextColor: core.White})
	})
}

var LFSplit = regexp.MustCompile(`\n`)
var CRLFSplit = regexp.MustCompile(`\r?\n`)

func (l *LinesAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult)

	go func() {
		var res core.NodeActionResult
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				res = core.NodeActionResult{Err: fmt.Errorf("panic in node %s: %v", n.Name, r)}
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
		lines := util.Map(linesStrs, func(line string) core.FlowValue { return core.NewStringValue(line) })

		res = core.NodeActionResult{
			Outputs: []core.FlowValue{{
				Type:      &core.FlowType{Kind: core.FSKindList, ContainedType: &core.FlowType{Kind: core.FSKindBytes}},
				ListValue: lines,
			}},
		}
	}()
	return done
}

func (l *LinesAction) Run(n *core.Node) <-chan core.NodeActionResult {
	return l.RunContext(context.Background(), n)
}

func (n *LinesAction) Serialize(s *core.Serializer) bool {
	core.SBool(s, &n.IncludeCarriageReturns)
	return s.Ok()
}
