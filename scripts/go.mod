module github.com/chinmay-sawant/ownframe/scripts

// Maintenance tools have their own module. make tools-check tests them
// with GOWORK=off, and make lint includes that gate.
go 1.26.10

require golang.org/x/mod v0.42.0
