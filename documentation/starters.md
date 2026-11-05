# Starter examples

The homepage has three examples. Their names are Hello World, Split with friends, and Builder basics. The template versions pair an HTML file with a Go click handler.

The previews are WebAssembly and are not 1.0 hosts. 1.0 hosts are desktop `Run` on Linux, Windows, and macOS, and Android `BindMobile`.

The homepage can switch Hello World and Builder basics to a builder-pattern program. Split with friends stays a template. The share chart is SVG, and package `ui` does not emit `svg` or `path`.

## Hello World

The template is `examples/frontend/hello/template.html`. The Say hello button sends `greet`. The handler sets the message to "Hello from Go!".

[Hello World preview](frontend/public/demos/hello/index.html)

The builder-pattern program is `examples/frontend/hello-builder`. From the repo root, run `go run ./examples/frontend/hello-builder`.

## Split with friends

The template is `examples/frontend/budget/template.html`. Clicks switch among Split, Settle up, and Activity, toggle friends, add a friend, split the bill, choose who paid, and save the split.

[Split with friends preview](frontend/public/demos/expense-tracker/index.html)

## Builder basics

The template is `examples/frontend/builder/template.html`. The button counts clicks. The Bold toggle and the Notifications switch flip host state. Each click renders the card again.

[Builder basics preview](frontend/public/demos/builder/index.html)

The builder-pattern program is `examples/frontend/builder-mount`. From the repo root, run `go run ./examples/frontend/builder-mount`.

## UI builder

The gallery source is `examples/builder/ui-builder`. The gallery is the component set running in one window. From the repo root, run `go run ./examples/builder/ui-builder`.

[UI builder preview](frontend/public/demos/ui-builder/index.html)

Call patterns are in [Builder patterns](builder.md). Constructors are in [Supported components](components.md).
