# Independent maintenance review

Give the prompt below to a reviewer who has not implemented the changes. Record the source commit and dirty state before running checks. A second agent repeating the author's claims is not an independent maintainer handoff.

## Review prompt

Review OwnFrame as a Go library maintained by someone who may disappear. Be opinionated and substantiate every finding with current source, a reproducible command, or an observed artifact.

Start with `AGENTS.md`, `CONTRIBUTING.md`, `go.mod`, `examples/go.mod`, `Makefile`, `.golangci.yml`, and the last 20 commits. Read `documentation/features.md`, `documentation/platforms.md`, and `plans/README.md`. Distinguish implementation, deferred scope, superseded plans, and behavior that still needs validation.

Investigate these questions:

1. Do public API examples compile against the checkout and a module consumer with `GOWORK=off`? Are library fixtures included in the module zip? Can examples releases resolve a published library without filesystem replacements?
2. Do tests protect retained state, software pictures, GPU window repaint, idle scheduling, and translucent replay separately? Which claims still lack real device or native host evidence?
3. Do packages own clear responsibilities? Identify duplicated state, unnecessary forwarding layers, weak ownership, and file splits that make one behavior harder to follow.
4. Are documentation claims consistent with runtime defaults, build tags, debug exposure, dependency pins, and measured limits? Check shell commands as well as Go fences.
5. Are maintenance checks strict on failures and missing inputs? Does the declared linter version actually run? Do workflow gates validate generated website output and release candidates?
6. Are release archives, checksums, provenance, and signatures validated? Separate prepared candidates from published releases and local checks from successful tag workflows.
7. Can one new maintainer build, debug, change, and validate the maintained teaching examples using the guides? Identify onboarding steps that require undocumented local state.
8. What evidence supports each platform claim? WSLg is not native Windows. A successful APK build is not device proof. Synthetic typing is not complete IME composition testing. No native screen-reader adapter should be called shipped without live validation.

For each finding give severity, mechanism, source reference, user impact, smallest coherent fix, and a concrete completion check. Separate craft from sprawl. Identify the strongest sign of agent-generated residue. Rate maintainability out of 10 and explain the remaining gap without rewarding file or test counts alone.

Begin with read-only inspection and cheap commands. Do not run `make test` during an initial cheap review unless authorized. Do not modify device settings, install apps, commit, push, tag, or publish without the task's authorization. For an implementation task, respect the user's current Git boundary and run the documented relevant gates before declaring completion.

## Handoff exercise

Use a clean checkout with the Go version in `go.mod` and the documented Linux graphics prerequisites. Install website dependencies with `npm ci --ignore-scripts` in `frontend`, then run the validation targets in [testing.md](testing.md). Make one small public-API example change, identify its owning package without author assistance, and explain one regression failure before fixing it. Record commands, host, commit, and anything missing from the guides. The current ledger keeps this human exercise open until a teammate supplies that evidence.
