package app

import (
	"context"

	"github.com/bvisness/flowshell/app/core"
)

func RunGraph(ctx context.Context, g *core.Graph, onComplete func(error)) error {
	nodes, err := core.Toposort(g.Nodes, g.Wires)
	if err != nil {
		return err
	}

	// 1. Validate & Update types based on connectivity
	for _, n := range nodes {
		n.Action.UpdateAndValidate(n)
	}

	// 2. Run nodes
	go func() {
		var finalErr error
		defer func() {
			if onComplete != nil {
				onComplete(finalErr)
			}
		}()

		// We need to wait for all nodes to finish.
		// Since they run independently (waiting on their inputs), we can just start them all
		// and collect their done channels.
		var doneChans []<-chan struct{}
		for _, n := range nodes {
			doneChans = append(doneChans, n.Run(ctx, false))
		}

		// Wait for all
		for _, ch := range doneChans {
			select {
			case <-ch:
			case <-ctx.Done():
				finalErr = ctx.Err()
				return
			}
		}
	}()

	return nil
}
