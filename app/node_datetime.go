package app

import (
	"context"
	"fmt"
	"time"

	"github.com/bvisness/flowshell/clay"
)

// GEN:NodeAction
type ParseTimeAction struct {
	Format string
}

func NewParseTimeNode() *Node {
	return &Node{
		Name: "Parse Time",
		InputPorts: []NodePort{
			{Name: "Text", Type: FlowType{Kind: FSKindBytes}},
			{Name: "Format", Type: FlowType{Kind: FSKindBytes}}, // Optional override
		},
		OutputPorts: []NodePort{
			{Name: "Timestamp", Type: FlowType{Kind: FSKindInt64, WellKnownType: FSWKTTimestamp}},
		},
		Action: &ParseTimeAction{Format: time.RFC3339},
	}
}

var _ NodeAction = &ParseTimeAction{}

func (a *ParseTimeAction) UpdateAndValidate(n *Node) { n.Valid = true }

func (a *ParseTimeAction) UI(n *Node) {
	clay.CLAY(clay.IDI("TimeUI", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.CLAY(clay.IDI("TimeRow1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER},
		}, func() {
			UIInputPort(n, 0)
			UISpacer(clay.IDI("TimeSpacer1", n.ID), GROWH)
			UIOutputPort(n, 0)
		})

		clay.CLAY(clay.IDI("TimeRow2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH},
		}, func() {
			UITextBox(clay.IDI("FormatStr", n.ID), &a.Format, UITextBoxConfig{
				El: clay.EL{Layout: clay.LAY{Sizing: GROWH}},
			})
		})

		clay.CLAY(clay.IDI("TimeRow3", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH},
		}, func() {
			clay.TEXT("Examples: 2006-01-02, Jan 2, 15:04:05", clay.TextElementConfig{FontSize: 12, TextColor: LightGray})
		})
	})
}

func (a *ParseTimeAction) Run(n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		valText, _, err := n.GetInputValue(0)
		if err != nil {
			done <- NodeActionResult{Err: err}
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
			done <- NodeActionResult{Err: fmt.Errorf("parse error: %v", err)}
			return
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{NewTimestampValue(t)},
		}
	}()
	return done
}

func (a *ParseTimeAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	return a.Run(n)
}

func (a *ParseTimeAction) Tag() string { return "ParseTime" }
func (a *ParseTimeAction) Serialize(s *Serializer) bool {
	SStr(s, &a.Format)
	return s.Ok()
}
