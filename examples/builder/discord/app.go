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

// dcAppearance is the settings Appearance choice. Dark matches the shipped
// Discord paint; light reuses the existing .of-scope--light overrides.
type dcAppearance string

const (
	dcDark  dcAppearance = "dark"
	dcLight dcAppearance = "light"
)

// Window geometry. Tests build app without a page and fall back to the same
// default size.
const (
	dcWindowWidth     = 1440
	dcWindowHeight    = 900
	dcWindowMinWidth  = 1100
	dcWindowMinHeight = 700
)

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

	appearance  dcAppearance
	themeDirty  bool
	chatScroll  chadcn.ScrollState
	chatViewH   int
	scrollStale bool

	profileOpen     string
	profileX        float64
	profileY        float64
	profileAnchored bool
	createOpen      bool

	toastOpen     bool
	toastTitle    string
	toastBody     string
	toastUntil    time.Time
	msgMenu       string
	msgX          int
	msgY          int
	msgAnchored   bool
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
		channels:    make(map[string][]dchannel),
		messages:    make(map[string][]dmessage),
		members:     make(map[string][]dmember),
		drafts:      make(map[string]string),
		collapsed:   make(map[string]bool),
		unread:      make(map[string]int),
		appearance:  dcDark,
		scrollStale: true,
		flags: map[string]bool{
			"overlay": true,
		},
	}
	seedData(a)
	view, err := ui.Mount(ownframe.Config{
		Title:       "Discord clone",
		Width:       dcWindowWidth,
		Height:      dcWindowHeight,
		MinWidth:    dcWindowMinWidth,
		MinHeight:   dcWindowMinHeight,
		Theme:       a.themeSheet(),
		RetainState: true,
		Scrollbar:   ownframe.ScrollbarHidden,
		// Fixed-pixel transcript windowing measures unscaled boxes, so
		// painted zoom would desync it. Ctrl+wheel falls through to the
		// chat scroll instead.
		DisableWheelZoom: true,
		DisableKeyZoom:   true,
		DisableTouchZoom: true,
	}, a.screen())
	if err != nil {
		return nil, err
	}
	a.view = view
	a.page = view.Page
	a.registerIcons()
	view.Handle(ownframe.Handlers{
		Click:     a.handleClick,
		Change:    a.handleChange,
		KeyDown:   a.handleKey,
		Hover:     a.handleHover,
		Wheel:     a.handleWheel,
		DragStart: a.startScrollDrag,
		DragMove:  a.moveScrollDrag,
		DragEnd:   a.endScrollDrag,
	})
	a.startTyping()
	return a, nil
}

func (a *app) refresh(ctx context.Context) error {
	if a.themeDirty && a.page != nil {
		if err := a.page.SetTheme(a.themeSheet()); err != nil {
			return err
		}
		a.themeDirty = false
	}
	if err := a.view.Update(ctx, a.screen()); err != nil {
		return err
	}
	clamped := a.updateScrollCache(a.currentMessages())
	if a.scrollStale {
		a.scrollStale = false
		a.chatScroll.ScrollTo(a.chatScroll.Metrics().MaxOffset())
		if err := a.view.Update(ctx, a.screen()); err != nil {
			return err
		}
		a.updateScrollCache(a.currentMessages())
		return nil
	}
	if clamped {
		if err := a.view.Update(ctx, a.screen()); err != nil {
			return err
		}
		a.updateScrollCache(a.currentMessages())
	}
	return nil
}

func (a *app) themeSheet() string {
	base := chadcn.Dark
	if a != nil && a.appearance == dcLight {
		base = chadcn.Light
	}
	return base + chadcn.Stylesheet + appCSS +
		searchCSS + profileCSS + railCSS + composerCSS + msghoverCSS + settingsCSS +
		scrollwinCSS
}

// selectTheme validates the Appearance choice and marks the sheet dirty.
func (a *app) selectTheme(name string) {
	switch dcAppearance(name) {
	case dcLight, dcDark:
		if a.appearance != dcAppearance(name) {
			a.appearance = dcAppearance(name)
			a.themeDirty = true
		}
	}
}

// windowSize returns the current viewport in CSS pixels. Tests build app
// without a page, so fall back to the default window size.
func (a *app) windowSize() (int, int) {
	const fallbackW, fallbackH = dcWindowWidth, dcWindowHeight
	if a == nil || a.page == nil {
		return fallbackW, fallbackH
	}
	w, h := a.page.Size()
	if w <= 0 || h <= 0 {
		return fallbackW, fallbackH
	}
	return w, h
}
