package main

import (
	"fmt"
	"os"

	"github.com/chinmay-sawant/ownframe/internal/animationpreview"
)

func main() {
	if err := animationpreview.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
