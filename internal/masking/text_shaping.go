package masking

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/chinmay-sawant/blinkless/layout"
	"golang.org/x/text/language"
)

// autospaceAbutTol is how far, in points, the end of one text op may sit from
// the start of the next and still count as the same line of text.
const autospaceAbutTol = 0.5

// fontTagLen is the byte length of an OpenType feature tag.
const fontTagLen = 4

// maxFeatureFields is a feature tag plus its optional value.
const maxFeatureFields = 2

// FontFeature is one OpenType feature setting of a text run.
type FontFeature struct {
	// Tag is the four-byte OpenType tag, such as "smcp".
	Tag   string
	Value uint32
}

// ParseFontFeatures reads the feature list the engine carries on a text op:
// comma-separated `"tag" value` entries such as `"smcp" 1, "liga" 0`. A bare
// tag means on, and on and off are accepted for the value. An entry whose tag
// is not four bytes or whose value is not a number is skipped rather than
// failing the run, so a malformed setting only loses that setting.
func ParseFontFeatures(list string) []FontFeature {
	if list == "" {
		return nil
	}
	var out []FontFeature
	for _, entry := range strings.Split(list, ",") {
		fields := strings.Fields(entry)
		if len(fields) == 0 || len(fields) > maxFeatureFields {
			continue
		}
		tag := strings.Trim(fields[0], `"'`)
		if len(tag) != fontTagLen {
			continue
		}
		value := uint32(1)
		if len(fields) == maxFeatureFields {
			switch fields[1] {
			case "on":
			case "off":
				value = 0
			default:
				n, convErr := strconv.ParseUint(fields[1], 10, 32)
				if convErr != nil {
					continue
				}
				value = uint32(n)
			}
		}
		out = append(out, FontFeature{Tag: tag, Value: value})
	}

	return out
}

// otLanguageTags maps OpenType language system tags (from CSS
// font-language-override) to BCP 47 language tags expected by Ebiten's
// text.GoTextFace.Language and go-text/typesetting. Several raw OpenType tags
// collide with unrelated BCP 47 codes when passed directly to language.Make
// (for example "SRB" is Sora rather than Serbian, "TRK" is Turkic collection,
// and "ROM" is Romany rather than Romanian).
//
//nolint:gochecknoglobals // read-only OpenType-to-BCP47 language tag lookup table
var otLanguageTags = map[string]string{
	"TRK": "tr", // Turkish
	"SRB": "sr", // Serbian
	"MKD": "mk", // Macedonian
	"AZE": "az", // Azerbaijani
	"ROM": "ro", // Romanian
	"MOL": "mo", // Moldavian
	"HUN": "hu", // Hungarian
	"POL": "pl", // Polish (legacy)
	"PLK": "pl", // Polish
	"CES": "cs", // Czech (legacy)
	"CSY": "cs", // Czech
	"SLV": "sl", // Slovenian
	"SKY": "sk", // Slovak
	"CAT": "ca", // Catalan
	"ESP": "es", // Spanish
	"GAL": "gl", // Galician
	"URD": "ur", // Urdu
	"SND": "sd", // Sindhi
	"KUR": "ku", // Kurdish
}

// ShapingLanguage converts a CSS font-language-override tag (an OpenType
// language system tag such as "SRB" or "TRK") into the BCP 47 language.Tag
// expected by text.GoTextFace. Empty and "normal" return language.Und.
func ShapingLanguage(tag string) language.Tag {
	tag = strings.TrimSpace(tag)
	if tag == "" || strings.EqualFold(tag, "normal") {
		return language.Und
	}
	if bcp47, ok := otLanguageTags[strings.ToUpper(tag)]; ok {
		return language.Make(bcp47)
	}
	return language.Make(tag)
}

// IdeographAlphaBoundary reports an ideograph next to a Latin letter in
// either order, the pair text-autospace opens a gap between. It mirrors the
// classifier layout measures with (blinkless internal/layout
// IsIdeographAlphaBoundary, which the public layout package does not export),
// so a painter that adds the gap lands each run where layout placed it.
func IdeographAlphaBoundary(a, b rune) bool {
	return ideographAlphaPair(a, b) || ideographAlphaPair(b, a)
}

func ideographAlphaPair(a, b rune) bool {
	return isIdeographic(a) && isLatinAlpha(b)
}

func isIdeographic(r rune) bool {
	switch {
	case r >= 0x3040 && r <= 0x30FF: // Hiragana and Katakana
		return true
	case r >= 0x3400 && r <= 0x9FFF: // CJK unified and extension A
		return true
	case r >= 0xF900 && r <= 0xFAFF: // CJK compatibility ideographs
		return true
	case r >= 0xFF66 && r <= 0xFF9D: // halfwidth Katakana
		return true
	default:
		return unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana)
	}
}

func isLatinAlpha(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

// softHyphen is U+00AD, which layout never opens an autospace gap beside.
const softHyphen = '­'

// AutospaceGapBetween is the gap, in the unit of gap, a run inserts between
// prev and next: gap at an ideograph and Latin boundary, else zero. Layout
// skips the start of a run and a soft hyphen on either side.
func AutospaceGapBetween(prev, next rune, gap float64) float64 {
	if gap == 0 || prev == 0 || next == 0 || prev == softHyphen || next == softHyphen {
		return 0
	}
	if IdeographAlphaBoundary(prev, next) {
		return gap
	}

	return 0
}

// LeadingAutospaceGap is the text-autospace gap, in points, that layout folded
// into the width of cur but did not draw. Layout splits a run into one text op
// per font, so an ideograph and Latin boundary can fall exactly between two
// ops: the second op's width includes the gap while its X does not move, and a
// painter that starts every op with no previous rune would drop the gap. It
// returns the gap when prev ends flush against cur on the same baseline and
// the last rune of prev and the first rune of cur form a boundary, else zero.
func LeadingAutospaceGap(prev, cur *layout.DisplayOp) float64 {
	if prev == nil || cur == nil || prev.Text == "" || cur.Text == "" {
		return 0
	}
	gap := cur.TextAutospaceGap()
	if gap == 0 {
		return 0
	}
	if abs(prev.X+prev.W-cur.X) > autospaceAbutTol || abs(prev.Y-cur.Y) > autospaceAbutTol {
		return 0
	}
	runes := []rune(prev.Text)
	last := runes[len(runes)-1]
	first := []rune(cur.Text)[0]

	return AutospaceGapBetween(last, first, gap)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}

	return v
}
