package app

import (
	"github.com/bvisness/flowshell/app/core"
	"fmt"

	_ "github.com/bvisness/flowshell/app/nodes"
	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const windowWidth = 1920

func Main() {
	CurrentSettings = LoadSettings()
	core.ApplyTheme(CurrentSettings.Theme)

	rl.SetTraceLogLevel(rl.LogError)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	// rl.InitWindow(windowWidth, windowHeight, "Flowshell")
	rl.InitWindow(int32(CurrentSettings.WindowWidth), int32(CurrentSettings.WindowHeight), "Flowshell")
	defer rl.CloseWindow()

	if CurrentSettings.WindowMaximized {
		rl.MaximizeWindow()
	}

	monitorWidth := float32(rl.GetMonitorWidth(rl.GetCurrentMonitor()))
	monitorHeight := float32(rl.GetMonitorHeight(rl.GetCurrentMonitor()))

	// Center if not maximized and looks like default or sane
	// Actually InitWindow usually centers or puts it somewhere.
	// If we want to restore position we need to save it too.
	// User only asked for window state (implied size/maximized).
	// Let's just center if it's the first run (default size).
	if CurrentSettings.WindowWidth == 1920 && CurrentSettings.WindowHeight == 1080 {
		rl.SetWindowPosition(int(monitorWidth/2-float32(CurrentSettings.WindowWidth)/2), int(monitorHeight/2-float32(CurrentSettings.WindowHeight)/2))
	}

	rl.SetTargetFPS(int32(rl.GetMonitorRefreshRate(rl.GetCurrentMonitor())))

	defer func() {
		// Save Settings on exit
		CurrentSettings.WindowWidth = rl.GetScreenWidth()
		CurrentSettings.WindowHeight = rl.GetScreenHeight()
		CurrentSettings.WindowMaximized = rl.IsWindowMaximized()
		if err := SaveSettings(CurrentSettings); err != nil {
			fmt.Printf("Error saving settings: %v\n", err)
		}
	}()

	initImages()

	// Load Plugins
	pluginNodes, err := LoadPlugins("plugins")
	if err != nil {
		fmt.Printf("Error loading plugins: %v\n", err)
	} else {
		nodeTypes = append(nodeTypes, pluginNodes...)
	}

	clay.SetMaxElementCount(1 << 16)
	arena := clay.CreateArenaWithCapacity(uintptr(clay.MinMemorySize()))
	clay.Initialize(
		arena,
		clay.Dimensions{Width: float32(CurrentSettings.WindowWidth), Height: float32(CurrentSettings.WindowHeight)}, // Initial size
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

	// Handle Zoom Input
	wheel := rl.GetMouseWheelMove()
	shouldZoom := wheel != 0 && !core.IsHoveringUI
	if shouldZoom {
		Camera.ZoomAt(rl.GetMousePosition(), util.Tern(wheel > 0, float32(1.1), float32(0.9)))
	}

	if rl.IsKeyPressed(rl.KeyF9) {
		clay.SetDebugModeEnabled(!clay.IsDebugModeEnabled())
	}

	// Update core.Graph Logic
	topoErr := UpdateGraph()

	// Reset global UI hover state for this frame.
	core.IsHoveringUI = false
	core.IsHoveringPanel = false

	clayPointerMouseDown := rl.IsMouseButtonDown(rl.MouseButtonLeft)
	if drag.Dragging {
		clayPointerMouseDown = false
	}
	core.UIInput.BeginFrame(clayPointerMouseDown)

	// --- Layout 1: Nodes (World Space -> Screen Space Mapped) ---
	// We map the Clay layout to the screen directly, but manually position nodes
	// using WorldToScreen. This allows us to handle infinite canvas interactions
	// correctly (Clay ignores inputs outside its layout bounds) while keeping
	// the UI elements at a constant pixel size (no semantic zoom distortion).
	clay.SetPointerState(
		clay.V2{X: float32(rl.GetMouseX()), Y: float32(rl.GetMouseY())},
		clayPointerMouseDown,
	)
	screenWidth := float32(rl.GetScreenWidth())
	screenHeight := float32(rl.GetScreenHeight())
	clay.SetLayoutDimensions(clay.D{Width: screenWidth, Height: screenHeight})

	scrollDelta := clay.Vector2{X: 0, Y: 0}
	if !shouldZoom {
		scrollDelta = clay.Vector2(rl.GetMouseWheelMoveV()).Times(4)
	}
	clay.UpdateScrollContainers(false, scrollDelta, rl.GetFrameTime())

	clay.BeginLayout()
	UINodes(topoErr)
	nodesRenderCommands := clay.EndLayout()

	// Update cached layout info based on the World Space layout
	afterLayout()

	// --- Layout 2: Overlay (Screen Space) ---
	clay.SetPointerState(
		clay.V2{X: float32(rl.GetMouseX()), Y: float32(rl.GetMouseY())},
		clayPointerMouseDown,
	)
	clay.SetLayoutDimensions(clay.D{Width: screenWidth, Height: screenHeight})

	// We don't update scroll containers again to avoid double application of scroll?
	// But if overlay has scrollable areas, they need it.
	// Assuming overlay doesn't scroll with same wheel event as nodes if they are separate.
	// For now, we only called it once above. It updates global state.
	// If overlay has scrollable elements, they might need a second update or just rely on the first one?
	// Clay's UpdateScrollContainers iterates ALL scroll containers.
	// It should be fine.

	clay.BeginLayout()
	UIOverlay(topoErr)
	overlayRenderCommands := clay.EndLayout()

	processInput()

	core.UIInput.EndFrame()

	rl.BeginDrawing()
	rl.ClearBackground(core.Night.RGBA())

	// World Space (Mapped to Screen Space)
	renderWorldOverlays()
	renderClayCommands(nodesRenderCommands)

	// Screen Space
	renderClayCommands(overlayRenderCommands)
	renderScreenOverlays()

	rl.EndDrawing()
	clay.ReleaseFrameMemory()

	// Update focus tracking
	if core.UIFocus != nil {
		core.LastUIFocus = *core.UIFocus
		core.LastUIFocusValid = true
	} else {
		core.LastUIFocusValid = false
	}
}

func handleClayErrors(errorData clay.ErrorData) {
	fmt.Printf("CLAY ERROR: %s\n", errorData.ErrorText)
}
