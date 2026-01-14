package nodes

import (
	"context"
	"fmt"
	"time"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/app/core"
)

// GEN:NodeAction
type ParseTimeAction struct {
	Format string
}

func NewParseTimeNode() *core.Node {
	return &core.Node{
		Name: "Parse Time",
		InputPorts: []core.NodePort{
			{Name: "Text", Type: core.FlowType{Kind: core.FSKindBytes}},
			{Name: "Format", Type: core.FlowType{Kind: core.FSKindBytes}}, // Optional override
		},
		OutputPorts: []core.NodePort{
			{Name: "Timestamp", Type: core.FlowType{Kind: core.FSKindInt64, WellKnownType: core.FSWKTTimestamp}},
		},
		Action: &ParseTimeAction{Format: time.RFC3339},
	}
}

var _ core.NodeAction = &ParseTimeAction{}

func (a *ParseTimeAction) UpdateAndValidate(n *core.Node) { n.Valid = true }

func (a *ParseTimeAction) UI(n *core.Node) {
	clay.CLAY(clay.IDI("TimeUI", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: core.GROWH, ChildGap: core.S2},
	}, func() {
		clay.CLAY(clay.IDI("TimeRow1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: core.GROWH, ChildAlignment: core.YCENTER},
		}, func() {
			core.UIInputPort(n, 0)
			core.UISpacer(clay.IDI("TimeSpacer1", n.ID), core.GROWH)
			core.UIOutputPort(n, 0)
		})

		clay.CLAY(clay.IDI("TimeRow2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: core.GROWH},
		}, func() {
			core.UITextBox(clay.IDI("FormatStr", n.ID), &a.Format, core.UITextBoxConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: core.GROWH}},
			})
		})

		clay.CLAY(clay.IDI("TimeRow3", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: core.GROWH},
		}, func() {
			clay.TEXT("Examples: 2006-01-02, Jan 2, 15:04:05", clay.TextElementConfig{FontSize: 12, TextColor: core.LightGray})
		})
	})
}

func (a *ParseTimeAction) Run(n *core.Node) <-chan core.NodeActionResult {
	done := make(chan core.NodeActionResult, 1)
	go func() {
		valText, _, err := n.GetInputValue(0)
		if err != nil {
			done <- core.NodeActionResult{Err: err}
			return
		}

		// Optional Format override from input
		format := a.Format
		if valFmt, wired, err := n.GetInputValue(1); err == nil && wired {
			format = string(valFmt.BytesValue)
		}

		text := string(valText.BytesValue)

		// Helper to try multiple common formats if user format fails?
		// Or strictly follow format. The prompt says "with format strings", implies strict control.

		t, err := time.Parse(format, text)
		if err != nil {
			// Try some fallbacks if default RFC3339 failed?
			// Only if using default?
			// Let's stick to explicit format for now.
			done <- core.NodeActionResult{Err: fmt.Errorf("parse error: %v", err)}
			return
		}

		done <- core.NodeActionResult{
			Outputs: []core.FlowValue{core.NewTimestampValue(t)},
		}
	}()
	return done
}

func (a *ParseTimeAction) RunContext(ctx context.Context, n *core.Node) <-chan core.NodeActionResult {
	return a.Run(n)
}

func (a *ParseTimeAction) Serialize(s *core.Serializer) bool {
	core.SStr(s, &a.Format)
	return s.Ok()
}