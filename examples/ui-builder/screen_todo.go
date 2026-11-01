package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
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
		Next:    4,
	}
}

// todoTransition lists only the shipped transition fields from
// documentation/css-motion.md, so a toggle cross-fades instead of snapping.
const todoTransition = "background-color 180ms ease, color 180ms ease, border-color 180ms ease, opacity 180ms ease, transform 180ms ease"

func (a *builderApp) todoPanel() *ui.Element {
	return ui.Section().CSS("flex", "1").CSS("min-width", "0px").CSS("min-height", "0px").
		Display("flex").FlexDirection("column").Gap("14px").
		Children(
			ui.Div().Children(
				ui.Text("Todo & counter").FontSize("24px").FontWeight("bold").Color("#f4f7ff"),
				ui.P().Text("A stateful panel: the counter mutates an integer and the list mutates a slice.").
					Margin("7px 0 0").Color("#a9bbd1").FontSize("12px"),
			),
			ui.Div().Display("flex").CSS("min-width", "0px").Gap("12px").
				Children(a.todoCounterCard(), a.todoListCard()),
		)
}

func (a *builderApp) todoCounterCard() *ui.Element {
	return ui.Div().CSS("width", "200px").CSS("flex-shrink", "0").
		Padding("16px").BorderRadius("11px").BackgroundColor("#202d42").
		Children(
			ui.Text("Counter").FontSize("13px").Color("#b5c7df"),
			ui.P().Text(fmt.Sprintf("%d", a.todos.Counter)).Margin("10px 0 5px").
				FontSize("29px").FontWeight("bold").Color("#f4f7ff"),
			ui.P().Text("Increment, decrement, or reset the value.").Margin("0 0 12px").
				FontSize("11px").Color("#a0b6d0"),
			ui.Div().Display("flex").Gap("8px").Children(
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
			Margin("0").FontSize("12px").Color("#a0b6d0"))
	}
	card := ui.Div().CSS("flex", "1").CSS("min-width", "0px").
		Padding("16px").BorderRadius("11px").BackgroundColor("#202d42").
		Children(
			ui.Text(fmt.Sprintf("Tasks (%d)", len(a.todos.Items))).
				FontSize("13px").FontWeight("bold").Color("#b5c7df"),
			a.todoProgress(),
			ui.Div().Display("flex").FlexDirection("column").Gap("8px").
				CSS("margin-top", "12px").Children(rows...),
		)
	if a.allTodoDone() {
		card.Children(a.todoDoneBanner())
	}
	return card.Children(
		ui.Div().Display("flex").Gap("8px").CSS("margin-top", "12px").Children(
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
		Margin("6px 0 0").FontSize("11px").Color("#8fa3bb")
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
		Display("flex").AlignItems("center").Gap("8px").
		CSS("margin-top", "12px").
		Padding("10px 12px").BorderRadius("8px").
		BackgroundColor("#17452e").Color("#c9f2d8").
		CSS("border", "1px solid #2b7a4b").
		FontSize("12px").FontWeight("bold").
		CSS("animation-name", "todo-done-pop").
		CSS("animation-duration", "450ms").
		CSS("animation-timing-function", "ease").
		CSS("animation-fill-mode", "both").
		CSS("transform-origin", "center center").
		Children(ui.Text("All tasks done — nice work!"))
}

func (a *builderApp) todoRow(index int, item string) *ui.Element {
	label := "○ " + item
	decoration := "none"
	itemColor := "#ecf2fb"
	rowBG := "#26364e"
	btnBG := "#26364e"
	btnBorder := "1px solid #3a4b62"
	rowOpacity := "1"
	if a.todos.Done[index] {
		label = "● " + item
		decoration = "line-through"
		itemColor = "#8fa3bb"
		rowBG = "#1f2f47"
		btnBG = "#1e2c44"
		btnBorder = "1px solid #2e3f58"
		rowOpacity = "0.85"
	}
	// The row and toggle ids let the automatic transition driver match the
	// old and new boxes across the re-render. The transition lists only
	// shipped fields; text-decoration snaps while colour and opacity fade.
	return ui.Div().ID(fmt.Sprintf("todo-row-%d", index)).Class("todo-row").
		Display("flex").AlignItems("center").Gap("10px").
		Padding("10px 12px").BackgroundColor(rowBG).BorderRadius("8px").
		CSS("opacity", rowOpacity).
		CSS("transition", todoTransition).
		Children(
			ui.Button(label).ID(fmt.Sprintf("todo-toggle-%d", index)).Class("todo-toggle").
				Attr("data-action", "todo:toggle:"+strconv.Itoa(index)).
				CSS("flex", "1").CSS("text-align", "left").
				BackgroundColor(btnBG).Color(itemColor).
				CSS("border", btnBorder).
				Padding("6px 8px").BorderRadius("6px").
				CSS("text-decoration", decoration).
				CSS("transition", todoTransition),
		)
}

func (a *builderApp) todoButton(label, action string, primary bool) *ui.Element {
	bg, fg, border := "#2b3950", "#dce6f6", "1px solid #42536b"
	if primary {
		bg, fg, border = "#4677ed", "#ffffff", "1px solid #4677ed"
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", action).
		BackgroundColor(bg).Color(fg).CSS("border", border).
		Padding("9px 12px").BorderRadius("8px").
		FontSize("12px").FontWeight("bold")
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
