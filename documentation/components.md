# Supported components

Package `components/chadcn` has 110 constructors. Each one returns `*ui.Element`. `FilterSearchItems` returns search rows, and `DialogFocus` is a focus scope. Neither is in that count.

The host owns open, selected, and checked state. Pass that state back into the constructor and rebuild with `View.Update`. Text inputs keep the live value on `Page.Form`. The value you pass to the constructor is the initial value.

## Quick start

`chadcn.Dark` or `chadcn.Light` sets the `--of-*` tokens. Append `chadcn.Stylesheet` for component paint and state. `config` is an `ownframe.Config`. Leave `HTML` and `File` empty, and set `Theme` to the sheet. Handle `err` before using `view`.

```go
sheet := chadcn.Dark + chadcn.Stylesheet
config.Theme = sheet
view, err := ui.Mount(config, chadcn.Screen(
	chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Deployments"),
			chadcn.CardDescription("Three environments"),
		),
		chadcn.CardContent(
			chadcn.Button("Deploy", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}),
		),
	),
))
```

## Typography

| Constructor | Role |
|---|---|
| **Blockquote** | Blockquote marks quoted prose. |
| **H1** | H1 is the largest document heading. |
| **H2** | H2 is a major section heading. |
| **H3** | H3 is a subsection heading. |
| **H4** | H4 is a minor heading. |
| **InlineCode** | InlineCode marks a short code span inside prose. |
| **Kbd** | Kbd displays one keyboard key or shortcut. |
| **Lead** | Lead is the larger introductory paragraph. |
| **List** | List is an unordered list with one item per string. |
| **Muted** | Muted is small secondary text. |
| **P** | P is a body paragraph. |

## Action

| Constructor | Role |
|---|---|
| **Button** | Button builds an HTML button with a label, an optional leading icon, and a loading state. |
| **ButtonGroup** | ButtonGroup joins its button children into one segmented control. |
| **Toggle** | Toggle builds a two-state button. |
| **ToggleGroup** | ToggleGroup builds a set of toggle buttons that share a selection. |

## Form

| Constructor | Role |
|---|---|
| **Checkbox** | Checkbox is a controlled checkbox button with an accessible checked state. |
| **CheckboxWithOptions** | CheckboxWithOptions builds a Checkbox with a description and disabled state. |
| **Field** | Field stacks a label, one control, and its messages in a column. |
| **FieldDescription** | FieldDescription returns the muted help line under a control. |
| **FieldError** | FieldError returns a validation message. |
| **FieldLabel** | FieldLabel returns the label part of a field. |
| **Input** | Input builds a single-line text field with an initial value. |
| **InputGroup** | InputGroup wraps one text input with optional prefix, suffix, and addon slots. |
| **InputOTP** | InputOTP renders one input per digit in a shared bordered row. |
| **InputWithOptions** | InputWithOptions builds an Input with a type, size, placeholder, and state. |
| **Label** | Label names a form control. |
| **NativeSelect** | NativeSelect builds the platform select and marks the option that matches selected. |
| **Questionnaire** | Questionnaire builds questions with single-choice answers and a progress line. |
| **RadioGroup** | RadioGroup renders one radio button per item, with radiogroup semantics. |
| **SearchField** | SearchField builds an accessible search control. |
| **Slider** | Slider renders a range track, the filled portion, and a thumb. |
| **Switch** | Switch is a controlled switch. |
| **SwitchWithOptions** | SwitchWithOptions builds a Switch with a disabled state. |
| **Textarea** | Textarea builds a multi-line field with an initial value. |
| **TextareaWithOptions** | TextareaWithOptions builds a Textarea with a row count, placeholder, and state. |
| **TextField** | TextField creates a labeled text input. |

## Selection

| Constructor | Role |
|---|---|
| **Calendar** | Calendar renders a month header, a weekday row, and a fixed six-by-seven day grid. |
| **Choice** | Choice is a selectable, labeled option with a secondary description. |
| **Combobox** | Combobox renders a text input with a chevron button and, when Open, a filtered listbox. |
| **DatePicker** | DatePicker shows the selected date or a placeholder, and a calendar panel when Open. |
| **FilterBar** | FilterBar builds a legend and pill row. |
| **Select** | Select renders a trigger button and, when Open, its listbox panel. |

## Overlay

| Constructor | Role |
|---|---|
| **AlertDialog** | AlertDialog builds a modal confirmation surface with cancel and confirm controls. |
| **Command** | Command builds a command palette with search, groups, one active row, and an empty state. |
| **ContextMenu** | ContextMenu renders a menu at the host pointer coordinates. |
| **Dialog** | Dialog builds a centered modal surface. |
| **Drawer** | Drawer builds a bottom modal panel with a grab handle and an action footer. |
| **DropdownMenu** | DropdownMenu renders a host-supplied trigger and, when Open is true, an anchored menu. |
| **HoverCard** | HoverCard renders a trigger and, when Open is true, an anchored card holding the content. |
| **Popover** | Popover renders a trigger and, when Open is true, an anchored panel holding the children. |
| **Search** | Search opens a backdrop card and filters its prefilled rows as Query changes. |
| **Sheet** | Sheet builds a modal panel anchored to one window edge over a backdrop. |
| **Tooltip** | Tooltip renders a trigger and, when Open is true, a label above, below, or beside it. |

## Navigation

| Constructor | Role |
|---|---|
| **Breadcrumb** | Breadcrumb builds a navigational trail. |
| **Menubar** | Menubar renders menu triggers and opens the menu whose ID matches Open. |
| **NavigationMenu** | NavigationMenu builds navigation triggers with one open content panel. |
| **Pagination** | Pagination builds previous, next, and numbered page controls. |
| **Sidebar** | Sidebar builds the vertical navigation column. |
| **SidebarNavItem** | SidebarNavItem builds one navigation row. |
| **SidebarSection** | SidebarSection groups navigation rows under a label. |

## Disclosure

| Constructor | Role |
|---|---|
| **Accordion** | Accordion builds a stack of disclosure items. |
| **Collapsible** | Collapsible renders a host-supplied trigger and, when Open is true, its content. |
| **Tabs** | Tabs builds a tab list and the active panel. |

## Feedback

| Constructor | Role |
|---|---|
| **Alert** | Alert builds a bordered status panel with `role="alert"`. |
| **Empty** | Empty fills a region with a centered icon, title, description, and actions. |
| **Notice** | Notice displays a heading and supporting text on a themed inset surface. |
| **Progress** | Progress displays a percentage, clamped to 0..100, with progressbar semantics. |
| **Skeleton** | Skeleton is a pulsing placeholder for content that is still loading. |
| **Spinner** | Spinner is an indeterminate loading ring driven by a CSS keyframe. |
| **Toast** | Toast builds a notification in the bottom-right corner of its containing block. |

## Data display

| Constructor | Role |
|---|---|
| **Avatar** | Avatar shows a page image, initials, or a fallback character inside a circle. |
| **Badge** | Badge is a short non-interactive label. |
| **Carousel** | Carousel shows one slide with previous, next, and dot indicators. |
| **CodeBlock** | CodeBlock builds a code frame with a filename header and a copy label above a preformatted body. |
| **DataTable** | DataTable builds a bordered table with a header, rows, and an empty state. |
| **Item** | Item lays out media, text content, and trailing actions on one row. |
| **Marker** | Marker builds a small status marker with a ring icon, a label, and optional detail text. |
| **StatCard** | StatCard builds a KPI row cell for a dashboard header strip. |
| **Table** | Table is the root of a simple data table. |
| **TableBody** | TableBody groups the data rows. |
| **TableCaption** | TableCaption is the table summary below the rows. |
| **TableCell** | TableCell is a data cell. |
| **TableFooter** | TableFooter groups summary rows below the body. |
| **TableHead** | TableHead is a column heading cell. |
| **TableHeader** | TableHeader groups the column heading row. |
| **TableRow** | TableRow is one table row. |
| **Trace** | Trace builds a list of timed steps with a status ring and a CSS timing bar. |

## Layout

| Constructor | Role |
|---|---|
| **AspectRatio** | AspectRatio keeps its children inside a box with a fixed width-to-height ratio. |
| **Card** | Card groups related content on a bordered surface. |
| **CardContent** | CardContent holds the main body of a card. |
| **CardDescription** | CardDescription is the secondary text below a CardTitle. |
| **CardFooter** | CardFooter holds actions below a top border. |
| **CardHeader** | CardHeader stacks the title and description above the card content. |
| **CardTitle** | CardTitle is the card heading. |
| **Direction** | Direction wraps children with a dir attribute so a subtree can override the page direction. |
| **Panel** | Panel is a themed surface without the card's border or padding. |
| **Resizable** | Resizable builds a flex split with a focusable, labeled divider. |
| **Screen** | Screen fills the viewport with the global theme's canvas and foreground. |
| **ScrollArea** | ScrollArea clips a scrolling child list and overlays a themed scrollbar. |
| **Scrollbar** | Scrollbar overlays a vertical scrollbar on a positioned, clipped viewport. |
| **Separator** | Separator draws a hairline divider with separator semantics. |

## Chart

| Constructor | Role |
|---|---|
| **AreaChart** | AreaChart builds a div trend plot with a legend, one dot per series, and a hatched band. |
| **Chart** | Chart builds a bar-chart frame with totals, axis labels, and a legend. |
| **HexCellElement** | HexCellElement builds one honeycomb cell from three solid parts. |
| **HexGrid** | HexGrid builds a grid of pointy-top hexagon cells colored by tone. |
| **StatusSegments** | StatusSegments builds proportional status bars with a count legend. |

## Chat

| Constructor | Role |
|---|---|
| **Attachment** | Attachment builds a file tile with an icon, the file name, metadata, and an optional upload bar. |
| **Bubble** | Bubble builds a rounded message bubble around its children. |
| **Message** | Message builds one chat row with an avatar, a name and time line, and the content. |
| **MessageScroller** | MessageScroller builds a scrollable message viewport with a top fade. |

## Gallery

Run `go run ./examples/builder/ui-builder` from the repo root, and open the wasm preview, which is not a 1.0 host, at [UI builder preview](frontend/public/demos/ui-builder/index.html).

`examples/builder/loglens` is the dashboard that uses the set.

## Limits

- No JavaScript and no canvas.
- Charts do not fetch or sort by themselves.
- The set does not navigate with `href`.
- Overlays do not trap focus unless the host uses `DialogFocus`.
