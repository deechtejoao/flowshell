package tests

import (
	"context"
	"testing"
	"time"

	
	"github.com/stretchr/testify/assert"
	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
)

func createTestInputNode(val core.FlowValue) *core.Node {
	n := &core.Node{ID: 1, OutputPorts: []core.NodePort{{Type: core.FlowType{Kind: core.FSKindAny}}}}
	n.SetResult(core.NodeActionResult{Outputs: []core.FlowValue{val}})
	return n
}

func setupTestGraph(testNode *core.Node, inputs ...core.FlowValue) *core.Graph {
	g := core.NewGraph()
	g.AddNode(testNode)

	for i, val := range inputs {
		inputNode := createTestInputNode(val)
		inputNode.ID = i + 100 // Unique ID
		g.AddNode(inputNode)
		g.AddWire(inputNode, 0, testNode, i)
	}
	return g
}

func TestGateNode(t *testing.T) {
	val := core.NewStringValue("data")

	t.Run("Condition True", func(t *testing.T) {
		node := nodes.NewGateNode()
		action := node.Action.(*nodes.GateAction)

		cond := core.NewInt64Value(1, 0) // Truthy
		setupTestGraph(node, val, cond)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Len(t, res.Outputs, 1)
		assert.Equal(t, "data", string(res.Outputs[0].BytesValue))
		assert.False(t, res.Outputs[0].Skipped)
	})

	t.Run("Condition False", func(t *testing.T) {
		node := nodes.NewGateNode()
		action := node.Action.(*nodes.GateAction)

		cond := core.NewInt64Value(0, 0) // Falsy
		setupTestGraph(node, val, cond)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Len(t, res.Outputs, 1)
		assert.True(t, res.Outputs[0].Skipped)
	})
}

func TestMergeNode(t *testing.T) {
	valA := core.NewStringValue("A")
	valB := core.NewStringValue("B")
	valSkipped := core.FlowValue{Type: &core.FlowType{Kind: core.FSKindAny}, Skipped: true}

	t.Run("A Valid, B Skipped", func(t *testing.T) {
		node := nodes.NewMergeNode()
		action := node.Action.(*nodes.MergeAction)

		setupTestGraph(node, valA, valSkipped)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Len(t, res.Outputs, 1)
		assert.Equal(t, "A", string(res.Outputs[0].BytesValue))
		assert.False(t, res.Outputs[0].Skipped)
	})

	t.Run("A Skipped, B Valid", func(t *testing.T) {
		node := nodes.NewMergeNode()
		action := node.Action.(*nodes.MergeAction)

		setupTestGraph(node, valSkipped, valB)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Len(t, res.Outputs, 1)
		assert.Equal(t, "B", string(res.Outputs[0].BytesValue))
		assert.False(t, res.Outputs[0].Skipped)
	})

	t.Run("Both Skipped", func(t *testing.T) {
		node := nodes.NewMergeNode()
		action := node.Action.(*nodes.MergeAction)

		setupTestGraph(node, valSkipped, valSkipped)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Len(t, res.Outputs, 1)
		assert.True(t, res.Outputs[0].Skipped)
	})

	t.Run("Both Valid (Priority A)", func(t *testing.T) {
		node := nodes.NewMergeNode()
		action := node.Action.(*nodes.MergeAction)

		setupTestGraph(node, valA, valB)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		assert.NoError(t, res.Err)
		assert.Len(t, res.Outputs, 1)
		assert.Equal(t, "A", string(res.Outputs[0].BytesValue))
		assert.False(t, res.Outputs[0].Skipped)
	})
}