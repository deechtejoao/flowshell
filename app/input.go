package app

// Global state to track if the mouse is currently hovering over any HIGH PRIORITY UI element
// (buttons, textboxes, dropdowns).
// This blocks ALL global actions (Node Drag, Wire Drag, Panning).
var IsHoveringUI bool

// Global state to track if the mouse is currently hovering over any LOW PRIORITY UI element
// (panels, backgrounds, node bodies).
// This blocks Panning, but allows specific global actions like Node Dragging (if the geometry matches).
var IsHoveringPanel bool
