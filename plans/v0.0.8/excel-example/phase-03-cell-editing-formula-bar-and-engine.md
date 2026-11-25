# v0.0.8 - Phase 03: Cell editing, formula bar, and formula engine

> **Parent:** [README.md](README.md)
> **Status:** Open (0 complete, 0 partial, 16 open)
> **Estimated effort:** 4-5 days

---

## Overview

Phase 03 turns the grid into a working calculation engine. Users can type directly into any active cell or into the synchronized formula bar, click other cells while editing a formula to insert color-coded references, pick functions from an autocomplete popup, and see dependent cells recalculate immediately across sheets.

## Executive summary

- Implement inline cell editing (`F2`, double-click, or direct typing) synchronized two-way with the formula bar (`fx`, Cancel `×`, Commit `✓`).
- Build a pure-Go formula lexer, recursive-descent AST parser, and evaluator supporting arithmetic, comparisons, string concatenation (`&`), cell and range references, cross-sheet references (`Sheet2!B4`, `'Sales Q1-Q4'!C2:C20`), and unary/binary precedence.
- Build a dependency graph with incremental topological recalculation and cycle detection (`#CIRC!`).
- Ship 35+ standard Excel functions across Math/Aggregation, Logical, Lookup/Reference, Text, and Date/Time categories.

---

## Phase 3: Cell editing, formula bar, and formula engine

### 3.1 Formula bar and inline cell editor

- [ ] `EXL-29` Build `examples/builder/excel/formula_bar.go` with the Name Box (96 px), vertical separator, Cancel (`×`), Commit (`✓`), Insert Function (`fx`) buttons, and the expanding formula text input synced to the active cell's `Raw` string. Proof: `TestFormulaBarSync` selects value and formula cells and verifies `Raw` text appears in the formula input.
- [ ] `EXL-30` Implement inline cell editing in `grid.go` and `app.go`: typing a printable character replaces the active cell and enters edit mode, `F2` or double-click enters edit mode preserving existing content, `Enter` commits and moves down, `Tab` commits and moves right, `Escape` cancels and restores prior `Raw`, and `Delete`/`Backspace` clears the selected range without entering edit mode. Proof: `TestInlineEditLifecycle` verifies all five commit/cancel/clear paths.
- [ ] `EXL-31` Implement literal type inference on commit in `cell.go`: parse plain numbers (`42`, `-3.14`), percentages (`12.5%` -> `0.125` with percent format), currency (`$1,250.00` -> `1250` with currency format), booleans (`TRUE`/`FALSE`), ISO/short dates (`2026-11-25`), leading apostrophe forced text (`'00123`), and leading `=` formulas. Proof: table-driven tests in `cell_test.go` covering every input kind.
- [ ] `EXL-32` Support point-and-click cell/range reference insertion while editing a formula (when the caret follows `=`, `(`, `,`, `:`, or an operator) and `F4` cycling through `$A$1`, `A$1`, `$A1`, `A1`. Proof: `TestFormulaClickInsertAndF4` simulates editing `=SUM(`, clicking/dragging `B2:B5`, and pressing `F4`.
- [ ] `EXL-33` Paint color-coded reference highlights on the grid using the 6-color `--xl-ref-1..6` palette from `README.md` whenever the user is editing a formula, matching each parsed cell/range token in the formula bar. Proof: `TestFormulaRefHighlights` verifies highlight overlays render for `=A1+B2:B4`.
- [ ] `EXL-34` Add a function autocomplete popover anchored beneath the active cell or formula bar when typing `=SU...`, showing matching function names and signatures (`SUM(number1, [number2], ...)`), selectable with Up/Down + Tab/Enter. Proof: `TestFormulaAutocomplete` types `=AV`, selects `AVERAGE`, and confirms `=AVERAGE(` insertion.

### 3.2 Formula lexer, parser, and evaluator

- [ ] `EXL-35` Implement `examples/builder/excel/formula_lexer.go` tokenizing numbers (including scientific notation `1.5e-3`), double-quoted strings with `""` escaping, booleans, cell/range refs, quoted and unquoted sheet prefixes (`'Sales Q1-Q4'!A1:B4`), identifiers, operators (`+`, `-`, `*`, `/`, `^`, `&`, `=`, `<>`, `<`, `<=`, `>`, `>=`, `%`), commas, and parentheses. Proof: `formula_lexer_test.go` covers valid formulas, edge cases, and malformed tokens.
- [ ] `EXL-36` Implement `examples/builder/excel/formula_parser.go` as a recursive-descent parser with exact Excel operator precedence (reference `:`, unary `-`/`+`, `%`, `^`, `*`/`/`, `+`/`-`, `&`, comparisons) producing an AST and collecting all referenced `CellRef`/`RangeRef` dependencies. Proof: `formula_parser_test.go` tests precedence (`-2^2`, `1+2*3`, `"a"&1+2`, nested calls, range unions) and syntax error recovery.
- [ ] `EXL-37` Implement `examples/builder/excel/formula_eval.go` evaluating AST nodes against a `Workbook`, handling type coercion (numeric strings, booleans in arithmetic, empty cells as 0 or `""`), error propagation, division-by-zero (`#DIV/0!`), unknown functions (`#NAME?`), invalid values (`#VALUE!`), and missing references (`#REF!`). Proof: `formula_eval_test.go` verifies all operators, coercions, and error propagation rules.

### 3.3 Dependency graph and incremental recalculation

- [ ] `EXL-38` Implement `examples/builder/excel/depgraph.go` maintaining forward (`dependsOn`) and reverse (`dependents`) edges per `(sheet, cell)`, performing topological sort on edit so only dirty downstream cells recalculate, and detecting direct and indirect cycles (`A1 -> B1 -> A1`) by marking involved cells `#CIRC!`. Proof: `depgraph_test.go` tests deep chains, diamond dependencies, cross-sheet updates, and circular references without infinite recursion.
- [ ] `EXL-39` Implement formula reference rewriting (`ShiftFormula(raw, dCol, dRow)`) so relative references shift while `$`-locked components stay fixed (used later by AutoFill, Copy/Paste, and Insert/Delete Row/Col). Proof: `formula_shift_test.go` covers `$A$1`, `$A1`, `A$1`, `A1:B5`, and cross-sheet refs.

### 3.4 Core Excel function library

- [ ] `EXL-40` Implement Math and Aggregation functions in `examples/builder/excel/formula_funcs.go`: `SUM`, `AVERAGE`, `MIN`, `MAX`, `COUNT`, `COUNTA`, `COUNTBLANK`, `ROUND`, `ROUNDUP`, `ROUNDDOWN`, `ABS`, `MOD`, `POWER`, `SQRT`, `INT`, `SUMIF`, `COUNTIF`, `SUMIFS`. Proof: unit tests in `formula_funcs_test.go` matching Excel results for ranges, empty cells, text in ranges, and criteria strings (`">100"`, `"North"`).
- [ ] `EXL-41` Implement Logical functions in `formula_funcs.go`: `IF`, `IFS`, `AND`, `OR`, `NOT`, `XOR`, `IFERROR`, `IFNA`, `ISBLANK`, `ISNUMBER`, `ISTEXT`, `ISERROR`. Proof: unit tests in `formula_funcs_test.go` including short-circuit behavior in `IF` and `IFERROR`.
- [ ] `EXL-42` Implement Lookup and Reference functions in `formula_funcs.go`: `VLOOKUP` (exact and approximate match), `HLOOKUP`, `XLOOKUP` (with `if_not_found`), `INDEX`, `MATCH` (match types -1, 0, 1), `ROW`, `COLUMN`, `ROWS`, `COLUMNS`. Proof: unit tests in `formula_funcs_test.go` verifying `#N/A` and `#REF!` edge cases.
- [ ] `EXL-43` Implement Text and Date/Time functions in `formula_funcs.go`: `CONCAT`, `TEXTJOIN`, `LEFT`, `RIGHT`, `MID`, `LEN`, `TRIM`, `UPPER`, `LOWER`, `PROPER`, `SUBSTITUTE`, `FIND`, `TEXT`, `TODAY`, `NOW`, `DATE`, `YEAR`, `MONTH`, `DAY`, `DATEDIF`. Proof: unit tests in `formula_funcs_test.go` with deterministic clock injection for `TODAY`/`NOW`.

### 3.5 Phase gate

- [ ] `EXL-44` Run Phase 03 closure gates: `gofmt`, `go vet -p 1 ./examples/builder/excel`, `go test -race -v ./examples/builder/excel/...`, and `make lint`, verifying all seeded formulas on `Summary`, `Sales Q1-Q4`, and `Expenses` evaluate with zero unexpected errors. Proof: command outputs recorded beside this row.

---

## Dependencies

- Requires Phase 02 (`EXL-21`, `EXL-22`, `EXL-24`, `EXL-27`).
- `EXL-35` and `EXL-36` (lexer and parser) must land before `EXL-33` (reference highlights), `EXL-37` (evaluator), and `EXL-38` (dependency graph).
- `EXL-38` (recalculation engine) must land before Phase 04 conditional formatting and Phase 05 sort/filter/status aggregations.
