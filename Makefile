GO ?= go

# Test binaries run one package at a time by default so the suite does not
# build every package at once. Raise the limit for a faster run:
# make test TEST_P=4
TEST_P ?= 1

# Pin golangci-lint for local + CI reproducibility. Override:
# make lint GOLANGCI_LINT_VERSION=vX.Y.Z
# The v2 module path needs a Go at least as new as the module version
# (go 1.26), so the install uses the default toolchain resolution.
GOLANGCI_LINT_VERSION ?= v2.15.0

.PHONY: test
test:
	$(GO) test -p $(TEST_P) ./... ./examples/...

# Race detector on the library module only. Examples stay off this target
# until the library job is green; full examples race is a later gate.
.PHONY: test-race
test-race:
	$(GO) test -race -p 1 ./...

# `go build ./...` links an executable for every example on each run, takes
# about a minute, and leaves the binaries in this directory. `go vet` compiles
# every package, including every example, without linking. Raise the limit for
# a faster run: make build BUILD_P=4
BUILD_P ?= 1

.PHONY: build
build:
	$(GO) vet -p $(BUILD_P) ./... ./examples/...

# Use exactly the declared runner version, including when PATH has another version.
.PHONY: lint lint-version docs-check plans-check graphics-check examples-check tools-check
lint-version:
	bash scripts/lint_test.sh
	bash scripts/lint.sh $(GOLANGCI_LINT_VERSION) --version

# VERBOSE=1 prints the complete output from successful tools as well.
VERBOSE ?= 0
lint:
	@python3 scripts/run-lint.py $(GOLANGCI_LINT_VERSION) $(if $(filter 1,$(VERBOSE)),--verbose)

tools-check:
	@python3 -m unittest discover -s scripts -p 'test_*.py'
	@cd scripts && GOWORK=off $(GO) test -p 1 ./...

docs-check:
	python3 scripts/check-doc-snippets.py

plans-check:
	python3 scripts/check-plans.py

examples-check:
	python3 scripts/check-examples.py

graphics-check:
	bash scripts/check-graphics.sh

# Freeze gate: no new examples import of ownframe/internal/. The current
# set is scripts/example-internal-allowlist.txt. Stale or extra pairs fail.
.PHONY: internal-import-check
internal-import-check:
	bash scripts/check-example-internal.sh

# Vulnerability scan via the pinned tool directive in go.mod
# (golang.org/x/vuln/cmd/govulncheck). Runs against both modules through
# the workspace. Usage: make vuln
.PHONY: vuln
vuln:
	go tool govulncheck ./...

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

.PHONY: module-check
module-check:
	@cd scripts && GOWORK=off $(GO) run ./cmd/module-check -root ..

.PHONY: site-check
site-check:
	@npm --prefix frontend test
	@python3 scripts/check-generated-site.py
	@npm --prefix frontend run check:site

.PHONY: release-check
release-check:
	@sh scripts/package_test.sh
	@sh scripts/package.sh login
	@python3 scripts/check-release-bundle.py dist

SOAK_SECONDS ?= 600
.PHONY: soak-check
soak-check:
	@OWNFRAME_SOAK_SECONDS=$(SOAK_SECONDS) OWNFRAME_SOAK_COMMIT="$$(git rev-parse HEAD)" $(GO) test -p 1 -count=1 -v -timeout=35m -run '^TestPageSoak$$' ./internal/page
