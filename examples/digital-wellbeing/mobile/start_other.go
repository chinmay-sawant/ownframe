//go:build !android && !ios

package mobile

func start() error { return nil }

func init() { _ = start() }
