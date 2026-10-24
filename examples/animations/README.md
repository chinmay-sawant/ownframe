# CSS Animations Knowledge Lab & Showcase

An interactive, single-component HTML and CSS animation showcase and knowledge challenge.
It demonstrates all 27 standard and WebKit animation properties defined in the Blinkless compatibility matrix (`documentation/compatibility-matrix.md`).

## What it does

The application presents an interactive knowledge challenge covering CSS animation capabilities, paired with live, visual animated stages rendered directly by the OwnFrame layout and paint pipeline:

1. **Timing Functions & Curves**: Demonstrates `ease-in-out`, `linear`, `cubic-bezier(0.68, -0.55, 0.27, 1.55)` (overshoot), and `steps(5, jump-end)` (stepped clock motion).
2. **Direction & Cycles**: Demonstrates `animation-direction` (`normal`, `reverse`, `alternate`) and `animation-iteration-count`.
3. **Delays & Offsets**: Demonstrates positive delays, zero delay, and negative start delays (`animation-delay: -1.0s`, starting mid-cycle).
4. **Fill Modes**: Demonstrates `none`, `forwards` (freezing at 100%), and `both`.
5. **WebKit Prefixes**: Demonstrates `-webkit-animation` vendor aliases and their 1:1 normalization in the pipeline.
6. **Advanced Properties**: Demonstrates `animation-composition`, `animation-timeline`, `animation-range`, and `animation-trigger`.

## Running the example

Run as a native desktop window:

```sh
go run ./examples/animations
```

Run as a browser-rendered web preview:

```sh
go run ./examples/animations -web
```

Optionally specify a custom listen address with `-addr`:

```sh
go run ./examples/animations -web -addr=127.0.0.1:8142
```

## Running tests

```sh
go test -v -p 1 ./examples/animations/...
```
