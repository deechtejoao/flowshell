package app

import (
	"embed"
	"path/filepath"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/util"
	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed assets
var assets embed.FS

var ImgPlay rl.Texture2D
var ImgRetry rl.Texture2D
var ImgPushpin rl.Texture2D
var ImgPushpinOutline rl.Texture2D
var ImgDropdownDown rl.Texture2D
var ImgDropdownUp rl.Texture2D
var ImgToggleRight rl.Texture2D
var ImgToggleDown rl.Texture2D

// In a separate function because raylib must be initialized first.
func initImages() {
	ImgPlay = LoadAssetTexture("assets/play-white.png")
	core.ImgPlay = ImgPlay

	ImgRetry = LoadAssetTexture("assets/retry-white.png")
	core.ImgRetry = ImgRetry

	ImgPushpin = LoadAssetTexture("assets/pushpin-white.png")
	core.ImgPushpin = ImgPushpin

	ImgPushpinOutline = LoadAssetTexture("assets/pushpin-outline-white.png")
	core.ImgPushpinOutline = ImgPushpinOutline

	ImgDropdownDown = LoadAssetTexture("assets/dropdown-down.png")
	core.ImgDropdownDown = ImgDropdownDown

	ImgDropdownUp = LoadAssetTexture("assets/dropdown-up.png")
	core.ImgDropdownUp = ImgDropdownUp

	ImgToggleRight = LoadAssetTexture("assets/toggle-right.png")
	core.ImgToggleRight = ImgToggleRight

	ImgToggleDown = LoadAssetTexture("assets/toggle-down.png")
	core.ImgToggleDown = ImgToggleDown

	// Initialize Fonts
	core.InterRegular = uint16(InterRegular)
	core.InterSemibold = uint16(InterSemibold)
	core.InterBold = uint16(InterBold)
	core.JetBrainsMono = uint16(JetBrainsMono)
	core.CodeFont = uint16(JetBrainsMono)
	core.LogoFont = uint16(InterBold)
}

func LoadAssetTexture(path string) rl.Texture2D {
	data := util.Must1(assets.ReadFile(path))
	img := rl.LoadImageFromMemory(filepath.Ext(path), data, int32(len(data)))
	return rl.LoadTextureFromImage(img)
}

func LoadAssetFont(path string, fontSize int32) rl.Font {
	data := util.Must1(assets.ReadFile(path))
	return rl.LoadFontFromMemory(filepath.Ext(path), data, fontSize, nil)
}
