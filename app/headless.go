package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func HeadlessRun(path string) {
	fmt.Printf("Starting Flowshell HEADLESS mode...\n")
	fmt.Printf("Loading graph: %s\n", path)

	g, err := LoadGraph(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading graph: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Graph loaded. %d nodes, %d wires.\n", len(g.Nodes), len(g.Wires))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, shutting down...")
		cancel()
	}()

	// Start all nodes
	// We start all nodes because in a flow-based system, we don't necessarily know
	// which ones are "entry points" (some might be listeners, some might be timers).
	// Nodes that need inputs will block until inputs are available.

	fmt.Println("Starting nodes...")
	activeNodes := 0
	for _, n := range g.Nodes {
		if n.Action != nil {
			activeNodes++
			// We discard the result channel for now in headless mode,
			// unless we want to log errors from it.
			// Ideally specific nodes should log to stdout/stderr or a file.
			// For now, let's just let them run.
			go func(node *Node) {
				// We call n.Run which internally handles context and calls Run/RunContext
				done := node.Run(ctx, false)
				<-done
				if res, ok := node.GetResult(); ok && res.Err != nil {
					fmt.Fprintf(os.Stderr, "[Node %d %s] Error: %v\n", node.ID, node.Name, res.Err)
				}
			}(n)
		}
	}

	fmt.Printf("Started %d nodes. Running (Ctrl+C to stop)...\n", activeNodes)

	// Wait for context cancellation (interrupt)
	<-ctx.Done()

	// Give a small grace period for cleanup if needed?
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Shutdown complete.")
}
