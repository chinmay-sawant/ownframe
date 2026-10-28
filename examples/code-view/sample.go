package main

import "embed"

// sampleFiles is the bundled demo project the viewer opens by default.
// `-root` replaces it with a real directory.
//
//go:embed testdata/sample
var sampleFiles embed.FS
