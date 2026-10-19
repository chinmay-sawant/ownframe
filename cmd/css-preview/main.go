package main

import (
	"fmt"
	"os"

	"github.com/chinmay-sawant/ownframe/internal/csspreview"
)

func main() {
	kind := "all"
	if len(os.Args) > 1 {
		kind = os.Args[1]
	}
	if err := csspreview.Run(kind); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
