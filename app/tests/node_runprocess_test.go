package tests

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	
	"github.com/bvisness/flowshell/app/core"
	"github.com/bvisness/flowshell/app/nodes"
)

// TestRunProcess_ExitCode verifies that the RunProcess node correctly outputs the exit code.
func TestRunProcess_ExitCode(t *testing.T) {
	var cmdExit0, cmdExit1 string
	if runtime.GOOS == "windows" {
		cmdExit0 = "powershell -Command \"exit 0\""
		cmdExit1 = "powershell -Command \"exit 1\""
	} else {
		cmdExit0 = "sh -c \"exit 0\""
		cmdExit1 = "sh -c \"exit 1\""
	}

	// 1. Test successful exit (0)
	t.Run("Exit Code 0", func(t *testing.T) {
		node := nodes.NewRunProcessNode(cmdExit0)
		action := node.Action.(*nodes.RunProcessAction)

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
			if exitCodeVal.Type.Kind != core.FSKindInt64 {
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
		node := nodes.NewRunProcessNode(cmdExit1)
		action := node.Action.(*nodes.RunProcessAction)

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

func TestRunProcess_UseShell(t *testing.T) {
	var cmdPipe string
	if runtime.GOOS == "windows" {
		cmdPipe = "echo 'hello' | Write-Output"
	} else {
		cmdPipe = "echo 'hello' | cat"
	}

	t.Run("Shell Pipe", func(t *testing.T) {
		node := nodes.NewRunProcessNode(cmdPipe)
		action := node.Action.(*nodes.RunProcessAction)
		action.UseShell = true

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		done := action.RunContext(ctx, node)
		select {
		case res := <-done:
			if res.Err != nil {
				t.Fatalf("Node execution failed: %v", res.Err)
			}
			stdout := string(res.Outputs[0].BytesValue)
			// PowerShell might output CRLF
			if strings.TrimSpace(stdout) != "hello" {
				t.Errorf("Expected 'hello', got '%s'", stdout)
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