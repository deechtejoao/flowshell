package app

import (
	"github.com/bvisness/flowshell/clay"
)

// Global state to track if the mouse is currently hovering over any HIGH PRIORITY UI element
// (buttons, textboxes, dropdowns).
// This blocks ALL global actions (Node Drag, Wire Drag, Panning).
var IsHoveringUI bool

// Global state to track if the mouse is currently hovering over any LOW PRIORITY UI element
// (panels, backgrounds, node bodies).
// This blocks Panning, but allows specific global actions like Node Dragging (if the geometry matches).
var IsHoveringPanel bool

var UIInput InputManager

type InputManager struct {
	pointerDown              bool
	pointerPressedThisFrame  bool
	pointerReleasedThisFrame bool

	pointerDownOwnerID  uint32
	pointerDownOwnerSet bool
}

func (m *InputManager) BeginFrame(pointerDown bool) {
	m.pointerPressedThisFrame = pointerDown && !m.pointerDown
	m.pointerReleasedThisFrame = !pointerDown && m.pointerDown
	m.pointerDown = pointerDown

	if m.pointerPressedThisFrame {
		m.pointerDownOwnerID = 0
		m.pointerDownOwnerSet = false
	}
}

func (m *InputManager) EndFrame() {
	m.pointerPressedThisFrame = false
	m.pointerReleasedThisFrame = false
}

func (m *InputManager) RegisterPointerDown(owner clay.ElementID, pointerData clay.PointerData) {
	if pointerData.State != clay.PointerDataPressedThisFrame {
		return
	}
	if m.pointerDownOwnerSet {
		return
	}
	// We trust Clay's event timing even if it's lagging behind our InputManager's frame detection.

	m.pointerDownOwnerID = owner.ID
	m.pointerDownOwnerSet = true
}

func (m *InputManager) IsClick(owner clay.ElementID, pointerData clay.PointerData) bool {
	if pointerData.State != clay.PointerDataReleasedThisFrame {
		return false
	}
	// We don't check m.pointerReleasedThisFrame because Clay might be late.

	if !m.pointerDownOwnerSet {
		return false
	}
	return m.pointerDownOwnerID == owner.ID
}
