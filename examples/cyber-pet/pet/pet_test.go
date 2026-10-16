package pet

import (
	"strings"
	"testing"
)

func TestHTMLTemplate(t *testing.T) {
	if !strings.Contains(indexHTML, "<!-- stylesheet -->") {
		t.Fatal("stylesheet marker missing in indexHTML")
	}
	html := buildHTML()
	if !strings.Contains(html, "<style>") {
		t.Fatal("style tag not inlined in buildHTML")
	}
}

func TestInitialization(t *testing.T) {
	a, _ := newT(t)
	s := a.State()
	if s.Energy != 80 || s.Hunger != 80 || s.Happiness != 80 {
		t.Fatalf("unexpected vitals: %+v", s)
	}
	if s.Mood != MoodHappy {
		t.Fatalf("unexpected mood: %s", s.Mood)
	}
	if a.Page() == nil {
		t.Fatal("page is nil")
	}
	if len(a.Page().Boxes()) == 0 {
		t.Fatal("no element boxes laid out")
	}
}

func TestFeedAction(t *testing.T) {
	a, _ := newT(t)
	a.state.Hunger = 40.0
	a.Feed()
	s := a.State()
	if s.Hunger != 65.0 {
		t.Fatalf("expected hunger 65, got %f", s.Hunger)
	}
	a.Feed()
	a.Feed()
	s = a.State()
	if s.Hunger != MaxStat {
		t.Fatalf("expected hunger clamped to 100, got %f", s.Hunger)
	}
}

func TestPlayAction(t *testing.T) {
	a, _ := newT(t)
	a.state.Energy = 50.0
	a.state.Happiness = 50.0
	a.Play()
	s := a.State()
	if s.Energy != 35.0 || s.Happiness != 70.0 {
		t.Fatalf("unexpected vitals after play: %+v", s)
	}

	a.state.Energy = 10.0
	a.Play()
	s = a.State()
	if !strings.Contains(s.Speech, "Too tired") {
		t.Fatalf("expected tired speech, got %s", s.Speech)
	}
}

func TestPetAction(t *testing.T) {
	a, _ := newT(t)
	a.state.Happiness = 50.0
	a.Pet()
	s := a.State()
	if s.Happiness != 65.0 {
		t.Fatalf("expected happiness 65, got %f", s.Happiness)
	}
	if s.Mood != MoodPurring {
		t.Fatalf("expected purring mood, got %s", s.Mood)
	}
	if !a.pm.heartActive {
		t.Fatal("expected floating heart active after pet")
	}
}
