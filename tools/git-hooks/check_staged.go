package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Get staged files
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		// If git fails, maybe no repo or other error, let it pass or fail
		fmt.Printf("Warning: could not check staged files: %v\n", err)
		os.Exit(0)
	}

	files := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(files) == 0 || (len(files) == 1 && files[0] == "") {
		os.Exit(0)
	}

	dirs := make(map[string]bool)
	for _, f := range files {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		// Get top level dir
		parts := strings.Split(f, "/")
		if len(parts) > 1 {
			dirs[parts[0]] = true
		} else {
			// Root file
			dirs["root"] = true
		}
	}

	// Heuristic: If modifying > 2 top level components, warn
	// Exceptions: docs, tools, .vscode usually don't count as "app logic" mixed with other things if they are the only ones
	// But simple rule: if app/ and clay/ and util/ are all modified, that's suspicious.

	count := 0
	components := []string{}
	for d := range dirs {
		if d == "app" || d == "clay" || d == "corpus" || d == "trace" || d == "util" {
			count++
			components = append(components, d)
		}
	}

	if count > 2 {
		fmt.Println("WARNING: You are modifying multiple components in a single commit:")
		for _, c := range components {
			fmt.Printf(" - %s\n", c)
		}
		fmt.Println("Atomic commits should ideally affect only one component.")
		fmt.Println("If this is a refactor, please ensure the commit message reflects that.")
		// We don't block, just warn. To block, use os.Exit(1)
		// User asked to "Prevent commits with mixed concerns", so maybe I should block.
		// I'll block if it's really wide (>2 components).
		os.Exit(1)
	}

	os.Exit(0)
}
