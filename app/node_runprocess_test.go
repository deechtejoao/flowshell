package app

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

// TestRunProcess_ExitCode verifies that the RunProcess node correctly outputs the exit code.
func TestRunProcess_ExitCode(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows OS")
	}

	// 1. Test successful exit (0)
	t.Run("Exit Code 0", func(t *testing.T) {
		// PowerShell command that exits with 0
		cmd := "powershell -Command \"exit 0\""
		node := NewRunProcessNode(cmd)
		action := node.Action.(*RunProcessAction)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		select {
		case res := <-done:
			if res.Err != nil {
				t.Fatalf("Node execution failed: %v", res.Err)
			}
			if len(res.Outputs) != 4 {
				t.Fatalf("Expected 4 outputs, got %d", len(res.Outputs))
			}
			exitCodeVal := res.Outputs[3]
			if exitCodeVal.Type.Kind != FSKindInt64 {
				t.Errorf("Expected Exit Code type Int64, got %v", exitCodeVal.Type.Kind)
			}
			if exitCodeVal.Int64Value != 0 {
				t.Errorf("Expected Exit Code 0, got %d", exitCodeVal.Int64Value)
			}
		case <-ctx.Done():
			t.Fatal("Test timed out")
		}
	})

	// 2. Test failure exit (1)
	t.Run("Exit Code 1", func(t *testing.T) {
		// PowerShell command that exits with 1
		cmd := "powershell -Command \"exit 1\""
		node := NewRunProcessNode(cmd)
		action := node.Action.(*RunProcessAction)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		select {
		case res := <-done:
			if res.Err != nil {
				// Note: RunProcess captures ExitError and treats it as success with non-zero exit code.
				// However, if the command fails to start, it returns error.
				// Here we expect it to run but exit with 1.
				// The implementation says:
				// if exitErr, ok := runErr.(*exec.ExitError); ok {
				//     runErr = nil // We captured the exit code
				// }
				t.Fatalf("Node execution reported error (should be nil with exit code 1): %v", res.Err)
			}
			exitCodeVal := res.Outputs[3]
			if exitCodeVal.Int64Value != 1 {
				t.Errorf("Expected Exit Code 1, got %d", exitCodeVal.Int64Value)
			}
		case <-ctx.Done():
			t.Fatal("Test timed out")
		}
	})
}

// Helper to check if we can run powershell
func init() {
	_, err := exec.LookPath("powershell")
	if err != nil {
		fmt.Println("WARNING: powershell not found, tests might fail")
	}
}
