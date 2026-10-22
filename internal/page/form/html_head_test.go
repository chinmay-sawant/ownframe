package form

import (
	"strings"
	"testing"
)

func TestRewriteHead(t *testing.T) {
	t.Parallel()

	src := "<html><HEAD><title>t</title></HEAD><body>X</body></html>"
	at := strings.Index(src, "X")
	sp := Span{Control: Control{ID: "a", Type: "text", Value: "Z"}, Start: at, End: at + 1}
	got := Rewrite(src, []Span{sp}, nil, "", false)
	style := strings.Index(got, "<style>")
	head := strings.Index(strings.ToLower(got), "</head>")
	if style < 0 || head < 0 || style > head || !strings.Contains(got, ">Z</span>") {
		t.Fatalf("%s", got)
	}
	if !strings.Contains(got, "<HEAD>"+formCSS) {
		t.Fatalf("%s", got)
	}
	if Rewrite("hi", nil, nil, "a", false) != "hi" {
		t.Fatal("empty spans changed")
	}
}

func TestRewriteHeaderNotHead(t *testing.T) {
	t.Parallel()

	src := `<header id="h">x</header><input id="e" value="v">`
	ctrl := Control{ID: "e", Tag: "input", Type: "text", Value: "v"}
	at := strings.Index(src, "<input")
	got := Rewrite(src, []Span{{Control: ctrl, Start: at, End: len(src)}}, nil, "", false)
	if !strings.HasPrefix(got, "<style>") || strings.Contains(got, `<header id="h">`+formCSS) {
		t.Fatalf("%s", got)
	}
}

func TestRewriteFile(t *testing.T) {
	t.Parallel()

	src := `<input type="file" id="f">`
	ctrl := Control{ID: "f", Tag: "input", Type: "file"}
	got := Rewrite(src, []Span{whole(src, ctrl)}, nil, "", false)
	if !strings.Contains(got, ">No file</span>") {
		t.Fatalf("%s", got)
	}
	ctrl.Value = "a.txt"
	got = Rewrite(src, []Span{whole(src, ctrl)}, nil, "", false)
	if !strings.Contains(got, ">a.txt</span>") {
		t.Fatalf("%s", got)
	}
}
