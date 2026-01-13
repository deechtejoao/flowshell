package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func createInputNode(val FlowValue) *Node {
	n := &Node{ID: 1, OutputPorts: []NodePort{{Type: FlowType{Kind: FSKindAny}}}}
	n.SetResult(NodeActionResult{Outputs: []FlowValue{val}})
	return n
}

func setupGraph(testNode *Node, inputs ...FlowValue) *Graph {
	g := NewGraph()
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
		node := NewMakeDirNode()
		action := node.Action.(*MakeDirAction)

		setupGraph(node, FlowValue{Type: &FlowType{Kind: FSKindBytes}, BytesValue: []byte(dirPath)})

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

		node := NewCopyFileNode()
		action := node.Action.(*CopyFileAction)

		setupGraph(node,
			FlowValue{Type: &FlowType{Kind: FSKindBytes}, BytesValue: []byte(srcPath)},
			FlowValue{Type: &FlowType{Kind: FSKindBytes}, BytesValue: []byte(dstPath)},
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

		node := NewMoveFileNode()
		action := node.Action.(*MoveFileAction)

		setupGraph(node,
			FlowValue{Type: &FlowType{Kind: FSKindBytes}, BytesValue: []byte(srcPath)},
			FlowValue{Type: &FlowType{Kind: FSKindBytes}, BytesValue: []byte(dstPath)},
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

		node := NewDeleteFileNode()
		action := node.Action.(*DeleteFileAction)

		setupGraph(node, FlowValue{Type: &FlowType{Kind: FSKindBytes}, BytesValue: []byte(delPath)})

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
