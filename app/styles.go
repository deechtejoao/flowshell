package app

import (
	"math"

	"github.com/bvisness/flowshell/clay"
)

var Night = clay.Color{R: 12, G: 14, B: 17, A: 255}
var Charcoal = clay.Color{R: 20, G: 22, B: 25, A: 255}
var DarkGray = clay.Color{R: 33, G: 36, B: 40, A: 255}
var Gray = clay.Color{R: 67, G: 71, B: 79, A: 255}
var LightGray = clay.Color{R: 157, G: 161, B: 170, A: 255}
var White = clay.Color{R: 250, G: 250, B: 252, A: 255}
var Red = clay.Color{R: 214, G: 25, B: 50, A: 255}
var Blue = clay.Color{R: 11, G: 88, B: 183, A: 255}
var Yellow = clay.Color{R: 242, G: 199, B: 68, A: 255}

var PlayButtonGreen = clay.Color{R: 61, G: 159, B: 72, A: 255}
var HoverWhite = clay.Color{R: 255, G: 255, B: 255, A: 20}

const S1 = 4
const S2 = 8
const S3 = 16

var W1 = clay.Sizing{Width: clay.SizingFixed(S1)}
var W2 = clay.Sizing{Width: clay.SizingFixed(S2)}
var W3 = clay.Sizing{Width: clay.SizingFixed(S3)}

func WH(w, h float32) clay.Sizing {
	return clay.Sizing{
		Width:  clay.SizingFixed(w),
		Height: clay.SizingFixed(h),
	}
}

func PX(n float32) clay.SizingAxis {
	return clay.SizingFixed(n)
}

var PA1 = clay.PaddingAll(S1)
var PA2 = clay.PaddingAll(S2)
var PA3 = clay.PaddingAll(S3)

func PVH(v, h uint16) clay.Padding {
	return clay.Padding{
		Left:   h,
		Right:  h,
		Top:    v,
		Bottom: v,
	}
}

func PD(t, r, b, l uint16, padding clay.Padding) clay.Padding {
	return clay.Padding{
		Left:   padding.Left + l,
		Right:  padding.Right + r,
		Top:    padding.Top + t,
		Bottom: padding.Bottom + b,
	}
}

var BL = clay.BW{Left: 1}
var BR = clay.BW{Right: 1}
var BT = clay.BW{Top: 1}
var BB = clay.BW{Bottom: 1}
var BA = clay.BW{Left: 1, Right: 1, Top: 1, Bottom: 1}
var BA2 = clay.BW{Left: 2, Right: 2, Top: 2, Bottom: 2}
var BTW = clay.BW{BetweenChildren: 1}
var BA_BTW = clay.BW{Left: 1, Right: 1, Top: 1, Bottom: 1, BetweenChildren: 1}

func BW(t, r, b, l uint16) clay.BW {
	return clay.BW{
		Left:   l,
		Right:  r,
		Top:    t,
		Bottom: b,
	}
}

const F1 = 12
const F2 = 16
const F3 = 18

const R1 = 2
const R2 = 4
const R3 = 6

func RA(radius float32) clay.CornerRadius {
	return clay.CornerRadius{
		TopLeft:     radius,
		TopRight:    radius,
		BottomRight: radius,
		BottomLeft:  radius,
	}
}

var RA1 = RA(R1)
var RA2 = RA(R2)
var RA3 = RA(R3)

var GROWH = clay.Sizing{Width: clay.SizingGrow(1, 0)}
var GROWV = clay.Sizing{Height: clay.SizingGrow(1, 0)}
var GROWALL = clay.Sizing{Width: clay.SizingGrow(1, 0), Height: clay.SizingGrow(1, 0)}

var XRIGHT = clay.ChildAlignment{X: clay.AlignXRight}
var XCENTER = clay.ChildAlignment{X: clay.AlignXCenter}
var YCENTER = clay.ChildAlignment{Y: clay.AlignYCenter}
var ALLCENTER = clay.ChildAlignment{X: clay.AlignXCenter, Y: clay.AlignYCenter}

const ZTOP = math.MaxInt16
