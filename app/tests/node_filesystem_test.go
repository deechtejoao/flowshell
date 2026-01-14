package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	
	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
)

func createInputNode(val core.FlowValue) *core.Node {
	n := &core.Node{ID: 1, OutputPorts: []core.NodePort{{Type: core.FlowType{Kind: core.FSKindAny}}}}
	n.SetResult(core.NodeActionResult{Outputs: []core.FlowValue{val}})
	return n
}

func setupGraph(testNode *core.Node, inputs ...core.FlowValue) *core.Graph {
	g := core.NewGraph()
	g.AddNode(testNode)

	for i, val := range inputs {
		inputNode := createInputNode(val)
		inputNode.ID = i + 100 // Unique ID
		g.AddNode(inputNode)
		g.AddWire(inputNode, 0, testNode, i)
	}
	return g
}

func TestFilesystemNodes(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Make Directory", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "new_dir")
		node := nodes.NewMakeDirNode()
		action := node.Action.(*nodes.MakeDirAction)

		setupGraph(node, core.FlowValue{Type: &core.FlowType{Kind: core.FSKindBytes}, BytesValue: []byte(dirPath)})

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		if res.Err != nil {
			t.Fatalf("Execution failed: %v", res.Err)
		}

		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("Directory was not created at %s", dirPath)
		}
	})

	t.Run("Copy File", func(t *testing.T) {
		srcPath := filepath.Join(tmpDir, "src.txt")
		dstPath := filepath.Join(tmpDir, "dst.txt")
		if err := os.WriteFile(srcPath, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}

		node := nodes.NewCopyFileNode()
		action := node.Action.(*nodes.CopyFileAction)

		setupGraph(node,
			core.FlowValue{Type: &core.FlowType{Kind: core.FSKindBytes}, BytesValue: []byte(srcPath)},
			core.FlowValue{Type: &core.FlowType{Kind: core.FSKindBytes}, BytesValue: []byte(dstPath)},
		)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		if res.Err != nil {
			t.Fatalf("Execution failed: %v", res.Err)
		}

		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}
		if string(content) != "hello" {
			t.Errorf("Expected 'hello', got '%s'", content)
		}
	})

	t.Run("Move File", func(t *testing.T) {
		srcPath := filepath.Join(tmpDir, "move_src.txt")
		dstPath := filepath.Join(tmpDir, "move_dst.txt")
		if err := os.WriteFile(srcPath, []byte("move me"), 0644); err != nil {
			t.Fatal(err)
		}

		node := nodes.NewMoveFileNode()
		action := node.Action.(*nodes.MoveFileAction)

		setupGraph(node,
			core.FlowValue{Type: &core.FlowType{Kind: core.FSKindBytes}, BytesValue: []byte(srcPath)},
			core.FlowValue{Type: &core.FlowType{Kind: core.FSKindBytes}, BytesValue: []byte(dstPath)},
		)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		if res.Err != nil {
			t.Fatalf("Execution failed: %v", res.Err)
		}

		if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
			t.Errorf("Source file should have been removed")
		}
		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("Failed to read destination file: %v", err)
		}
		if string(content) != "move me" {
			t.Errorf("Expected 'move me', got '%s'", content)
		}
	})

	t.Run("Delete File", func(t *testing.T) {
		delPath := filepath.Join(tmpDir, "del.txt")
		if err := os.WriteFile(delPath, []byte("bye"), 0644); err != nil {
			t.Fatal(err)
		}

		node := nodes.NewDeleteFileNode()
		action := node.Action.(*nodes.DeleteFileAction)

		setupGraph(node, core.FlowValue{Type: &core.FlowType{Kind: core.FSKindBytes}, BytesValue: []byte(delPath)})

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		res := <-done

		if res.Err != nil {
			t.Fatalf("Execution failed: %v", res.Err)
		}

		if _, err := os.Stat(delPath); !os.IsNotExist(err) {
			t.Errorf("File should have been deleted")
		}
	})
}