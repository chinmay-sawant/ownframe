package remote

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/lg-remote/tv"
)

type recordLink struct {
	fakeLink
	mu    sync.Mutex
	specs []string
}

func (r *recordLink) Exec(host, spec string) (string, error) {
	r.mu.Lock()
	r.specs = append(r.specs, spec)
	r.mu.Unlock()

	return "ok", nil
}

func (r *recordLink) taken() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := r.specs
	r.specs = nil

	return out
}

func keysApp(t *testing.T) (*App, *recordLink) {
	t.Helper()
	a := newTest(t, WithPhone(true))
	link := &recordLink{}
	a.SetLink(link)
	clickID(t, a, "tab-keys")
	if a.view.Panel != "keys" {
		t.Fatalf("panel %q", a.view.Panel)
	}
	link.taken()

	return a, link
}

func TestKeyboardSendsTypedTextAsItChanges(t *testing.T) {
	a, link := keysApp(t)
	steps := []struct {
		field string
		want  []string
	}{
		{"cat", []string{"ime:text:cat"}},
		{"cats", []string{"ime:text:s"}},
		{"cat", []string{"ime:delete:1"}},
		{"dog", []string{"ime:delete:3", "ime:text:dog"}},
		{"dog:a b", []string{"ime:text::a b"}},
	}
	for _, s := range steps {
		a.page.Form().SetFormValue(typingField, s.field)
		if err := a.onTick(context.Background()); err != nil {
			t.Fatal(err)
		}
		a.runJobs()
		if got := link.taken(); !reflect.DeepEqual(got, s.want) {
			t.Fatalf("field %q sent %v, want %v", s.field, got, s.want)
		}
	}
}

func TestKeyboardButtonsSendTheirCommands(t *testing.T) {
	a, link := keysApp(t)
	a.page.Form().SetFormValue(typingField, "ab")
	if err := a.onTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	a.runJobs()
	link.taken()

	for _, c := range []struct {
		id   string
		want []string
	}{
		{"kb-delete", []string{"ime:delete:1"}},
		{"kb-enter", []string{"ime:enter:"}},
		{"kb-clear", []string{"ime:clear:"}},
	} {
		clickID(t, a, c.id)
		a.runJobs()
		if got := link.taken(); !reflect.DeepEqual(got, c.want) {
			t.Fatalf("%s sent %v, want %v", c.id, got, c.want)
		}
	}
	if got := a.page.Form().FormValue(typingField); got != "" {
		t.Fatalf("field after clear %q", got)
	}
}

func TestKeyboardNeedsWiFi(t *testing.T) {
	a, link := keysApp(t)
	a.view.Mode = "Bluetooth"
	a.page.Form().SetFormValue(typingField, "x")
	if err := a.onTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	a.runJobs()
	if got := link.taken(); len(got) != 0 {
		t.Fatalf("bluetooth sent %v", got)
	}
}

func TestSettingsButtonsWork(t *testing.T) {
	a := newTest(t, WithPhone(true))
	clickID(t, a, "settings")
	if a.view.Panel != "settings" {
		t.Fatal("settings did not open")
	}
	for _, id := range []string{"accent-red", "accent-orange", "accent-green", "accent-blue", "accent-purple", "accent-pink"} {
		clickID(t, a, id)
		if want := id[len("accent-"):]; a.view.Accent != want {
			t.Fatalf("%s set accent %q", id, a.view.Accent)
		}
	}
	clickID(t, a, "set-light")
	if a.dark {
		t.Fatal("light did not apply")
	}
	clickID(t, a, "set-dark")
	if !a.dark {
		t.Fatal("dark did not apply")
	}
	clickID(t, a, "layout-comfort")
	if a.view.Layout != layoutComfort {
		t.Fatal("comfort did not apply")
	}
	clickID(t, a, "settings-done")
	if a.view.Panel == "settings" {
		t.Fatal("done did not close settings")
	}
	for _, tab := range []string{"tab-media", "tab-keys", "tab-nums", "tab-pad", "tab-remote"} {
		clickID(t, a, tab)
		if a.view.Panel != tab[len("tab-"):] {
			t.Fatalf("%s opened %q", tab, a.view.Panel)
		}
	}
}

type stateLink struct {
	recordLink
	state tv.KeyboardState
}

func (s *stateLink) KeyboardState() tv.KeyboardState { return s.state }

func TestKeysTabFocusesTheFieldAfterTheClick(t *testing.T) {
	a, _ := keysApp(t)
	if err := a.onTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := a.page.Form().FocusID(); got != typingField {
		t.Fatalf("focus %q, want %q so the phone keyboard opens", got, typingField)
	}
}

func TestKeysNoteFollowsTheTVKeyboard(t *testing.T) {
	a := newTest(t, WithPhone(true))
	link := &stateLink{}
	a.SetLink(link)
	clickID(t, a, "tab-keys")

	for _, c := range []struct {
		state tv.KeyboardState
		want  string
	}{
		{tv.KeyboardOpen, "The TV keyboard is open. Type here."},
		{tv.KeyboardClosed, "No LG text box is open. Press OK on a search box to open the TV keyboard, then type here. Apps with their own keyboard (YouTube, Disney+) cannot take typing."},
		{tv.KeyboardUnreported, "Open a text box on the TV, then type here."},
	} {
		link.state = c.state
		if err := a.onTick(context.Background()); err != nil {
			t.Fatal(err)
		}
		if a.view.KeyboardNote != c.want {
			t.Fatalf("state %d note %q, want %q", c.state, a.view.KeyboardNote, c.want)
		}
	}
}

func TestKeysPanelNavButtonsSendTVKeys(t *testing.T) {
	a, link := keysApp(t)
	for id, want := range map[string]string{
		"left": "button:LEFT", "up": "button:UP", "kb-back": "button:BACK", "ok": "button:ENTER",
		"down": "button:DOWN", "right": "button:RIGHT",
	} {
		clickID(t, a, id)
		a.runJobs()
		got := link.taken()
		if len(got) != 1 || got[0] != want {
			t.Fatalf("%s sent %v, want %s", id, got, want)
		}
	}
}

func TestClearReachesTheTVEvenWhenThePhoneFieldIsEmpty(t *testing.T) {
	a, link := keysApp(t)
	clickID(t, a, "kb-clear")
	a.runJobs()
	if got := link.taken(); !reflect.DeepEqual(got, []string{"ime:clear:"}) {
		t.Fatalf("clear sent %v", got)
	}
}

func TestTappingTheFieldKeepsItFocused(t *testing.T) {
	a, _ := keysApp(t)
	if err := a.onTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	clickID(t, a, typingField)
	if err := a.onTick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := a.page.Form().FocusID(); got != typingField {
		t.Fatalf("focus %q after tapping the field", got)
	}
	if a.view.PressedID == typingField {
		t.Fatal("the text field was flashed like a button")
	}
}
