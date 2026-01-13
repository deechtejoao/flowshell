package tests

import (
	"context"
	"testing"
	"time"

	"github.com/bvisness/flowshell/app"
	"github.com/stretchr/testify/assert"
)

func createTestInputNode(val app.FlowValue) *app.Node {
	n := &app.Node{ID: 1, OutputPorts: []app.NodePort{{Type: app.FlowType{Kind: app.FSKindAny}}}}
	n.SetResult(app.NodeActionResult{Outputs: []app.FlowValue{val}})
	return n
}

func setupTestGraph(testNode *app.Node, inputs ...app.FlowValue) *app.Graph {
	g := app.NewGraph()
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
	val := app.NewStringValue("data")

	t.Run("Condition True", func(t *testing.T) {
		node := app.NewGateNode()
		action := node.Action.(*app.GateAction)

		cond := app.NewInt64Value(1, 0) // Truthy
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
		node := app.NewGateNode()
		action := node.Action.(*app.GateAction)

		cond := app.NewInt64Value(0, 0) // Falsy
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
	valA := app.NewStringValue("A")
	valB := app.NewStringValue("B")
	valSkipped := app.FlowValue{Type: &app.FlowType{Kind: app.FSKindAny}, Skipped: true}

	t.Run("A Valid, B Skipped", func(t *testing.T) {
		node := app.NewMergeNode()
		action := node.Action.(*app.MergeAction)

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
		node := app.NewMergeNode()
		action := node.Action.(*app.MergeAction)

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
		node := app.NewMergeNode()
		action := node.Action.(*app.MergeAction)

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
		node := app.NewMergeNode()
		action := node.Action.(*app.MergeAction)

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
