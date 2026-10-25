package cssanimation

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ErrColor reports a colour value a keyframe cannot parse. A malformed value
// fails the bind instead of snapping to the element's base paint.
var ErrColor = errors.New("cssanimation: invalid colour")

// ErrColorSpace reports a colour function whose interpolation space this phase
// does not implement. The declaration is left at its base paint and the binder
// counts one fidelity gap for it instead of failing the element.
var ErrColorSpace = errors.New("cssanimation: colour space is not supported yet")

// ErrCurrentColor reports currentColor on an element with no text operation to
// resolve it against. The binder counts it as a bind error.
var ErrCurrentColor = errors.New("cssanimation: currentColor needs an element text colour")

// ErrNumber reports a numeric keyframe value that does not parse.
var ErrNumber = errors.New("cssanimation: invalid numeric value")

// Colour parsing constants. Every form resolves to the same 0..1 sRGB scale,
// and CSS legacy colour interpolation stays in that space.
const (
	hexLenShort      = 3
	hexLenShortAlpha = 4
	hexLenFull       = 6
	hexLenFullAlpha  = 8
	hexRadix         = 16
	channelMax       = 255
	hexLetterOffset  = 10
	hueCircle        = 360
	degreesPerGrad   = 0.9
	degreesPerTurn   = 360
	radiansToDegrees = 180 / math.Pi

	minChannel = 0.0
	maxChannel = 1.0
	halfValue  = 0.5
	oneValue   = 1.0
	twoValue   = 2.0
	sixValue   = 6.0
	thirdValue = oneValue / 3
	twoThirds  = twoValue / 3
	sixthValue = oneValue / 6
)

// Color is one parsed CSS colour in sRGB floats with its own alpha. Current
// marks the currentColor keyword, which the binder resolves against the
// element's own text colour before sampling; it is not a number to
// interpolate.
type Color struct {
	R, G, B, A float64
	Current    bool
}

// parseColor parses the colour forms a keyframe may declare: #rgb, #rgba,
// #rrggbb, #rrggbbaa, rgb(), rgba(), hsl(), hsla(), the CSS named colours,
// transparent, and currentColor. The cascade keywords (inherit and its
// siblings) are an error: a keyframe has no element context to resolve them.
func parseColor(value string) (Color, error) {
	text := strings.ToLower(strings.TrimSpace(value))
	if text == "" {
		return Color{}, fmt.Errorf("%w: empty value", ErrColor)
	}
	switch text {
	case "transparent":
		return Color{A: minChannel}, nil
	case "currentcolor":
		return Color{A: maxChannel, Current: true}, nil
	case "inherit", "initial", "unset", "revert", "revert-layer":
		return Color{}, fmt.Errorf("%w: %q is not animatable in a keyframe", ErrColor, text)
	}
	if digits, ok := strings.CutPrefix(text, "#"); ok {
		return parseHex(digits)
	}
	if name, rest, ok := strings.Cut(text, "("); ok {
		if !strings.HasSuffix(rest, ")") {
			return Color{}, fmt.Errorf("%w: %q", ErrColor, value)
		}
		return parseFunction(name, strings.TrimSuffix(rest, ")"))
	}
	return namedColor(text)
}

func parseHex(digits string) (Color, error) {
	var r, g, b float64
	alpha := float64(channelMax)
	var ok1, ok2, ok3, ok4 bool
	switch len(digits) {
	case hexLenShort, hexLenShortAlpha:
		r, ok1 = hexPair(digits[0], digits[0])
		g, ok2 = hexPair(digits[1], digits[1])
		b, ok3 = hexPair(digits[2], digits[2])
		ok4 = true
		if len(digits) == hexLenShortAlpha {
			alpha, ok4 = hexPair(digits[3], digits[3])
		}
	case hexLenFull, hexLenFullAlpha:
		r, ok1 = hexPair(digits[0], digits[1])
		g, ok2 = hexPair(digits[2], digits[3])
		b, ok3 = hexPair(digits[4], digits[5])
		ok4 = true
		if len(digits) == hexLenFullAlpha {
			alpha, ok4 = hexPair(digits[6], digits[7])
		}
	default:
		return Color{}, fmt.Errorf("%w: #%s", ErrColor, digits)
	}
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return Color{}, fmt.Errorf("%w: #%s", ErrColor, digits)
	}
	return Color{
		R: r / channelMax, G: g / channelMax, B: b / channelMax, A: alpha / channelMax,
	}, nil
}

// hexPair reads two hex nibbles as one channel value in 0..255.
func hexPair(high, low byte) (float64, bool) {
	hi, ok := hexNibble(high)
	if !ok {
		return 0, false
	}
	lo, ok := hexNibble(low)
	if !ok {
		return 0, false
	}
	return hi*hexRadix + lo, true
}

func hexNibble(digit byte) (float64, bool) {
	switch {
	case digit >= '0' && digit <= '9':
		return float64(digit - '0'), true
	case digit >= 'a' && digit <= 'f':
		return float64(digit-'a') + hexLetterOffset, true
	}
	return 0, false
}

func parseFunction(name, body string) (Color, error) {
	switch name {
	case "rgb", "rgba":
		return parseRGB(body)
	case "hsl", "hsla":
		return parseHSL(body)
	case "color", "lab", "lch", "oklab", "oklch", "color-mix", "light-dark", "device-cmyk":
		return Color{}, fmt.Errorf("%w: %s()", ErrColorSpace, name)
	}
	return Color{}, fmt.Errorf("%w: %s()", ErrColor, name)
}

func parseRGB(body string) (Color, error) {
	channels, alpha, err := colorArgs(body)
	if err != nil {
		return Color{}, err
	}
	r, err := rgbChannel(channels[0])
	if err != nil {
		return Color{}, err
	}
	g, err := rgbChannel(channels[1])
	if err != nil {
		return Color{}, err
	}
	b, err := rgbChannel(channels[2])
	if err != nil {
		return Color{}, err
	}
	a, err := alphaValue(alpha)
	if err != nil {
		return Color{}, err
	}
	return Color{R: r, G: g, B: b, A: a}, nil
}

func parseHSL(body string) (Color, error) {
	channels, alpha, err := colorArgs(body)
	if err != nil {
		return Color{}, err
	}
	hue, err := hueDegrees(channels[0])
	if err != nil {
		return Color{}, err
	}
	saturation, err := percentage(channels[1], "saturation")
	if err != nil {
		return Color{}, err
	}
	lightness, err := percentage(channels[2], "lightness")
	if err != nil {
		return Color{}, err
	}
	a, err := alphaValue(alpha)
	if err != nil {
		return Color{}, err
	}
	r, g, b := hslToRGB(hue, saturation, lightness)
	return Color{R: r, G: g, B: b, A: a}, nil
}

// colorArgs splits the inside of a colour function into three channel strings
// and an optional alpha string. Both the legacy comma form and the modern
// space form with a slash alpha are accepted.
const colorArgsWithAlpha = 4

func colorArgs(body string) ([3]string, string, error) {
	var channels [3]string
	body = strings.TrimSpace(body)
	if body == "" {
		return channels, "", fmt.Errorf("%w: empty arguments", ErrColor)
	}
	if strings.Contains(body, ",") {
		parts := strings.Split(body, ",")
		if len(parts) == colorArgsWithAlpha {
			alpha := strings.TrimSpace(parts[3])
			parts = parts[:3]
			if err := copyChannels(&channels, parts); err != nil {
				return channels, "", err
			}
			return channels, alpha, nil
		}
		if err := copyChannels(&channels, parts); err != nil {
			return channels, "", err
		}
		return channels, "", nil
	}
	channelPart, alphaPart, hasAlpha := strings.Cut(body, "/")
	fields := strings.Fields(channelPart)
	if err := copyChannels(&channels, fields); err != nil {
		return channels, "", err
	}
	if !hasAlpha {
		return channels, "", nil
	}
	return channels, strings.TrimSpace(alphaPart), nil
}

func copyChannels(channels *[3]string, parts []string) error {
	if len(parts) != len(channels) {
		return fmt.Errorf("%w: want 3 channels, got %d", ErrColor, len(parts))
	}
	for i := range channels {
		part := strings.TrimSpace(parts[i])
		if part == "" {
			return fmt.Errorf("%w: empty channel", ErrColor)
		}
		channels[i] = part
	}
	return nil
}

// rgbChannel parses one rgb() channel: a percentage of the full scale or a
// 0..255 number. The result is clamped, as CSS requires.
func rgbChannel(text string) (float64, error) {
	if digits, ok := strings.CutSuffix(text, "%"); ok {
		value, err := finite(digits)
		if err != nil {
			return 0, err
		}
		return clamp01(value / percentScale), nil
	}
	value, err := finite(text)
	if err != nil {
		return 0, err
	}
	return clamp01(value / channelMax), nil
}

// percentage parses a number with a required percent sign.
func percentage(text, name string) (float64, error) {
	digits, ok := strings.CutSuffix(strings.TrimSpace(text), "%")
	if !ok {
		return 0, fmt.Errorf("%w: %s needs a percentage", ErrColor, name)
	}
	value, err := finite(digits)
	if err != nil {
		return 0, err
	}
	return clamp01(value / percentScale), nil
}

// alphaValue parses an optional alpha channel. An empty alpha means opaque.
func alphaValue(text string) (float64, error) {
	if text == "" {
		return maxChannel, nil
	}
	if digits, ok := strings.CutSuffix(text, "%"); ok {
		value, err := finite(digits)
		if err != nil {
			return 0, err
		}
		return clamp01(value / percentScale), nil
	}
	value, err := finite(text)
	if err != nil {
		return 0, err
	}
	return clamp01(value), nil
}

// hueDegrees parses a hue with an optional angle unit.
func hueDegrees(text string) (float64, error) {
	text = strings.TrimSpace(text)
	switch {
	case strings.HasSuffix(text, "grad"):
		value, err := finite(strings.TrimSuffix(text, "grad"))
		return value * degreesPerGrad, err
	case strings.HasSuffix(text, "deg"):
		return finite(strings.TrimSuffix(text, "deg"))
	case strings.HasSuffix(text, "rad"):
		value, err := finite(strings.TrimSuffix(text, "rad"))
		return value * radiansToDegrees, err
	case strings.HasSuffix(text, "turn"):
		value, err := finite(strings.TrimSuffix(text, "turn"))
		return value * degreesPerTurn, err
	}
	return finite(text)
}

// hslToRGB converts a hue in degrees and the 0..1 saturation and lightness to
// sRGB channels. This is the CSS Color 4 algorithm for legacy hsl().
func hslToRGB(hue, saturation, lightness float64) (float64, float64, float64) {
	h := math.Mod(hue, hueCircle) / hueCircle
	if h < minChannel {
		h += maxChannel
	}
	if saturation == minChannel {
		return clamp01(lightness), clamp01(lightness), clamp01(lightness)
	}
	q := lightness * (oneValue + saturation)
	if lightness >= halfValue {
		q = lightness + saturation - lightness*saturation
	}
	p := twoValue*lightness - q
	return clamp01(hueChannel(p, q, h+thirdValue)),
		clamp01(hueChannel(p, q, h)),
		clamp01(hueChannel(p, q, h-thirdValue))
}

func hueChannel(p, q, t float64) float64 {
	if t < minChannel {
		t += maxChannel
	}
	if t > maxChannel {
		t -= maxChannel
	}
	switch {
	case t < sixthValue:
		return p + (q-p)*sixValue*t
	case t < halfValue:
		return q
	case t < twoThirds:
		return p + (q-p)*(twoThirds-t)*sixValue
	}
	return p
}

func clamp01(value float64) float64 {
	return math.Max(minChannel, math.Min(maxChannel, value))
}

// finite parses one number and rejects the non-finite forms ParseFloat
// accepts, so a NaN cannot reach a paint channel.
func finite(text string) (float64, error) {
	value, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("%w: %q is not a number", ErrColor, text)
	}
	return value, nil
}

// number parses a unitless number such as opacity. It rejects the non-finite
// forms ParseFloat accepts, so a NaN never reaches a paint field.
func number(value string) (float64, bool) {
	v, err := finite(value)
	return v, err == nil
}

// namedColourData is the CSS Color 4 named-colour list, one "name=rrggbb"
// entry per line. It stays a constant, so the package owns no mutable global
// table, and the scan runs at bind time only.
const namedColourData = `
aliceblue=f0f8ff
antiquewhite=faebd7
aqua=00ffff
aquamarine=7fffd4
azure=f0ffff
beige=f5f5dc
bisque=ffe4c4
black=000000
blanchedalmond=ffebcd
blue=0000ff
blueviolet=8a2be2
brown=a52a2a
burlywood=deb887
cadetblue=5f9ea0
chartreuse=7fff00
chocolate=d2691e
coral=ff7f50
cornflowerblue=6495ed
cornsilk=fff8dc
crimson=dc143c
cyan=00ffff
darkblue=00008b
darkcyan=008b8b
darkgoldenrod=b8860b
darkgray=a9a9a9
darkgreen=006400
darkgrey=a9a9a9
darkkhaki=bdb76b
darkmagenta=8b008b
darkolivegreen=556b2f
darkorange=ff8c00
darkorchid=9932cc
darkred=8b0000
darksalmon=e9967a
darkseagreen=8fbc8f
darkslateblue=483d8b
darkslategray=2f4f4f
darkslategrey=2f4f4f
darkturquoise=00ced1
darkviolet=9400d3
deeppink=ff1493
deepskyblue=00bfff
dimgray=696969
dimgrey=696969
dodgerblue=1e90ff
firebrick=b22222
floralwhite=fffaf0
forestgreen=228b22
fuchsia=ff00ff
gainsboro=dcdcdc
ghostwhite=f8f8ff
gold=ffd700
goldenrod=daa520
gray=808080
green=008000
greenyellow=adff2f
grey=808080
honeydew=f0fff0
hotpink=ff69b4
indianred=cd5c5c
indigo=4b0082
ivory=fffff0
khaki=f0e68c
lavender=e6e6fa
lavenderblush=fff0f5
lawngreen=7cfc00
lemonchiffon=fffacd
lightblue=add8e6
lightcoral=f08080
lightcyan=e0ffff
lightgoldenrodyellow=fafad2
lightgray=d3d3d3
lightgreen=90ee90
lightgrey=d3d3d3
lightpink=ffb6c1
lightsalmon=ffa07a
lightseagreen=20b2aa
lightskyblue=87cefa
lightslategray=778899
lightslategrey=778899
lightsteelblue=b0c4de
lightyellow=ffffe0
lime=00ff00
limegreen=32cd32
linen=faf0e6
magenta=ff00ff
maroon=800000
mediumaquamarine=66cdaa
mediumblue=0000cd
mediumorchid=ba55d3
mediumpurple=9370db
mediumseagreen=3cb371
mediumslateblue=7b68ee
mediumspringgreen=00fa9a
mediumturquoise=48d1cc
mediumvioletred=c71585
midnightblue=191970
mintcream=f5fffa
mistyrose=ffe4e1
moccasin=ffe4b5
navajowhite=ffdead
navy=000080
oldlace=fdf5e6
olive=808000
olivedrab=6b8e23
orange=ffa500
orangered=ff4500
orchid=da70d6
palegoldenrod=eee8aa
palegreen=98fb98
paleturquoise=afeeee
palevioletred=db7093
papayawhip=ffefd5
peachpuff=ffdab9
peru=cd853f
pink=ffc0cb
plum=dda0dd
powderblue=b0e0e6
purple=800080
rebeccapurple=663399
red=ff0000
rosybrown=bc8f8f
royalblue=4169e1
saddlebrown=8b4513
salmon=fa8072
sandybrown=f4a460
seagreen=2e8b57
seashell=fff5ee
sienna=a0522d
silver=c0c0c0
skyblue=87ceeb
slateblue=6a5acd
slategray=708090
slategrey=708090
snow=fffafa
springgreen=00ff7f
steelblue=4682b4
tan=d2b48c
teal=008080
thistle=d8bfd8
tomato=ff6347
turquoise=40e0d0
violet=ee82ee
wheat=f5deb3
white=ffffff
whitesmoke=f5f5f5
yellow=ffff00
yellowgreen=9acd32
`

// namedColor resolves one CSS named colour.
func namedColor(name string) (Color, error) {
	rest := namedColourData
	for rest != "" {
		var line string
		line, rest, _ = strings.Cut(rest, "\n")
		entry, digits, ok := strings.Cut(line, "=")
		if !ok || entry != name {
			continue
		}
		return parseHex(digits)
	}
	return Color{}, fmt.Errorf("%w: %q", ErrColor, name)
}
