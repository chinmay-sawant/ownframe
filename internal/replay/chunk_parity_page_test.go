package replay

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// cardParityPage is the chunk parity fixture: an opaque page background, a
// translucent card, a rounded card with a border and text, and a plain card.
func cardParityPage() (*layout.Display, error) {
	source := `<body style="margin:0;width:640px;height:200px;background:#ddeeff">` +
		`<div style="position:absolute;left:20px;top:20px;width:160px;height:100px;` +
		`background:rgba(51,102,204,0.5)"></div>` +
		`<div style="position:absolute;left:220px;top:20px;width:160px;height:100px;` +
		`background:#ffddaa;border:3px solid #aa6622;border-radius:8px;padding:6px;font-size:13px">Card B detail</div>` +
		`<div style="position:absolute;left:420px;top:20px;width:160px;height:100px;` +
		`background:#ccffcc;border:1px solid #338833"></div>` +
		`</body>`

	d, err := render.DisplayList(context.Background(), source, 640, 200)
	if err != nil {
		return nil, err
	}

	if !render.Replayable(d) {
		return nil, fmt.Errorf("the card fixture should replay")
	}

	return d, nil
}
