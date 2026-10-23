package draw

import (
	"strings"
	"testing"
)

func TestEmojiPassInsertsRule(t *testing.T) {
	t.Parallel()

	const plain = `<p>hi</p>`
	if EmojiPass(plain) != plain {
		t.Fatal("plain changed")
	}

	withHead := EmojiPass("<head><title>t</title></head>😂")
	if !strings.Contains(withHead, "<head><style>") || !strings.Contains(withHead, "data-ownframe-emoji") {
		t.Fatalf("head: %s", withHead)
	}
	if !strings.Contains(withHead, `src="emoji/`) {
		t.Fatalf("image: %s", withHead)
	}

	bare := EmojiPass("😂")
	if !strings.HasPrefix(bare, "<style>") {
		t.Fatalf("bare: %s", bare)
	}
}
