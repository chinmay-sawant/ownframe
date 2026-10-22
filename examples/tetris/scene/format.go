package scene

import (
	"fmt"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

// padScore prints nine digits, the width of game.MaxScore, so a value
// change never moves the text run.
func padScore(n int) string { return fmt.Sprintf("%09d", clampInt(n, 0, game.MaxScore)) }

// padLines prints four digits, the width of game.MaxLines.
func padLines(n int) string { return fmt.Sprintf("%04d", clampInt(n, 0, game.MaxLines)) }

// padLevel prints two digits, the width of game.MaxLevel.
func padLevel(n int) string { return fmt.Sprintf("%02d", clampInt(n, 1, game.MaxLevel)) }

func clampInt(v, lo, hi int) int {
	return min(max(v, lo), hi)
}

// pageLabel names a history page. A plus marks a page that may have more
// entries after it.
func pageLabel(page int, more bool) string {
	label := "PAGE " + fmt.Sprint(page+1)
	if more {
		label += "+"
	}

	return label
}

// themeLabel is the theme button caption: the state it switches to.
func themeLabel(dark bool) string {
	if dark {
		return "THEME LIGHT"
	}

	return "THEME DARK"
}

// entryLine formats one history row in monospace columns. The stable game
// id names the player, and seeded demo entries carry a DEMO tag.
func entryLine(e ScoreEntry) string {
	tag := ""
	if e.Dummy {
		tag = "  DEMO"
	}

	return fmt.Sprintf("%5d  %9d  %5d  L%02d  %s%s",
		e.Rank, e.Score, e.Lines, e.Level, shortID(e.ID), tag)
}

// shortID trims a game id to the columns a row can carry.
func shortID(id string) string {
	if id == "" {
		return "-"
	}

	if len(id) > shortIDLen {
		return id[:shortIDLen]
	}

	return id
}
