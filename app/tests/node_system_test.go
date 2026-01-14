package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
	"github.com/stretchr/testify/assert"
)

func TestParseTimeNode(t *testing.T) {
	node := nodes.NewParseTimeNode()
	action := node.Action.(*nodes.ParseTimeAction)

	// Default RFC3339
	tVal, _ := time.Parse(time.RFC3339, "2023-10-01T12:00:00Z")
	setupGraph(node, core.NewStringValue("2023-10-01T12:00:00Z"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := action.RunContext(ctx, node)
	res := <-done

	assert.NoError(t, res.Err)
	assert.Equal(t, tVal.Unix(), res.Outputs[0].Int64Value)

	// Custom Format override
	t.Run("Custom Format", func(t *testing.T) {
		setupGraph(node,
			core.NewStringValue("01 Oct 23 12:00 UTC"),
			core.NewStringValue(time.RFC822),
		)

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Equal(t, tVal.Unix(), res.Outputs[0].Int64Value)
	})
}

func TestGetVariableNode(t *testing.T) {
	node := nodes.NewGetVariableNode()
	action := node.Action.(*nodes.GetVariableAction)
	action.VariableName = "TEST_VAR"

	// Test Env Var
	os.Setenv("TEST_VAR", "env_value")
	defer os.Unsetenv("TEST_VAR")

	setupGraph(node)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := action.RunContext(ctx, node)
	res := <-done

	assert.NoError(t, res.Err)
	assert.Equal(t, "env_value", string(res.Outputs[0].BytesValue))

	// Test Graph Var
	os.Unsetenv("TEST_VAR")
	node.Graph.Variables = map[string]string{"TEST_VAR": "graph_value"}

	done = action.RunContext(ctx, node)
	res = <-done

	assert.NoError(t, res.Err)
	assert.Equal(t, "graph_value", string(res.Outputs[0].BytesValue))
}

func TestValueNode(t *testing.T) {
	val := core.NewInt64Value(123, 0)
	node := nodes.NewValueNode(val)
	action := node.Action.(*nodes.ValueAction)

	setupGraph(node)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := action.RunContext(ctx, node)
	res := <-done

	assert.NoError(t, res.Err)
	assert.Equal(t, int64(123), res.Outputs[0].Int64Value)
}

func TestIfElseNode(t *testing.T) {
	t.Run("True Condition", func(t *testing.T) {
		node := nodes.NewIfElseNode()
		action := node.Action.(*nodes.IfElseAction)

		val := core.NewStringValue("val")
		cond := core.NewInt64Value(1, 0) // True

		setupGraph(node, val, cond)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		// True port (0) should have value
		assert.Equal(t, "val", string(res.Outputs[0].BytesValue))
		assert.False(t, res.Outputs[0].Skipped)
		// False port (1) should be skipped
		assert.True(t, res.Outputs[1].Skipped)
	})

	t.Run("False Condition", func(t *testing.T) {
		node := nodes.NewIfElseNode()
		action := node.Action.(*nodes.IfElseAction)

		val := core.NewStringValue("val")
		cond := core.NewInt64Value(0, 0) // False

		setupGraph(node, val, cond)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		// True port (0) should be skipped
		assert.True(t, res.Outputs[0].Skipped)
		// False port (1) should have value
		assert.Equal(t, "val", string(res.Outputs[1].BytesValue))
		assert.False(t, res.Outputs[1].Skipped)
	})
}
