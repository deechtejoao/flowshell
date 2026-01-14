package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
)

func TestLoadFileNode_CSVPortVisibility(t *testing.T) {
	// Setup temp file
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "data.csv")
	err := os.WriteFile(csvPath, []byte("name,age\nalice,30\nbob,25"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create node
	n := nodes.NewLoadFileNode(csvPath)
	action := n.Action.(*nodes.LoadFileAction)

	// Ensure format is CSV
	action.Format.SelectByValue("csv")

	// 1. Initial state (no result yet)
	// NewLoadFileNode initializes OutputPorts with FSKindBytes, but UpdateAndValidate should clear it for CSV if no result.
	action.UpdateAndValidate(n)
	if len(n.OutputPorts) != 1 {
		t.Errorf("Expected 1 OutputPort initially for CSV, got %d", len(n.OutputPorts))
	}

	// 2. Run successfully
	resCh := action.Run(n)
	res := <-resCh
	if res.Err != nil {
		t.Fatalf("Run failed: %v", res.Err)
	}
	n.SetResult(res)

	// 3. Verify port visible
	action.UpdateAndValidate(n)
	if len(n.OutputPorts) != 1 {
		t.Errorf("Expected 1 output port after success, got %d", len(n.OutputPorts))
	} else if n.OutputPorts[0].Type.Kind != core.FSKindTable {
		t.Errorf("Expected FSKindTable, got %v", n.OutputPorts[0].Type.Kind)
	}

	// 4. Run failure (file not found)
	action.Path = filepath.Join(tmpDir, "nonexistent.csv")
	resCh = action.Run(n)
	res = <-resCh
	if res.Err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
	n.SetResult(res)

	// 5. Verify port visible as generic table
	action.UpdateAndValidate(n)
	if len(n.OutputPorts) != 1 {
		t.Errorf("Expected 1 output port after error, got %d", len(n.OutputPorts))
	} else if n.OutputPorts[0].Type.Kind != core.FSKindTable {
		t.Errorf("Expected FSKindTable, got %v", n.OutputPorts[0].Type.Kind)
	}

	// 6. Switch to Raw
	action.Format.SelectByValue("raw")
	action.UpdateAndValidate(n)
	if len(n.OutputPorts) != 1 {
		t.Errorf("Expected 1 output port for Raw, got %d", len(n.OutputPorts))
	} else if n.OutputPorts[0].Type.Kind != core.FSKindBytes {
		t.Errorf("Expected FSKindBytes, got %v", n.OutputPorts[0].Type.Kind)
	}
}
