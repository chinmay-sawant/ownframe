package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// catalogEntry is one card in the ChadCN catalog. Every entry is interactive:
// controlled components read builderApp catalog state and report through
// On* callbacks, while display-only components carry a data-action tap target
// so clicks, hovers, and notices keep working. Text inputs keep stable
// Page.Form IDs so the page carries user edits.
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

	catalogCarouselLast    = 2
	catalogContextX        = 24
	catalogContextY        = 40
	catalogHexColumns      = 8
	catalogScrollContentH  = 400
	catalogScrollViewportH = 120
	catalogTone1           = 1
	catalogTone2           = 2
	catalogTone3           = 3
	catalogTone4           = 4
	catalogStatusNew       = 24
	catalogStatusProgress  = 61
	catalogStatusScheduled = 38
	catalogStatusCompleted = 112
	catalogProgressStep    = 10
	catalogProgressMax     = 100
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
		{"Accordion", func(a *builderApp) *ui.Element {
			open := append([]string(nil), a.catalogAccordion...)
			return chadcn.Accordion(chadcn.AccordionOptions{
				Items: []chadcn.AccordionItem{
					{ID: "one", Title: "Is it accessible?", Content: ui.Text("Yes. It follows the button pattern.")},
					{ID: "two", Title: "Can I try it?", Content: ui.Text("Click the trigger to expand.")},
				},
				Open: open,
				OnToggle: func(ctx context.Context, id string) error {
					kept := []string{}
					found := false
					for _, v := range a.catalogAccordion {
						if v == id {
							found = true
							continue
						}
						kept = append(kept, v)
					}
					if !found {
						kept = append(kept, id)
					}
					a.catalogAccordion = kept
					a.notice = "Catalog accordion " + id + " toggled."
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Alert", func(*builderApp) *ui.Element {
			return chadcn.Alert(chadcn.AlertOptions{
				Title:       "Heads up",
				Description: "The component set is ready.",
				Action:      chadcn.Button("View", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:alert:view"),
			})
		}},
		{"Alert Dialog", func(a *builderApp) *ui.Element {
			return chadcn.AlertDialog(chadcn.AlertDialogOptions{
				ID:          "catalog-alert-dialog",
				Title:       "Delete file?",
				Description: "This cannot be undone.",
				Trigger:     chadcn.Button("Open alert dialog", chadcn.ButtonOptions{}),
				Open:        a.catalogAlertDialogOpen,
				OnOpenChange: func(ctx context.Context, open bool) error {
					a.catalogAlertDialogOpen = open
					a.notice = fmt.Sprintf("Catalog alert dialog open=%v.", open)
					return a.refreshBlur(ctx)
				},
				OnConfirm: func(ctx context.Context) error {
					a.catalogAlertDialogOpen = false
					a.notice = "Catalog file deleted."
					return a.refreshBlur(ctx)
				},
				OnCancel: func(ctx context.Context) error {
					a.catalogAlertDialogOpen = false
					a.notice = "Catalog delete cancelled."
					return a.refreshBlur(ctx)
				},
			})
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
			}).ID("catalog-area-chart").Attr("data-action", "catalog:area-chart:tap")
		}},
		{"Aspect Ratio", func(*builderApp) *ui.Element {
			return chadcn.AspectRatio(chadcn.AspectRatioOptions{}, ui.Text("16:9")).Width(ui.Percent(100)).BackgroundColor(theme.Control()).ID("catalog-aspect").Attr("data-action", "catalog:aspect:tap")
		}},
		{"Attachment", func(*builderApp) *ui.Element {
			return chadcn.Attachment(chadcn.AttachmentOptions{Name: "report.pdf", Action: "catalog:attachment:dismiss"})
		}},
		{"Avatar", func(*builderApp) *ui.Element {
			return chadcn.Avatar(chadcn.AvatarOptions{Initials: "AK"}).ID("catalog-avatar").Attr("data-action", "catalog:avatar:tap")
		}},
		{"Badge", func(*builderApp) *ui.Element {
			return chadcn.Badge("Badge").ID("catalog-badge").Attr("data-action", "catalog:badge:tap")
		}},
		{"Breadcrumb", func(*builderApp) *ui.Element {
			return chadcn.Breadcrumb([]chadcn.BreadcrumbItem{{Label: "Home", Action: "catalog:crumb:home"}, {Label: "Library", Action: "catalog:crumb:library"}, {Label: "Catalog", Current: true}})
		}},
		{"Bubble", func(*builderApp) *ui.Element {
			return chadcn.Bubble(chadcn.BubbleOptions{}, ui.Text("Hello")).ID("catalog-bubble").Attr("data-action", "catalog:bubble:tap")
		}},
		{"Button", func(*builderApp) *ui.Element {
			return chadcn.Button("Button", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).ID("catalog-button").Attr("data-action", "catalog:button:tap")
		}},
		{"Button Group", func(*builderApp) *ui.Element {
			return chadcn.ButtonGroup(chadcn.ButtonGroupOptions{},
				chadcn.Button("Left", chadcn.ButtonOptions{}).Attr("data-action", "catalog:button-group:left"),
				chadcn.Button("Right", chadcn.ButtonOptions{}).Attr("data-action", "catalog:button-group:right"))
		}},
		{"Calendar", func(a *builderApp) *ui.Element {
			return chadcn.Calendar(chadcn.CalendarOptions{
				Month:    a.catalogCalendarMonth,
				Selected: a.catalogCalendarSelected,
				OnSelect: func(ctx context.Context, day time.Time) error {
					a.catalogCalendarSelected = day
					a.catalogCalendarMonth = day
					a.notice = "Catalog date " + day.Format("January 2, 2006") + " picked."
					return a.refreshBlur(ctx)
				},
				OnMonthChange: func(ctx context.Context, month time.Time) error {
					a.catalogCalendarMonth = month
					a.notice = "Catalog calendar " + month.Format("January 2006") + "."
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Card", func(*builderApp) *ui.Element {
			return chadcn.Card(
				chadcn.CardHeader(chadcn.CardTitle("Card")),
				chadcn.CardContent(ui.Text("Content")),
				chadcn.CardFooter(chadcn.Button("Open", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:card:open")),
			)
		}},
		{"Carousel", func(a *builderApp) *ui.Element {
			return chadcn.Carousel(chadcn.CarouselOptions{
				ID:     "catalog-carousel",
				Index:  a.catalogCarouselIndex,
				Slides: []*ui.Element{ui.Text("Slide one"), ui.Text("Slide two"), ui.Text("Slide three")},
				OnPrev: func(ctx context.Context) error {
					if a.catalogCarouselIndex > 0 {
						a.catalogCarouselIndex--
					}
					a.notice = fmt.Sprintf("Catalog carousel slide %d.", a.catalogCarouselIndex+1)
					return a.refreshBlur(ctx)
				},
				OnNext: func(ctx context.Context) error {
					if a.catalogCarouselIndex < catalogCarouselLast {
						a.catalogCarouselIndex++
					}
					a.notice = fmt.Sprintf("Catalog carousel slide %d.", a.catalogCarouselIndex+1)
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Chart", func(*builderApp) *ui.Element {
			return chadcn.Chart(chadcn.ChartOptions{
				Title:  "Visitors",
				Series: []chadcn.ChartSeries{{Label: "Desktop", Values: []float64{32, 48, 61, 44, 72}}},
				Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
			}).ID("catalog-chart").Attr("data-action", "catalog:chart:tap")
		}},
		{"Checkbox", func(a *builderApp) *ui.Element {
			return chadcn.Checkbox("Checkbox", a.catalogCheck).Attr("data-action", "catalog:check")
		}},
		{"Choice", func(a *builderApp) *ui.Element {
			return chadcn.Choice("Nightly backups", "Keeps 14 days of snapshots.", a.catalogChoiceSelected).Attr("data-action", "catalog:choice:toggle")
		}},
		{"Code Block", func(*builderApp) *ui.Element {
			return chadcn.CodeBlock(chadcn.CodeBlockOptions{Filename: "main.go", Code: "package main"}).ID("catalog-code").Attr("data-action", "catalog:code:tap")
		}},
		{"Collapsible", func(a *builderApp) *ui.Element {
			return chadcn.Collapsible(chadcn.CollapsibleOptions{
				ID:      "catalog-collapsible",
				Open:    a.catalogCollapsibleOpen,
				Trigger: chadcn.Button("Toggle", chadcn.ButtonOptions{}),
				OnOpenChange: func(ctx context.Context, open bool) error {
					a.catalogCollapsibleOpen = open
					a.notice = fmt.Sprintf("Catalog collapsible open=%v.", open)
					return a.refreshBlur(ctx)
				},
			}, ui.Text("Hidden content"))
		}},
		{"Combobox", func(a *builderApp) *ui.Element {
			return chadcn.Combobox(chadcn.ComboboxOptions{
				ID:    "catalog-combobox",
				Open:  a.catalogComboboxOpen,
				Query: a.catalogComboboxQuery,
				Value: a.catalogComboboxValue,
				Items: []chadcn.ComboboxItem{{Value: "next", Label: "Next.js"}, {Value: "remix", Label: "Remix"}},
				OnOpenChange: func(ctx context.Context, open bool) error {
					a.catalogComboboxOpen = open
					a.notice = fmt.Sprintf("Catalog combobox open=%v.", open)
					return a.refreshBlur(ctx)
				},
				OnQuery: func(ctx context.Context, query string) error {
					a.catalogComboboxQuery = query
					return a.refresh(ctx)
				},
				OnSelect: func(ctx context.Context, value string) error {
					a.catalogComboboxValue = value
					a.catalogComboboxOpen = false
					a.notice = "Catalog combobox picked " + value + "."
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Command", func(a *builderApp) *ui.Element {
			query := a.catalogCommandQuery
			groups := []chadcn.CommandGroup{{Label: "Suggestions", Items: []chadcn.CommandItem{{Value: "calendar", Label: "Calendar"}, {Value: "search", Label: "Search"}}}}
			filtered := catalogFilterCommand(groups, query)
			return chadcn.Command(chadcn.CommandOptions{
				ID:     "catalog-command",
				Query:  query,
				Active: a.catalogCommandActive,
				Groups: filtered,
				OnQuery: func(ctx context.Context, next string) error {
					a.catalogCommandQuery = next
					return a.refresh(ctx)
				},
				OnSelect: func(ctx context.Context, value string) error {
					a.catalogCommandActive = value
					a.notice = "Catalog command picked " + value + "."
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Context Menu", func(a *builderApp) *ui.Element {
			opener := chadcn.Button("Open menu", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:context:toggle")
			if !a.catalogContextOpen {
				return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).Position(ui.PositionRelative).Children(
					opener,
					ui.Text("Right-click menu is closed.").FontSize(ui.Px(12)).Color(theme.Muted()),
				)
			}
			menu := chadcn.ContextMenu(chadcn.ContextMenuOptions{
				ID:    "catalog-context",
				Open:  true,
				X:     catalogContextX,
				Y:     catalogContextY,
				Items: []chadcn.MenuItem{{Label: "Copy"}, {Label: "Paste"}},
				OnSelect: func(ctx context.Context, label string) error {
					a.catalogContextOpen = false
					a.notice = "Catalog context picked " + label + "."
					return a.refreshBlur(ctx)
				},
			})
			return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).Position(ui.PositionRelative).Children(opener, menu)
		}},
		{"Data Table", func(a *builderApp) *ui.Element {
			if a.catalogDataSelected == nil {
				a.catalogDataSelected = map[int]bool{}
			}
			return chadcn.DataTable(chadcn.DataTableOptions{
				ID:       "catalog-data-table",
				Columns:  []chadcn.DataColumn{{Label: "Name"}, {Label: "Status"}},
				Rows:     [][]string{{"Ada", "Active"}, {"Linus", "Away"}},
				Selected: a.catalogDataSelected,
				OnSelect: func(ctx context.Context, row int) error {
					if a.catalogDataSelected[row] {
						delete(a.catalogDataSelected, row)
					} else {
						a.catalogDataSelected[row] = true
					}
					a.notice = fmt.Sprintf("Catalog row %d selected.", row+1)
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Date Picker", func(a *builderApp) *ui.Element {
			return chadcn.DatePicker(chadcn.DatePickerOptions{
				ID:       "catalog-date-picker",
				Open:     a.catalogDateOpen,
				Month:    a.catalogDateMonth,
				Selected: a.catalogDateSelected,
				OnOpenChange: func(ctx context.Context, open bool) error {
					a.catalogDateOpen = open
					a.notice = fmt.Sprintf("Catalog date picker open=%v.", open)
					return a.refreshBlur(ctx)
				},
				OnSelect: func(ctx context.Context, day time.Time) error {
					a.catalogDateSelected = day
					a.catalogDateMonth = day
					a.catalogDateOpen = false
					a.notice = "Catalog date " + day.Format("January 2, 2006") + " picked."
					return a.refreshBlur(ctx)
				},
				OnMonthChange: func(ctx context.Context, month time.Time) error {
					a.catalogDateMonth = month
					a.notice = "Catalog date picker " + month.Format("January 2006") + "."
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Dialog", func(a *builderApp) *ui.Element {
			return chadcn.Dialog(chadcn.DialogOptions{
				ID:      "catalog-dialog",
				Title:   "Dialog",
				Trigger: chadcn.Button("Open dialog", chadcn.ButtonOptions{}),
				Open:    a.catalogDialogOpen,
				OnOpenChange: func(ctx context.Context, open bool) error {
					a.catalogDialogOpen = open
					a.notice = fmt.Sprintf("Catalog dialog open=%v.", open)
					return a.refreshBlur(ctx)
				},
			}, ui.Text("Dialog body. Close to return to the catalog."))
		}},
		{"Direction", func(*builderApp) *ui.Element {
			return chadcn.Direction(chadcn.DirectionOptions{}, ui.Text("RTL")).ID("catalog-direction").Attr("data-action", "catalog:direction:tap")
		}},
		{"Drawer", func(a *builderApp) *ui.Element {
			return chadcn.Drawer(chadcn.DrawerOptions{
				ID:      "catalog-drawer",
				Title:   "Drawer",
				Trigger: chadcn.Button("Open drawer", chadcn.ButtonOptions{}),
				Open:    a.catalogDrawerOpen,
				OnOpenChange: func(ctx context.Context, open bool) error {
					a.catalogDrawerOpen = open
					a.notice = fmt.Sprintf("Catalog drawer open=%v.", open)
					return a.refreshBlur(ctx)
				},
			}, ui.Text("Drawer body."))
		}},
		{"Dropdown Menu", func(a *builderApp) *ui.Element {
			return chadcn.DropdownMenu(chadcn.DropdownMenuOptions{
				ID:      "catalog-dropdown",
				Open:    a.catalogDropdownOpen,
				Trigger: chadcn.Button("Open", chadcn.ButtonOptions{}),
				Items:   []chadcn.MenuItem{{Label: "Profile"}, {Label: "Settings"}},
				OnOpenChange: func(ctx context.Context, open bool) error {
					a.catalogDropdownOpen = open
					a.notice = fmt.Sprintf("Catalog dropdown open=%v.", open)
					return a.refreshBlur(ctx)
				},
				OnSelect: func(ctx context.Context, label string) error {
					a.catalogDropdownOpen = false
					a.notice = "Catalog dropdown picked " + label + "."
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Empty", func(*builderApp) *ui.Element {
			return chadcn.Empty(chadcn.EmptyOptions{Title: "No results", Description: "Try another search.", Actions: []*ui.Element{chadcn.Button("Clear", chadcn.ButtonOptions{}).Attr("data-action", "catalog:empty:clear")}})
		}},
		{"Field", func(*builderApp) *ui.Element {
			return chadcn.Field(chadcn.FieldOptions{Label: "Email", For: "catalog-email", Control: chadcn.Input("catalog-email", "")})
		}},
		{"Filter Bar", func(a *builderApp) *ui.Element {
			active := a.catalogFilter
			if active == "" {
				active = "weekly"
			}
			return chadcn.FilterBar(chadcn.FilterBarOptions{
				Label:  "Trend filters",
				Legend: []chadcn.FilterBarLegend{{Label: "Opened", Tone: catalogTone1}, {Label: "Resolved", Tone: catalogTone2}, {Label: "Backlog gap", Tone: catalogTone3}},
				Options: []chadcn.FilterBarOption{
					{Label: "Daily", Active: active == "daily", Action: "catalog:filter:daily"},
					{Label: "Weekly", Active: active == "weekly", Action: "catalog:filter:weekly"},
					{Label: "Monthly", Active: active == "monthly", Action: "catalog:filter:monthly"},
				},
			})
		}},
		{"Hex Grid", func(*builderApp) *ui.Element {
			return chadcn.HexGrid(chadcn.HexGridOptions{
				ID:      "catalog-hex",
				Label:   "Portfolio health",
				Columns: catalogHexColumns,
				Cells: []chadcn.HexCell{
					{Tone: catalogTone1}, {Tone: catalogTone1}, {Tone: catalogTone2}, {Tone: catalogTone2}, {Tone: catalogTone2}, {Tone: 0}, {Tone: 0}, {Tone: catalogTone1},
					{Tone: catalogTone1}, {Tone: catalogTone3}, {Tone: catalogTone3}, {Tone: catalogTone2}, {Tone: catalogTone4}, {Tone: catalogTone4}, {Tone: 0}, {Tone: catalogTone1},
					{Tone: catalogTone3}, {Tone: catalogTone3}, {Tone: catalogTone4}, {Tone: catalogTone4}, {Tone: catalogTone2}, {Tone: catalogTone1}, {Tone: 0}, {Tone: 0},
					{Tone: 0}, {Tone: catalogTone4}, {Tone: catalogTone4}, {Tone: catalogTone3}, {Tone: catalogTone3}, {Tone: catalogTone2}, {Tone: 0}, {Tone: 0},
				},
			}).ID("catalog-hex-grid").Attr("data-action", "catalog:hex:tap")
		}},
		{"Hover Card", func(a *builderApp) *ui.Element {
			return chadcn.HoverCard(chadcn.HoverCardOptions{ID: "catalog-hover", Open: a.catalogHoverOpen}, chadcn.Button("Hover", chadcn.ButtonOptions{}).ID("catalog-hover-trigger").Attr("data-action", "catalog:hover:toggle"), ui.Text("Preview"))
		}},
		{"Input", func(*builderApp) *ui.Element { return chadcn.Input("catalog-input", "Value") }},
		{"Input Group", func(*builderApp) *ui.Element {
			return chadcn.InputGroup(chadcn.InputGroupOptions{ID: "catalog-group", Label: "Handle", Value: "acme", Prefix: ui.Text("@")})
		}},
		{"Input OTP", func(*builderApp) *ui.Element {
			return chadcn.InputOTP(chadcn.InputOTPOptions{ID: "catalog-otp", Length: catalogOTPLength})
		}},
		{"Item", func(*builderApp) *ui.Element {
			return chadcn.Item(chadcn.ItemOptions{Media: ui.Text("A"), Title: "Ada Lovelace", Description: "First programmer", Actions: chadcn.Button("Open", chadcn.ButtonOptions{}).Attr("data-action", "catalog:item:open")})
		}},
		{"Kbd", func(*builderApp) *ui.Element {
			return chadcn.Kbd("⌘K").ID("catalog-kbd").Attr("data-action", "catalog:kbd:tap")
		}},
		{"Label", func(*builderApp) *ui.Element {
			return chadcn.Label("Label").ID("catalog-label").Attr("data-action", "catalog:label:tap")
		}},
		{"Marker", func(*builderApp) *ui.Element {
			return chadcn.Marker(chadcn.MarkerOptions{Label: "Deployed"}).ID("catalog-marker").Attr("data-action", "catalog:marker:tap")
		}},
		{"Menubar", func(a *builderApp) *ui.Element {
			return chadcn.Menubar(chadcn.MenubarOptions{
				ID:    "catalog-menubar",
				Open:  a.catalogMenubarOpen,
				Menus: []chadcn.MenubarMenu{{ID: "file", Label: "File", Items: []chadcn.MenuItem{{Label: "New"}, {Label: "Open"}}}},
				OnOpenChange: func(ctx context.Context, id string) error {
					a.catalogMenubarOpen = id
					a.notice = "Catalog menubar open=" + id + "."
					return a.refreshBlur(ctx)
				},
				OnSelect: func(ctx context.Context, label string) error {
					a.catalogMenubarOpen = ""
					a.notice = "Catalog menubar picked " + label + "."
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Message", func(*builderApp) *ui.Element {
			return chadcn.Message(chadcn.MessageOptions{
				Name:     "Ada",
				Time:     "4:55 PM",
				Children: []*ui.Element{ui.Text("Hello from the catalog."), chadcn.Button("Reply", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:message:reply")},
			})
		}},
		{"Message Scroller", func(*builderApp) *ui.Element {
			return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).Children(
				chadcn.MessageScroller(chadcn.MessageScrollerOptions{}),
				chadcn.Button("Load messages", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:messages:load"),
			)
		}},
		{"Native Select", func(*builderApp) *ui.Element {
			return chadcn.NativeSelect("catalog-native", []chadcn.SelectValue{{Value: "a", Label: "Apple"}, {Value: "b", Label: "Banana"}}, "a")
		}},
		{"Navigation Menu", func(a *builderApp) *ui.Element {
			return chadcn.NavigationMenu(chadcn.NavigationMenuOptions{
				ID:    "catalog-navmenu",
				Open:  a.catalogNavOpen,
				Items: []chadcn.NavigationItem{{ID: "docs", Label: "Docs"}},
				OnOpenChange: func(ctx context.Context, id string) error {
					a.catalogNavOpen = id
					a.notice = "Catalog navigation open=" + id + "."
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Notice", func(*builderApp) *ui.Element {
			return chadcn.Notice("Trial ends in 3 days", "Add a payment method to keep syncing.").ID("catalog-notice").Attr("data-action", "catalog:notice:tap")
		}},
		{"Pagination", func(a *builderApp) *ui.Element {
			page := a.catalogPage
			if page < 1 {
				page = 1
			}
			return chadcn.Pagination(chadcn.PaginationOptions{
				Page:       page,
				TotalPages: catalogPageCount,
				OnSelect: func(ctx context.Context, next int) error {
					a.catalogPage = next
					a.notice = fmt.Sprintf("Catalog page %d.", next)
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Popover", func(a *builderApp) *ui.Element {
			return chadcn.Popover(chadcn.PopoverOptions{
				ID:      "catalog-popover",
				Open:    a.catalogPopoverOpen,
				Trigger: chadcn.Button("Open", chadcn.ButtonOptions{}),
				OnOpenChange: func(ctx context.Context, open bool) error {
					a.catalogPopoverOpen = open
					a.notice = fmt.Sprintf("Catalog popover open=%v.", open)
					return a.refreshBlur(ctx)
				},
			}, ui.Text("Popover content"))
		}},
		{"Progress", func(a *builderApp) *ui.Element {
			value := a.catalogProgress
			if value == 0 {
				value = catalogProgress
			}
			return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).Children(
				chadcn.Progress("Progress", value).ID("catalog-progress"),
				ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).Children(
					chadcn.Button("−", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:progress:dec"),
					ui.Text(fmt.Sprintf("%d%%", value)).FontSize(ui.Px(12)).Color(theme.Muted()),
					chadcn.Button("+", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:progress:inc"),
				),
			)
		}},
		{"Questionnaire", func(a *builderApp) *ui.Element {
			return chadcn.Questionnaire(chadcn.QuestionnaireOptions{
				Title: "Feedback",
				Questions: []chadcn.Question{{
					Label:    "How was the setup?",
					Selected: a.catalogQuestionAnswer,
					Choices:  []chadcn.ChoiceItem{{Value: "easy", Label: "Easy"}, {Value: "hard", Label: "Hard"}},
				}},
				OnSelect: func(ctx context.Context, question, value string) error {
					a.catalogQuestionAnswer = value
					a.notice = "Catalog feedback " + value + "."
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Radio Group", func(a *builderApp) *ui.Element {
			value := a.catalogRadioValue
			if value == "" {
				value = "medium"
			}
			return chadcn.RadioGroup(chadcn.RadioGroupOptions{
				ID:    "catalog-radio",
				Label: "Size",
				Value: value,
				Items: []chadcn.RadioItem{{Value: "small", Label: "Small"}, {Value: "medium", Label: "Medium"}, {Value: "large", Label: "Large"}},
				OnSelect: func(ctx context.Context, next string) error {
					a.catalogRadioValue = next
					a.notice = "Catalog size " + next + "."
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Resizable", func(*builderApp) *ui.Element {
			return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).Children(
				chadcn.Resizable(chadcn.ResizableOptions{ID: "catalog-resizable", First: ui.Text("One"), Second: ui.Text("Two")}),
				chadcn.Button("Reset split", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:resizable:reset"),
			)
		}},
		{"Scroll Area", func(*builderApp) *ui.Element {
			return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).Children(
				chadcn.ScrollArea(chadcn.ScrollAreaOptions{ID: "catalog-scroll", Height: ui.Px(120), Children: []*ui.Element{
					ui.Div().Children(ui.Text("Row one")),
					ui.Div().Children(ui.Text("Row two")),
					ui.Div().Children(ui.Text("Row three")),
				}}),
				chadcn.Button("Jump to top", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:scroll:top"),
			)
		}},
		{"Scrollbar", func(*builderApp) *ui.Element {
			state := &chadcn.ScrollState{}
			state.SetRange(catalogScrollContentH, catalogScrollViewportH)
			return ui.Div().Position(ui.PositionRelative).Height(ui.Px(140)).Overflow(ui.OverflowHidden).Children(
				ui.P().Text("Scrollable region with an overlaid scrollbar.").Margin(ui.Px(0)).FontSize(ui.Px(12)),
				chadcn.Scrollbar(chadcn.ScrollbarOptions{ID: "catalog-scrollbar", ControlsID: "catalog-scroll", State: state, TrackPixels: catalogScrollViewportH}),
				chadcn.Button("Nudge", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:scrollbar:nudge"),
			)
		}},
		{"Search", func(a *builderApp) *ui.Element {
			return chadcn.Search(chadcn.SearchOptions{
				ID:          "catalog-search",
				Query:       a.catalogSearchQuery,
				Placeholder: "Search orders",
				Open:        a.catalogSearchOpen,
				Items:       catalogSearchItems(),
				Value:       a.catalogSearchValue,
				OnOpenChange: func(ctx context.Context, open bool) error {
					a.catalogSearchOpen = open
					if open {
						a.notice = "Catalog search opened over its backdrop."
						return a.refresh(ctx)
					}
					a.notice = "Catalog search dismissed."
					return a.refreshBlur(ctx)
				},
				OnQuery: func(ctx context.Context, query string) error {
					a.catalogSearchQuery = query
					a.catalogSearchOpen = true
					return a.refresh(ctx)
				},
				OnSelect: func(ctx context.Context, value string) error {
					a.catalogSearchValue = value
					a.catalogSearchOpen = false
					a.notice = "Catalog search picked " + value + "."
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Search Field", func(*builderApp) *ui.Element {
			return chadcn.SearchField("catalog-search-field", "Search orders", "")
		}},
		{"Select", func(a *builderApp) *ui.Element {
			return chadcn.Select(chadcn.SelectOptions{
				ID:    "catalog-select",
				Open:  a.catalogSelectOpen,
				Value: a.catalogSelect,
				Items: []chadcn.SelectItem{{Value: "a", Label: "Apple"}, {Value: "b", Label: "Banana"}},
				OnOpenChange: func(ctx context.Context, open bool) error {
					a.catalogSelectOpen = open
					return a.refreshBlur(ctx)
				},
				OnSelect: func(ctx context.Context, value string) error {
					a.catalogSelect = value
					a.catalogSelectOpen = false
					a.notice = "Catalog select picked " + value + "."
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Separator", func(*builderApp) *ui.Element {
			return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).Children(
				chadcn.Separator(chadcn.SeparatorOptions{}),
				ui.Text("Section divider").FontSize(ui.Px(12)).Color(theme.Muted()),
			).ID("catalog-separator").Attr("data-action", "catalog:separator:tap")
		}},
		{"Sheet", func(a *builderApp) *ui.Element {
			return chadcn.Sheet(chadcn.SheetOptions{
				ID:      "catalog-sheet",
				Title:   "Sheet",
				Trigger: chadcn.Button("Open sheet", chadcn.ButtonOptions{}),
				Open:    a.catalogSheetOpen,
				OnOpenChange: func(ctx context.Context, open bool) error {
					a.catalogSheetOpen = open
					a.notice = fmt.Sprintf("Catalog sheet open=%v.", open)
					return a.refreshBlur(ctx)
				},
			}, ui.Text("Sheet body."))
		}},
		{"Sidebar", func(a *builderApp) *ui.Element {
			toggle := chadcn.Button("Toggle sidebar", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:sidebar:toggle")
			return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).Children(
				toggle,
				ui.Div().Height(ui.Px(180)).Overflow(ui.OverflowHidden).Children(
					chadcn.Sidebar(chadcn.SidebarOptions{ID: "catalog-sidebar", Open: true, Collapsed: !a.catalogSidebarOpen, Width: ui.Px(220), Header: ui.Text("Acme"), Content: ui.Text("Workspace"), Footer: ui.Text("shadcn@example.com")}),
				),
			)
		}},
		{"Sidebar Section", func(*builderApp) *ui.Element {
			return chadcn.SidebarSection("Operations",
				chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "Dashboard", Active: true, Action: "catalog:nav:dashboard"}),
				chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "Work orders", Badge: "137", Action: "catalog:nav:work-orders"}),
				chadcn.SidebarNavItem(chadcn.SidebarItem{Label: "Maintenance", Action: "catalog:nav:maintenance"}),
			)
		}},
		{"Skeleton", func(*builderApp) *ui.Element {
			return ui.Div().Children(chadcn.Skeleton(chadcn.SkeletonOptions{})).ID("catalog-skeleton").Attr("data-action", "catalog:skeleton:tap")
		}},
		{"Slider", func(a *builderApp) *ui.Element {
			// The host owns Value. Nudge buttons step it, and a horizontal
			// drag maps the pointer through SliderValueAt.
			return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).Children(
				chadcn.Slider(a.catalogSliderOptions()).ID(catalogSliderID),
				ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).Children(
					chadcn.Button("−", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:slider:dec"),
					ui.Text(fmt.Sprintf("%.0f", a.catalogSlider)).FontSize(ui.Px(12)).Color(theme.Muted()),
					chadcn.Button("+", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:slider:inc"),
				),
			)
		}},
		{"Spinner", func(*builderApp) *ui.Element {
			return ui.Div().Children(chadcn.Spinner(chadcn.SpinnerOptions{})).ID("catalog-spinner").Attr("data-action", "catalog:spinner:tap")
		}},
		{"Stat Card", func(*builderApp) *ui.Element {
			return chadcn.StatCard(chadcn.StatCardOptions{Label: "Open work orders", Value: "137", Delta: "23 urgent", Tone: chadcn.StatDeltaWarn}).ID("catalog-stat").Attr("data-action", "catalog:stat:tap")
		}},
		{"Status Segments", func(*builderApp) *ui.Element {
			return chadcn.StatusSegments(chadcn.StatusSegmentsOptions{
				ID:      "catalog-segments",
				Label:   "Work order status",
				Overdue: "Overdue work orders: 7",
				Segments: []chadcn.StatusSegment{
					{Label: "New", Value: catalogStatusNew, Count: "24", Tone: catalogTone1},
					{Label: "In progress", Value: catalogStatusProgress, Count: "61", Tone: catalogTone3},
					{Label: "Scheduled", Value: catalogStatusScheduled, Count: "38", Tone: catalogTone2},
					{Label: "Completed", Value: catalogStatusCompleted, Count: "112", Tone: catalogTone4},
				},
			}).ID("catalog-status").Attr("data-action", "catalog:segments:tap")
		}},
		{"Switch", func(a *builderApp) *ui.Element {
			return chadcn.Switch("Switch", a.catalogSwitch).Attr("data-action", "catalog:switch")
		}},
		{"Table", func(*builderApp) *ui.Element {
			return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).ID("catalog-table").Attr("data-action", "catalog:table:tap").Children(
				chadcn.Table(chadcn.TableHeader(chadcn.TableRow(chadcn.TableHead("Name"))), chadcn.TableBody(chadcn.TableRow(chadcn.TableCell("Ada")))),
			)
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
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Textarea", func(*builderApp) *ui.Element { return chadcn.Textarea("catalog-textarea", "Text") }},
		{"Text Field", func(*builderApp) *ui.Element {
			return chadcn.TextField("catalog-text-field", "Company", "Acme")
		}},
		{"Toast", func(a *builderApp) *ui.Element {
			if !a.catalogToastOpen {
				return chadcn.Button("Show toast", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:toast:show")
			}
			return chadcn.Toast(chadcn.ToastOptions{
				Title:       "Saved",
				Description: "Catalog toast is open.",
				Open:        true,
				Action:      chadcn.Button("Undo", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).Attr("data-action", "catalog:toast:undo"),
				CloseAction: "catalog:toast:close",
			})
		}},
		{"Toggle", func(a *builderApp) *ui.Element {
			return chadcn.Toggle("Toggle", a.catalogTogglePressed).Attr("data-action", "catalog:toggle:tap")
		}},
		{"Toggle Group", func(a *builderApp) *ui.Element {
			return chadcn.ToggleGroup(chadcn.ToggleGroupOptions{
				Items:    []chadcn.ToggleItem{{Value: "bold", Label: "Bold"}, {Value: "italic", Label: "Italic"}},
				Selected: a.catalogToggleGroupSelected,
				OnSelect: func(ctx context.Context, value string) error {
					kept := []string{}
					found := false
					for _, v := range a.catalogToggleGroupSelected {
						if v == value {
							found = true
							continue
						}
						kept = append(kept, v)
					}
					if !found {
						kept = append(kept, value)
					}
					a.catalogToggleGroupSelected = kept
					a.notice = "Catalog toggle " + value + "."
					return a.refreshBlur(ctx)
				},
			})
		}},
		{"Tooltip", func(a *builderApp) *ui.Element {
			return chadcn.Tooltip(chadcn.TooltipOptions{ID: "catalog-tooltip", Open: a.catalogTooltipOpen, Content: "Add to library"}, chadcn.Button("Hover", chadcn.ButtonOptions{}).ID("catalog-tooltip-trigger").Attr("data-action", "catalog:tooltip:toggle"))
		}},
		{"Trace", func(*builderApp) *ui.Element {
			return chadcn.Trace(chadcn.TraceOptions{Rows: []chadcn.TraceRow{{Label: "Build", Detail: "42 ms", Start: 0, Width: catalogTraceWidth, Status: "success"}, {Label: "Deploy", Detail: "18 ms", Start: catalogTraceStart, Width: catalogTraceShort, Status: "warning"}}}).ID("catalog-trace").Attr("data-action", "catalog:trace:tap")
		}},
		{"Typography", func(*builderApp) *ui.Element {
			return ui.Div().Children(chadcn.H1("Heading")).ID("catalog-type").Attr("data-action", "catalog:type:tap")
		}},
	}
}

// catalogSearchItems is the prefilled card data for the catalog Search demo.
// Typing filters these rows locally; a query with no match is the signal for
// a background fetch of remote rows.
func catalogSearchItems() []chadcn.SearchItem {
	return []chadcn.SearchItem{
		{Value: "order-1001", Label: "Order #1001", Description: "Ada · pending"},
		{Value: "order-1002", Label: "Order #1002", Description: "Linus · shipped"},
		{Value: "order-1003", Label: "Order #1003", Description: "Grace · pending"},
		{Value: "order-1004", Label: "Order #1004", Description: "Alan · delivered"},
		{Value: "order-1005", Label: "Order #1005", Description: "Katherine · pending"},
		{Value: "order-1006", Label: "Order #1006", Description: "Margaret · shipped"},
		{Value: "refund-2001", Label: "Refund #2001", Description: "Ada · review"},
		{Value: "invoice-3001", Label: "Invoice #3001", Description: "Acme · paid"},
	}
}

// catalogFilterCommand keeps only the command rows matching the query,
// case-insensitively on label or value. An empty query keeps every group, so
// typing filters the palette and clearing restores it. A query with no match
// yields no groups and the component renders its "No results" state.
func catalogFilterCommand(groups []chadcn.CommandGroup, query string) []chadcn.CommandGroup {
	query = strings.TrimSpace(query)
	if query == "" {
		return groups
	}
	lower := strings.ToLower(query)
	filtered := make([]chadcn.CommandGroup, 0, len(groups))
	for _, group := range groups {
		kept := make([]chadcn.CommandItem, 0, len(group.Items))
		for _, item := range group.Items {
			if strings.Contains(strings.ToLower(item.Label), lower) || strings.Contains(strings.ToLower(item.Value), lower) {
				kept = append(kept, item)
			}
		}
		if len(kept) > 0 {
			filtered = append(filtered, chadcn.CommandGroup{Label: group.Label, Items: kept})
		}
	}
	return filtered
}
