package tv

import "time"

// Network and protocol constants for TV discovery and control.
const (
	ssdpPort      = 1900
	ssdpTimeout   = 1500 * time.Millisecond
	ssdpBufSize   = 2048
	nameTimeout   = 700 * time.Millisecond
	nameReadLimit = 8192
	helloTimeout  = 4 * time.Second
	sysTimeout    = 4 * time.Second
	mainTimeout   = 2 * time.Second
	tlsTimeout    = 2 * time.Second
	pairChanSize  = 4
	pairTimeout   = 45 * time.Second
	probeWorkers  = 40
	probeTimeout  = 200 * time.Millisecond
	roundTimeout  = 8 * time.Second
	stepsChanSize = 16
	stepTimeout   = 8 * time.Second
	storeDirPerm  = 0o700
	storeFilePerm = 0o600
	wakePktSize   = 102
)
