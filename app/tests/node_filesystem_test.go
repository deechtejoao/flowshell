package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bvisness/flowshell/app"
)

func createInputNode(val app.FlowValue) *app.Node {
	n := &app.Node{ID: 1, OutputPorts: []app.NodePort{{Type: app.FlowType{Kind: app.FSKindAny}}}}
	n.SetResult(app.NodeActionResult{Outputs: []app.FlowValue{val}})
	return n
}

func setupGraph(testNode *app.Node, inputs ...app.FlowValue) *app.Graph {
	g := app.NewGraph()
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
		node := app.NewMakeDirNode()
		action := node.Action.(*app.MakeDirAction)

		setupGraph(node, app.FlowValue{Type: &app.FlowType{Kind: app.FSKindBytes}, BytesValue: []byte(dirPath)})

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

		node := app.NewCopyFileNode()
		action := node.Action.(*app.CopyFileAction)

		setupGraph(node,
			app.FlowValue{Type: &app.FlowType{Kind: app.FSKindBytes}, BytesValue: []byte(srcPath)},
			app.FlowValue{Type: &app.FlowType{Kind: app.FSKindBytes}, BytesValue: []byte(dstPath)},
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

		node := app.NewMoveFileNode()
		action := node.Action.(*app.MoveFileAction)

		setupGraph(node,
			app.FlowValue{Type: &app.FlowType{Kind: app.FSKindBytes}, BytesValue: []byte(srcPath)},
			app.FlowValue{Type: &app.FlowType{Kind: app.FSKindBytes}, BytesValue: []byte(dstPath)},
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

		node := app.NewDeleteFileNode()
		action := node.Action.(*app.DeleteFileAction)

		setupGraph(node, app.FlowValue{Type: &app.FlowType{Kind: app.FSKindBytes}, BytesValue: []byte(delPath)})

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
