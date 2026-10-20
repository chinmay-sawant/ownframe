package main

import (
	"fmt"
	"os"

	"github.com/chinmay-sawant/ownframe/internal/transitionpreview"
)

func main() {
	if err := transitionpreview.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
