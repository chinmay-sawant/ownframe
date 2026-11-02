package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// todoState holds the todo tab's counter and task list.
// It is stored on builderApp as a.todos by the app owner.
type todoState struct {
	Items   []string
	Done    map[int]bool
	Counter int
	Next    int
}

func defaultTodoState() todoState {
	return todoState{
		Items:   []string{"Wire the counter buttons", "Add a todo row", "Clear completed tasks"},
		Done:    map[int]bool{},
		Counter: 0,
		Next:    todoInitialNext,
	}
}

// todoInitialNext is one past the three seeded items, so the first task
// added with "+ Add task" is numbered 4.
const todoInitialNext = 4

// todoTransition lists only the shipped transition fields from
// documentation/css-motion.md, so a toggle cross-fades instead of snapping.
const todoTransition = "background-color 180ms ease, color 180ms ease, border-color 180ms ease, opacity 180ms ease, transform 180ms ease"

func (a *builderApp) todoPanel() *ui.Element {
	return ui.Section().Flex(ui.Number(1)).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).
		Children(
			ui.Div().Children(
				a.panelHeading("Todo & counter"),
				ui.P().Text("A stateful panel: the counter mutates an integer and the list mutates a slice.").
					Margin(ui.Edges(ui.Px(7), ui.Px(0), ui.Px(0))).Color(theme.Muted()).FontSize(ui.Px(12)),
			),
			ui.Div().Display(ui.DisplayFlex).MinWidth(ui.Px(0)).Gap(ui.Px(12)).
				Children(a.todoCounterCard(), a.todoListCard()),
		)
}

func (a *builderApp) todoCounterCard() *ui.Element {
	return ui.Div().Width(ui.Px(200)).FlexShrink(ui.Number(0)).
		Padding(ui.Px(16)).BorderRadius(ui.Px(11)).BackgroundColor(theme.Surface()).
		Children(
			ui.Text("Counter").FontSize(ui.Px(13)).Color(theme.Muted()),
			ui.P().Text(fmt.Sprintf("%d", a.todos.Counter)).Margin(ui.Edges(ui.Px(10), ui.Px(0), ui.Px(5))).
				FontSize(ui.Px(29)).FontWeight(ui.FontWeightBold).Color(theme.Text()),
			ui.P().Text("Increment, decrement, or reset the value.").Margin(ui.Edges(ui.Px(0), ui.Px(0), ui.Px(12))).
				FontSize(ui.Px(11)).Color(theme.Muted()),
			ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).Children(
				a.todoButton("−", "todo:dec", false),
				a.todoButton("+", "todo:inc", true),
				a.todoButton("Reset", "todo:reset", false),
			),
		)
}

func (a *builderApp) todoListCard() *ui.Element {
	rows := make([]*ui.Element, 0, len(a.todos.Items)+1)
	for i, item := range a.todos.Items {
		rows = append(rows, a.todoRow(i, item))
	}
	if len(rows) == 0 {
		rows = append(rows, ui.P().Text("No tasks — add one below.").
			Margin(ui.Px(0)).FontSize(ui.Px(12)).Color(theme.Muted()))
	}
	card := ui.Div().Flex(ui.Number(1)).MinWidth(ui.Px(0)).
		Padding(ui.Px(16)).BorderRadius(ui.Px(11)).BackgroundColor(theme.Surface()).
		Children(
			ui.Text(fmt.Sprintf("Tasks (%d)", len(a.todos.Items))).
				FontSize(ui.Px(13)).FontWeight(ui.FontWeightBold).Color(theme.Muted()),
			a.todoProgress(),
			ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(8)).
				MarginTop(ui.Px(12)).Children(rows...),
		)
	if a.allTodoDone() {
		card.Children(a.todoDoneBanner())
	}
	return card.Children(
		ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).MarginTop(ui.Px(12)).Children(
			a.todoButton("+ Add task", "todo:add", true),
			a.todoButton("Clear done", "todo:clear", false),
		),
	)
}

// todoProgress reports how many tasks are done. It re-renders with the list,
// so completing a row visibly moves the count.
func (a *builderApp) todoProgress() *ui.Element {
	done := 0
	for i := range a.todos.Items {
		if a.todos.Done[i] {
			done++
		}
	}
	return ui.P().Text(fmt.Sprintf("%d of %d done", done, len(a.todos.Items))).
		Margin(ui.Edges(ui.Px(6), ui.Px(0), ui.Px(0))).FontSize(ui.Px(11)).Color(theme.Muted())
}

// allTodoDone reports whether the list is non-empty and every row is done.
func (a *builderApp) allTodoDone() bool {
	if len(a.todos.Items) == 0 {
		return false
	}
	for i := range a.todos.Items {
		if !a.todos.Done[i] {
			return false
		}
	}
	return true
}

// todoDoneBanner celebrates a fully completed list. The pop keyframes live in
// the document stylesheet (renderBuilderDocument); the inline animation
// longhands below only reference them, so the automatic CSS driver samples the
// entrance without any Go tick callback.
func (a *builderApp) todoDoneBanner() *ui.Element {
	return ui.Div().ID("todo-all-done").Class("todo-done-banner").
		Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(8)).
		MarginTop(ui.Px(12)).
		Padding(ui.Edges(ui.Px(10), ui.Px(12))).BorderRadius(ui.Px(8)).
		BackgroundColor(theme.Control()).Color(theme.Text()).
		Border(theme.Border()).
		FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold).
		AnimationName(ui.RawValue("todo-done-pop")).
		AnimationDuration(ui.RawValue("450ms")).
		AnimationTimingFunction(ui.RawValue("ease")).
		AnimationFillMode(ui.RawValue("both")).
		TransformOrigin(ui.RawValue("center center")).
		Children(ui.Text("All tasks done — nice work!"))
}

func (a *builderApp) todoRow(index int, item string) *ui.Element {
	label := "○ " + item
	decoration := ui.TextDecorationNone
	itemColor := theme.Text()
	rowBG := theme.Control()
	btnBG := theme.Control()
	btnBorder := theme.Border()
	rowOpacity := 1.0
	if a.todos.Done[index] {
		label = "● " + item
		decoration = ui.TextDecorationLineThrough
		itemColor = theme.Muted()
		rowBG = theme.Selected()
		btnBG = theme.Selected()
		btnBorder = theme.Border()
		rowOpacity = 0.85
	}
	// The row and toggle ids let the automatic transition driver match the
	// old and new boxes across the re-render. The transition lists only
	// shipped fields; text-decoration snaps while colour and opacity fade.
	return ui.Div().ID(fmt.Sprintf("todo-row-%d", index)).Class("todo-row").
		Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).
		Padding(ui.Edges(ui.Px(10), ui.Px(12))).BackgroundColor(rowBG).BorderRadius(ui.Px(8)).
		Opacity(ui.Number(rowOpacity)).
		Transition(ui.RawValue(todoTransition)).
		Children(
			ui.Button(label).ID(fmt.Sprintf("todo-toggle-%d", index)).Class("todo-toggle").
				Attr("data-action", "todo:toggle:"+strconv.Itoa(index)).
				Flex(ui.Number(1)).TextAlign(ui.TextAlignLeft).
				BackgroundColor(btnBG).Color(itemColor).
				Border(btnBorder).
				Padding(ui.Edges(ui.Px(6), ui.Px(8))).BorderRadius(ui.Px(6)).
				TextDecoration(decoration).
				Transition(ui.RawValue(todoTransition)),
		)
}

func (a *builderApp) todoButton(label, action string, primary bool) *ui.Element {
	return widgetsBtn(label, action, primary)
}

// handleTodo handles todo:* actions, updating a.todos and a.notice.
// It reports whether the action was consumed.
func (a *builderApp) handleTodo(action string) bool {
	if action == "todo:inc" {
		a.todos.Counter++
		a.notice = fmt.Sprintf("Counter incremented to %d.", a.todos.Counter)
		return true
	}
	if action == "todo:dec" {
		a.todos.Counter--
		a.notice = fmt.Sprintf("Counter decremented to %d.", a.todos.Counter)
		return true
	}
	if action == "todo:reset" {
		a.todos.Counter = 0
		a.notice = "Counter reset to 0."
		return true
	}
	if action == "todo:add" {
		a.todos.Items = append(a.todos.Items, fmt.Sprintf("Task %d", a.todos.Next))
		a.todos.Next++
		a.notice = fmt.Sprintf("Added task %d of %d.", len(a.todos.Items), len(a.todos.Items))
		return true
	}
	if action == "todo:clear" {
		kept := make([]string, 0, len(a.todos.Items))
		for i, item := range a.todos.Items {
			if !a.todos.Done[i] {
				kept = append(kept, item)
			}
		}
		removed := len(a.todos.Items) - len(kept)
		a.todos.Items = kept
		a.todos.Done = map[int]bool{}
		a.notice = fmt.Sprintf("Cleared %d completed task(s); %d remaining.", removed, len(kept))
		return true
	}
	if index, ok := strings.CutPrefix(action, "todo:toggle:"); ok {
		i, err := strconv.Atoi(strings.TrimSpace(index))
		if err != nil || i < 0 || i >= len(a.todos.Items) {
			return true
		}
		if a.todos.Done == nil {
			a.todos.Done = map[int]bool{}
		}
		a.todos.Done[i] = !a.todos.Done[i]
		if a.todos.Done[i] {
			a.notice = fmt.Sprintf("Task %d marked done.", i+1)
		} else {
			a.notice = fmt.Sprintf("Task %d marked open.", i+1)
		}
		return true
	}
	return false
}
