package app

import "github.com/bvisness/flowshell/clay"

var ColorLight = clay.Color{224, 215, 210, 255}
var ColorRed = clay.Color{168, 66, 28, 255}
var ColorOrange = clay.Color{225, 138, 50, 255}

func ui() {
	clay.CLAY(clay.ID("OuterContainer"), clay.EL{Layout: clay.LAY{Sizing: clay.Sizing{clay.SizingGrow(0, 0), clay.SizingGrow(0, 0)}, Padding: clay.PaddingAll(16), ChildGap: 16}, BackgroundColor: clay.Color{250, 250, 255, 255}}, func() {
		clay.CLAY(clay.ID("Sidebar"), clay.EL{
			Layout:          clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: clay.Sizing{Width: clay.SizingFixed(300), Height: clay.SizingGrow(0, 0)}, Padding: clay.PaddingAll(16), ChildGap: 16},
			BackgroundColor: ColorLight,
		}, func() {
			clay.CLAY(clay.ID("ProfilePictureOuter"), clay.EL{Layout: clay.LAY{Sizing: clay.Sizing{Width: clay.SizingGrow(0, 0)}, Padding: clay.PaddingAll(16), ChildGap: 16, ChildAlignment: clay.ChildAlignment{Y: clay.AlignYCenter}}, BackgroundColor: ColorRed}, func() {
				clay.TEXT("Clay - UI Library", clay.TextElementConfig{FontID: InterBold, FontSize: 24, TextColor: clay.Color{255, 255, 255, 255}})
			})

			for range 5 {
				sidebarItemComponent()
			}
		})
		clay.CLAY(clay.ID("MainContent"), clay.EL{Layout: clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: clay.Sizing{Width: clay.SizingGrow(0, 0), Height: clay.SizingGrow(0, 0)}, Padding: clay.PaddingAll(16), ChildGap: 8}, BackgroundColor: ColorLight}, func() {
			for f := range FontsEnd {
				clay.TEXT(fontFiles[f], clay.TextElementConfig{FontID: f, TextColor: clay.Color{0, 0, 0, 255}})
			}
		})
	})
}

func sidebarItemComponent() {
	clay.CLAY_AUTO_ID(clay.EL{Layout: clay.LAY{Sizing: clay.Sizing{Width: clay.SizingGrow(0, 0), Height: clay.SizingFixed(50)}}, BackgroundColor: ColorOrange})
}
