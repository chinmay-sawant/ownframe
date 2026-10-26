# Phase 02: syntax

Goal: tokenize every displayed line in Go before `SetData`, then print plain
spans. The page never parses a language and never receives trusted HTML.

## Where it lives

- `syntax.go` has the segment model, the language map, and the shared C-family
  lexer for Go, Python, and JavaScript.
- `syntax_markup.go` has CSS, JSON, HTML, YAML, and shell.
- `syntax_markdown.go` has Markdown blocks and inline spans.
- `components/base.css` maps the token classes to colours.
- `syntax_test.go` covers detection, tokens, and text preservation.

## Design

1. A token is `segment{Text, Class}`. The template prints
   `<span class="t-{{.Class}}">{{.Text}}</span>`. html/template escapes the
   text, so a source file cannot inject markup.
2. A whole file goes through one pass. That keeps block comments, Python triple
   quotes, and Markdown fences open across lines.
3. A diff line goes through `highlightLine` on its own, because unified hunks
   interleave lines from different sides of a file.
4. The shared lexer tracks:
   - block comments and multiline strings,
   - the identifier after `func`, `def`, `class`, `function`, `const`, `let`,
     `var`, and `type`,
   - Python decorators,
   - numbers with bases, exponents, and suffixes.
5. Markdown handles headings, front matter, fences, block quotes, lists,
   tables, inline code, bold, italic, and links.
6. Colour only. No italic, because fake oblique pushes the page to the bitmap
   fallback, and the fallback refuses some text shaping options.
7. JavaScript is a supported language with no sample file. The bundled
   workspace carries no script, per the request; `syntax_test.go` is the proof
   that `.js` and `.ts` files highlight when opened with `-root`.

## Checklist

- [x] Language detection for Go, Python, JavaScript, TypeScript, Markdown,
      CSS, JSON, HTML, YAML, shell, and plain text.
- [x] Shared lexer for Go, Python, and JavaScript.
- [x] Python triple-quoted strings that span lines.
- [x] Go receiver methods and `type` names.
- [x] Markdown headings, fences, emphasis, links, and lists.
- [x] CSS selectors, properties, values, and custom properties.
- [x] JSON keys, strings, numbers, and literals.
- [x] HTML tags, attributes, values, and comments.
- [x] YAML keys, comments, anchors, and scalars.
- [x] Shell keywords, variables, strings, and comments.
- [x] Token classes in the stylesheet.
- [x] Text preservation tests for every language.
- [x] A test that the sample carries no script file.
- [-] Add a JavaScript sample under `testdata` only if a reviewer wants a
      visible demo; it must stay out of the embedded workspace. Reconsider if
      the catalog needs a visible JavaScript card.

## Gate

`TestJavaScriptTokens`, `TestPythonTripleQuotedString`,
`TestGoFunctionReceivers`, `TestMarkdownStructure`, and `TestMarkupTokens`
pass, and `TestHighlightLineAlwaysHasText` covers every language key.

## Risks

- A new language keyword set can steal a name that the surrounding code uses.
  The name-after-keyword state is the fragile part; add a test per language.
- An unsupported font feature or a fake italic can drop the whole page to the
  bitmap painter. Keep the tokens to colour and weight.
- Markdown inline rules are heuristics. A line with unmatched markers should
  stay plain, and the text preservation test catches dropped runes.
