package tests

import (
	"context"
	"testing"
	"time"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
	"github.com/stretchr/testify/assert"
)

func TestLinesNode(t *testing.T) {
	node := nodes.NewLinesNode()
	action := node.Action.(*nodes.LinesAction)
	// Force LF split for consistency across platforms in this test, unless we want to test platform-specific
	action.IncludeCarriageReturns = false // Treat \r as char unless we use it to split? Wait, CRLFSplit handles \r?\n.
	// Actually implementation uses util.Tern(l.IncludeCarriageReturns, CRLFSplit, LFSplit)
	// If IncludeCarriageReturns is true (Windows default), it splits on \r?\n.
	// If false, it splits on \n.

	// Let's explicitly set it to false and test with \n
	action.IncludeCarriageReturns = false

	setupGraph(node, core.NewStringValue("line1\nline2\nline3"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := action.RunContext(ctx, node)
	res := <-done

	assert.NoError(t, res.Err)
	assert.Len(t, res.Outputs[0].ListValue, 3)
	assert.Equal(t, "line1", string(res.Outputs[0].ListValue[0].BytesValue))
	assert.Equal(t, "line2", string(res.Outputs[0].ListValue[1].BytesValue))
	assert.Equal(t, "line3", string(res.Outputs[0].ListValue[2].BytesValue))
}

func TestRegexNodes(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		node := nodes.NewRegexMatchNode()
		action := node.Action.(*nodes.RegexMatchAction)

		setupGraph(node, core.NewStringValue("hello world"), core.NewStringValue(`^hello`))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Equal(t, int64(1), res.Outputs[0].Int64Value)
	})

	t.Run("FindAll", func(t *testing.T) {
		node := nodes.NewRegexFindAllNode()
		action := node.Action.(*nodes.RegexFindAllAction)

		setupGraph(node, core.NewStringValue("a1 b2 c3"), core.NewStringValue(`\d`))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		list := res.Outputs[0].ListValue
		assert.Len(t, list, 3)
		assert.Equal(t, "1", string(list[0].BytesValue))
		assert.Equal(t, "3", string(list[2].BytesValue))
	})

	t.Run("Replace", func(t *testing.T) {
		node := nodes.NewRegexReplaceNode()
		action := node.Action.(*nodes.RegexReplaceAction)

		setupGraph(node,
			core.NewStringValue("hello world"),
			core.NewStringValue(`world`),
			core.NewStringValue("gopher"),
		)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Equal(t, "hello gopher", string(res.Outputs[0].BytesValue))
	})

	t.Run("Split", func(t *testing.T) {
		node := nodes.NewRegexSplitNode()
		action := node.Action.(*nodes.RegexSplitAction)

		setupGraph(node,
			core.NewStringValue("a,b,c"),
			core.NewStringValue(`,`),
		)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		list := res.Outputs[0].ListValue
		assert.Len(t, list, 3)
		assert.Equal(t, "a", string(list[0].BytesValue))
		assert.Equal(t, "b", string(list[1].BytesValue))
	})
}

func TestTextOpsNodes(t *testing.T) {
	t.Run("JoinText", func(t *testing.T) {
		node := nodes.NewJoinTextNode()
		action := node.Action.(*nodes.JoinTextAction)

		list := core.NewListValue(core.FlowType{Kind: core.FSKindBytes}, []core.FlowValue{
			core.NewStringValue("a"),
			core.NewStringValue("b"),
		})

		setupGraph(node, list, core.NewStringValue("-"))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Equal(t, "a-b", string(res.Outputs[0].BytesValue))
	})

	t.Run("SplitText", func(t *testing.T) {
		node := nodes.NewSplitTextNode()
		action := node.Action.(*nodes.SplitTextAction)

		setupGraph(node, core.NewStringValue("a-b"), core.NewStringValue("-"))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Equal(t, "a", string(res.Outputs[0].ListValue[0].BytesValue))
	})

	t.Run("CaseConvert", func(t *testing.T) {
		node := nodes.NewCaseConvertNode()
		action := node.Action.(*nodes.CaseConvertAction)
		action.Mode = nodes.CaseUpper

		setupGraph(node, core.NewStringValue("hello"))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Equal(t, "HELLO", string(res.Outputs[0].BytesValue))
	})

	t.Run("FormatString", func(t *testing.T) {
		node := nodes.NewFormatStringNode()
		action := node.Action.(*nodes.FormatStringAction)
		action.Format = "Num: %d"

		setupGraph(node, core.NewInt64Value(42, 0))

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Equal(t, "Num: 42", string(res.Outputs[0].BytesValue))
	})
}

func TestTrimSpacesNode(t *testing.T) {
	node := nodes.NewTrimSpacesNode()
	action := node.Action.(*nodes.TrimSpacesAction)

	setupGraph(node, core.NewStringValue("  hello  "))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := action.RunContext(ctx, node)
	res := <-done

	assert.NoError(t, res.Err)
	assert.Equal(t, "hello", string(res.Outputs[0].BytesValue))
}

func TestMinifyHTMLNode(t *testing.T) {
	node := nodes.NewMinifyHTMLNode()
	action := node.Action.(*nodes.MinifyHTMLAction)

	setupGraph(node, core.NewStringValue("<div>  <span>  </span>  </div>"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := action.RunContext(ctx, node)
	res := <-done

	assert.NoError(t, res.Err)
	assert.Equal(t, "<div><span></span></div>", string(res.Outputs[0].BytesValue))
}
