package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// Card groups related content on a bordered surface. Compose it with
// CardHeader, CardTitle, CardDescription, CardContent, and CardFooter.
func Card(children ...*ui.Element) *ui.Element {
	return ui.Section().Class("of-card").Children(children...)
}

// CardHeader stacks the title and description above the card content.
func CardHeader(children ...*ui.Element) *ui.Element {
	return ui.Div().Class("of-card-header").Children(children...)
}

// CardTitle is the card heading. Extra children follow the text.
func CardTitle(text string, children ...*ui.Element) *ui.Element {
	return ui.Tag("h3").Class("of-card-title").Text(text).Children(children...)
}

// CardDescription is the secondary text below a CardTitle.
func CardDescription(text string, children ...*ui.Element) *ui.Element {
	return ui.P().Class("of-card-description").Text(text).Children(children...)
}

// CardContent holds the main body of a card.
func CardContent(children ...*ui.Element) *ui.Element {
	return ui.Div().Class("of-card-content").Children(children...)
}

// CardFooter holds actions below a top border.
func CardFooter(children ...*ui.Element) *ui.Element {
	return ui.Div().Class("of-card-footer").Children(children...)
}
