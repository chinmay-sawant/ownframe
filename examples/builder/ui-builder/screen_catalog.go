package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// catalogEntry is one card in the ChadCN catalog. Most entries render a
// representative state; the Tabs, Slider, Select, Checkbox, and Switch entries
// are live and read builderApp catalog state. Text inputs keep stable Page.Form
// IDs so the page carries user edits. Display-only entries (Chart, AreaChart,
// HexGrid, StatCard, StatusSegments) stay static: they expose no action API.
type catalogEntry struct {
	name  string
	build func(a *builderApp) *ui.Element
}

const (
	catalogColumns    = 2
	catalogOTPLength  = 4
	catalogPageCount  = 3
	catalogProgress   = 60
	catalogSliderMax  = 100
	catalogSliderStep = 10
	catalogTraceStart = 40
	catalogTraceWidth = 40
	catalogTraceShort = 20
)

func (a *builderApp) catalogPanel() *ui.Element {
	rows := a.catalogWindowedRows()
	return ui.Section().Flex(ui.Number(1)).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).
		Children(
			ui.Div().Children(
				a.panelHeading("ChadCN catalog"),
				ui.P().Text("Every constructor in the set, shown in a representative state. Open overlays and focus behavior are covered by the dialog tab and the package tests.").
					Margin(ui.Edges(ui.Px(7), ui.Px(0), ui.Px(0))).Color(theme.Muted()).FontSize(ui.Px(12)),
			),
			ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(12)).Children(rows...),
		)
}

func catalogCard(a *builderApp, entry catalogEntry) *ui.Element {
	body := safeCatalogElement(func() *ui.Element { return entry.build(a) })
	return widgetShell(entry.name, "ChadCN constructor", body)
}

// safeCatalogElement keeps a card visible when a closed overlay returns nil.
func safeCatalogElement(build func() *ui.Element) *ui.Element {
	element := build()
	if element == nil {
		return ui.P().Text("Closed until the host opens it.").Margin(ui.Px(0)).FontSize(ui.Px(12)).Color(theme.Muted())
	}
	return element
}

func catalogEntries() []catalogEntry {
	return []catalogEntry{
		{"Accordion", func(*builderApp) *ui.Element {
			return chadcn.Accordion(chadcn.AccordionOptions{
				Items: []chadcn.AccordionItem{{ID: "one", Title: "Is it accessible?", Content: ui.Text("Yes. It follows the button pattern.")}},
				Open:  []string{"one"},
			})
		}},
		{"Alert", func(*builderApp) *ui.Element {
			return chadcn.Alert(chadcn.AlertOptions{Title: "Heads up", Description: "The component set is ready."})
		}},
		{"Alert Dialog", func(*builderApp) *ui.Element {
			return chadcn.AlertDialog(chadcn.AlertDialogOptions{Title: "Delete file?", Description: "This cannot be undone.", Trigger: chadcn.Button("Open alert dialog", chadcn.ButtonOptions{})})
		}},
		{"Area Chart", func(*builderApp) *ui.Element {
			return chadcn.AreaChart(chadcn.AreaChartOptions{
				Title:       "Requests trend",
				Description: "Opened versus resolved across the week.",
				Series: []chadcn.ChartSeries{
					{Label: "Opened", Values: []float64{12, 18, 15, 22, 26, 19, 24}},
					{Label: "Resolved", Values: []float64{8, 12, 11, 16, 18, 15, 21}},
				},
				Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
			})
		}},
		{"Aspect Ratio", func(*builderApp) *ui.Element {
			return chadcn.AspectRatio(chadcn.AspectRatioOptions{}, ui.Text("16:9")).Width(ui.Percent(100)).BackgroundColor(theme.Control())
		}},
		{"Attachment", func(*builderApp) *ui.Element { return chadcn.Attachment(chadcn.AttachmentOptions{Name: "report.pdf"}) }},
		{"Avatar", func(*builderApp) *ui.Element { return chadcn.Avatar(chadcn.AvatarOptions{Initials: "AK"}) }},
		{"Badge", func(*builderApp) *ui.Element { return chadcn.Badge("Badge") }},
		{"Breadcrumb", func(*builderApp) *ui.Element {
			return chadcn.Breadcrumb([]chadcn.BreadcrumbItem{{Label: "Home"}, {Label: "Library", Current: true}})
		}},
		{"Bubble", func(*builderApp) *ui.Element { return chadcn.Bubble(chadcn.BubbleOptions{}, ui.Text("Hello")) }},
		{"Button", func(*builderApp) *ui.Element {
			return chadcn.Button("Button", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled})
		}},
		{"Button Group", func(*builderApp) *ui.Element {
			return chadcn.ButtonGroup(chadcn.ButtonGroupOptions{},
				chadcn.Button("Left", chadcn.ButtonOptions{}),
				chadcn.Button("Right", chadcn.ButtonOptions{}))
		}},
		{"Calendar", func(*builderApp) *ui.Element {
			now := time.Now()
			return chadcn.Calendar(chadcn.CalendarOptions{Month: now, Selected: now})
		}},
		{"Card", func(*builderApp) *ui.Element {
			return chadcn.Card(chadcn.CardHeader(chadcn.CardTitle("Card")), chadcn.CardContent(ui.Text("Content")))
		}},
		{"Carousel", func(*builderApp) *ui.Element {
			return chadcn.Carousel(chadcn.CarouselOptions{Slides: []*ui.Element{ui.Text("Slide one"), ui.Text("Slide two")}})
		}},
		{"Chart", func(*builderApp) *ui.Element {
			return chadcn.Chart(chadcn.ChartOptions{
				Title:  "Visitors",
				Series: []chadcn.ChartSeries{{Label: "Desktop", Values: []float64{32, 48, 61, 44, 72}}},
				Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
			})
		}},
		{"Checkbox", func(a *builderApp) *ui.Element {
			return chadcn.Checkbox("Checkbox", a.catalogCheck).Attr("data-action", "catalog:check")
		}},
		{"Choice", func(*builderApp) *ui.Element {
			return chadcn.Choice("Nightly backups", "Keeps 14 days of snapshots.", true)
		}},
		{"Code Block", func(*builderApp) *ui.Element {
			return chadcn.CodeBlock(chadcn.CodeBlockOptions{Filename: "main.go", Code: "package main"})
		}},
		{"Collapsible", func(*builderApp) *ui.Element {
			return chadcn.Collapsible(chadcn.CollapsibleOptions{Open: true, Trigger: chadcn.Button("Toggle", chadcn.ButtonOptions{})}, ui.Text("Hidden content"))
		}},
		{"Combobox", func(*builderApp) *ui.Element {
			return chadcn.Combobox(chadcn.ComboboxOptions{Open: true, Items: []chadcn.ComboboxItem{{Value: "next", Label: "Next.js"}, {Value: "remix", Label: "Remix"}}})
		}},
		{"Command", func(*builderApp) *ui.Element {
			return chadcn.Command(chadcn.CommandOptions{Active: "calendar", Groups: []chadcn.CommandGroup{{Label: "Suggestions", Items: []chadcn.CommandItem{{Value: "calendar", Label: "Calendar"}}}}})
		}},
		{"Context Menu", func(*builderApp) *ui.Element {
			return chadcn.ContextMenu(chadcn.ContextMenuOptions{Open: true, Items: []chadcn.MenuItem{{Label: "Copy"}}})
		}},
		{"Data Table", func(*builderApp) *ui.Element {
			return chadcn.DataTable(chadcn.DataTableOptions{
				Columns: []chadcn.DataColumn{{Label: "Name"}, {Label: "Status"}},
				Rows:    [][]string{{"Ada", "Active"}, {"Linus", "Away"}},
			})
		}},
		{"Date Picker", func(*builderApp) *ui.Element { return chadcn.DatePicker(chadcn.DatePickerOptions{Open: true}) }},
		{"Dialog", func(*builderApp) *ui.Element {
			return chadcn.Dialog(chadcn.DialogOptions{Title: "Dialog", Trigger: chadcn.Button("Open dialog", chadcn.ButtonOptions{})})
		}},
		{"Direction", func(*builderApp) *ui.Element { return chadcn.Direction(chadcn.DirectionOptions{}, ui.Text("RTL")) }},
		{"Drawer", func(*builderApp) *ui.Element {
			return chadcn.Drawer(chadcn.DrawerOptions{Title: "Drawer", Trigger: chadcn.Button("Open drawer", chadcn.ButtonOptions{})})
		}},
		{"Dropdown Menu", func(*builderApp) *ui.Element {
			return chadcn.DropdownMenu(chadcn.DropdownMenuOptions{Open: true, Trigger: chadcn.Button("Open", chadcn.ButtonOptions{}), Items: []chadcn.MenuItem{{Label: "Profile"}}})
		}},
		{"Empty", func(*builderApp) *ui.Element {
			return chadcn.Empty(chadcn.EmptyOptions{Title: "No results", Description: "Try another search.", Actions: []*ui.Element{chadcn.Button("Clear", chadcn.ButtonOptions{})}})
		}},
		{"Field", func(*builderApp) *ui.Element {
			return chadcn.Field(chadcn.FieldOptions{Label: "Email", Control: chadcn.Input("catalog-email", "")})
		}},
		{"Filter Bar", func(*builderApp) *ui.Element {
			return chadcn.FilterBar(chadcn.FilterBarOptions{
				Label:  "Trend filters",
				Legend: []chadcn.FilterBarLegend{{Label: "Opened", Tone: 1}, {Label: "Resolved", Tone: 2}, {Label: "Backlog gap", Tone: 3}},
				Options: []chadcn.FilterBarOption{
					{Label: "Daily", Action: "catalog:filter:daily"},
					{Label: "Weekly", Active: true, Action: "catalog:filter:weekly"},
					{Label: "Monthly", Action: "catalog:filter:monthly"},
				},
			})
		}},
		{"Hex Grid", func(*builderApp) *ui.Element {
			return chadcn.HexGrid(chadcn.HexGridOptions{
				Label:   "Portfolio health",
				Columns: 8,
				Cells: []chadcn.HexCell{
					{Tone: 1}, {Tone: 1}, {Tone: 2}, {Tone: 2}, {Tone: 2}, {Tone: 0}, {Tone: 0}, {Tone: 1},
					{Tone: 1}, {Tone: 3}, {Tone: 3}, {Tone: 2}, {Tone: 4}, {Tone: 4}, {Tone: 0}, {Tone: 1},
					{Tone: 3}, {Tone: 3}, {Tone: 4}, {Tone: 4}, {Tone: 2}, {Tone: 1}, {Tone: 0}, {Tone: 0},
					{Tone: 0}, {Tone: 4}, {Tone: 4}, {Tone: 3}, {Tone: 3}, {Tone: 2}, {Tone: 0}, {Tone: 0},
				},
			})
		}},
		{"Hover Card", func(*builderApp) *ui.Element {
			return chadcn.HoverCard(chadcn.HoverCardOptions{Open: true}, chadcn.Button("Hover", chadcn.ButtonOptions{}), ui.Text("Preview"))
		}},
		{"Input", func(*builderApp) *ui.Element { return chadcn.Input("catalog-input", "Value") }},
		{"Input Group", func(*builderApp) *ui.Element {
			return chadcn.InputGroup(chadcn.InputGroupOptions{ID: "catalog-group", Label: "Handle", Value: "acme", Prefix: ui.Text("@")})
		}},
		{"Input OTP", func(*builderApp) *ui.Element {
			return chadcn.InputOTP(chadcn.InputOTPOptions{ID: "catalog-otp", Length: catalogOTPLength})
		}},
		{"Item", func(*builderApp) *ui.Element {
			return chadcn.Item(chadcn.ItemOptions{Media: ui.Text("A"), Title: "Ada Lovelace", Description: "First programmer", Actions: chadcn.Button("Open", chadcn.ButtonOptions{})})
		}},
		{"Kbd", func(*builderApp) *ui.Element { return chadcn.Kbd("⌘K") }},
		{"Label", func(*builderApp) *ui.Element { return chadcn.Label("Label") }},
		{"Marker", func(*builderApp) *ui.Element { return chadcn.Marker(chadcn.MarkerOptions{Label: "Deployed"}) }},
		{"Menubar", func(*builderApp) *ui.Element {
			return chadcn.Menubar(chadcn.MenubarOptions{Open: "file", Menus: []chadcn.MenubarMenu{{ID: "file", Label: "File", Items: []chadcn.MenuItem{{Label: "New"}}}}})
		}},
		{"Message", func(*builderApp) *ui.Element {
			return chadcn.Message(chadcn.MessageOptions{Name: "Ada", Time: "4:55 PM"})
		}},
		{"Message Scroller", func(*builderApp) *ui.Element { return chadcn.MessageScroller(chadcn.MessageScrollerOptions{}) }},
		{"Native Select", func(*builderApp) *ui.Element {
			return chadcn.NativeSelect("catalog-native", []chadcn.SelectValue{{Value: "a", Label: "Apple"}}, "a")
		}},
		{"Navigation Menu", func(*builderApp) *ui.Element {
			return chadcn.NavigationMenu(chadcn.NavigationMenuOptions{Open: "docs", Items: []chadcn.NavigationItem{{ID: "docs", Label: "Docs"}}})
		}},
		{"Notice", func(*builderApp) *ui.Element {
			return chadcn.Notice("Trial ends in 3 days", "Add a payment method to keep syncing.")
		}},
		{"Pagination", func(*builderApp) *ui.Element {
			return chadcn.Pagination(chadcn.PaginationOptions{Page: 1, TotalPages: catalogPageCount})
		}},
		{"Popover", func(*builderApp) *ui.Element {
			return chadcn.Popover(chadcn.PopoverOptions{Open: true, Trigger: chadcn.Button("Open", chadcn.ButtonOptions{})}, ui.Text("Popover content"))
		}},
		{"Progress", func(*builderApp) *ui.Element { return chadcn.Progress("Progress", catalogProgress) }},
		{"Questionnaire", func(*builderApp) *ui.Element {
			return chadcn.Questionnaire(chadcn.QuestionnaireOptions{
				Title: "Feedback",
				Questions: []chadcn.Question{{
					Label: "How was the setup?", Choices: []chadcn.ChoiceItem{{Value: "easy", Label: "Easy"}, {Value: "hard", Label: "Hard"}},
				}},
			})
		}},
		{"Radio Group", func(*builderApp) *ui.Element { return chadcn.RadioGroup(chadcn.RadioGroupOptions{}) }},
		{"Resizable", func(*builderApp) *ui.Element {
			return chadcn.Resizable(chadcn.ResizableOptions{First: ui.Text("One"), Second: ui.Text("Two")})
		}},
		{"Scroll Area", func(*builderApp) *ui.Element {
			return chadcn.ScrollArea(chadcn.ScrollAreaOptions{ID: "catalog-scroll", Height: ui.Px(120), Children: []*ui.Element{ui.Text("Row one"), ui.Text("Row two"), ui.Text("Row three")}})
		}},
		{"Scrollbar", func(*builderApp) *ui.Element {
			state := &chadcn.ScrollState{}
			state.SetRange(400, 120)
			return ui.Div().Position(ui.PositionRelative).Height(ui.Px(140)).Overflow(ui.OverflowHidden).Children(
				ui.P().Text("Scrollable region with an overlaid scrollbar.").Margin(ui.Px(0)).FontSize(ui.Px(12)),
				chadcn.Scrollbar(chadcn.ScrollbarOptions{ID: "catalog-scrollbar", ControlsID: "catalog-scroll", State: state, TrackPixels: 120}),
			)
		}},
		{"Search Field", func(*builderApp) *ui.Element {
			return chadcn.SearchField("catalog-search", "Search orders", "")
		}},
		{"Select", func(a *builderApp) *ui.Element {
			return chadcn.Select(chadcn.SelectOptions{
				ID:    "catalog-select",
				Open:  a.catalogSelectOpen,
				Value: a.catalogSelect,
				Items: []chadcn.SelectItem{{Value: "a", Label: "Apple"}},
				OnOpenChange: func(ctx context.Context, open bool) error {
					a.catalogSelectOpen = open
					return a.refresh(ctx)
				},
				OnSelect: func(ctx context.Context, value string) error {
					a.catalogSelect = value
					a.catalogSelectOpen = false
					a.notice = "Catalog select picked " + value + "."
					return a.refresh(ctx)
				},
			})
		}},
		{"Separator", func(*builderApp) *ui.Element { return chadcn.Separator(chadcn.SeparatorOptions{}) }},
		{"Sheet", func(*builderApp) *ui.Element {
			return chadcn.Sheet(chadcn.SheetOptions{Title: "Sheet", Trigger: chadcn.Button("Open sheet", chadcn.ButtonOptions{})})
		}},
		{"Sidebar", func(*builderApp) *ui.Element {
			return chadcn.Sidebar(chadcn.SidebarOptions{Open: true, Header: chadcn.Button("Acme", chadcn.ButtonOptions{}), Content: ui.Text("Workspace"), Footer: ui.Text("shadcn@example.com")})
		}},
		{"Sidebar Section", func(*builderApp) *ui.Element {
			return chadcn.SidebarSection("Operations",
				chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "Dashboard", Active: true, Action: "catalog:nav:dashboard"}),
				chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "Work orders", Badge: "137", Action: "catalog:nav:work-orders"}),
				chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "Maintenance", Action: "catalog:nav:maintenance"}),
			)
		}},
		{"Skeleton", func(*builderApp) *ui.Element { return chadcn.Skeleton(chadcn.SkeletonOptions{}) }},
		{"Slider", func(a *builderApp) *ui.Element {
			// Slider reports only its value, so the host owns the value and two
			// nudge buttons move it; the window has no slider drag gesture.
			return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).Children(
				chadcn.Slider(chadcn.SliderOptions{Label: "Volume", Value: a.catalogSlider}).ID("catalog-slider"),
				ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).Children(
					chadcn.Button("−", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:slider:dec"),
					ui.Text(fmt.Sprintf("%.0f", a.catalogSlider)).FontSize(ui.Px(12)).Color(theme.Muted()),
					chadcn.Button("+", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:slider:inc"),
				),
			)
		}},
		{"Spinner", func(*builderApp) *ui.Element { return chadcn.Spinner(chadcn.SpinnerOptions{}) }},
		{"Stat Card", func(*builderApp) *ui.Element {
			return chadcn.StatCard(chadcn.StatCardOptions{Label: "Open work orders", Value: "137", Delta: "23 urgent", Tone: chadcn.StatDeltaWarn})
		}},
		{"Status Segments", func(*builderApp) *ui.Element {
			return chadcn.StatusSegments(chadcn.StatusSegmentsOptions{
				Label:   "Work order status",
				Overdue: "Overdue work orders: 7",
				Segments: []chadcn.StatusSegment{
					{Label: "New", Value: 24, Count: "24", Tone: 1},
					{Label: "In progress", Value: 61, Count: "61", Tone: 3},
					{Label: "Scheduled", Value: 38, Count: "38", Tone: 2},
					{Label: "Completed", Value: 112, Count: "112", Tone: 4},
				},
			})
		}},
		{"Switch", func(a *builderApp) *ui.Element {
			return chadcn.Switch("Switch", a.catalogSwitch).Attr("data-action", "catalog:switch")
		}},
		{"Table", func(*builderApp) *ui.Element {
			return chadcn.Table(chadcn.TableHeader(chadcn.TableRow(chadcn.TableHead("Name"))), chadcn.TableBody(chadcn.TableRow(chadcn.TableCell("Ada"))))
		}},
		{"Tabs", func(a *builderApp) *ui.Element {
			return chadcn.Tabs(chadcn.TabsOptions{
				Active: a.catalogTabSel,
				Items: []chadcn.TabItem{
					{ID: "account", Label: "Account", Content: ui.Text("Account settings")},
					{ID: "password", Label: "Password", Content: ui.Text("Password settings")},
				},
				OnSelect: func(ctx context.Context, id string) error {
					a.catalogTabSel = id
					a.notice = "Catalog tab " + id + " active."
					return a.refresh(ctx)
				},
			})
		}},
		{"Textarea", func(*builderApp) *ui.Element { return chadcn.Textarea("catalog-textarea", "Text") }},
		{"Text Field", func(*builderApp) *ui.Element {
			return chadcn.TextField("catalog-text-field", "Company", "Acme")
		}},
		{"Toast", func(*builderApp) *ui.Element { return chadcn.Toast(chadcn.ToastOptions{Title: "Saved"}) }},
		{"Toggle", func(*builderApp) *ui.Element { return chadcn.Toggle("Toggle", true) }},
		{"Toggle Group", func(*builderApp) *ui.Element { return chadcn.ToggleGroup(chadcn.ToggleGroupOptions{}) }},
		{"Tooltip", func(*builderApp) *ui.Element {
			return chadcn.Tooltip(chadcn.TooltipOptions{Open: true, Content: "Add to library"}, chadcn.Button("Hover", chadcn.ButtonOptions{}))
		}},
		{"Trace", func(*builderApp) *ui.Element {
			return chadcn.Trace(chadcn.TraceOptions{Rows: []chadcn.TraceRow{{Label: "Build", Detail: "42 ms", Start: 0, Width: catalogTraceWidth, Status: "success"}, {Label: "Deploy", Detail: "18 ms", Start: catalogTraceStart, Width: catalogTraceShort, Status: "warning"}}})
		}},
		{"Typography", func(*builderApp) *ui.Element { return chadcn.H1("Heading") }},
	}
}
