package app

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var Camera = NewCameraState()

type CameraState struct {
	rl.Camera2D
}

func NewCameraState() *CameraState {
	return &CameraState{
		Camera2D: rl.Camera2D{
			Zoom: 1.0,
		},
	}
}

// WorldToScreen converts a world position to screen coordinates
func (c *CameraState) WorldToScreen(worldPos rl.Vector2) rl.Vector2 {
	return rl.GetWorldToScreen2D(worldPos, c.Camera2D)
}

// ScreenToWorld converts a screen position to world coordinates
func (c *CameraState) ScreenToWorld(screenPos rl.Vector2) rl.Vector2 {
	return rl.GetScreenToWorld2D(screenPos, c.Camera2D)
}

func (c *CameraState) Pan(delta rl.Vector2) {
	// Delta is in screen pixels (mouse movement).
	// We need to move the camera Target by the delta / zoom.
	// But wait, if we move the mouse right, we want the world to move right.
	// That means the camera Target (which is where we are looking) should move LEFT.
	// So we subtract.
	c.Target = rl.Vector2Subtract(c.Target, rl.Vector2Scale(delta, 1.0/c.Zoom))
}

func (c *CameraState) ZoomAt(screenPos rl.Vector2, factor float32) {
	worldPos := c.ScreenToWorld(screenPos)
	
	c.Zoom *= factor
	if c.Zoom < 0.1 {
		c.Zoom = 0.1
	}
	if c.Zoom > 5.0 {
		c.Zoom = 5.0
	}
	
	// Adjust Target so that worldPos remains under screenPos
	// NewTarget = WorldPos - (ScreenPos - Offset) / NewZoom
	// Raylib formula: Screen = (World - Target) * Zoom + Offset
	// => World - Target = (Screen - Offset) / Zoom
	// => Target = World - (Screen - Offset) / Zoom
	
	term := rl.Vector2Scale(rl.Vector2Subtract(screenPos, c.Offset), 1.0/c.Zoom)
	c.Target = rl.Vector2Subtract(worldPos, term)
}
