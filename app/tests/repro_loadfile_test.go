package tests

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
)

func TestLoadFileNode_CSV(t *testing.T) {
	// Create temp CSV
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")
	f, err := os.Create(csvPath)
	if err != nil {
		t.Fatal(err)
	}
	w := csv.NewWriter(f)
	w.Write([]string{"Name", "Age"})
	w.Write([]string{"Alice", "30"})
	w.Write([]string{"Bob", "25"})
	w.Flush()
	f.Close()

	// Create node
	node := nodes.NewLoadFileNode(csvPath)
	action := node.Action.(*nodes.LoadFileAction)
	// Force CSV format
	action.Format.SelectByValue("csv")

	// Run
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := action.RunContext(ctx, node)

	select {
	case res := <-done:
		if res.Err != nil {
			t.Fatalf("Node failed: %v", res.Err)
		}
		if len(res.Outputs) != 1 {
			t.Fatalf("Expected 1 output, got %d", len(res.Outputs))
		}
		table := res.Outputs[0]
		if table.Type.Kind != core.FSKindTable {
			t.Fatalf("Expected Table output, got %v", table.Type.Kind)
		}
		rows := table.TableValue
		if len(rows) != 2 {
			t.Errorf("Expected 2 rows, got %d", len(rows))
		}
	case <-ctx.Done():
		t.Fatal("Node execution timed out")
	}
}
