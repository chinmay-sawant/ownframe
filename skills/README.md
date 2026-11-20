# Repository Skills Inventory

This directory holds local agent workflow skills and templates. Nested `skills/go.mod` keeps this folder out of the library module archive.

## Active ownframe workflows

| Skill | Purpose |
|---|---|
| **PR** (`PR/`) | Pull request (`PR_TEMPLATE.md`) and issue (`ISSUE_TEMPLATE.md`) templates referenced by `CONTRIBUTING.md` |
| **phase-wise-checklist** (`phase-wise-checklist/`) | Evidence-backed phase-wise execution checklist convention used under `plans/` |
| **release-note** (`release-note/`) | Structured release note authoring for library and examples tags |
| **debug-html-template** (`debug-html-template/`) | Debugging HTML/CSS template rendering in ownframe |
| **diagnose-fixture-picture** / **diagnose-golden-fixture** / **generate-fixture-test** | Visual fixture diagnosis and golden test generation |
| **chrome-debug** / **chrome-debug-v2** / **chrome-flex-pdf-closure** | Chrome visual parity reference debugging |
| **css-best-practices** | CSS authoring rules compatible with Blinkless |
| **perf-patterns** / **perf-review** | Zero-allocation Go and frame-loop performance review |
| **critical-go-review** / **solid-go-review** / **golang-anti-patterns** / **diff-verify** / **improve-codebase** | Go code quality, SOLID review, and diff verification |
| **feynman** / **html-report** | Technical explanation and standalone HTML report generation |

## External / vendored templates

| Skill | Notes |
|---|---|
| **ponytail**, **ponytail-audit**, **ponytail-debt**, **ponytail-gain**, **ponytail-help**, **ponytail-review** (`PONYTAIL_VERSION.txt`) | Vendored Ponytail technical-debt workflow pack; optional and not required by CI or `Makefile` |
