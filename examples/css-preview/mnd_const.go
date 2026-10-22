package main

import "time"

// partialLimit is the most samples whose damage still fits the window's
// eight-rect partial replay. Past it the union covers the page, and a full
// replay is cheaper than a tile rebuild.
const partialLimit = 8

// frameWait is the idle delay before the demo asks for another frame.
const frameWait = 50 * time.Millisecond
