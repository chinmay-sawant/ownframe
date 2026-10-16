package pet

import "time"

func (s *State) updateMood(now time.Time) {
	if s.Sleeping {
		s.Mood = MoodSleeping
		return
	}
	if now.Before(s.PettingUntil) {
		s.Mood = MoodPurring
		return
	}
	if now.Before(s.PlayUntil) {
		s.Mood = MoodPlayful
		return
	}
	if s.Hunger < 30.0 {
		s.Mood = MoodHungry
		return
	}
	if s.Energy < 25.0 {
		s.Mood = MoodTired
		return
	}
	if s.Happiness < 30.0 {
		s.Mood = MoodSad
		return
	}
	if s.Happiness >= 80.0 && s.Energy >= 60.0 {
		s.Mood = MoodPlayful
		return
	}
	s.Mood = MoodHappy
}

func moodEmoji(mood string) string {
	switch mood {
	case MoodSleeping:
		return "💤"
	case MoodPurring:
		return "😻"
	case MoodPlayful:
		return "😸"
	case MoodHungry:
		return "😿"
	case MoodTired:
		return "🥱"
	case MoodSad:
		return "😿"
	default:
		return "😺"
	}
}

func statusHint(s State) string {
	if s.Sleeping {
		return "Pet is sleeping and recharging..."
	}
	if s.Mood == MoodHungry {
		return "Pet is hungry! Press 1 to feed."
	}
	if s.Mood == MoodTired {
		return "Pet is tired. Press 4 to let it sleep."
	}
	if s.Mood == MoodSad {
		return "Pet is lonely. Press 2 to play or 3 to pet."
	}
	return "Pet is thriving in cyberspace!"
}
