package main

import (
	"fmt"
	"os"

	"github.com/bvisness/flowshell/app"
)

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "run" {
			if len(os.Args) < 3 {
				fmt.Println("Usage: flowshell run <file.flow>")
				return
			}
			app.HeadlessRun(os.Args[2])
			return
		}
	}
	app.Main()
}
