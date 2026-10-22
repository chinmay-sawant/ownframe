//go:build !js && !android && !ios

package main

import "flag"

type options struct {
	web, stdin, random, media, snapshot bool
	addr, notifyAddr, mediaEndpoint     string
	margin, variant                     int
}

func flags() options {
	var o options
	flag.BoolVar(&o.web, "web", false, "serve a picture preview")
	flag.StringVar(&o.addr, "addr", "127.0.0.1:8128", "preview listen address (keep loopback; plaintext HTTP)")
	flag.IntVar(&o.margin, "margin", defaultMargin, "inset from monitor edges in CSS pixels")
	flag.IntVar(&o.variant, "variant", 0, "expression number, or 0 to cycle")
	flag.StringVar(&o.notifyAddr, "notify-addr", "127.0.0.1:6969", "notification listen address, or off (keep loopback; plaintext HTTP; non-loopback needs auth)")
	flag.BoolVar(&o.stdin, "stdin-notifications", false, "read launcher notifications from stdin")
	flag.BoolVar(&o.random, "random-behavior", false, "randomize idle cat expressions")
	flag.BoolVar(&o.media, "media", true, "watch native Chrome playback on Windows")
	flag.BoolVar(&o.snapshot, "media-snapshot", false, "print native media metadata and exit")
	flag.StringVar(&o.mediaEndpoint, "media-endpoint", "", "launcher media notification URL")
	flag.Parse()
	return o
}
