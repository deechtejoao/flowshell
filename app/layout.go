package app

import (
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// LayoutGraph applies a naive Sugiyama-like hierarchical layout to the graph.
func LayoutGraph(g *Graph) {
	if len(g.Nodes) == 0 {
		return
	}

	// 1. Build Adjacency List (ignoring back-edges for acyclic layering?)
	// For simplicity, we'll assign ranks based on longest path from roots.
	// Since we handle cycles in execution, layout might just do BFS/DFS depth.

	// Map NodeID -> Rank (Layer index)
	ranks := make(map[int]int)
	// Map NodeID -> In-degree (for topological processing)
	inDegree := make(map[int]int)
	adj := make(map[int][]int)

	// Initialize
	for _, n := range g.Nodes {
		ranks[n.ID] = 0
		inDegree[n.ID] = 0
		adj[n.ID] = []int{}
	}

	// Build edges
	for _, w := range g.Wires {
		adj[w.StartNode.ID] = append(adj[w.StartNode.ID], w.EndNode.ID)
		inDegree[w.EndNode.ID]++
	}

	// 2. Assign Ranks (Longest Path Layering)
	// Nodes with 0 in-degree are sources at rank 0.
	// We use a queue for processing.
	queue := []int{}
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	// "Push" approach (BFS-like) works better with established adj list
	// Iterate queue.
	nodeCount := len(g.Nodes)

	// Reset ranks to -1 to detect unvisited
	for id := range ranks {
		ranks[id] = 0
	}

	// Kahn's algorithm adapted for layering
	// We need to keep queueing children when all parents are processed?
	// Simplified: Just push updates. Rank(Child) = max(Rank(Child), Rank(Parent) + 1)

	// We'll run relaxation for NodeCount rounds (Bellman-Ford style for longest path in DAG)
	for i := 0; i < nodeCount; i++ {
		changed := false
		for _, w := range g.Wires {
			if ranks[w.EndNode.ID] <= ranks[w.StartNode.ID] {
				ranks[w.EndNode.ID] = ranks[w.StartNode.ID] + 1
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// 3. Group Nodes by Layer
	layers := make(map[int][]*Node)
	maxRank := 0
	for _, n := range g.Nodes {
		r := ranks[n.ID]
		if r > maxRank {
			maxRank = r
		}
		layers[r] = append(layers[r], n)
	}

	// 4. Order Nodes within Layers & Position
	// Constant values for spacing
	const NodeXSpacing = 400
	const NodeYSpacing = 200
	const StartX = 100
	const StartY = 100

	// Sort layers by ID initially for stability
	for r := 0; r <= maxRank; r++ {
		layerNodes := layers[r]
		sort.Slice(layerNodes, func(i, j int) bool {
			return layerNodes[i].ID < layerNodes[j].ID
		})

		// Assign Positions
		// Center alignment? Or simple left-align?
		// Simple left-align is fine for v1.

		for i, n := range layerNodes {
			n.Pos = rl.Vector2{
				X: float32(StartX + r*NodeXSpacing), // Left-to-Right layout
				Y: float32(StartY + i*NodeYSpacing),
			}
		}
	}
}
