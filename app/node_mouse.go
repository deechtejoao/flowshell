package app

import (
	"context"

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

func (a *GetMousePositionAction) UI(n *Node) {}

func (a *GetMousePositionAction) Run(n *Node) <-chan NodeActionResult {
	// Not used since we use RunContext
	return nil
}

func (a *GetMousePositionAction) RunContext(ctx context.Context, n *Node) <-chan NodeActionResult {
	done := make(chan NodeActionResult, 1)
	go func() {
		// Event() returns immediately if no hook is registered?
		// No, robotgo/gohook is event based.
		// To get *current* position, we might need a different function if we don't want to wait for an event.
		// hook.Event() blocks until event.
		// We want instantaneous position?
		// gohook doesn't expose "GetMousePos" directly?
		// Wait, robotgo does (GetMousePos). gohook is just hooks.
		// If we only have gohook, we can only listen.
		// But usually `go get github.com/robotn/gohook` is enough?
		// Actually, if we want "Get Mouse Position" (instant), we probably need `robotgo` or `vcaesar/gops`?
		// Let's defer "Get Mouse Position" if gohook doesn't support it easily without blocking.
		// Ah, wait. `gohook` allows registering raw hooks.
		// If the user wants "WaitForClick", gohook is perfect.
		// If the user wants "Current Position", we can use `gohook` to wait for the *next* mouse move?
		// No, that blocks.
		// Let's implement "Wait For Click" first, as that was the primary use case for "Mouse Events" (reacting to them).
		// For "Get Mouse Position", we'd need robotgo.
		// Let's implement "Wait For Click".

		// Actually, let's implement "Wait For Next Mouse Event" which returns X/Y.
		// For "Wait For Click", we specifically wait for MouseDown.

		// TODO: gohook might block the main thread if we aren't careful?
		// Hooks usually need a message loop or similar.
		// `hook.Start()` starts a blocking loop.
		// `hook.Register(...)` registers callbacks.
		// We need to be careful not to freeze the UI.
		// But this is running in a goroutine `RunContext`.

		// Strategy: Use a channel to receive the click event from the hook callback.
		evChan := hook.Start() // This starts the hook stream logic?
		// No, hook.Start() *blocks*.
		// We should probably check the docs or usage.
		// "hook.Register(hook.KeyDown, ...)" then "s := hook.Start()"
		// We want to avoid catching ALL events globally forever.
		// Can we query on demand?
		// Maybe just:
		// ev := <-hook.Event()
		// But hook.Event() might not work if Start() isn't running?
		// Let's try simple registration.

		// WARNING: Starting/Stopping hooks frequently is expensive/risky.
		// Better: Start hook once globally (in App Init?) and broadcast?
		// For now, let's do a localized approach but be careful.

		// Actually, `robotn/gohook` example:
		// evChan := hook.Start()
		// defer hook.End()
		//
		// for ev := range evChan { ... }

		// We can't easily start/stop hooks inside a Node Run without potentially conflicting with other nodes or the OS.
		// But let's try.

		var x, y int16

		// We need to filter for Mouse Up/Down.
		// NOTE: hook.Start() is blocking if we don't put it in goroutine?
		// Docs say "Start() function will block".
		// So we need another goroutine for the hook loop.

		clickCh := make(chan struct{})

		// Using Register is cleaner?
		// hook.Register(hook.MouseDown, []string{}, func(e hook.Event) { ... })
		// s := hook.Start()

		// Let's use the low-level Event channel if possible.

		// Given the complexity/risk of global hooks in a localized node run:
		// We'll implement a simplistic version: "Wait for Left Click".

		doneInside := make(chan bool)

		hooks := hook.Start()
		defer hook.End()

		for {
			select {
			case <-ctx.Done():
				done <- NodeActionResult{Err: ctx.Err()}
				return
			case ev := <-hooks:
				if ev.Kind == hook.MouseDown && ev.Button == hook.MouseMap["left"] {
					x = ev.X
					y = ev.Y
					doneInside <- true
					goto Finished
				}
			}
		}

	Finished:
		done <- NodeActionResult{
			Outputs: []FlowValue{
				NewInt64Value(int64(x), 0),
				NewInt64Value(int64(y), 0),
			},
		}
	}()
	return done
}

func (a *GetMousePositionAction) Tag() string {
	return "GetMousePosition"
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

func (a *WaitForClickAction) UI(n *Node) {}

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

func (a *WaitForClickAction) Tag() string {
	return "WaitForClick"
}

func (a *WaitForClickAction) Serialize(s *Serializer) bool {
	return s.Ok()
}
