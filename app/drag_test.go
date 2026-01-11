package app

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type MockInputProvider struct {
	KeysPressed     map[int32]bool
	ButtonsReleased map[rl.MouseButton]bool
	ButtonsUp       map[rl.MouseButton]bool
	ButtonsDown     map[rl.MouseButton]bool
	MousePos        rl.Vector2
}

func (m *MockInputProvider) IsKeyPressed(key int32) bool {
	return m.KeysPressed[key]
}
func (m *MockInputProvider) IsMouseButtonReleased(button rl.MouseButton) bool {
	return m.ButtonsReleased[button]
}
func (m *MockInputProvider) IsMouseButtonUp(button rl.MouseButton) bool {
	return m.ButtonsUp[button]
}
func (m *MockInputProvider) IsMouseButtonDown(button rl.MouseButton) bool {
	return m.ButtonsDown[button]
}
func (m *MockInputProvider) GetMousePosition() rl.Vector2 {
	return m.MousePos
}

func NewMockInput() *MockInputProvider {
	return &MockInputProvider{
		KeysPressed:     make(map[int32]bool),
		ButtonsReleased: make(map[rl.MouseButton]bool),
		ButtonsUp:       make(map[rl.MouseButton]bool),
		ButtonsDown:     make(map[rl.MouseButton]bool),
	}
}

func TestDragState_Lifecycle(t *testing.T) {
	mock := NewMockInput()
	d := DragState{Input: mock}

	// 1. Idle state
	mock.ButtonsUp[rl.MouseLeftButton] = true
	d.Update()
	if d.Pending || d.Dragging {
		t.Error("Should be idle")
	}

	// 2. Mouse Down (Pending)
	mock.ButtonsUp[rl.MouseLeftButton] = false
	mock.ButtonsDown[rl.MouseLeftButton] = true
	mock.MousePos = rl.Vector2{X: 100, Y: 100}
	d.Update()
	if !d.Pending {
		t.Error("Should be pending")
	}
	if d.MouseStart != (rl.Vector2{X: 100, Y: 100}) {
		t.Error("MouseStart not recorded")
	}

	// 3. TryStartDrag (Fail - moved too little)
	thing := "my object"
	mock.MousePos = rl.Vector2{X: 101, Y: 101}
	ok := d.TryStartDrag(thing, rl.Rectangle{X: 0, Y: 0, Width: 200, Height: 200}, rl.Vector2{})
	if ok {
		t.Error("Should not start drag (movement too small)")
	}

	// 4. TryStartDrag (Success)
	mock.MousePos = rl.Vector2{X: 105, Y: 105}
	ok = d.TryStartDrag(thing, rl.Rectangle{X: 0, Y: 0, Width: 200, Height: 200}, rl.Vector2{X: 10, Y: 10})
	if !ok {
		t.Error("Should start drag")
	}
	if !d.Dragging || d.Pending {
		t.Error("Should be dragging, not pending")
	}
	if d.Thing != thing {
		t.Error("Thing mismatch")
	}

	// 5. Dragging update
	d.Update() // Mouse still down
	if !d.Dragging {
		t.Error("Should still be dragging")
	}

	// 6. Release
	mock.ButtonsDown[rl.MouseLeftButton] = false
	mock.ButtonsReleased[rl.MouseLeftButton] = true
	d.Update()
	if d.Dragging {
		t.Error("Should stop dragging on release")
	}

	// 7. Cleanup (Next frame Up)
	mock.ButtonsReleased[rl.MouseLeftButton] = false
	mock.ButtonsUp[rl.MouseLeftButton] = true
	d.Update()
	if d.Thing != nil {
		t.Error("Thing should be cleared")
	}
}

func TestDragState_Cancel(t *testing.T) {
	mock := NewMockInput()
	d := DragState{Input: mock}

	// Start dragging
	mock.ButtonsDown[rl.MouseLeftButton] = true
	mock.MousePos = rl.Vector2{X: 100, Y: 100}
	d.Update() // Pending
	mock.MousePos = rl.Vector2{X: 110, Y: 110}
	d.TryStartDrag("thing", rl.Rectangle{X: 0, Y: 0, Width: 200, Height: 200}, rl.Vector2{})

	if !d.Dragging {
		t.Fatal("Failed to start drag")
	}

	// Cancel with Escape
	mock.KeysPressed[rl.KeyEscape] = true
	d.Update()

	if d.Dragging {
		t.Error("Should stop dragging on Escape")
	}
	if !d.Canceled {
		t.Error("Should be marked as canceled")
	}
}
