package atspi

import "github.com/chinmay-sawant/ownframe/internal/a11yqueue"

// ActionHandler executes one accessibility request on the page loop. The queue
// itself lives in a11yqueue so Windows and Android bridges do not import this
// Linux bus package.
type ActionHandler = a11yqueue.Handler

const actionCapacity = a11yqueue.Capacity
