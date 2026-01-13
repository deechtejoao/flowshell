package app

import (
	"context"

	"github.com/bvisness/flowshell/clay"
	hook "github.com/robotn/gohook"
)

// GEN:NodeAction
type GetMousePositionAction struct{}

func NewGetMousePositionNode() *Node {
	return &Node{
		Name: "Get Mouse Position",
		OutputPorts: []NodePort{
			{Name: "X", Type: FlowType{Kind: FSKindInt64}},
			{Name: "Y", Type: FlowType{Kind: FSKindInt64}},
		},
		Action: &GetMousePositionAction{},
	}
}

var _ NodeAction = &GetMousePositionAction{}

func (a *GetMousePositionAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *GetMousePositionAction) UI(n *Node) {
	clay.CLAY(clay.IDI("GetMousePosition", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.CLAY(clay.IDI("Row1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER, ChildGap: S2},
		}, func() {
			UISpacer(clay.AUTO_ID, GROWH)
			clay.TEXT("X", clay.TextElementConfig{TextColor: White})
			UIOutputPort(n, 0)
		})
		clay.CLAY(clay.IDI("Row2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER, ChildGap: S2},
		}, func() {
			UISpacer(clay.AUTO_ID, GROWH)
			clay.TEXT("Y", clay.TextElementConfig{TextColor: White})
			UIOutputPort(n, 1)
		})
	})
}

func (a *GetMousePositionAction) Run(n *Node) <-chan NodeActionResult {
	// Not used since we use RunContext
	return nil
}

func (a *GetMousePositionAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	// Not implemented yet
	return nil
}

func (a *GetMousePositionAction) Serialize(s *Serializer) bool {
	return s.Ok()
}

// GEN:NodeAction
type WaitForClickAction struct{}

func NewWaitForClickNode() *Node {
	return &Node{
		Name: "Wait For Click",
		OutputPorts: []NodePort{
			{Name: "X", Type: FlowType{Kind: FSKindInt64}},
			{Name: "Y", Type: FlowType{Kind: FSKindInt64}},
		},
		Action: &WaitForClickAction{},
	}
}

var _ NodeAction = &WaitForClickAction{}

func (a *WaitForClickAction) UpdateAndValidate(n *Node) {
	n.Valid = true
}

func (a *WaitForClickAction) UI(n *Node) {
	clay.CLAY(clay.IDI("WaitForClick", n.ID), clay.EL{
		Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: GROWH, ChildGap: S2},
	}, func() {
		clay.CLAY(clay.IDI("Row1", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER, ChildGap: S2},
		}, func() {
			UISpacer(clay.AUTO_ID, GROWH)
			clay.TEXT("X", clay.TextElementConfig{TextColor: White})
			UIOutputPort(n, 0)
		})
		clay.CLAY(clay.IDI("Row2", n.ID), clay.EL{
			Layout: clay.LAY{Sizing: GROWH, ChildAlignment: YCENTER, ChildGap: S2},
		}, func() {
			UISpacer(clay.AUTO_ID, GROWH)
			clay.TEXT("Y", clay.TextElementConfig{TextColor: White})
			UIOutputPort(n, 1)
		})
	})
}

func (a *WaitForClickAction) Run(n *Node) <-chan NodeActionResult {
	return nil
}

func (a *WaitForClickAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		// Just reuse the logic above?
		// Wait, GetMousePosition should ideally result *immediately*.
		// But without robotgo (automation), we can't query.
		// So `NewGetMousePositionNode` is misnamed if it waits.
		// Let's rename the first one to "Wait For Click" and maybe drop the second,
		// or make one that waits for *ANY* mouse movement (Wait For Move).

		// I'll implement "Wait For Click" specifically.

		hooks := hook.Start()
		defer hook.End()

		var x, y int16

		found := false
		for !found {
			select {
			case <-ctx.Done():
				done <- NodeActionResult{Err: ctx.Err()}
				return
			case ev := <-hooks:
				if ev.Kind == hook.MouseDown && ev.Button == hook.MouseMap["left"] {
					x = ev.X
					y = ev.Y
					found = true
				}
			}
		}

		done <- NodeActionResult{
			Outputs: []FlowValue{
				NewInt64Value(int64(x), 0),
				NewInt64Value(int64(y), 0),
			},
		}
	}()
	return done
}

func (a *WaitForClickAction) Serialize(s *Serializer) bool {
	return s.Ok()
}
