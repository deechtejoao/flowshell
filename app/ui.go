package app

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bvisness/flowshell/clay"
	"github.com/bvisness/flowshell/util"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

const MenuMinWidth = 200
const MenuMaxWidth = 0 // no max

const GridSize = 20

func SnapToGrid(v V2) V2 {
	return V2{
		X: float32(int(v.X/GridSize) * GridSize),
		Y: float32(int(v.Y/GridSize) * GridSize),
	}
}

const NodeMinWidth = 360

var currentGraph = NewGraph()
var History *HistoryManager

func InitHistory() {
	History = NewHistoryManager(currentGraph)
}

func PushHistory() {
	if History == nil {
		InitHistory()
	}
	History.Push(currentGraph)
}

type NodeType struct {
	Name         string
	Create       func() *Node
	ShortcutKey  int32
	ShortcutMods []int32
}

var nodeTypes = []NodeType{
	{Name: "Run Process", Create: func() *Node { return NewRunProcessNode(util.Tern(runtime.GOOS == "Windows", "dir", "ls")) }},
	{Name: "List Files", Create: func() *Node { return NewListFilesNode(".") }},
	{Name: "Lines", Create: func() *Node { return NewLinesNode() }},
	{Name: "Load File", Create: func() *Node { return NewLoadFileNode("") }},
	{Name: "Save File", Create: func() *Node { return NewSaveFileNode() }},
	{Name: "Trim Spaces", Create: func() *Node { return NewTrimSpacesNode() }},
	{Name: "Min", Create: func() *Node { return NewAggregateNode("Min") }},
	{Name: "Max", Create: func() *Node { return NewAggregateNode("Max") }},
	{Name: "Mean (Average)", Create: func() *Node { return NewAggregateNode("Mean") }},
	{Name: "Concatenate Tables (Combine Rows)", Create: func() *Node { return NewConcatTablesNode() }},
	{Name: "Filter Empty", Create: func() *Node { return NewFilterEmptyNode() }},
	{Name: "Sort", Create: func() *Node { return NewSortNode() }},
	{Name: "Select Columns", Create: func() *Node { return NewSelectColumnsNode() }},
	{Name: "Extract Column", Create: func() *Node { return NewExtractColumnNode() }},
	{Name: "Add Column", Create: func() *Node { return NewAddColumnNode() }},
	{Name: "Convert Type", Create: func() *Node { return NewConvertNode() }},
	{Name: "Transpose", Create: func() *Node { return NewTransposeNode() }},
	{Name: "Minify HTML", Create: func() *Node { return NewMinifyHTMLNode() }},
	{Name: "Wait For Click", Create: func() *Node { return NewWaitForClickNode() }},
	{Name: "Regex Match", Create: func() *Node { return NewRegexMatchNode() }},
	{Name: "Regex Find All", Create: func() *Node { return NewRegexFindAllNode() }},
	{Name: "Regex Replace", Create: func() *Node { return NewRegexReplaceNode() }},
	{Name: "Regex Split", Create: func() *Node { return NewRegexSplitNode() }},
	{Name: "HTTP Request", Create: func() *Node { return NewHTTPRequestNode() }},
	{Name: "Formula", Create: func() *Node { return NewFormulaNode() }},
	{Name: "Join Text", Create: func() *Node { return NewJoinTextNode() }},
	{Name: "Split Text", Create: func() *Node { return NewSplitTextNode() }},
	{Name: "Change Case", Create: func() *Node { return NewCaseConvertNode() }},
	{Name: "Format String", Create: func() *Node { return NewFormatStringNode() }},
	{Name: "Parse Time", Create: func() *Node { return NewParseTimeNode() }},
	{Name: "JSON Query", Create: func() *Node { return NewJsonQueryNode() }},
	{Name: "XML Query", Create: func() *Node { return NewXmlQueryNode() }},
}

func CreateGroup() {
	PushHistory()
	// Calculate bounding box of selected nodes
	if len(selectedNodes) == 0 {
		return
	}

	minX, minY := float32(1e9), float32(1e9)
	maxX, maxY := float32(-1e9), float32(-1e9)

	for id := range selectedNodes {
		if n, ok := currentGraph.GetNode(id); ok {
			minX = min(minX, n.Pos.X)
			minY = min(minY, n.Pos.Y)
			maxX = max(maxX, n.Pos.X+NodeMinWidth) // Approx width
			// Height is variable, but let's assume a safe buffer or use n.DragRect if available
			// n.DragRect is updated in layout, so it might be stale or 0 if not rendered yet?
			// But since we are selecting them, they must be rendered.
			h := n.DragRect.Height
			if h == 0 {
				h = 100
			}
			maxY = max(maxY, n.Pos.Y+h)
		}
	}

	// Add padding
	pad := float32(20)
	rect := rl.Rectangle{
		X:      minX - pad,
		Y:      minY - pad - 40, // Extra top padding for header
		Width:  (maxX - minX) + pad*2,
		Height: (maxY - minY) + pad*2 + 40,
	}

	currentGraph.AddGroup(&Group{
		Title: "New Group",
		Rect:  rect,
		Color: clay.Color{R: 100, G: 100, B: 100, A: 255}, // Default gray
	})
}

func parseNodeSearch(input string) (string, int) {
	input = strings.TrimSpace(input)
	parts := strings.SplitN(input, " ", 2)
	if len(parts) > 1 {
		if count, err := strconv.Atoi(parts[0]); err == nil && count > 0 {
			return parts[1], count
		}
	}
	return input, 1
}

func SearchNodeTypes(searchInput string) []NodeType {
	search, _ := parseNodeSearch(searchInput)
	var names []string
	for _, t := range nodeTypes {
		names = append(names, t.Name)
	}

	if search == "" || search == "?" || search == "*" {
		return nodeTypes
	}

	ranks := fuzzy.RankFindFold(search, names)
	slices.SortStableFunc(ranks, func(a, b fuzzy.Rank) int {
		return a.Distance - b.Distance
	})
	var res []NodeType
nextrank:
	for _, rank := range ranks {
		for _, t := range nodeTypes {
			if t.Name == rank.Target {
				res = append(res, t)
				continue nextrank
			}
		}
	}
	return res
}

var selectedNodeID = 0
var selectedNodes = make(map[int]struct{})

func IsNodeSelected(id int) bool {
	_, ok := selectedNodes[id]
	return ok
}

func SelectNode(id int, multi bool) {
	if !multi {
		for k := range selectedNodes {
			delete(selectedNodes, k)
		}
	}
	selectedNodes[id] = struct{}{}
	selectedNodeID = id
}

func ToggleSelectNode(id int) {
	if _, ok := selectedNodes[id]; ok {
		delete(selectedNodes, id)
		if selectedNodeID == id {
			selectedNodeID = 0
			// Pick another one?
			for k := range selectedNodes {
				selectedNodeID = k
				break
			}
		}
	} else {
		selectedNodes[id] = struct{}{}
		selectedNodeID = id
	}
}

func GetSelectedNode() (*Node, bool) {
	return currentGraph.GetNode(selectedNodeID)
}

func DeleteNode(id int) {
	currentGraph.DeleteNode(id)
}

func DeleteSelectedNodes() {
	PushHistory()
	for id := range selectedNodes {
		DeleteNode(id)
	}
	for k := range selectedNodes {
		delete(selectedNodes, k)
	}
	selectedNodeID = 0
}

func DuplicateNode(original *Node) {
	PushHistory()
	// Clone via Serialization
	s := NewEncoder(1) // version 1
	original.Serialize(s)
	data := s.Bytes()

	sRead := NewDecoder(data)

	// Create new "empty" node structure
	clone := &Node{}
	// We need to initialize Action via helper or manually meta-alloc?
	// Serialize() on decode expects n.Action to be allocated?
	// No, the default Serialize implementation allocates it:
	// n.Action = meta.Alloc()
	// So we just need an empty Node.

	if clone.Serialize(sRead) {
		// Post-processing
		clone.ID = 0 // Will be assigned by AddNode
		clone.Graph = nil
		clone.Pos = V2(clay.V2(clone.Pos).Plus(clay.V2{X: 20, Y: 20})) // Offset
		clone.Pos = SnapToGrid(clone.Pos)
		clone.outputState = nil // Reset state
		clone.Running = false
		clone.Valid = false

		currentGraph.AddNode(clone)
		SelectNode(clone.ID, false)
	} else {
		fmt.Println("Failed to duplicate node: serialization error")
	}
}

var UICursor rl.MouseCursor
var UIFocus *clay.ElementID

var CurrentZIndex int16

func WithZIndex(z int16, f func()) {
	prev := CurrentZIndex
	CurrentZIndex = z
	defer func() { CurrentZIndex = prev }()
	f()
}

func IsFocused(id clay.ElementID) bool {
	return UIFocus != nil && id.ID == UIFocus.ID
}

func processInput() {
	if rl.IsMouseButtonPressed(rl.MouseRightButton) && !IsHoveringPanel && !IsHoveringUI {
		// Check if we clicked on a node for Context Menu
		clickedNode := false
		for _, n := range currentGraph.Nodes {
			// Hit test using DragRect (header) or full body if we can guess it?
			// DragRect is safest for "Select/Interact".
			// But user might right click the body.
			// Let's rely on clay.GetElementData for the node's main ID "Node<ID>"
			if data, ok := clay.GetElementData(n.ClayID()); ok {
				if rl.CheckCollisionPointRec(rl.GetMousePosition(), rl.Rectangle(data.BoundingBox)) {
					clickedNode = true

					// Open Context Menu
					items := []ContextMenuItem{
						{Label: "Run", Action: func() { n.Run(context.Background(), false) }},
						{Label: util.Tern(n.Pinned, "Unpin", "Pin"), Action: func() {
							PushHistory()
							n.Pinned = !n.Pinned
						}},
						{Label: "Duplicate", Action: func() { DuplicateNode(n) }}, // DuplicateNode calls PushHistory
						{Label: "Delete", Action: func() {
							// DeleteSelectedNodes calls PushHistory, but here we might delete a single node
							// that isn't selected? Or we should select it first?
							// Logic below deletes node n.ID.
							PushHistory()
							DeleteNode(n.ID)
							if IsNodeSelected(n.ID) {
								delete(selectedNodes, n.ID)
								selectedNodeID = 0
							}
						}},
					}

					ContextMenu = &ContextMenuState{
						Pos:    V2(rl.GetMousePosition()),
						NodeID: n.ID,
						Items:  items,
					}
					break
				}
			}
		}

		if !clickedNode {
			ContextMenu = nil // Close if clicking background

			isGroup := false
			for _, grp := range currentGraph.Groups {
				if rl.CheckCollisionPointRec(rl.GetMousePosition(), grp.Rect) {
					isGroup = true
					break
				}
			}

			if !isGroup {
				NewNodeName = ""
				id := clay.ID("NewNodeName")
				UIFocus = &id
			}
		}
	} else if rl.IsMouseButtonPressed(rl.MouseLeftButton) && !IsHoveringPanel && !IsHoveringUI {
		// Close context menu on left click outside
		ContextMenu = nil
	}

	// Double check to clear focus if clicking background
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && !IsHoveringPanel && !IsHoveringUI {
		UIFocus = nil
	}

	// Undo/Redo
	actions := rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
	if actions && rl.IsKeyPressed(rl.KeyZ) {
		if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
			if g := History.Redo(); g != nil {
				currentGraph = g
				// Clear selection on undo/redo to avoid ghost selections
				// Or try to restore? Restoring is hard.
				clear(selectedNodes)
				selectedNodeID = 0
			}
		} else {
			if g := History.Undo(); g != nil {
				currentGraph = g
				clear(selectedNodes)
				selectedNodeID = 0
			}
		}
	}
	if actions && rl.IsKeyPressed(rl.KeyY) {
		if g := History.Redo(); g != nil {
			currentGraph = g
			clear(selectedNodes)
			selectedNodeID = 0
		}
	}

	if rl.IsKeyPressed(rl.KeyS) && rl.IsKeyDown(rl.KeyLeftControl) {
		_ = SaveGraph("saved.flow", currentGraph)
	}
	if rl.IsKeyPressed(rl.KeyL) && rl.IsKeyDown(rl.KeyLeftControl) {
		if g, err := LoadGraph("saved.flow"); err == nil {
			currentGraph = g
		}
	}

	if rl.IsKeyPressed(rl.KeyC) && (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)) {
		Copy()
	}
	if rl.IsKeyPressed(rl.KeyV) && (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)) {
		PushHistory()
		Paste()
	}

	if rl.IsKeyPressed(rl.KeyG) && (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)) {
		CreateGroup()
	}

	// Create node shortcuts
	if !IsHoveringUI && UIFocus == nil {
		for _, nt := range nodeTypes {
			if nt.ShortcutKey != 0 {
				modsDown := true
				for _, mod := range nt.ShortcutMods {
					if !rl.IsKeyDown(mod) {
						modsDown = false
						break
					}
				}
				if modsDown && rl.IsKeyPressed(nt.ShortcutKey) {
					PushHistory()
					newNode := nt.Create()
					// Create at mouse position
					newNode.Pos = SnapToGrid(V2(rl.GetMousePosition()))
					currentGraph.AddNode(newNode)
					selectedNodeID = newNode.ID
				}
			}
		}
	}

	if rl.IsKeyPressed(rl.KeyDelete) && UIFocus == nil && len(selectedNodes) > 0 {
		DeleteSelectedNodes()
	}

	if rl.IsFileDropped() {
		files := rl.LoadDroppedFiles()
		handled := false

		// Check if we dropped onto a Load File node
		if len(files) > 0 && IsHoveringPanel {
			mousePos := rl.GetMousePosition()
			// Find the top-most node under the mouse
			for i := len(currentGraph.Nodes) - 1; i >= 0; i-- {
				n := currentGraph.Nodes[i]
				if data, ok := clay.GetElementData(n.ClayID()); ok {
					if rl.CheckCollisionPointRec(mousePos, rl.Rectangle(data.BoundingBox)) {
						if loadAction, ok := n.Action.(*LoadFileAction); ok {
							loadAction.path = files[0]
							// Auto-select format based on extension
							if ext := strings.ToLower(filepath.Ext(files[0])); ext != "" {
								loadAction.format.SelectByValue(ext[1:])
							}
							handled = true
							// Force re-validation
							n.ClearResult()
						}
						break // Only handle for the top-most node (or first found)
					}
				}
			}
		}

		if !handled && len(files) > 0 {
			// Check for .flow file import
			ext := strings.ToLower(filepath.Ext(files[0]))
			if ext == ".flow" {
				PushHistory()
				if g, err := LoadGraph(files[0]); err == nil {
					MergeGraph(currentGraph, g)
					handled = true
				} else {
					fmt.Printf("Failed to load graph: %v\n", err)
				}
			}
		}

		if !handled && !IsHoveringUI && !IsHoveringPanel {
			PushHistory()
			for i, filename := range files {
				n := NewLoadFileNode(filename)
				n.Pos = V2(clay.V2(rl.GetMousePosition()).Plus(clay.V2{X: 20, Y: 20}.Times(float32(i))))
				if !rl.IsKeyDown(rl.KeyLeftShift) && !rl.IsKeyDown(rl.KeyRightShift) {
					n.Pos = SnapToGrid(n.Pos)
				}
				currentGraph.AddNode(n)
				SelectNode(n.ID, i > 0)
			}
		}
	}

	// Selection on Mouse Down (Immediate feedback)
	for _, n := range currentGraph.Nodes {
		if UIInput.IsPressed(n.ClayID()) {
			multi := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) || rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
			if !IsNodeSelected(n.ID) {
				SelectNode(n.ID, multi)
			}
		}
	}

	for _, n := range currentGraph.Nodes {
		// Starting new wires (Prioritize over node drag)
		for i, portPos := range n.OutputPortPositions {
			portRect := rl.Rectangle{
				X:      portPos.X - PortDragRadius,
				Y:      portPos.Y - PortDragRadius,
				Width:  PortDragRadius * 2,
				Height: PortDragRadius * 2,
			}
			if !IsHoveringUI && drag.TryStartDrag(NewWireDragKey, portRect, V2{}) {
				NewWireSourceNode = n
				NewWireSourcePort = i
			}
		}
		for i, portPos := range n.InputPortPositions {
			portRect := rl.Rectangle{
				X:      portPos.X - PortDragRadius,
				Y:      portPos.Y - PortDragRadius,
				Width:  PortDragRadius * 2,
				Height: PortDragRadius * 2,
			}
			if wire, hasWire := n.GetInputWire(i); hasWire {
				if !IsHoveringUI && drag.TryStartDrag(NewWireDragKey, portRect, V2{}) {
					currentGraph.Wires = slices.DeleteFunc(currentGraph.Wires, func(w *Wire) bool { return w == wire })
					NewWireSourceNode = wire.StartNode
					NewWireSourcePort = wire.StartPort
				}
			}
		}

		// Node drag and drop
		if !IsHoveringUI {
			if drag.TryStartDrag(n, n.DragRect, n.Pos) {
				PushHistory()
			}
		}
		if draggingThisNode, done, canceled := drag.State(n); draggingThisNode {
			// If we start dragging a node that isn't selected, select it (and deselect others)
			// UNLESS we are dragging a selected node, then we move all selected nodes.
			if !IsNodeSelected(n.ID) {
				SelectNode(n.ID, false)
			}

			// Calculate delta
			delta := rl.Vector2Subtract(drag.NewObjPosition(), n.Pos)

			// Apply delta to all selected nodes
			for id := range selectedNodes {
				if node, ok := currentGraph.GetNode(id); ok {
					node.Pos = rl.Vector2Add(node.Pos, delta)
					// Snap to grid (hold Shift to disable)
					if !rl.IsKeyDown(rl.KeyLeftShift) && !rl.IsKeyDown(rl.KeyRightShift) {
						node.Pos = SnapToGrid(node.Pos)
					}
				}

			}

			if done {
				if canceled {
					n.Pos = drag.ObjStart
				}
			}
		}

		// Selected node keyboard shortcuts
		if IsNodeSelected(n.ID) {
			if rl.IsKeyPressed(rl.KeyR) && (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)) {
				n.Run(context.Background(), rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift))
			}
		}
	}

	// Dropping new wires
	if draggingNewWire, done, canceled := drag.State(NewWireDragKey); draggingNewWire {
		if done && !canceled {
			// Loop over nodes to find any you may have dropped on
			for _, node := range currentGraph.Nodes {
				for port, portPos := range node.InputPortPositions {
					portRect := rl.Rectangle{
						X:      portPos.X - PortDragRadius,
						Y:      portPos.Y - PortDragRadius,
						Width:  PortDragRadius * 2,
						Height: PortDragRadius * 2,
					}

					if rl.CheckCollisionPointRec(rl.GetMousePosition(), portRect) && node != NewWireSourceNode {
						// Check types
						sourceType := NewWireSourceNode.OutputPorts[NewWireSourcePort].Type
						targetType := node.InputPorts[port].Type
						if err := Typecheck(sourceType, targetType); err != nil {
							fmt.Printf("Cannot connect: %v\n", err)
							ConnectionError = fmt.Sprintf("Cannot connect: %s", err.Error())
							ConnectionErrorTime = time.Now()
						} else {
							// Delete existing wires into that port
							if len(currentGraph.Wires) > 0 { // Optimization or check required?
								// Just push history on wire modify
								PushHistory()
							}
							currentGraph.Wires = slices.DeleteFunc(currentGraph.Wires, func(wire *Wire) bool {
								return wire.EndNode == node && wire.EndPort == port
							})
							currentGraph.Wires = append(currentGraph.Wires, &Wire{
								StartNode: NewWireSourceNode, EndNode: node,
								StartPort: NewWireSourcePort, EndPort: port,
							})
						}
						node.ClearResult()
					}
				}
			}
		}
	}

	// Dropping FlowValue
	if draggingValue, done, canceled := drag.State("FLOW_VALUE_DRAG"); draggingValue {
		if done && !canceled {
			if dragValue, ok := drag.Thing.(FlowValueDrag); ok {
				n := NewValueNode(dragValue.Value)
				n.Pos = V2(rl.GetMousePosition())
				if !rl.IsKeyDown(rl.KeyLeftShift) && !rl.IsKeyDown(rl.KeyRightShift) {
					n.Pos = SnapToGrid(n.Pos)
				}
				PushHistory()
				currentGraph.AddNode(n)
				selectedNodeID = n.ID
			}
		}
	}

	// Resizing output window
	{
		outputWindowRect := rl.Rectangle{
			X:      float32(rl.GetScreenWidth()) - OutputWindowWidth - OutputWindowDragWidth/2,
			Y:      0,
			Width:  OutputWindowDragWidth,
			Height: float32(rl.GetScreenHeight()),
		}
		drag.TryStartDrag(OutputWindowDragKey, outputWindowRect, V2{X: OutputWindowWidth, Y: 0})

		resizing, done, canceled := drag.State(OutputWindowDragKey)
		if resizing {
			if done {
				if canceled {
					OutputWindowWidth = drag.ObjStart.X
				}
			} else {
				OutputWindowWidth = float32(rl.GetScreenWidth()) - float32(rl.GetMouseX())
			}
		}

		if rl.CheckCollisionPointRec(rl.GetMousePosition(), outputWindowRect) || resizing {
			UICursor = rl.MouseCursorResizeEW
		}
	}

	// Box Selection & Panning
	{
		if background, ok := clay.GetElementData(clay.ID("Background")); ok {
			isShift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)

			if !IsHoveringUI && !IsHoveringPanel {
				if isShift {
					// Snapshot selection
					initial := make(map[int]struct{})
					for k, v := range selectedNodes {
						initial[k] = v
					}
					drag.TryStartDrag(BoxSelectDrag{InitialSelection: initial}, rl.Rectangle(background.BoundingBox), V2{})
				} else {
					if drag.TryStartDrag(PanDragKey, rl.Rectangle(background.BoundingBox), V2{}) {
						LastPanMousePosition = rl.GetMousePosition()
					}
				}
			}

			if boxSelecting, _, canceled := drag.State(BoxSelectDragKey); boxSelecting {
				if bs, ok := drag.Thing.(BoxSelectDrag); ok {
					if canceled {
						// Restore
						clear(selectedNodes)
						for k, v := range bs.InitialSelection {
							selectedNodes[k] = v
						}
					} else {
						// Update selection
						start := drag.MouseStart
						end := rl.GetMousePosition()

						x := min(start.X, end.X)
						y := min(start.Y, end.Y)
						w := max(start.X, end.X) - x
						h := max(start.Y, end.Y) - y
						box := rl.Rectangle{X: x, Y: y, Width: w, Height: h}

						clear(selectedNodes)
						for k, v := range bs.InitialSelection {
							selectedNodes[k] = v
						}

						for _, n := range currentGraph.Nodes {
							if rl.CheckCollisionRecs(box, n.DragRect) {
								selectedNodes[n.ID] = struct{}{}
								selectedNodeID = n.ID
							}
						}
					}
				}
			}

			if panning, _, _ := drag.State(PanDragKey); panning {
				mousePos := rl.GetMousePosition()
				delta := rl.Vector2Subtract(mousePos, LastPanMousePosition)
				for _, n := range currentGraph.Nodes {
					n.Pos = rl.Vector2Add(n.Pos, delta)
				}
				LastPanMousePosition = mousePos
			}
		}
	}
}

const OutputWindowDragKey = "OUTPUT_WINDOW"
const OutputWindowDragWidth = 8

const PanDragKey = "PAN"
const BoxSelectDragKey = "BOX_SELECT"

type BoxSelectDrag struct {
	InitialSelection map[int]struct{}
}

func (b BoxSelectDrag) DragKey() string {
	return BoxSelectDragKey
}

type FlowValueDrag struct {
	Value FlowValue
}

func (f FlowValueDrag) DragKey() string {
	return "FLOW_VALUE_DRAG"
}

var LastPanMousePosition V2

const NewWireDragKey = "NEW_WIRE"
const PortDragRadius = 5

var NewWireSourceNode *Node
var NewWireSourcePort int

var NewNodeName string

var ShowLoadConfirmation bool
var ConnectionError string
var ConnectionErrorTime time.Time

var ContextMenu *ContextMenuState

type ContextMenuItem struct {
	Label  string
	Action func()
}

type ContextMenuState struct {
	Pos    V2
	NodeID int
	Items  []ContextMenuItem
}

var OutputWindowWidth float32 = windowWidth * 0.30

func UpdateGraph() error {
	// Sweep the graph, validating all nodes
	sortedNodes, topoErr := Toposort(currentGraph.Nodes, currentGraph.Wires)
	if topoErr != nil {
		// If there is a cycle, we can't toposort. Just use the default order.
		sortedNodes = currentGraph.Nodes
	}
	for _, node := range sortedNodes {
		node.Action.UpdateAndValidate(node)
	}
	return topoErr
}

func UINodes(topoErr error) {
	clay.CLAY(clay.ID("NodeCanvas"), clay.EL{
		Layout: clay.LAY{Sizing: GROWALL},
	}, func() {
		for _, group := range currentGraph.Groups {
			UIGroup(group)
		}
		for _, node := range currentGraph.Nodes {
			UINode(node, topoErr != nil)
		}
	})
}

func UIOverlay(topoErr error) {
	clay.CLAY(clay.ID("OverlayRoot"), clay.EL{
		Layout: clay.LAY{Sizing: GROWALL},
		Floating: clay.FLOAT{
			AttachTo:           clay.AttachToRoot,
			PointerCaptureMode: clay.PointercaptureModePassthrough, // Important: let clicks pass through to nodes if not hitting UI
		},
	}, func() {
		UIMinimap()

		if topoErr != nil {
			WithZIndex(ZTOP, func() {
				clay.CLAY(clay.ID("CycleWarning"), clay.EL{
					Layout:          clay.LAY{Sizing: clay.Sizing{Width: clay.SizingGrow(1, 0), Height: clay.SizingFixed(30)}, ChildAlignment: ALLCENTER},
					BackgroundColor: Red,
					Floating: clay.FLOAT{
						AttachTo: clay.AttachToParent,
						AttachPoints: clay.FloatingAttachPoints{
							Element: clay.AttachPointCenterTop,
							Parent:  clay.AttachPointCenterTop,
						},
						ZIndex: ZTOP,
					},
				}, func() {
					clay.TEXT("Cycle Detected! Graph execution disabled.", clay.TextElementConfig{TextColor: White, FontID: InterBold})
				})
			})
		}

		if ShowLoadConfirmation {
			WithZIndex(ZTOP, func() {
				clay.CLAY(clay.ID("LoadConfirmation"), clay.EL{
					Layout:          clay.LAY{LayoutDirection: clay.TopToBottom, Sizing: clay.Sizing{Width: clay.SizingFixed(300)}, ChildAlignment: ALLCENTER, Padding: PA3, ChildGap: S2},
					BackgroundColor: Charcoal,
					Border:          clay.BorderElementConfig{Width: BA2, Color: Red},
					CornerRadius:    RA2,
					Floating: clay.FLOAT{
						AttachTo: clay.AttachToParent,
						AttachPoints: clay.FloatingAttachPoints{
							Element: clay.AttachPointCenterCenter,
							Parent:  clay.AttachPointCenterCenter,
						},
						ZIndex:             ZTOP,
						PointerCaptureMode: clay.PointercaptureModePassthrough,
					},
				}, func() {
					clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
						IsHoveringUI = true
					}, nil)
					clay.TEXT("Discard changes and load?", clay.TextElementConfig{TextColor: White, FontID: InterBold, FontSize: F2})
					clay.CLAY(clay.AUTO_ID, clay.EL{Layout: clay.LAY{ChildGap: S3}}, func() {
						UIButton(clay.ID("ConfirmLoad"), UIButtonConfig{
							El: clay.EL{Layout: clay.LAY{Padding: PA2}, BackgroundColor: Red, CornerRadius: RA1},
							OnClick: func(_ clay.ElementID, _ clay.PointerData, _ any) {
								ShowLoadConfirmation = false
								_, _ = LoadGraph("saved.flow")
							},
						}, func() {
							clay.TEXT("Load", clay.TextElementConfig{TextColor: White})
						})
						UIButton(clay.ID("CancelLoad"), UIButtonConfig{
							El: clay.EL{Layout: clay.LAY{Padding: PA2}, BackgroundColor: Gray, CornerRadius: RA1},
							OnClick: func(_ clay.ElementID, _ clay.PointerData, _ any) {
								ShowLoadConfirmation = false
							},
						}, func() {
							clay.TEXT("Cancel", clay.TextElementConfig{TextColor: White})
						})
					})
				})
			})
		}

		// Check expansion state for New Node Menu
		textboxID := clay.ID("NewNodeName")
		isFocused := IsFocused(textboxID)

		shortcut := rl.IsKeyPressed(rl.KeySpace) && (rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl))

		if rl.IsKeyPressed(rl.KeyEscape) {
			UIFocus = nil
			isFocused = false
		}

		containerWidth := clay.SizingFit(0, 0)
		if isFocused || shortcut {
			containerWidth = clay.SizingFixed(300)
		}

		WithZIndex(10, func() {
			clay.CLAY(clay.ID("NewNodeContainer"), clay.EL{
				Layout: clay.LAY{
					Sizing:  clay.Sizing{Width: containerWidth, Height: clay.SizingFit(0, 0)},
					Padding: PA3,
				},
				Floating: clay.FLOAT{
					AttachTo: clay.AttachToParent,
					AttachPoints: clay.FloatingAttachPoints{
						Element: clay.AttachPointLeftBottom,
						Parent:  clay.AttachPointLeftBottom,
					},
					ZIndex:             10,
					PointerCaptureMode: clay.PointercaptureModePassthrough,
				},
			}, func() {
				clay.CLAY_AUTO_ID(clay.EL{
					Layout: clay.LAY{
						ChildGap: S2,
						Sizing:   clay.Sizing{Width: clay.SizingGrow(0, 300)}, // Capped grow? Or just Grow.
					},
				}, func() {
					clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
						IsHoveringUI = true
					}, nil)

					// textboxID already defined above
					// shortcut already defined above

					// Before UI: defocus textbox
					if shortcut && isFocused { // Was IsFocused(textboxID)
						UIFocus = nil
					}

					UIButton(clay.ID("NewNode"), UIButtonConfig{
						El: clay.EL{
							Layout: clay.LAY{
								Sizing:         WH(36, 36),
								ChildAlignment: ALLCENTER,
							},
							Border: clay.B{Width: BA, Color: Gray},
						},
						OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
							NewNodeName = ""
							id := textboxID
							UIFocus = &id
						},
					}, func() {
						clay.TEXT("+", clay.T{FontID: InterBold, FontSize: 36, TextColor: White})
					})
					if IsFocused(textboxID) {
						addNodeFromMatch := func(nt NodeType) {
							_, count := parseNodeSearch(NewNodeName)

							prevPos := V2{}

							for i := 0; i < count; i++ {
								newNode := nt.Create()

								var pos V2
								if i == 0 {
									centerWorld := Camera.ScreenToWorld(rl.Vector2{X: float32(rl.GetScreenWidth()) / 2, Y: float32(rl.GetScreenHeight()) / 2})
									pos = SnapToGrid(centerWorld)
									// Collision avoidance with bounding box check
									testRect := rl.Rectangle{X: pos.X, Y: pos.Y, Width: NodeMinWidth, Height: 150} // 150 is estimated height
									for j := 0; j < 50; j++ {
										overlap := false
										for _, n := range currentGraph.Nodes {
											// Use the actual drag rect if available (it might be 0 if not rendered yet, but we have n.Pos)
											// Or fallback to a default size check
											otherRect := n.DragRect
											if otherRect.Width == 0 {
												otherRect = rl.Rectangle{X: n.Pos.X, Y: n.Pos.Y, Width: NodeMinWidth, Height: 150}
											}
											if rl.CheckCollisionRecs(testRect, otherRect) {
												overlap = true
												break
											}
										}
										if !overlap {
											break
										}
										pos.X += GridSize
										pos.Y += GridSize
										testRect.X = pos.X
										testRect.Y = pos.Y
									}
								} else {
									// Place next to previous
									// We don't know the exact width until first frame, but we can guess or use a standard width
									// NodeMinWidth is 360.
									pos = prevPos
									pos.X += NodeMinWidth + GridSize*2 // Add some padding
								}

								newNode.Pos = pos
								prevPos = pos

								// Special handling for batch SaveFile nodes
								if count > 1 {
									if saveAction, ok := newNode.Action.(*SaveFileAction); ok {
										// Split extension
										ext := filepath.Ext(saveAction.Path)
										base := strings.TrimSuffix(saveAction.Path, ext)
										saveAction.Path = fmt.Sprintf("%s_%d%s", base, i+1, ext)
									}
								}

								currentGraph.AddNode(newNode)
								selectedNodeID = newNode.ID
							}
						}

						UITextBox(textboxID, &NewNodeName, UITextBoxConfig{
							El: clay.ElementDeclaration{
								Layout: clay.LAY{
									Sizing: GROWALL,
								},
							},
							OnSubmit: func(val string) {
								matches := SearchNodeTypes(val)
								if len(matches) > 0 {
									addNodeFromMatch(matches[0])
								}
								UIFocus = nil
							},
						}, func() {
							WithZIndex(20, func() {
								clay.CLAY(clay.ID("NewNodeMatches"), clay.EL{
									Layout: clay.LAY{
										LayoutDirection: clay.TopToBottom,
										Sizing:          GROWH,
									},
									BackgroundColor: DarkGray,
									Border:          clay.B{Width: BA_BTW, Color: Gray},
									Floating: clay.FLOAT{
										AttachTo: clay.AttachToParent,
										AttachPoints: clay.FloatingAttachPoints{
											Parent:  clay.AttachPointLeftTop,
											Element: clay.AttachPointLeftBottom,
										},
										ZIndex:             20,
										PointerCaptureMode: clay.PointercaptureModePassthrough,
									},
								}, func() {
									clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
										IsHoveringUI = true
									}, nil)

									matches := SearchNodeTypes(NewNodeName)
									for i := len(matches) - 1; i >= 0; i-- {
										UIButton(clay.IDI("MatchButton", i), UIButtonConfig{
											El: clay.EL{
												Layout: clay.LAY{
													Padding: PVH(S1, S2),
													Sizing:  GROWH,
												},
												BackgroundColor: util.Tern(clay.Hovered(), HoverWhite, clay.Color{}),
											},
											OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
												addNodeFromMatch(matches[userData.(int)])
												UIFocus = nil
											},
											OnClickUserData: i,
										}, func() {
											clay.TEXT(matches[i].Name, clay.T{TextColor: White})
											_, count := parseNodeSearch(NewNodeName)
											if count > 1 {
												clay.TEXT(fmt.Sprintf(" (x%d)", count), clay.T{TextColor: Yellow, FontSize: 14})
											}
											if matches[i].ShortcutKey != 0 {
												// Just render the text, no spacer for now to debug
												keyName := ""
												if matches[i].ShortcutKey >= 32 && matches[i].ShortcutKey <= 126 {
													keyName = string(rune(matches[i].ShortcutKey))
												} else {
													keyName = fmt.Sprintf("K%d", matches[i].ShortcutKey)
												}

												modStr := ""
												for _, mod := range matches[i].ShortcutMods {
													switch mod {
													case rl.KeyLeftControl, rl.KeyRightControl:
														modStr += "Ctrl+"
													case rl.KeyLeftAlt, rl.KeyRightAlt:
														modStr += "Alt+"
													case rl.KeyLeftShift, rl.KeyRightShift:
														modStr += "Shift+"
													}
												}
												clay.TEXT("  "+modStr+keyName, clay.T{TextColor: Gray, FontSize: 10})
											}
										})
									}
								})
							})
						})
					}

					// After UI: focus textbox
					if shortcut && !IsFocused(textboxID) {
						UIFocus = &textboxID
						NewNodeName = ""
					}
				})
			})
		})

		clay.CLAY_LATE(clay.ID("Output"), func() clay.EL {
			return clay.EL{
				Layout: clay.LAY{
					LayoutDirection: clay.TopToBottom,
					Sizing:          clay.Sizing{Width: clay.SizingFixed(OutputWindowWidth), Height: clay.SizingGrow(1, 0)},
					Padding:         PA2,
				},
				Clip: clay.ClipElementConfig{
					Vertical:    true,
					Horizontal:  true,
					ChildOffset: clay.GetScrollOffset(),
				},
			}
		}, func() {
			clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
				IsHoveringUI = true
			}, nil)

			if selectedNode, ok := GetSelectedNode(); ok {
				if selectedNode.ResultAvailable {
					result := selectedNode.Result
					if result.Err == nil {
						for outputIndex, output := range result.Outputs {
							port := selectedNode.OutputPorts[outputIndex]
							if err := Typecheck(*output.Type, port.Type); err != nil {
								// panic(err)
								clay.TEXT(fmt.Sprintf("Type Error: %v", err), clay.TextElementConfig{TextColor: Red})
								continue
							}

							outputState := selectedNode.GetOutputState(port.Name)

							clay.CLAY_AUTO_ID(clay.EL{
								Layout: clay.LAY{ChildGap: S1, ChildAlignment: YCENTER},
							}, func() {
								UIButton(clay.AUTO_ID, UIButtonConfig{
									OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
										outputState.Collapsed = !outputState.Collapsed
									},
								}, func() {
									UIImage(clay.AUTO_ID, util.Tern(outputState.Collapsed, ImgToggleRight, ImgToggleDown), clay.EL{})
								})
								clay.TEXT(port.Name, clay.TextElementConfig{FontID: InterSemibold, TextColor: White})
							})
							if !outputState.Collapsed {
								clay.CLAY_AUTO_ID(clay.EL{
									Layout: clay.LAY{ChildGap: S1},
								}, func() {
									clay.CLAY_AUTO_ID(clay.EL{
										Layout: clay.LAY{
											Sizing:         clay.Sizing{Width: PX(float32(ImgToggleDown.Width)), Height: GROWV.Height},
											ChildAlignment: XCENTER,
										},
									}, func() {
										clay.CLAY_AUTO_ID(clay.EL{
											Layout: clay.LAY{
												Sizing: clay.Sizing{Width: PX(1), Height: GROWV.Height},
											},
											Border: clay.B{Color: Gray, Width: BR},
										})
									})
									UIFlowValue(output)
								})
							}
						}
					} else {
						clay.TEXT(result.Err.Error(), clay.TextElementConfig{TextColor: Red})
					}
				}
			}
		})

		// Box Selection & Panning
		if overlay, ok := clay.GetElementData(clay.ID("OverlayRoot")); ok {
			isShift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)

			if !IsHoveringUI && !IsHoveringPanel {
				if isShift {
					// Snapshot selection
					initial := make(map[int]struct{})
					for k, v := range selectedNodes {
						initial[k] = v
					}
					drag.TryStartDrag(BoxSelectDrag{InitialSelection: initial}, rl.Rectangle(overlay.BoundingBox), V2{})
				} else {
					if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
						clear(selectedNodes)
						selectedNodeID = 0
						UIFocus = nil
					}
					if drag.TryStartDrag(PanDragKey, rl.Rectangle(overlay.BoundingBox), V2{}) {
						LastPanMousePosition = rl.GetMousePosition()
					}
				}
			}

			if boxSelecting, _, canceled := drag.State(BoxSelectDragKey); boxSelecting {
				if bs, ok := drag.Thing.(BoxSelectDrag); ok {
					if canceled {
						// Restore
						clear(selectedNodes)
						for k, v := range bs.InitialSelection {
							selectedNodes[k] = v
						}
					} else {
						// Update selection
						start := drag.MouseStart
						end := rl.GetMousePosition()

						startWorld := Camera.ScreenToWorld(start)
						endWorld := Camera.ScreenToWorld(end)

						x := min(startWorld.X, endWorld.X)
						y := min(startWorld.Y, endWorld.Y)
						w := max(startWorld.X, endWorld.X) - x
						h := max(startWorld.Y, endWorld.Y) - y
						box := rl.Rectangle{X: x, Y: y, Width: w, Height: h}

						clear(selectedNodes)
						for k, v := range bs.InitialSelection {
							selectedNodes[k] = v
						}

						for _, n := range currentGraph.Nodes {
							if rl.CheckCollisionRecs(box, n.DragRect) {
								selectedNodes[n.ID] = struct{}{}
								selectedNodeID = n.ID
							}
						}
					}
				}
			}

			if panning, _, _ := drag.State(PanDragKey); panning {
				mousePos := rl.GetMousePosition()
				delta := rl.Vector2Subtract(mousePos, LastPanMousePosition)
				Camera.Pan(delta)
				LastPanMousePosition = mousePos
			}

			// Render Connection Error
			if time.Since(ConnectionErrorTime) < 2*time.Second {
				clay.CLAY_AUTO_ID(clay.EL{
					Floating: clay.FloatingElementConfig{
						AttachTo: clay.AttachToRoot,
						Offset:   clay.V2(rl.GetMousePosition()).Plus(clay.V2{X: 16, Y: 16}),
						ZIndex:   ZTOP,
					},
					Layout:          clay.LAY{Padding: PA1},
					BackgroundColor: Red,
					Border:          clay.BorderElementConfig{Color: White, Width: BA},
					CornerRadius:    RA1,
				}, func() {
					clay.TEXT(ConnectionError, clay.TextElementConfig{TextColor: White, FontID: InterBold})
				})
			}

			// Render Context Menu
			if ContextMenu != nil {
				// Close if clicking outside (handled in processInput generally, but let's be safe?)
				// Actually clay.OnHover only works if we cover screen?
				// processInput handles the "click outside" via global mouse check.
				UIContextMenu()
			}
		}
	})

	rl.SetMouseCursor(UICursor)
	UICursor = rl.MouseCursorDefault
}

func afterLayout() {
	for _, node := range currentGraph.Nodes {
		node.UpdateLayoutInfo()
	}
}

func renderWorldOverlays() {
	// Render wires
	for _, wire := range currentGraph.Wires {
		color := util.Tern(wire.StartNode.ResultAvailable && wire.StartNode.Result.Err != nil, Red, LightGray)
		rl.DrawLineBezier(
			rl.Vector2(wire.StartNode.OutputPortPositions[wire.StartPort]),
			rl.Vector2(wire.EndNode.InputPortPositions[wire.EndPort]),
			1,
			color.RGBA(),
		)
	}
	if draggingNewWire, _, _ := drag.State(NewWireDragKey); draggingNewWire {
		// Calculate end position in screen space
		mousePos := rl.GetMousePosition()

		rl.DrawLineBezier(
			rl.Vector2(NewWireSourceNode.OutputPortPositions[NewWireSourcePort]),
			mousePos,
			1,
			LightGray.RGBA(),
		)
	}

	for _, node := range currentGraph.Nodes {
		for _, portPos := range append(node.InputPortPositions, node.OutputPortPositions...) {
			rl.DrawCircle(int32(portPos.X), int32(portPos.Y), 4, White.RGBA())
		}
	}
}

func renderScreenOverlays() {
	if draggingBox, _, _ := drag.State(BoxSelectDragKey); draggingBox {
		start := drag.MouseStart
		end := rl.GetMousePosition()

		x := min(start.X, end.X)
		y := min(start.Y, end.Y)
		w := max(start.X, end.X) - x
		h := max(start.Y, end.Y) - y
		rect := rl.Rectangle{X: x, Y: y, Width: w, Height: h}

		c := rl.Color{R: uint8(Blue.R), G: uint8(Blue.G), B: uint8(Blue.B), A: uint8(Blue.A)}
		rl.DrawRectangleRec(rect, rl.Fade(c, 0.2))
		rl.DrawRectangleLinesEx(rect, 1, c)
	}
}

const GroupDragKey = "GROUP_DRAG"

type GroupDrag struct {
	Group *Group
	Nodes []*Node
}

func (g GroupDrag) DragKey() string { return GroupDragKey }

func UIGroup(group *Group) {
	clay.CLAY(clay.IDI("Group", group.ID), clay.EL{
		Floating: clay.FloatingElementConfig{
			AttachTo: clay.AttachToParent,
			Offset:   clay.Vector2(Camera.WorldToScreen(group.XY())),
			ClipTo:   clay.ClipToAttachedParent,
		},
		Layout: clay.LAY{
			LayoutDirection: clay.TopToBottom,
			Sizing:          clay.Sizing{Width: clay.SizingFixed(group.Rect.Width * Camera.Zoom), Height: clay.SizingFixed(group.Rect.Height * Camera.Zoom)},
		},
		BackgroundColor: clay.Color{R: group.Color.R, G: group.Color.G, B: group.Color.B, A: 50},
		Border:          clay.B{Color: group.Color, Width: BA},
		CornerRadius:    RA2,
	}, func() {
		clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
			// Don't set IsHoveringPanel so we can still pan if we click the background of a group
			// But we do want to catch clicks to select/drag
			UIInput.RegisterPointerDown(elementID, pointerData, 0)
		}, nil)

		// Header
		clay.CLAY_AUTO_ID(clay.EL{
			Layout: clay.LAY{
				Sizing:         GROWH,
				Padding:        PD(1, 0, 0, 0, PVH(S1, S2)),
				ChildAlignment: clay.ChildAlignment{Y: clay.AlignYCenter},
			},
			BackgroundColor: clay.Color{R: group.Color.R, G: group.Color.G, B: group.Color.B, A: 200},
		}, func() {
			clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, _ any) {
				UIInput.RegisterPointerDown(elementID, pointerData, 0)

				if drag.TryStartDrag(GroupDrag{Group: group}, rl.Rectangle{}, V2{}) {
					// Find nodes inside
					var nodes []*Node
					for _, n := range currentGraph.Nodes {
						// Simple center point check
						center := rl.Vector2Add(n.Pos, rl.Vector2{X: NodeMinWidth / 2, Y: 50}) // Approx center
						if rl.CheckCollisionPointRec(center, group.Rect) {
							nodes = append(nodes, n)
						}
					}
					// Update the drag thing with found nodes
					drag.Thing = GroupDrag{Group: group, Nodes: nodes}
				}
			}, nil)

			clay.TEXT(group.Title, clay.TextElementConfig{FontID: InterSemibold, FontSize: F3, TextColor: White})
		})
	})

	if dragging, _, _ := drag.State(GroupDragKey); dragging {
		if gd, ok := drag.Thing.(GroupDrag); ok && gd.Group.ID == group.ID {
			targetPos := drag.NewObjPosition()
			if !rl.IsKeyDown(rl.KeyLeftShift) && !rl.IsKeyDown(rl.KeyRightShift) {
				targetPos = SnapToGrid(targetPos)
			}

			delta := rl.Vector2Subtract(targetPos, group.XY())
			group.Rect.X = targetPos.X
			group.Rect.Y = targetPos.Y

			for _, n := range gd.Nodes {
				n.Pos = rl.Vector2Add(n.Pos, delta)
			}
		}
	}
}

func (r *Group) XY() rl.Vector2 {
	return rl.Vector2{X: r.Rect.X, Y: r.Rect.Y}
}

func UINode(node *Node, disabled bool) {
	// Node header font size
	headerFontSize := uint16(F3)

	border := clay.B{
		Color: Gray,
		Width: BA,
	}
	if node.Result.Err != nil {
		border = clay.B{
			Color: Red,
			Width: BA2,
		}
	} else if IsNodeSelected(node.ID) {
		border = clay.B{
			Color: Blue,
			Width: BA2,
		}
	}

	zIndex := int16(0)
	if IsNodeSelected(node.ID) {
		zIndex = 10
	}

	WithZIndex(zIndex, func() {
		clay.CLAY(node.ClayID(), clay.EL{
			Floating: clay.FloatingElementConfig{
				AttachTo: clay.AttachToParent,
				Offset:   clay.Vector2(Camera.WorldToScreen(node.Pos)),
				ClipTo:   clay.ClipToAttachedParent,
				ZIndex:   zIndex,
			},

			Layout: clay.LAY{
				LayoutDirection: clay.TopToBottom,
				Sizing:          clay.Sizing{Width: clay.SizingFit(NodeMinWidth, 0)},
			},
			BackgroundColor: DarkGray,
			Border:          border,
			CornerRadius:    RA2,
		}, func() {
			clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
				IsHoveringPanel = true
				UIInput.RegisterPointerDown(elementID, pointerData, zIndex)
				if UIInput.IsClick(elementID, pointerData) && !drag.WasDragging {
					UIFocus = nil
					multi := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) || rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
					SelectNode(node.ID, multi)
				}
			}, nil)

			clay.CLAY(clay.IDI("NodeHeader", node.ID), clay.EL{ // Node header
				Layout: clay.LAY{
					Sizing:         GROWH,
					Padding:        PD(1, 0, 0, 0, PVH(S1, S2)),
					ChildAlignment: clay.ChildAlignment{Y: clay.AlignYCenter},
				},
				BackgroundColor: Charcoal,
			}, func() {
				clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, _ any) {
					IsHoveringPanel = true
					UIInput.RegisterPointerDown(elementID, pointerData, zIndex)
					if UIInput.IsClick(elementID, pointerData) && !drag.WasDragging {
						UIFocus = nil
						multi := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) || rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
						SelectNode(node.ID, multi)
					}
				}, nil)

				clay.TEXT(node.Name, clay.TextElementConfig{FontID: InterSemibold, FontSize: headerFontSize, TextColor: White})
				UISpacer(node.DragHandleClayID(), GROWALL)
				// ... (rest of header using scaled vars if needed)
				// For now, let's just make sure the main container and header scale.
				// Buttons might need scaling too.

				// ...
				if node.Running {
					clay.TEXT("Running...", clay.TextElementConfig{TextColor: White})
				}

				playButtonDisabled := !node.Valid || node.Running || disabled

				UIButton(clay.IDI("NodePin", node.ID), // Pin button
					UIButtonConfig{
						El: clay.EL{Layout: clay.LAY{Padding: PA1}},
						OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
							node.Pinned = !node.Pinned
						},
					},
					func() {
						UIImage(clay.IDI("NodePinIcon", node.ID), util.Tern(node.Pinned, ImgPushpin, ImgPushpinOutline), clay.EL{
							BackgroundColor: Red,
						})

						if clay.Hovered() {
							UITooltip("Pin command (prevent automatic re-runs)")
						}
					},
				)
				UIButton(clay.IDI("NodeRetry", node.ID), // Retry / play-all button
					UIButtonConfig{
						El:       clay.EL{Layout: clay.LAY{Padding: PA1}},
						Disabled: playButtonDisabled,
						OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
							node.Run(context.Background(), true)
						},
					},
					func() {
						UIImage(clay.IDI("NodeRetryIcon", node.ID), ImgRetry, clay.EL{
							BackgroundColor: util.Tern(playButtonDisabled, LightGray, Blue),
						})

						if clay.Hovered() {
							UITooltip("Run command and inputs")
						}
					},
				)
				UIButton(clay.IDI("NodePlay", node.ID), // Play button
					UIButtonConfig{
						El:       clay.EL{Layout: clay.LAY{Padding: PA1}},
						Disabled: playButtonDisabled,
						OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
							node.Run(context.Background(), false)
						},
					},
					func() {
						UIImage(clay.IDI("NodePlayIcon", node.ID), ImgPlay, clay.EL{
							BackgroundColor: util.Tern(playButtonDisabled, LightGray, PlayButtonGreen),
						})

						if clay.Hovered() {
							UITooltip("Run command")
						}
					},
				)
				UIButton(clay.IDI("NodeDelete", node.ID), // Delete button
					UIButtonConfig{
						El: clay.EL{Layout: clay.LAY{Padding: PA1}},
						// Ensure delete button is always clickable above other node elements
						ZIndex: zIndex + 100,
						OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
							DeleteNode(node.ID)
							delete(selectedNodes, node.ID)
							if selectedNodeID == node.ID {
								selectedNodeID = 0
							}
						},
					},
					func() {
						clay.TEXT("X", clay.TextElementConfig{TextColor: util.Tern(clay.Hovered(), Red, LightGray), FontID: InterBold})
						if clay.Hovered() {
							UITooltip("Delete Node")
						}
					},
				)
			})
			clay.CLAY(clay.IDI("NodeBody", node.ID), clay.EL{ // Node body
				Layout: clay.LAY{Sizing: GROWH, Padding: PA2},
			}, func() {
				node.Action.UI(node)
			})
		})
	})
}

func UIInputPort(n *Node, port int) {
	clay.CLAY(clay.IDI(fmt.Sprintf("InputPort%d", port), n.ID), clay.EL{
		Layout: clay.LAY{ChildAlignment: YCENTER},
	}, func() {
		PortAnchor(n, false, port)
		clay.TEXT(n.InputPorts[port].Name, clay.TextElementConfig{TextColor: White})
	})
}

func UIOutputPort(n *Node, port int) {
	clay.CLAY(clay.IDI(fmt.Sprintf("OutputPort%d", port), n.ID), clay.EL{
		Layout: clay.LAY{ChildAlignment: YCENTER},
	}, func() {
		clay.TEXT(n.OutputPorts[port].Name, clay.TextElementConfig{TextColor: White})
		PortAnchor(n, true, port)
	})
}

func UIFlowValue(v FlowValue) {
	clay.CLAY(clay.AUTO_ID, clay.EL{}, func() {
		clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, _ any) {
			if data, ok := clay.GetElementData(elementID); ok {
				drag.TryStartDrag(FlowValueDrag{Value: v}, rl.Rectangle(data.BoundingBox), V2{})
			}
		}, nil)

		switch v.Type.Kind {
		case FSKindBytes:
			if len(v.BytesValue) == 0 {
				clay.TEXT("<no data>", clay.TextElementConfig{FontID: JetBrainsMono, TextColor: LightGray})
			} else {
				clay.TEXT(string(v.BytesValue), clay.TextElementConfig{FontID: JetBrainsMono, TextColor: White})
			}
		case FSKindInt64:
			var str string
			if v.Type.WellKnownType == FSWKTTimestamp {
				str = time.Unix(v.Int64Value, 0).Format(time.RFC1123)
			} else if v.Type.Unit == FSUnitBytes {
				str = FormatBytes(v.Int64Value)
			} else {
				str = fmt.Sprintf("%d", v.Int64Value)
			}
			clay.TEXT(str, clay.TextElementConfig{TextColor: White})
		case FSKindFloat64:
			var str string
			str = fmt.Sprintf("%v", v.Float64Value)
			clay.TEXT(str, clay.TextElementConfig{TextColor: White})
		case FSKindList:
			clay.CLAY_AUTO_ID(clay.EL{ // list items
				Layout: clay.LAY{LayoutDirection: clay.TopToBottom, ChildGap: S1},
			}, func() {
				for i, item := range v.ListValue {
					clay.CLAY_AUTO_ID(clay.EL{ // list item
						Layout: clay.LAY{ChildGap: S2},
					}, func() {
						clay.TEXT(fmt.Sprintf("%d", i), clay.TextElementConfig{FontID: InterSemibold, TextColor: White})
						UIFlowValue(item)
					})
				}
			})
		case FSKindTable:
			clay.CLAY_AUTO_ID(clay.EL{ // Table
				Border: clay.B{Width: BA_BTW, Color: Gray},
			}, func() {
				for col, field := range v.Type.ContainedType.Fields {
					clay.CLAY_AUTO_ID(clay.EL{ // Table col
						Layout: clay.LAY{LayoutDirection: clay.TopToBottom},
						Border: clay.B{Width: BTW, Color: Gray},
					}, func() {
						clay.CLAY_AUTO_ID(clay.EL{ // Header cell
							Layout: clay.LAY{Padding: PVH(S2, S3)},
						}, func() {
							clay.TEXT(field.Name, clay.TextElementConfig{FontID: InterSemibold, TextColor: White})
						})

						for _, val := range v.ColumnValues(col) {
							clay.CLAY_AUTO_ID(clay.EL{
								Layout: clay.LAY{Padding: PVH(S2, S3)},
							}, func() {
								UIFlowValue(val)
							})
						}
					})
				}
			})
		default:
			clay.TEXT("Unknown data type", clay.TextElementConfig{TextColor: White})
		}
	})
}

type UIButtonConfig struct {
	El       clay.ElementDeclaration
	Disabled bool

	OnHover         clay.OnHoverFunc
	OnHoverUserData any

	OnClick         clay.OnHoverFunc
	OnClickUserData any

	ZIndex int16
}

func UIButton(id clay.ElementID, config UIButtonConfig, children ...func()) {
	clay.CLAY_LATE(id, func() clay.ElementDeclaration {
		config.El.CornerRadius = RA1
		config.El.BackgroundColor = util.Tern(clay.Hovered() && !config.Disabled, HoverWhite, clay.Color{})

		return config.El
	}, func() {
		clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, _ any) {
			IsHoveringUI = true

			if !config.Disabled {
				UICursor = rl.MouseCursorPointingHand
				z := config.ZIndex
				if CurrentZIndex > z {
					z = CurrentZIndex
				}
				// Bump Z-index to ensure buttons are always above their background context (e.g. Node body)
				UIInput.RegisterPointerDown(elementID, pointerData, z+1)
			}

			if config.OnHover != nil {
				config.OnHover(elementID, pointerData, config.OnHoverUserData)
			}

			if !config.Disabled && UIInput.IsClick(elementID, pointerData) {
				UIFocus = nil
				if config.OnClick != nil {
					config.OnClick(elementID, pointerData, config.OnClickUserData)
				}
			}
		}, nil)

		for _, f := range children {
			f()
		}
	})
}

type UITextBoxConfig struct {
	El       clay.EL
	Disabled bool

	OnSubmit func(val string)

	ZIndex int16
}

func UITextBox(id clay.ElementID, str *string, config UITextBoxConfig, children ...func()) {
	if IsFocused(id) {
		if config.Disabled {
			UIFocus = nil
		} else {
			if rl.IsKeyPressed(rl.KeyEnter) {
				if config.OnSubmit != nil {
					config.OnSubmit(*str)
				}

				UIFocus = nil
			} else {
				for r := rl.GetCharPressed(); r != 0; r = rl.GetCharPressed() {
					*str = *str + string(rune(r))
				}
				if rl.IsKeyPressed(rl.KeyBackspace) || rl.IsKeyPressedRepeat(rl.KeyBackspace) {
					if len(*str) > 0 {
						*str = (*str)[:len(*str)-1]
					}
				}
			}
		}
	}

	clay.CLAY_LATE(id, func() clay.EL {
		config.El.Border = clay.B{Width: BA, Color: Gray}
		config.El.Layout.Padding = PVH(S1, S2)
		config.El.Layout.ChildAlignment.Y = clay.AlignYCenter
		config.El.BackgroundColor = DarkGray
		config.El.Clip = clay.CLIP{Horizontal: true}

		if clay.Hovered() {
			UICursor = rl.MouseCursorIBeam
		}

		clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
			IsHoveringUI = true
			z := config.ZIndex
			if CurrentZIndex > z {
				z = CurrentZIndex
			}
			UIInput.RegisterPointerDown(elementID, pointerData, z)

			if UIInput.IsClick(elementID, pointerData) {
				UIFocus = &elementID
			}
		}, nil)

		return config.El
	}, func() {
		clay.TEXT(*str, clay.T{TextColor: util.Tern(config.Disabled, LightGray, White)})
		UISpacer(clay.AUTO_ID, WH(1, 16))
		if IsFocused(id) {
			clay.CLAY_AUTO_ID(clay.EL{
				Layout:          clay.LAY{Sizing: WH(2, 16)},
				BackgroundColor: White,
			})
		}

		for _, f := range children {
			f()
		}
	})
}

type OnChangeFunc func(before, after any)

type UIDropdown struct {
	Options  []UIDropdownOption
	Selected int

	open bool
}

type UIDropdownOption struct {
	Name  string
	Value any
}

func (d *UIDropdown) GetOption(i int) UIDropdownOption {
	if len(d.Options) == 0 {
		return UIDropdownOption{}
	}
	if i >= len(d.Options) {
		return d.Options[0]
	}
	return d.Options[i]
}

func (d *UIDropdown) GetSelectedOption() UIDropdownOption {
	return d.GetOption(d.Selected)
}

func (d *UIDropdown) SelectByName(name string) bool {
	for i, opt := range d.Options {
		if opt.Name == name {
			d.Selected = i
			return true
		}
	}
	return false
}

func (d *UIDropdown) SelectByValue(v any) bool {
	for i, opt := range d.Options {
		if opt.Value == v {
			d.Selected = i
			return true
		}
	}
	return false
}

type UIDropdownConfig struct {
	El       clay.EL
	OnChange OnChangeFunc
}

func (d *UIDropdown) Do(id clay.ElementID, config UIDropdownConfig) {
	config.El.Layout.Padding = clay.Padding{}
	config.El.Layout.ChildAlignment.Y = clay.AlignYCenter
	config.El.Border = clay.BorderElementConfig{Width: BA, Color: Gray}
	config.El.BackgroundColor = DarkGray

	clay.CLAY(id, config.El, func() {
		clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
			IsHoveringUI = true
		}, nil)

		clay.CLAY_AUTO_ID(clay.EL{
			Layout: clay.LAY{
				Padding: PVH(S1, S2),
				Sizing:  GROWH,
			},
		}, func() {
			clay.TEXT(d.GetSelectedOption().Name, clay.TextElementConfig{TextColor: White})
		})
		UIButton(clay.AUTO_ID, UIButtonConfig{
			El: clay.EL{
				Layout: clay.LAY{
					ChildAlignment: ALLCENTER,
					Sizing:         GROWV,
					Padding:        PA2,
				},
				Border: clay.B{Width: clay.BW{Left: 1}, Color: Gray},
			},
			OnClick: func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
				d.open = !d.open
			},
		}, func() {
			UIImage(clay.AUTO_ID, util.Tern(d.open, ImgDropdownUp, ImgDropdownDown), clay.EL{
				BackgroundColor: LightGray,
			})
		})

		if d.open {
			WithZIndex(ZTOP, func() {
				clay.CLAY_AUTO_ID(clay.EL{
					Layout: clay.LAY{
						LayoutDirection: clay.TopToBottom,
						Sizing:          GROWH,
					},
					Floating: clay.FLOAT{
						AttachTo:           clay.AttachToParent,
						AttachPoints:       clay.FloatingAttachPoints{Parent: clay.AttachPointLeftBottom},
						ZIndex:             ZTOP,
						PointerCaptureMode: clay.PointercaptureModePassthrough,
					},
					Border: clay.B{
						Width: clay.BW{
							Top:  0,
							Left: 1, Right: 1,
							Bottom: 1,

							BetweenChildren: 1,
						},
						Color: Gray,
					},
					BackgroundColor: DarkGray,
				}, func() {
					clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
						IsHoveringUI = true
					}, nil)

					for i, opt := range d.Options {
						clay.CLAY_AUTO_ID_LATE(func() clay.EL {
							return clay.EL{
								Layout: clay.LAY{
									Padding: PVH(S1, S2),
									Sizing:  GROWH,
								},
								BackgroundColor: util.Tern(clay.Hovered(), HoverWhite, clay.Color{}),
							}
						}, func() {
							clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
								IsHoveringUI = true
								UIInput.RegisterPointerDown(elementID, pointerData, ZTOP)

								if UIInput.IsClick(elementID, pointerData) {
									selectedBefore := d.Selected
									d.Selected = userData.(int)
									d.open = false
									if config.OnChange != nil {
										config.OnChange(d.GetOption(selectedBefore).Value, d.GetOption(d.Selected).Value)
									}
								}
							}, i)
							clay.TEXT(opt.Name, clay.T{TextColor: White})
						})
					}
				})
			})
		}
	})
}

func UISpacer(id clay.ElementID, sizing clay.Sizing) {
	clay.CLAY(id, clay.EL{Layout: clay.LAY{Sizing: sizing}})
}

func UIImage(id clay.ElementID, img rl.Texture2D, decl clay.EL) {
	decl.Layout.Sizing = WH(float32(img.Width), float32(img.Height))
	decl.Image = clay.ImageElementConfig{ImageData: img}
	clay.CLAY(id, decl)
}

func UITooltip(msg string) {
	clay.CLAY_AUTO_ID(clay.EL{
		Floating: clay.FloatingElementConfig{
			AttachTo: clay.AttachToRoot,
			Offset:   clay.V2(rl.GetMousePosition()).Plus(clay.V2{X: 0, Y: 28}),
			ZIndex:   ZTOP,
		},
		Layout:          clay.LAY{Padding: PA1},
		BackgroundColor: DarkGray,
		Border:          clay.BorderElementConfig{Color: Gray, Width: BA},
	}, func() {
		clay.TEXT(msg, clay.TextElementConfig{TextColor: White})
	})
}

func UICheckbox(id clay.ElementID, checked *bool, label string) {
	clay.CLAY_AUTO_ID(clay.EL{
		Layout: clay.LAY{ChildGap: S1, ChildAlignment: YCENTER},
	}, func() {
		UIButton(id, UIButtonConfig{
			OnClick: func(_ clay.ElementID, _ clay.PointerData, _ any) {
				*checked = !*checked
			},
		}, func() {
			UIImage(clay.AUTO_ID, util.Tern(*checked, ImgToggleDown, ImgToggleRight), clay.EL{})
		})
		clay.TEXT(label, clay.TextElementConfig{TextColor: White})
	})
}

func PortAnchorID(node *Node, isOutput bool, port int) clay.ElementID {
	return clay.ID(fmt.Sprintf("N%d%s%d", node.ID, util.Tern(isOutput, "O", "I"), port))
}

func PortAnchor(node *Node, isOutput bool, port int) {
	// Give it a small non-zero size to ensure it has a valid bounding box for layout calculations
	clay.CLAY(PortAnchorID(node, isOutput, port), clay.EL{Layout: clay.LAY{Sizing: WH(1, 1)}})
}

func FormatBytes(n int64) string {
	if n < 1_000 {
		return fmt.Sprintf("%d B", n)
	} else if n < 1_000_000 {
		return fmt.Sprintf("%.1f KB", float32(n)/1_000)
	} else if n < 1_000_000_000 {
		return fmt.Sprintf("%.1f MB", float32(n)/1_000_000)
	} else if n < 1_000_000_000_000 {
		return fmt.Sprintf("%.1f GB", float32(n)/1_000_000_000)
	} else {
		return fmt.Sprintf("%.1f TB", float32(n)/1_000_000_000_000)
	}
}

func UIContextMenu() {
	if ContextMenu == nil {
		return
	}
	// Use ZTOP for the menu. Since ZTOP is MaxInt16, we can't go higher.
	// We rely on the fact that this is rendered late in the frame (overlay).
	WithZIndex(ZTOP, func() {
		clay.CLAY(clay.ID("ContextMenu"), clay.EL{
			Floating: clay.FloatingElementConfig{
				AttachTo: clay.AttachToRoot,
				Offset:   clay.Vector2(ContextMenu.Pos),
				ZIndex:   ZTOP,
			},
			Layout:          clay.LAY{LayoutDirection: clay.TopToBottom, Padding: PA1},
			BackgroundColor: DarkGray,
			Border:          clay.BorderElementConfig{Color: Gray, Width: BA},
			CornerRadius:    RA1,
		}, func() {
			clay.OnHover(func(elementID clay.ElementID, pointerData clay.PointerData, userData any) {
				IsHoveringUI = true
			}, nil)

			for i, item := range ContextMenu.Items {
				UIButton(clay.IDI("ContextItem", i), UIButtonConfig{
					El: clay.EL{
						Layout:          clay.LAY{Padding: PVH(S2, S3), Sizing: GROWH},
						BackgroundColor: util.Tern(clay.Hovered(), HoverWhite, clay.Color{}),
					},
					OnClick: func(_ clay.ElementID, _ clay.PointerData, _ any) {
						item.Action()
						ContextMenu = nil
					},
					ZIndex: ZTOP,
				}, func() {
					clay.TEXT(item.Label, clay.TextElementConfig{TextColor: White})
				})
			}
		})
	})
}
