package chadcn

import (
	"context"
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// ChoiceItem is one answer in a Question. Detail is optional supporting text.
type ChoiceItem struct {
	Value  string
	Label  string
	Detail string
}

// Question is one prompt with single-choice answers. Selected matches a
// ChoiceItem.Value and is supplied by the host on every build.
type Question struct {
	Label    string
	Help     string
	Choices  []ChoiceItem
	Selected string
}

// QuestionnaireOptions describes a question list. OnSelect receives the
// Question.Label and the chosen ChoiceItem.Value.
type QuestionnaireOptions struct {
	Title       string
	Description string
	Questions   []Question
	OnSelect    func(ctx context.Context, question, value string) error
}

// Questionnaire builds a list of questions with single-choice answers, a
// progress line, and per-question counts. The host stores the answers and
// rebuilds the tree after OnSelect returns.
func Questionnaire(options QuestionnaireOptions) *ui.Element {
	root := ui.Div().Class("of-questionnaire")
	if header := questionnaireHeader(options); header != nil {
		root.Children(header)
	}
	if len(options.Questions) > 0 {
		list := ui.Div().Class("of-questionnaire-list")
		for index, question := range options.Questions {
			list.Children(questionnaireQuestion(options, question, index))
		}
		root.Children(list)
	}
	return root
}

func questionnaireHeader(options QuestionnaireOptions) *ui.Element {
	if options.Title == "" && options.Description == "" && len(options.Questions) == 0 {
		return nil
	}
	header := ui.Div().Class("of-questionnaire-header")
	if len(options.Questions) > 0 {
		answered := 0
		for _, question := range options.Questions {
			if questionMatchesChoice(question) {
				answered++
			}
		}
		header.Children(ui.Span().Class("of-questionnaire-progress").
			Text(strconv.Itoa(answered) + " of " + strconv.Itoa(len(options.Questions)) + " answered"))
	}
	if options.Title != "" {
		header.Children(ui.Tag("h2").Class("of-questionnaire-title").Text(options.Title))
	}
	if options.Description != "" {
		header.Children(ui.P().Class("of-questionnaire-description").Text(options.Description))
	}
	return header
}

func questionnaireQuestion(options QuestionnaireOptions, question Question, index int) *ui.Element {
	item := ui.Div().Class("of-questionnaire-question")
	item.Children(ui.Span().Class("of-questionnaire-count").
		Text("Question " + strconv.Itoa(index+1) + " of " + strconv.Itoa(len(options.Questions))))
	if question.Label != "" {
		item.Children(ui.Tag("h3").Class("of-questionnaire-label").Text(question.Label))
	}
	if question.Help != "" {
		item.Children(ui.P().Class("of-questionnaire-help").Text(question.Help))
	}
	choices := ui.Div().Class("of-questionnaire-choices").Attr("role", "radiogroup").
		Attr("aria-label", question.Label)
	for choiceIndex, choice := range question.Choices {
		choices.Children(questionnaireChoice(options, question, choice, choiceIndex))
	}
	return item.Children(choices)
}

func questionnaireChoice(options QuestionnaireOptions, question Question, choice ChoiceItem, index int) *ui.Element {
	selected := question.Selected != "" && question.Selected == choice.Value
	modifier := ""
	if selected {
		modifier = "of-questionnaire-choice--selected"
	}
	button := ui.Button("").Class(classes("of-questionnaire-choice", modifier, "of-focus")).
		Attr("type", "button").Attr("role", "radio").Attr("aria-checked", ariaBool(selected))
	body := ui.Div().Class("of-questionnaire-choice-body")
	if choice.Label != "" {
		body.Children(ui.Span().Class("of-questionnaire-choice-label").Text(choice.Label))
	}
	if choice.Detail != "" {
		body.Children(ui.Span().Class("of-questionnaire-choice-detail").Text(choice.Detail))
	}
	button.Children(
		ui.Span().Class("of-questionnaire-choice-mark").Attr("aria-hidden", "true"),
		body,
		ui.Span().Class("of-questionnaire-choice-key").Attr("aria-hidden", "true").
			Text(questionnaireChoiceKey(index)),
	)
	if options.OnSelect != nil {
		handler := options.OnSelect
		button.OnClick(func(ctx context.Context) error {
			return handler(ctx, question.Label, choice.Value)
		})
	}
	return button
}

// questionMatchesChoice reports whether Selected is one of the question's
// choice values. A non-empty value that matches nothing is unanswered.
func questionMatchesChoice(question Question) bool {
	for _, choice := range question.Choices {
		if question.Selected == choice.Value {
			return true
		}
	}
	return false
}

func questionnaireChoiceKey(index int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if index >= 0 && index < len(letters) {
		return letters[index : index+1]
	}
	return strconv.Itoa(index + 1)
}
