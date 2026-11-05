package main

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// Ownership: the 5 model types and the app struct live ONLY in this file.
// data.go must not redeclare them; it provides seedData(a *app), which fills
// servers, channels, messages, members, drafts, collapsed, unread, typing,
// and the initial serverID/channelID. shell.go provides screen() and the
// action sibling provides handleClick/handleChange/handleKey/handleHover.

type server struct {
	ID, Name, Initial, Color string
	Unread                   bool
	Mentions                 int
}

type dchannel struct {
	ID, ServerID, Name, Topic, Category, Kind string
	Unread                                    bool
	Mentions                                  int
}

type reaction struct {
	Emoji   string
	Count   int
	Reacted bool
}

type dmessage struct {
	ID, ChannelID, Author, AuthorColor, Time, Text, ReplyTo string
	Reactions                                               []reaction
	Edited, Pinned, Mention                                 bool
}

type dmember struct {
	ID, Name, Color, Status, Activity, Role string
	Bot                                     bool
}

type app struct {
	view *ui.View
	page *ownframe.Page

	serverID, channelID string
	servers             []server
	channels            map[string][]dchannel
	messages            map[string][]dmessage
	members             map[string][]dmember
	drafts              map[string]string
	collapsed           map[string]bool
	unread              map[string]int

	searchOpen  bool
	searchQuery string

	settingsOpen bool
	settingsTab  string

	profileOpen string
	createOpen  bool

	toastOpen     bool
	toastTitle    string
	toastBody     string
	msgMenu       string
	replyTo       string
	voiceID       string
	voiceMuted    bool
	voiceDeafened bool
	memberHidden  bool
	typing        []string
	typingIdx     int
	typingUntil   time.Time
	typingNames   []string
	// flags holds generic overlay toggles flipped by the "flag:<key>"
	// action (e.g. Game Overlay switches). The settings agent owns keys.
	flags map[string]bool
}

func newApp() (*app, error) {
	a := &app{
		channels:  make(map[string][]dchannel),
		messages:  make(map[string][]dmessage),
		members:   make(map[string][]dmember),
		drafts:    make(map[string]string),
		collapsed: make(map[string]bool),
		unread:    make(map[string]int),
		flags: map[string]bool{
			"overlay": true,
		},
	}
	seedData(a)
	view, err := ui.Mount(ownframe.Config{
		Title:       "Discord clone",
		Width:       1440,
		Height:      900,
		MinWidth:    1100,
		MinHeight:   700,
		Theme:       a.themeSheet(),
		RetainState: true,
		Scrollbar:   ownframe.ScrollbarHidden,
	}, a.screen())
	if err != nil {
		return nil, err
	}
	a.view = view
	a.page = view.Page
	view.Handle(ownframe.Handlers{
		Click:   a.handleClick,
		Change:  a.handleChange,
		KeyDown: a.handleKey,
		Hover:   a.handleHover,
	})
	a.startTyping()
	return a, nil
}

func (a *app) refresh(ctx context.Context) error {
	return a.view.Update(ctx, a.screen())
}

func (a *app) themeSheet() string {
	return chadcn.Dark + chadcn.Stylesheet + appCSS +
		searchCSS + profileCSS + railCSS + composerCSS + msghoverCSS + settingsCSS
}
