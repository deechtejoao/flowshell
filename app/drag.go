package app

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var drag DragState

type DragState struct {
	Dragging bool
	Pending  bool
	Canceled bool

	Thing any
	Key   string

	MouseStart V2
	ObjStart   V2
}

// Call once per frame at the start of the frame.
func (d *DragState) Update() {
	if rl.IsKeyPressed(rl.KeyEscape) {
		d.Dragging = false
		d.Canceled = true
	} else if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		d.Dragging = false
	} else if rl.IsMouseButtonUp(rl.MouseLeftButton) {
		d.Dragging = false
		d.Pending = false
		d.Canceled = true
		d.Thing = nil
		d.Key = ""
		d.MouseStart = rl.Vector2{}
		d.ObjStart = rl.Vector2{}
	} else if rl.IsMouseButtonDown(rl.MouseLeftButton) {
		if !d.Dragging && !d.Pending {
			d.Pending = true
			d.MouseStart = rl.GetMousePosition()
		}
	}
}

func (d *DragState) TryStartDrag(thing any, dragRegion rl.Rectangle, objStart rl.Vector2) bool {
	if thing == nil {
		panic("you must provide a thing to drag")
	}

	if d.Dragging {
		// can't start a new drag while one is in progress
		return false
	}

	if !d.Pending {
		// can't start a new drag with this item unless we have a pending one
		return false
	}

	if rl.Vector2Length(rl.Vector2Subtract(rl.GetMousePosition(), d.MouseStart)) < 3 {
		// haven't dragged far enough
		return false
	}

	if !rl.CheckCollisionPointRec(d.MouseStart, dragRegion) {
		// not dragging from the right place
		return false
	}

	d.Dragging = true
	d.Pending = false
	d.Canceled = false
	d.Thing = thing
	d.Key = GetDragKey(thing)
	d.ObjStart = objStart

	return true
}

func (d *DragState) Offset() rl.Vector2 {
	if !d.Dragging && d.Key == "" {
		return rl.Vector2{}
	}
	return rl.Vector2Subtract(rl.GetMousePosition(), d.MouseStart)
}

func (d *DragState) NewObjPosition() rl.Vector2 {
	return rl.Vector2Add(d.ObjStart, d.Offset())
}

// Pass in an key and it will tell you the relevant drag state for that thing.
// matchesKey will be true if that object is the one currently being dragged.
// done will be true if the drag is complete this frame.
// canceled will be true if the drag is done but was canceled.
func (d *DragState) State(key any) (matchesKey bool, done bool, canceled bool) {
	matchesKey = true
	if key != nil {
		matchesKey = d.Key == GetDragKey(key)
	}

	if !d.Dragging && d.Key != "" && matchesKey {
		return matchesKey, true, d.Canceled
	} else {
		return matchesKey, false, false
	}
}

func GetDragKey(key any) string {
	switch kt := key.(type) {
	case string:
		return kt
	case DragKeyer:
		return kt.DragKey()
	default:
		panic(fmt.Errorf("cannot make drag key for %v", key))
	}
}

type DragKeyer interface {
	DragKey() string
}
