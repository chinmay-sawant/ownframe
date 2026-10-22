//go:build !js && !android && !ios

package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

// maxNotifyHeaderBytes caps notification request headers at 64 KiB.
const maxNotifyHeaderBytes = 1 << 16

// notifyTimeouts bounds the notification server I/O.
const (
	notifyReadHeaderTimeout = 3 * time.Second
	notifyReadTimeout       = 5 * time.Second
	notifyWriteTimeout      = 5 * time.Second
)

// checkNotifyAddr rejects empty and wildcard binds. Other non-loopback
// binds pass with a warning; the caller must front them with auth.
func checkNotifyAddr(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("notification server: %w", err)
	}
	if host == "" || port == "" {
		return fmt.Errorf("notification server: bind needs host and port")
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsUnspecified() {
			return fmt.Errorf("notification server: wildcard bind needs auth")
		}
		if !ip.IsLoopback() {
			log.Printf("notification server: non-loopback %s needs auth", addr)
		}
		return nil
	}
	if host != "localhost" {
		log.Printf("notification server: non-loopback %s needs auth", addr)
	}
	return nil
}
