package main

import (
	"fmt"
	"os"

	"github.com/zpenka/brief"
)

func main() {
	if err := brief.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "brief: %v\n", err)
		os.Exit(1)
	}
}
