package tests

import (
	"testing"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
	"github.com/bvisness/flowshell/clay"
	"github.com/stretchr/testify/assert"
)

func TestNodeUILayouts(t *testing.T) {
	// 1. Initialize Clay
	minMem := clay.MinMemorySize()
	arena := clay.CreateArenaWithCapacity(uintptr(minMem))

	dims := clay.Dimensions{Width: 1024, Height: 768}
	clay.Initialize(arena, dims, clay.ErrorHandler{})

	// 2. Define all node creators
	creators := []func() *core.Node{
		nodes.NewAddColumnNode,
		func() *core.Node { return nodes.NewAggregateNode("Sum") },
		nodes.NewConcatTablesNode,
		nodes.NewTransposeNode,
		nodes.NewFilterEmptyNode,
		nodes.NewLinesNode,
		nodes.NewRegexMatchNode,
		nodes.NewRegexFindAllNode,
		nodes.NewRegexReplaceNode,
		nodes.NewRegexSplitNode,
		nodes.NewJoinTextNode,
		nodes.NewSplitTextNode,
		nodes.NewCaseConvertNode,
		nodes.NewFormatStringNode,
		nodes.NewTrimSpacesNode,
		nodes.NewMinifyHTMLNode,
		nodes.NewParseTimeNode,
		nodes.NewGetVariableNode,
		nodes.NewIfElseNode,
		nodes.NewMakeDirNode,
		nodes.NewCopyFileNode,
		nodes.NewMoveFileNode,
		nodes.NewDeleteFileNode,
		func() *core.Node { return nodes.NewListFilesNode(".") },
		nodes.NewSaveFileNode,
		nodes.NewHTTPRequestNode,
		nodes.NewLineChartNode,
		nodes.NewBarChartNode,
		nodes.NewScatterPlotNode,
		nodes.NewGraphInputNode,
		nodes.NewGraphOutputNode,
		nodes.NewPromptUserNode,
		nodes.NewGetMousePositionNode,
		nodes.NewWaitForClickNode,
		nodes.NewConvertNode,
		nodes.NewJsonQueryNode,
		nodes.NewXmlQueryNode,
		nodes.NewExtractColumnNode,
		nodes.NewFormulaNode,
		func() *core.Node { return nodes.NewLoadFileNode("test.txt") },
		nodes.NewMapNode,
		nodes.NewSelectColumnsNode,
		nodes.NewSortNode,
		nodes.NewGateNode,
		nodes.NewMergeNode,
		func() *core.Node { return nodes.NewRunProcessNode("echo hello") },
	}

	// 3. Test Loop
	for _, creator := range creators {
		n := creator()

		t.Run(n.Name, func(t *testing.T) {
			assert.NotNil(t, n)
			assert.NotNil(t, n.Action)

			// Some nodes need wire connections to be "Valid" or show full UI
			// but UI() should be robust enough to handle disconnected state.
			// We wrap it in a safe layout begin/end just in case.

			clay.BeginLayout()

			// Mock getting element data?
			// Some UI implementations call clay.GetElementData().
			// Without a previous render, this might return false.
			// This is expected and should not panic.

			assert.NotPanics(t, func() {
				n.Action.UI(n)
			})

			clay.EndLayout()
		})
	}
}
