# Template content, HTML rewriting, and injection risks

## Runtime template loading

Location: `internal/page/page_nav.go`

`Load` and `Route` parse caller supplied HTML into a Go `html/template` with `template.New("page").Parse(html)` and execute it at redraw with `p.data`. This is expected for trusted author content, but it makes the `Page.Data` struct readable by any template that can execute.

### Vulnerability

- Severity: Medium
- Status: Open
- Boundary: trusted application templates versus runtime content

Any caller that loads remote or user controlled HTML can expose data through template expressions such as `{{.Secret}}` or `{{range .}}{{.}}{{end}}`. Go template execution also allows calls to exported zero argument methods on data objects.

### Valid scenario

```text
page.Load(ctx, trustedAuthorTemplate, nil)
```

The template comes from the shipped application. Contextual escaping protects values rendered in HTML, attributes, and JavaScript strings.

### Invalid scenario

```text
page.Load(ctx, networkFetchedTemplate, nil)
```

The template may read fields from `Page.Data` and render them. ownframe cannot distinguish author content from attacker content at parse time.

### Recommended fixes

1. Document that `Load`, `Route`, and `HTML` accept trusted template source only.
2. Add an option that parses plain HTML without template execution.
3. Provide examples that validate or reject runtime template requests before they reach `Page.Load`.

## Contextual escaping through Go templates

Location: `internal/page/page_draw.go`

Templates are executed with `p.tpl.Execute(&body, p.data)`, so write contexts receive normal `html/template` escaping. Form rewriting occurs after execution and preserves the escaped output. `rewriteButtons`, control scans, and emoji replacement run on the rendered source rather than on raw template source.

### Valid scenario

Untrusted user text rendered as `{{.Name}}` in a paragraph or attribute exits as escaped HTML.

### Invalid scenario to test

Malformed tags or escaped quote sequences could cause the post execution rewriters to rebuild an attribute boundary. Adversarial tests should cover tags such as `<input value="...">` containing entities, template escaped quotes, and nested quotes after `buttonFromInput` and emoji replacement.

### Status

No bypass found in the audit. Fuzzing is still needed.

## Dangerous outbound links

Location: `internal/page/page_link.go`

`parseOutboundLink` allows only `http`, `https`, and `mailto`. Other schemes, including `javascript`, return `ErrUnsupportedLinkScheme`, so `Handlers.OpenLink` does not receive dangerous URIs.

### Valid scenario

`<a href="https://example.com/path">` parses and is delivered to `OpenLink`.

### Invalid scenario

`<a href="javascript:alert(1)">` is rejected. `mailto:` remains allowed by design and can open a mail client outside the app trust domain.

## Configured deep link schemes

Location: `internal/page/page_deeplink_schemes.go`, `internal/page/page_deeplink.go`

`acceptedDeepLinkSchemes` rejects `javascript`, `data`, `vbscript`, `file`, `blob`, and `about` even when an app explicitly allowlists them.

### Valid scenario

`Config{DeepLinkSchemes: []string{"myapp"}}` dispatches only `myapp:` links.

### Invalid scenario

`Config{DeepLinkSchemes: []string{"javascript"}}` returns `ErrBadLinkScheme` during construction.

## Claimed malicious page cannot run script

ownframe renders to an Ebiten display list or PNG. It is not a browser. Template content cannot execute JavaScript inside `Run`. It can still expose data or prepare unsafe links for apps that handle them incorrectly.

### Valid scenario

A browser preview served by `Serve` receives a static PNG and generated click areas, not executable HTML from the template.

### Invalid scenario

A click handler that calls an unsafe `OpenLink` with `Box.Action` or user controlled data can still leave the trust boundary set by ownframe.

## Theme CSS

Location: `internal/page/theme.go`, `internal/page/print_source.go`, `internal/render/state.go`

ownframe accepts complete CSS strings through `Config.Theme`, `Config.ThemeFile`, and `Page.SetTheme`. It does not parse CSS for allowlisted rules or property limits. Theme resources resolve only through `Page.SetImage`, so `url(...)` performs no network fetch.

### Vulnerability

- Severity: Medium
- Status: Partially fixed

`injectTheme` now escapes theme content with `html.EscapeString` before inserting it into `<style>` in generated source. Plain theme rendering still passes CSS to the engine unsanitized. Because ownframe treats theme content as trusted application code, this risk appears when apps assemble themes from untrusted input.

### Valid scenario

An app authored CSS theme with no external imports. `url(logo)` only paints if the same name was registered with `SetImage`.

### Invalid scenario

An app stores a user chosen CSS string and calls `SetTheme` with it. A style can alter layout, hide controls, or cover clickable areas even though script cannot run.

### Recommended fixes

1. Keep the `injectTheme` escaping change and add a test with `</style>` inside a CSS string.
2. Add an optional size cap in `parseTheme`.
3. Document that themes are trusted application content.