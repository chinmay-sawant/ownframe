package form

import (
	"strings"
	"testing"
)

func TestHeadSearchSkipsComments(t *testing.T) {
	t.Parallel()

	if HeadOpen(`<!-- <head>fake</head> -->`) >= 0 || FindHead(`<!-- <head>fake</head> -->`) >= 0 {
		t.Fatal("found a head tag inside a comment")
	}
}

func TestRewriteFindsHeadAfterComment(t *testing.T) {
	t.Parallel()

	comment := `<!-- <head>fake</head> -->`
	src := comment + `<head><title>real</title></head><input id="e">`
	input := strings.Index(src, `<input`)
	ctrl := Control{ID: "e", Tag: "input", Type: "text"}
	sp := Span{Control: ctrl, Start: input, End: len(src)}
	got := Rewrite(src, []Span{sp}, nil, "", false)
	if !strings.Contains(got, comment+`<head>`+formCSS) {
		t.Fatalf("style was not inserted after the real head: %s", got)
	}
}
