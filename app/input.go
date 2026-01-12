package app

import "github.com/bvisness/flowshell/clay"

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
	if m.pointerReleasedThisFrame {
		m.pointerDownOwnerID = 0
		m.pointerDownOwnerSet = false
	}

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
	if !m.pointerPressedThisFrame {
		return
	}

	m.pointerDownOwnerID = owner.ID
	m.pointerDownOwnerSet = true
}

func (m *InputManager) IsClick(owner clay.ElementID, pointerData clay.PointerData) bool {
	if pointerData.State != clay.PointerDataReleasedThisFrame {
		return false
	}
	if !m.pointerReleasedThisFrame {
		return false
	}
	if !m.pointerDownOwnerSet {
		return false
	}
	return m.pointerDownOwnerID == owner.ID
}
