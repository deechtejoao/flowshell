package app

import (
	"math"

	"github.com/bvisness/flowshell/clay"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const MinimapSize = 200
const MinimapPadding = 10

func UIMinimap() {
	if len(currentGraph.Nodes) <= CurrentSettings.MinimapThreshold {
		return
	}

	// 1. Calculate Graph Bounds
	minX, minY := float32(math.Inf(1)), float32(math.Inf(1))
	maxX, maxY := float32(math.Inf(-1)), float32(math.Inf(-1))

	for _, n := range currentGraph.Nodes {
		minX = min(minX, n.Pos.X)
		minY = min(minY, n.Pos.Y)
		// Use estimated node size if not rendered yet, or actual size
		width, height := float32(200), float32(100)
		if data, ok := clay.GetElementData(n.ClayID()); ok {
			width = data.BoundingBox.Width
			height = data.BoundingBox.Height
		}
		maxX = max(maxX, n.Pos.X+width)
		maxY = max(maxY, n.Pos.Y+height)
	}

	// Add viewport to bounds to ensure we can always see where we are even if looking at nothing
	viewTL := Camera.ScreenToWorld(rl.Vector2{X: 0, Y: 0})
	viewBR := Camera.ScreenToWorld(rl.Vector2{X: float32(rl.GetScreenWidth()), Y: float32(rl.GetScreenHeight())})

	// Union with Viewport? Maybe not. Minimap usually shows content.
	// But if we are far away, we want to see the view rect relative to content.
	// So yes, verify bounds include view.
	// Actually, usually minimap is tied to Content. If view is outside, view rect is clipped or off-minimap.
	// Let's stick to Content Bounds + padding.

	padding := float32(500) // World units padding
	minX -= padding
	minY -= padding
	maxX += padding
	maxY += padding

	graphW := maxX - minX
	graphH := maxY - minY
	if graphW <= 0 {
		graphW = 1000
	}
	if graphH <= 0 {
		graphH = 1000
	}

	// Aspect ratio
	minimapW, minimapH := float32(MinimapSize), float32(MinimapSize)
	scaleX := minimapW / graphW
	scaleY := minimapH / graphH
	scale := min(scaleX, scaleY)

	// Center the conceptual graph in the minimap rect
	// If graph is wide, we have vert bars. If tall, horz bars.
	// But let's just align top-left for simplicity or center?
	// Let's center.

	// Minimap -> World (for clicking)
	miniToWorld := func(pos rl.Vector2) rl.Vector2 {
		relX := pos.X / scale
		relY := pos.Y / scale
		return rl.Vector2{X: minX + relX, Y: minY + relY}
	}

	clay.CLAY(clay.ID("Minimap"), clay.EL{
		Floating: clay.FloatingElementConfig{
			AttachTo: clay.AttachToParent,
			AttachPoints: clay.FloatingAttachPoints{
				Parent:  clay.AttachPointRightBottom,
				Element: clay.AttachPointRightBottom,
			},
			Offset: clay.Vector2{X: -20, Y: -20},
		},
		Layout: clay.LAY{
			Sizing: clay.Sizing{
				Width:  clay.SizingFixed(minimapW),
				Height: clay.SizingFixed(minimapH),
			},
		},
		BackgroundColor: clay.Color{R: 20, G: 22, B: 25, A: 200}, // Charcoal transparent
		Border:          clay.B{Color: Gray, Width: BA},
		CornerRadius:    RA2,
	}, func() {
		// Render Nodes
		// Since we can't use immediate mode Raylib drawing EASILY inside Clay structure (Clay renders later),
		// We should use Custom element or just use Clay rectangles for nodes if count is low?
		// Creating 1000 clay elements for minimap nodes might be heavy?
		// "Custom" element allows us to issue a callback during render phase.
		// Use clay.Custom!

		clay.CLAY(clay.ID("MinimapContent"), clay.EL{
			Layout: clay.LAY{Sizing: GROWALL},
			Custom: clay.CustomElementConfig{
				CustomData: &MinimapRenderData{
					Nodes: currentGraph.Nodes,
					MinX:  minX, MinY: minY,
					Scale:  scale,
					Graph:  currentGraph,
					ViewTL: viewTL, ViewBR: viewBR,
				},
			},
		}, func() {
			// Interaction
			clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
				// Click to jump / drag
				if pointerData.State == clay.PointerDataPressedThisFrame || pointerData.State == clay.PointerDataPressed {
					localPos := pointerData.Position
					// pointerData.Position is screen space. We need local space relative to this element.
					// But we don't easily know this element's Screen Pos until render?
					// Wait, clay.OnHover gives us data.
					// Actually, simpler: Use Raylib input directly if we know the rect?
					// Or assume top-left of this content is roughly known?

					// Better: The Custom renderer knows the bounds. But we are in layout phase.
					// Let's rely on standard logic.
					// We need to know where we clicked relative to the minimap box.
					// Since it's floating bottom-right...

					// Approximate approach: calculate screen pos of minimap based on screen size?
					screenW := float32(rl.GetScreenWidth())
					screenH := float32(rl.GetScreenHeight())
					miniX := screenW - 20 - minimapW
					miniY := screenH - 20 - minimapH

					clickX := localPos.X - miniX
					clickY := localPos.Y - miniY

					worldTarget := miniToWorld(rl.Vector2{X: clickX, Y: clickY})

					// Jump camera center to there
					// Camera Target is Top Left? No, Camera logic uses Offset/Target.
					// Target is "World coordinate that appears at Offset(Screen Center)".
					// So if we set Target = worldTarget, center of screen looks at worldTarget.
					// Assuming Offset is center.
					// Check Camera.go: Offset is usually set to ScreenWidth/2, ScreenHeight/2 in app.go?
					// Let's assume Offset is set correctly.
					// But usually Camera Target represents the world point at (0,0) of screen if Offset is 0.
					// Let's double check app.go camera init.

					// Safe bet: Center camera on worldTarget.
					// c.Target = worldTarget - (ScreenCenter / Zoom)?
					// Actually, Camera.Target is the point in world space that is mapped to Camera.Offset in screen space.
					// If Offset is (0,0), Target is TopLeft.
					// We likely want to center the viewport on the click.

					// Just set Target = worldTarget?
					// We need to check offsets.

					// Let's just set the target directly for now and refine.
					Camera.Target = worldTarget

					// Compensate for view center if needed.
					// If Camera.Offset is (0,0), we are setting Top-Left.
					// If we want to center, we need to subtrac half viewport in world units.
					viewW := (viewBR.X - viewTL.X)
					viewH := (viewBR.Y - viewTL.Y)
					Camera.Target.X -= viewW / 2
					Camera.Target.Y -= viewH / 2
				}
			}, nil)
		})
	})
}

type MinimapRenderData struct {
	Nodes          []*Node
	Graph          *Graph
	MinX, MinY     float32
	Scale          float32
	ViewTL, ViewBR rl.Vector2
}

// GEN:CustomRender
func RenderMinimap(bbox clay.BoundingBox, config clay.CustomElementConfig) {
	data := config.CustomData.(*MinimapRenderData)

	originX := bbox.X
	originY := bbox.Y

	// Draw Nodes
	for _, n := range data.Nodes {
		// Map node pos to local
		rectX := originX + (n.Pos.X-data.MinX)*data.Scale
		rectY := originY + (n.Pos.Y-data.MinY)*data.Scale

		// Map Size
		// We don't have exact size here easily without lookup, assume standard or small rect
		w := 200 * data.Scale
		h := 100 * data.Scale

		// Use actual size if available?
		// We are in render phase, we could technically look it up but `clay` context is locked?
		// Just use estimate for minimap.

		rl.DrawRectangleV(rl.Vector2{X: rectX, Y: rectY}, rl.Vector2{X: w, Y: h}, rl.Fade(rl.LightGray, 0.5))
	}

	// Draw Viewport
	vx := originX + (data.ViewTL.X-data.MinX)*data.Scale
	vy := originY + (data.ViewTL.Y-data.MinY)*data.Scale
	vw := (data.ViewBR.X - data.ViewTL.X) * data.Scale
	vh := (data.ViewBR.Y - data.ViewTL.Y) * data.Scale

	rl.DrawRectangleLinesEx(rl.Rectangle{X: vx, Y: vy, Width: vw, Height: vh}, 1, rl.White)
}
