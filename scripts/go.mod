module github.com/chinmay-sawant/ownframe/scripts

// Maintenance tools have their own module. make tools-check tests them
// with GOWORK=off, and make lint includes that gate.
go 1.26.11

require (
	golang.org/x/mod v0.42.0
	golang.org/x/perf v0.0.0-20260908200009-22c9c6c9d4da
)

require (
	github.com/aclements/go-moremath v0.0.0-20210112150236-f10218a38794 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/telemetry v0.0.0-20261010193011-f71dc92da45e // indirect
	golang.org/x/tools v0.52.0 // indirect
	golang.org/x/vuln v1.9.0 // indirect
)

tool golang.org/x/vuln/cmd/govulncheck
