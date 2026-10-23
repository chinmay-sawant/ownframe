package form

func moveLeft(value string, caret, start, end int, extend, word bool) int {
	switch {
	case word:
		return PrevWordStart(value, caret)
	case start != end && !extend:
		return start
	default:
		return caret - 1
	}
}

func moveRight(value string, caret, start, end int, extend, word bool) int {
	switch {
	case word:
		return NextWordStart(value, caret)
	case start != end && !extend:
		return end
	default:
		return caret + 1
	}
}
