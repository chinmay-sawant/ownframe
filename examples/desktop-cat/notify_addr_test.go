//go:build !js && !android && !ios

package main

import "testing"

func TestCheckNotifyAddr(t *testing.T) {
	t.Parallel()
	for addr, wantErr := range map[string]bool{
		"127.0.0.1:6969":   false,
		"localhost:6969":   false,
		"192.168.1.5:6969": false,
		":6969":            true,
		"0.0.0.0:6969":     true,
		"127.0.0.1":        true,
	} {
		if err := checkNotifyAddr(addr); (err != nil) != wantErr {
			t.Fatalf("%s err=%v", addr, err)
		}
	}
}
