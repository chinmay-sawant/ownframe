GO ?= go

# Test binaries run one package at a time by default so the suite does not
# build every package at once. Raise the limit for a faster run:
# make test TEST_P=4
TEST_P ?= 1

# Pin golangci-lint for local + CI reproducibility. Override:
# make lint GOLANGCI_LINT_VERSION=vX.Y.Z
# The binary must be built with a Go at least as new as the module version
# (go 1.26), so the install always uses GOTOOLCHAIN=local.
GOLANGCI_LINT_VERSION ?= v1.64.8

.PHONY: test
test:
	$(GO) test -p $(TEST_P) ./... ./examples/...

# `go build ./...` links an executable for every example on each run, takes
# about a minute, and leaves the binaries in this directory. `go vet` compiles
# every package, including every example, without linking. Raise the limit for
# a faster run: make build BUILD_P=4
BUILD_P ?= 1

.PHONY: build
build:
	$(GO) vet -p $(BUILD_P) ./... ./examples/...

# Lint both modules with the pinned golangci-lint and the policy in
# .golangci.yml. The library at the root and the examples module under
# examples/ are separate Go modules, so each needs its own pass.
# Installs the pinned binary into $(go env GOPATH)/bin when it is missing.
# Always builds with GOTOOLCHAIN=local so the binary matches go.mod.
# The run ends with the file-size gate (scripts/check-file-size.sh), so CI
# enforces the 2,000-character rule in AGENTS.md alongside the linters.
# The website in frontend/ is linted by ESLint in the same run, so one
# command covers the Go code and the React code.
# Every gate runs before the target fails, so one run reports the library
# findings, the examples findings, the size ledger, and the website findings
# together.
.PHONY: lint
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found; installing $(GOLANGCI_LINT_VERSION) with local Go toolchain..."; \
		GOTOOLCHAIN=local go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	}
	golangci-lint version
	@status=0; \
	golangci-lint run ./... || status=1; \
	golangci-lint run ./examples/... || status=1; \
	$(MAKE) --no-print-directory size-check || status=1; \
	if ! command -v npm >/dev/null 2>&1; then \
		echo "npm not found; the website lint needs Node.js 22.12 or newer" >&2; \
		status=1; \
	else \
		if [ ! -d frontend/node_modules ]; then \
			echo "frontend dependencies missing; running npm ci..."; \
			(cd frontend && npm ci) || status=1; \
		fi; \
		(cd frontend && npm run lint) || status=1; \
	fi; \
	exit $$status

# File-size gate (AGENTS.md "Go file size"). Scans every non-pruned .go
# file, test files included, and checks it against the recorded counts in
# scripts/file-size-allowlist.txt.
.PHONY: size-check
size-check:
	bash scripts/check-file-size.sh

# Every folder under examples/ that `make open` can walk.
EXAMPLES := $(patsubst examples/%/,%,$(wildcard examples/*/))

# Overlays and helpers that `make open` never walks. `desktop-cat` is a
# transparent click-through overlay, not a closable window, so run it alone:
#   go run ./examples/desktop-cat
# The form and input demos below now live in `input-lab`; the old folders
# stay for their tests, so open skips them. Run one alone when needed:
#   go run ./examples/login
SKIP_OPEN := desktop-cat bind bind-hooks clipboard controls editing forms input login states

# The folder named after `open`, if one was given.
OPEN_FROM := $(word 2,$(MAKECMDGOALS))

# Open examples one window at a time, from the named folder through the last
# folder under examples/:
#   make open crash              crash, then each later example, to the end
#   make open crash ARGS=-web    serve each in a browser instead
# A bare `make open` starts from the first folder. The next example opens
# after the previous window closes. Folders without Go files are skipped.
.PHONY: open
open:
	@start=0; \
	for dir in $(sort $(EXAMPLES)); do \
		if [ -z "$(OPEN_FROM)" ] || [ "$$dir" = "$(OPEN_FROM)" ]; then start=1; fi; \
		if [ $$start -eq 0 ]; then continue; fi; \
		case " $(SKIP_OPEN) " in *" $$dir "*) \
			if [ "$$dir" != "$(OPEN_FROM)" ]; then \
				echo "--- skip $$dir (run manually: go run ./examples/$$dir)"; \
				continue; \
			fi;; \
		esac; \
		if ! ls "examples/$$dir"/*.go >/dev/null 2>&1; then \
			echo "--- skip $$dir (no Go files)"; \
			continue; \
		fi; \
		if [ "$$($(GO) list -f '{{.Name}}' ./examples/$$dir 2>/dev/null)" != "main" ]; then \
			echo "--- skip $$dir (not a command)"; \
			continue; \
		fi; \
		echo "==> go run ./examples/$$dir $(ARGS)"; \
		$(GO) run ./examples/$$dir $(ARGS) || echo "--- examples/$$dir exited $$?"; \
	done; \
	if [ $$start -eq 0 ]; then \
		echo "make open: no folder examples/$(OPEN_FROM)"; \
		echo "examples: $(sort $(EXAMPLES))"; \
		exit 1; \
	fi

# `make open crash` turns crash into a second make goal, so each example
# folder needs a target of its own. This one does nothing.
.PHONY: $(EXAMPLES)
$(EXAMPLES):
	@:

# Unknown names fail with a pointer to `make open`.
%:
	@echo "make: unknown target '$@'; run 'make open <example>'" >&2; exit 1
