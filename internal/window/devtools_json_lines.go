package window

import "image/color"

// devJSONExtra is the indent + text spans devJSONLine always adds.
const devJSONExtra = 2

// devJSONLine builds indent + lead + text + tail.
func devJSONLine(indent string, lead devLine, text string, ink color.RGBA, tail string) devLine {
	spans := make([]devSpan, 0, len(lead.spans)+devJSONExtra)
	spans = append(spans, devSpan{text: indent, ink: devPunctInk})
	spans = append(spans, lead.spans...)
	spans = append(spans, devSpan{text: text, ink: ink})

	if tail != "" {
		spans = append(spans, devSpan{text: tail, ink: devPunctInk})
	}

	return devLine{spans: spans}
}

// devJSONPlain builds indent + text + tail in punctuation ink.
func devJSONPlain(indent, text, tail string) devLine {
	return devJSONLine(indent, devLine{}, text, devPunctInk, tail)
}

// devJSONLead is the `"key": ` member prefix, key in the key ink.
func devJSONLead(key string) devLine {
	q := string(devJSONMarshal(key))

	return devLine{spans: []devSpan{
		{text: `"`, ink: devPunctInk},
		{text: q[1 : len(q)-1], ink: devKeyInk},
		{text: `": `, ink: devPunctInk},
	}}
}
