package app

import (
	"fmt"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const windowWidth = 1920
const windowHeight = 1080

// const windowWidth = 1280
// const windowHeight = 720

func Main() {
	// rl.SetConfigFlags(rl.FlagWindowResizable)

	rl.InitWindow(windowWidth, windowHeight, "Flowshell")
	defer rl.CloseWindow()

	monitorWidth := float32(rl.GetMonitorWidth(rl.GetCurrentMonitor()))
	monitorHeight := float32(rl.GetMonitorHeight(rl.GetCurrentMonitor()))
	// rl.SetWindowSize(windowWidth, windowHeight)
	rl.SetWindowPosition(int(monitorWidth/2-windowWidth/2), int(monitorHeight/2-windowHeight/2))
	rl.SetTargetFPS(int32(rl.GetMonitorRefreshRate(rl.GetCurrentMonitor())))

	initImages()

	clay.SetMaxElementCount(1 << 18)
	arena := clay.CreateArenaWithCapacity(uintptr(clay.MinMemorySize()))
	clay.Initialize(
		arena,
		clay.Dimensions{Width: windowWidth, Height: windowHeight},
		clay.ErrorHandler{ErrorHandlerFunction: handleClayErrors},
	)
	clay.SetMeasureTextFunction(func(str string, config *clay.TextElementConfig, userData any) clay.Dimensions {
		fontSize := config.FontSize
		if fontSize == 0 {
			fontSize = DefaultFontSize
		}
		font := LoadFont(config.FontID, int(fontSize))
		dims := rl.MeasureTextEx(font, str, float32(fontSize), float32(config.LetterSpacing))
		return clay.Dimensions{Width: dims.X, Height: dims.Y}
	}, nil)

	rl.SetExitKey(0)
	for !rl.WindowShouldClose() {
		frame()
	}
}

func frame() {
	drag.Update()
	beforeLayout()

	// Handle Zoom Input
	wheel := rl.GetMouseWheelMove()
	if wheel != 0 {
		Camera.ZoomAt(rl.GetMousePosition(), util.Tern(wheel > 0, float32(1.1), float32(0.9)))
	}

	if rl.IsKeyPressed(rl.KeyF9) {
		clay.SetDebugModeEnabled(!clay.IsDebugModeEnabled())
	}

	// Update Graph Logic
	topoErr := UpdateGraph()

	// Reset global UI hover state for this frame.
	IsHoveringUI = false
	IsHoveringPanel = false

	clayPointerMouseDown := rl.IsMouseButtonDown(rl.MouseButtonLeft)
	if drag.Dragging {
		clayPointerMouseDown = false
	}
	UIInput.BeginFrame(clayPointerMouseDown)

	// --- Layout 1: Nodes (World Space) ---
	worldMouse := Camera.ScreenToWorld(rl.GetMousePosition())
	clay.SetPointerState(
		clay.V2{X: worldMouse.X, Y: worldMouse.Y},
		clayPointerMouseDown,
	)
	clay.SetLayoutDimensions(clay.D{
		Width:  windowWidth / Camera.Zoom,
		Height: windowHeight / Camera.Zoom,
	})

	clay.BeginLayout()
	UINodes(topoErr)
	nodeCommands := clay.EndLayout()

	// --- Layout 2: Overlay (Screen Space) ---
	clay.SetPointerState(
		clay.V2{X: float32(rl.GetMouseX()), Y: float32(rl.GetMouseY())},
		clayPointerMouseDown,
	)
	clay.SetLayoutDimensions(clay.D{Width: windowWidth, Height: windowHeight})
	clay.UpdateScrollContainers(false, clay.Vector2(rl.GetMouseWheelMoveV()).Times(4), rl.GetFrameTime())

	clay.BeginLayout()
	UIOverlay(topoErr)
	overlayCommands := clay.EndLayout()

	afterLayout()
	UIInput.EndFrame()

	rl.BeginDrawing()
	rl.ClearBackground(Night.RGBA())

	rl.BeginMode2D(Camera.Camera2D)
	renderClayCommands(nodeCommands)
	renderWorldOverlays()
	rl.EndMode2D()

	renderClayCommands(overlayCommands)
	renderScreenOverlays()

	rl.EndDrawing()
	clay.ReleaseFrameMemory()
}

func handleClayErrors(errorData clay.ErrorData) {
	fmt.Printf("CLAY ERROR: %s\n", errorData.ErrorText)
}
