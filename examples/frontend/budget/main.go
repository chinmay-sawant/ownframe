package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

const (
	windowWidth    = 640
	windowHeight   = 640
	maxFriends     = 8
	maxDollars     = 1000000
	centsPerDollar = 100
	decimalPlaces  = 2
	minimumFriends = 2
	maxNameLength  = 30
	chartCenter    = 60
	chartRadius    = 54
	fullTurn       = 2
	priyaID        = 2
	jordanID       = 3
	maxHistory     = 3
)

//go:embed template.html
var templateHTML string

//go:embed style.css
var stylesheet string

type friend struct {
	ID       int
	Name     string
	Selected bool
	Share    string
	Color    string
}

type segment struct {
	Path  string
	Color string
}

type savedSplit struct {
	Total  string
	Payer  string
	Shares []friend
}

type view struct {
	Screen    string
	PayerID   int
	History   []savedSplit
	Saved     bool
	Amount    string
	NewFriend string
	Friends   []friend
	Segments  []segment
	Error     string
	HasSplit  bool
	Selected  int
}

func money(cents int) string {
	return fmt.Sprintf("$%d.%02d", cents/centsPerDollar, cents%centsPerDollar)
}

func parseAmount(value string) (int, error) {
	value = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(value), "$"))
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, fmt.Errorf("enter a valid amount, such as 120.00")
	}
	for _, part := range parts {
		for _, digit := range part {
			if digit < '0' || digit > '9' {
				return 0, fmt.Errorf("enter a valid amount, such as 120.00")
			}
		}
	}
	dollars, err := strconv.Atoi(parts[0])
	if err != nil || dollars > maxDollars {
		return 0, fmt.Errorf("enter an amount up to $1,000,000")
	}
	cents := 0
	if len(parts) == decimalPlaces {
		if len(parts[1]) > decimalPlaces {
			return 0, fmt.Errorf("use at most two decimal places")
		}
		if parts[1] != "" {
			fraction := parts[1]
			if len(fraction) == 1 {
				fraction += "0"
			}
			cents, _ = strconv.Atoi(fraction)
		}
	}
	total := dollars*centsPerDollar + cents
	if total > maxDollars*centsPerDollar {
		return 0, fmt.Errorf("enter an amount up to $1,000,000")
	}
	if total == 0 {
		return 0, fmt.Errorf("enter an amount greater than zero")
	}
	return total, nil
}

func (v *view) clearSplit() {
	v.HasSplit = false
	v.Saved = false
	v.Segments = nil
	v.Error = ""
	v.Selected = 0
	for i := range v.Friends {
		v.Friends[i].Share = ""
		if v.Friends[i].Selected {
			v.Selected++
		}
	}
}

func (v *view) split() {
	v.clearSplit()
	amount, err := parseAmount(v.Amount)
	if err != nil {
		v.Error = err.Error()
		return
	}
	if v.Selected < minimumFriends {
		v.Error = "Select at least two friends."
		return
	}
	if v.PayerID >= len(v.Friends) || !v.Friends[v.PayerID].Selected {
		for i := range v.Friends {
			if v.Friends[i].Selected {
				v.PayerID = i
				break
			}
		}
	}
	share, remainder := amount/v.Selected, amount%v.Selected
	index := 0
	distributed := 0
	for i := range v.Friends {
		if !v.Friends[i].Selected {
			continue
		}
		part := share
		if index < remainder {
			part++
		}
		v.Friends[i].Share = money(part)
		if part > 0 {
			start := -math.Pi/fullTurn + fullTurn*math.Pi*float64(distributed)/float64(amount)
			end := -math.Pi/fullTurn + fullTurn*math.Pi*float64(distributed+part)/float64(amount)
			largeArc := 0
			if part*fullTurn > amount {
				largeArc = 1
			}
			path := fmt.Sprintf("M 60 60 L %.3f %.3f A 54 54 0 %d 1 %.3f %.3f Z", chartCenter+chartRadius*math.Cos(start), chartCenter+chartRadius*math.Sin(start), largeArc, chartCenter+chartRadius*math.Cos(end), chartCenter+chartRadius*math.Sin(end))
			if part == amount {
				path = "M 60 60 L 60 6 A 54 54 0 1 1 60 114 A 54 54 0 1 1 60 6 Z"
			}
			v.Segments = append(v.Segments, segment{Path: path, Color: v.Friends[i].Color})
		}
		distributed += part
		index++
	}
	v.HasSplit = true
}

func newPage() (*ownframe.Page, *view, error) {
	page, err := ownframe.New(ownframe.Config{Title: "Split with friends", Width: windowWidth, Height: windowHeight, HTML: templateHTML, Theme: stylesheet})
	if err != nil {
		return nil, nil, err
	}
	colors := [maxFriends]string{"#4e6fd8", "#48a793", "#e8a94b", "#d67683", "#9273c8", "#58a4ce", "#c58154", "#8c9a59"}
	state := &view{Screen: "split", Amount: "120.00", Friends: []friend{
		{ID: 0, Name: "Alex", Selected: true, Color: colors[0]},
		{ID: 1, Name: "Sam", Selected: true, Color: colors[1]},
		{ID: priyaID, Name: "Priya", Selected: true, Color: colors[priyaID]},
		{ID: jordanID, Name: "Jordan", Color: colors[jordanID]},
	}}
	state.split()
	page.SetData(state)
	page.Handle(ownframe.Handlers{Change: func(_ context.Context, box ownframe.Box) error {
		if box.ID == "amount" {
			state.clearSplit()
			page.SetData(state)
		}
		return nil
	}, Click: func(_ context.Context, box ownframe.Box) error {
		switch {
		case box.Action == "screen-split":
			state.Screen = "split"
		case box.Action == "screen-settle":
			state.Screen = "settle"
		case box.Action == "screen-activity":
			state.Screen = "activity"
		case strings.HasPrefix(box.Action, "payer-"):
			id, err := strconv.Atoi(strings.TrimPrefix(box.Action, "payer-"))
			if err == nil && id >= 0 && id < len(state.Friends) && state.Friends[id].Selected {
				if state.PayerID != id {
					state.PayerID = id
					state.Saved = false
				}
			}
		case box.Action == "save":
			if state.HasSplit && !state.Saved {
				amount, err := parseAmount(state.Amount)
				if err != nil {
					return err
				}
				var shares []friend
				for _, person := range state.Friends {
					if person.Selected {
						shares = append(shares, person)
					}
				}
				state.History = append([]savedSplit{{Total: money(amount), Payer: state.Friends[state.PayerID].Name, Shares: shares}}, state.History...)
				if len(state.History) > maxHistory {
					state.History = state.History[:maxHistory]
				}
				state.Saved = true
			}
		case box.Action == "split":
			state.split()
		case box.Action == "add":
			name := strings.TrimSpace(state.NewFriend)
			switch {
			case name == "":
				state.Error = "Enter a friend's name."
			case len([]rune(name)) > maxNameLength:
				state.Error = "Use a name of 30 characters or fewer."
			case len(state.Friends) >= maxFriends:
				state.Error = "This demo supports up to 8 friends."
			default:
				id := len(state.Friends)
				state.Friends = append(state.Friends, friend{ID: id, Name: name, Selected: true, Color: colors[id]})
				state.NewFriend = ""
				page.Form().SetFormValue("friend-name", "")
				state.clearSplit()
			}
		case strings.HasPrefix(box.Action, "friend-"):
			id, err := strconv.Atoi(strings.TrimPrefix(box.Action, "friend-"))
			if err == nil && id >= 0 && id < len(state.Friends) {
				state.Friends[id].Selected = !state.Friends[id].Selected
				if !state.Friends[state.PayerID].Selected {
					for i := range state.Friends {
						if state.Friends[i].Selected {
							state.PayerID = i
							break
						}
					}
				}
				state.clearSplit()
			}
		default:
			return nil
		}
		page.SetData(state)
		return nil
	}})
	return page, state, nil
}

func main() {
	page, _, err := newPage()
	if err != nil {
		log.Fatal(err)
	}
	if err := ownframe.Run(context.Background(), page); err != nil {
		log.Fatal(err)
	}
}
