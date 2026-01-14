package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	ActiveMenu  string // Name of the currently open menu ("" if none)
	ShowMenuBar = false
)

func UIMenuBar() {
	core.WithZIndex(core.Z_CONTEXT_MENU, func() {
		// Overlay to close menu on click outside
		if ActiveMenu != "" {
			clay.CLAY(clay.ID("MenuOverlay"), clay.EL{
				Layout: clay.LAY{Sizing: core.GROWALL},
				Floating: clay.FLOAT{
					AttachTo:           clay.AttachToRoot,
					ZIndex:             core.Z_CONTEXT_MENU - 1,        // Behind menu
					PointerCaptureMode: clay.PointercaptureModeCapture, // Block clicks
				},
			}, func() {
				// Click anywhere to close
				if core.UIInput.IsClick(clay.ID("MenuOverlay"), clay.PointerData{}) {
					ActiveMenu = ""
				}
				// Also handle standard mouse click if overlay isn't catching it?
				// Clay captures pointer if PointerCaptureModeCapture is on.
				core.UIInput.RegisterPointerDown(clay.ID("MenuOverlay"), clay.PointerData{}, core.Z_CONTEXT_MENU-1)
			})
		}

		clay.CLAY(clay.ID("MenuBar"), clay.EL{
			Layout:          clay.LAY{Sizing: clay.Sizing{Width: core.GROWALL.Width, Height: clay.SizingFixed(30)}, LayoutDirection: clay.LeftToRight, ChildAlignment: clay.ChildAlignment{Y: clay.AlignYCenter}, Padding: core.PVH(0, core.S2), ChildGap: core.S2},
			BackgroundColor: core.Charcoal,
			Border:          clay.B{Width: clay.BW{Bottom: 1}, Color: core.Gray},
			Floating: clay.FLOAT{
				AttachTo: clay.AttachToRoot,
				ZIndex:   core.Z_CONTEXT_MENU,
			},
		}, func() {
			// File Menu
			UIMenuItem("File", func() {
				UIMenuDropdownItem("New", func() {
					ActiveMenu = ""
					// Confirm discard?
					core.PushHistory()
					CurrentGraph = core.NewGraph()
					CurrentFilename = ""
				})
				UIMenuDropdownItem("Open... (Ctrl+L)", func() {
					ActiveMenu = ""
					cwd, _ := os.Getwd()
					filename, ok, err := core.OpenFileDialog("Open Flow", cwd, map[string]string{"flow": "Flow Files"})
					if err != nil {
						fmt.Printf("Load error: %v\n", err)
						core.ShowInfoDialog("Error", fmt.Sprintf("Failed to open file dialog: %v", err))
					} else if ok {
						core.PushHistory()
						if g, err := core.LoadGraph(filename); err == nil {
							CurrentGraph = g
							CurrentFilename = filename
						}
					}
				})
				UIMenuDropdownItem("Save (Ctrl+S)", func() {
					ActiveMenu = ""
					if CurrentFilename != "" {
						err := core.SaveGraph(CurrentFilename, CurrentGraph)
						if err != nil {
							fmt.Printf("Save error: %v\n", err)
						}
					} else {
						cwd, _ := os.Getwd()
						filename, ok, err := core.SaveFileDialog("Save Flow", cwd, map[string]string{"flow": "Flow Files"})
						if err != nil {
							fmt.Printf("Save error: %v\n", err)
						} else if ok {
							if filepath.Ext(filename) != ".flow" {
								filename += ".flow"
							}
							if err := core.SaveGraph(filename, CurrentGraph); err == nil {
								CurrentFilename = filename
							}
						}
					}
				})
				UIMenuDropdownItem("Save As... (Ctrl+Shift+S)", func() {
					ActiveMenu = ""
					initialDir, _ := os.Getwd()
					if CurrentFilename != "" {
						initialDir = filepath.Dir(CurrentFilename)
					}
					filename, ok, err := core.SaveFileDialog("Save Flow", initialDir, map[string]string{"flow": "Flow Files"})
					if err != nil {
						fmt.Printf("Save error: %v\n", err)
					} else if ok {
						if filepath.Ext(filename) != ".flow" {
							filename += ".flow"
						}
						if err := core.SaveGraph(filename, CurrentGraph); err == nil {
							CurrentFilename = filename
						}
					}
				})
				UIMenuSeparator()
				UIMenuDropdownItem("Quit", func() {
					ActiveMenu = ""
					ShouldQuit = true
				})
			})

			// View Menu
			UIMenuItem("View", func() {
				UIMenuDropdownItem("VariablesPanel", func() {
					ActiveMenu = ""
					ShowVariables = !ShowVariables
				})
				UIMenuDropdownItem("Reset View", func() {
					ActiveMenu = ""
					Camera.Zoom = 1.0
					Camera.Target = rl.Vector2{X: 0, Y: 0}
				})
			})

			// Help Menu
			UIMenuItem("Help", func() {
				UIMenuDropdownItem("About", func() {
					ActiveMenu = ""
					core.ShowInfoDialog("About Flowshell", "Flowshell\n\nA node-based shell automation tool.\n\nCreated by BVisness")
				})
			})
		})
	})
}

func UIMenuItem(label string, content func()) {
	id := clay.IDI("Menu"+label, 0)
	isOpen := ActiveMenu == label

	core.UIButton(id, core.UIButtonConfig{
		El: clay.EL{
			Layout:          clay.LAY{Padding: core.PVH(core.S1, core.S2)},
			BackgroundColor: util.Tern(isOpen, core.Gray, util.Tern(clay.Hovered(), core.HoverWhite, clay.Color{})),
			CornerRadius:    core.RA(4),
		},
		OnHover: func(_ clay.ElementID, _ clay.PointerData, _ any) {
			// Hover switching
			if ActiveMenu != "" && ActiveMenu != label {
				ActiveMenu = label
			}
		},
		OnClick: func(_ clay.ElementID, _ clay.PointerData, _ any) {
			if ActiveMenu == label {
				ActiveMenu = ""
			} else {
				ActiveMenu = label
			}
		},
	}, func() {
		clay.TEXT(label, clay.T{TextColor: core.White, FontSize: 14})
	})

	if isOpen {
		core.WithZIndex(core.Z_DROPDOWN_BG, func() {
			clay.CLAY(clay.IDI("MenuDropdown"+label, 0), clay.EL{
				Layout: clay.LAY{
					LayoutDirection: clay.TopToBottom,
					Padding:         core.PA1,
					ChildGap:        core.S1,
					Sizing:          clay.Sizing{Width: clay.SizingFixed(200)}, // Fixed width for now
				},
				BackgroundColor: core.Charcoal,
				Border:          clay.B{Width: core.BA, Color: core.Gray},
				CornerRadius:    core.RA(4),
				Floating: clay.FLOAT{
					AttachTo: clay.AttachToElementWithID,
					ParentID: id.ID,
					AttachPoints: clay.FloatingAttachPoints{
						Parent:  clay.AttachPointLeftBottom,
						Element: clay.AttachPointLeftTop,
					},
					ZIndex: core.Z_DROPDOWN_BG,
				},
			}, content)
		})
	}
}

func UIMenuDropdownItem(label string, onClick func()) {
	core.UIButton(clay.ID("MenuItem"+label), core.UIButtonConfig{
		El: clay.EL{
			Layout:          clay.LAY{Padding: core.PVH(core.S2, core.S2), Sizing: core.GROWH},
			BackgroundColor: util.Tern(clay.Hovered(), core.Blue, clay.Color{}),
			CornerRadius:    core.RA(2),
		},
		OnClick: func(_ clay.ElementID, _ clay.PointerData, _ any) {
			onClick()
		},
	}, func() {
		clay.TEXT(label, clay.T{TextColor: core.White, FontSize: 14})
	})
}

func UIMenuSeparator() {
	clay.CLAY(clay.AUTO_ID, clay.EL{
		Layout: clay.LAY{
			Sizing: clay.Sizing{
				Width:  core.GROWH.Width,
				Height: clay.SizingFixed(1),
			},
		},
		BackgroundColor: core.Gray,
	})
}
