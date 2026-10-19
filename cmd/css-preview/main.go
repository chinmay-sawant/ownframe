package main

import (
	"fmt"
	"os"

	"github.com/chinmay-sawant/ownframe/internal/csspreview"
)

func main() {
	if len(os.Args) > 1 {
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/css-preview/main.go")
		os.Exit(2)
	}
	if err := csspreview.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
