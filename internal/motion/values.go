package motion

import (
	"errors"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
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

// ErrTransform reports a transform list the parser rejects. NewComposite
// keeps the first one as the bind error for the element, so a malformed rule
// is counted once at bind rather than dropped per frame.
var ErrTransform = errors.New("cssanimation: invalid transform list")

// Unit conversions and argument counts for the transform grammar.
const (
	percentScale = 100
	degToRad     = math.Pi / 180
	gradToRad    = math.Pi / 200
	turnToRad    = 2 * math.Pi
	matrixArgs   = 6
	matrix3DArgs = 16
	scaleArgs    = 3
	rotate3DArgs = 4
	twoArgs      = 2
	threeArgs    = 3
	oneArg       = 1
)

// lengthAxis selects the reference box side for a percentage.
type lengthAxis int

const (
	axisX lengthAxis = iota
	axisY
	axisZ
)

// transformSize resolves lengths while a transform list is parsed. The box
// is in CSS pixels and pointsPerPixel converts the result to display points.
// hasBox is false when no element box was supplied: a percentage then fails
// the bind instead of silently resolving to zero.
type transformSize struct {
	width, height  float64
	pointsPerPixel float64
	hasBox         bool
}

// parseTransform parses one CSS transform list. "none" yields the identity
// value, and a list with any 3D function keeps the parsed functions for the
// restamp path.
func parseTransform(value string, size transformSize) (TransformValue, error) {
	functions, err := parseTransformFunctions(value, size)
	if err != nil {
		return TransformValue{}, err
	}
	return transformValueFromFunctions(functions), nil
}

// transformValueFromFunctions folds a parsed function list into a
// TransformValue, keeping planar 2D lists on the fast 2D decomposition path.
func transformValueFromFunctions(functions []TransformFunction) TransformValue {
	if len(functions) == 0 {
		return TransformValue{Decomposed: IdentityTransform()}
	}
	threeD := false
	for _, f := range functions {
		threeD = threeD || is3DFunction(f)
	}
	if threeD {
		// The 2D projection of a 3D list is not a matrix, but its translation
		// is still readable in X and Y until the restamp path lands.
		m := compose3D(functions)
		t := IdentityTransform()
		t.Tx, t.Ty = m[matrix3DTx], m[matrix3DTy]
		return TransformValue{Decomposed: t, Functions: functions, ThreeD: true}
	}
	m := compose2D(functions)
	return TransformValue{Decomposed: decompose2D(m), Functions: functions}
}

// parseIndividualTranslate parses the CSS `translate` property:
// none | <length-percentage> | <length-percentage> <length-percentage> | <length-percentage> <length-percentage> <length>.
func parseIndividualTranslate(value string, size transformSize) ([]TransformFunction, error) {
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "none") {
		return nil, nil
	}
	fields := strings.Fields(value)
	if len(fields) == 0 || len(fields) > threeArgs {
		return nil, transformErr("translate takes 1 to 3 values")
	}
	if len(fields) <= twoArgs {
		vals, err := transformLengths(fields, size, oneArg, twoArgs, axisX, axisY)
		if err != nil {
			return nil, err
		}
		return []TransformFunction{{Name: Translate, Args: vals}}, nil
	}
	vals, err := transformLengths(fields, size, threeArgs, threeArgs, axisX, axisY, axisZ)
	if err != nil {
		return nil, err
	}
	return []TransformFunction{{Name: Translate3D, Args: vals}}, nil
}

// parseIndividualRotate parses the CSS `rotate` property:
// none | <angle> | [ x | y | z | <number>{3} ] && <angle>.
func parseIndividualRotate(value string) ([]TransformFunction, error) {
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "none") {
		return nil, nil
	}
	fields := strings.Fields(value)
	switch len(fields) {
	case oneArg:
		angle, err := transformAngleArg(fields[0])
		if err != nil {
			return nil, err
		}
		return []TransformFunction{{Name: Rotate, Args: []float64{angle}}}, nil
	case twoArgs:
		axis, angleToken := strings.ToLower(fields[0]), fields[1]
		if axis != "x" && axis != "y" && axis != "z" {
			axis, angleToken = strings.ToLower(fields[1]), fields[0]
		}
		angle, err := transformAngleArg(angleToken)
		if err != nil {
			return nil, err
		}
		switch axis {
		case "x":
			return []TransformFunction{{Name: RotateX, Args: []float64{angle}}}, nil
		case "y":
			return []TransformFunction{{Name: RotateY, Args: []float64{angle}}}, nil
		case "z":
			return []TransformFunction{{Name: RotateZ, Args: []float64{angle}}}, nil
		}
		return nil, transformErr("rotate: invalid axis %q", fields[0])
	case rotate3DArgs:
		vals, err := transformRotate3D(fields)
		if err != nil {
			reordered := []string{fields[1], fields[2], fields[3], fields[0]}
			if vals2, err2 := transformRotate3D(reordered); err2 == nil {
				vals, err = vals2, nil
			}
		}
		if err != nil {
			return nil, err
		}
		return []TransformFunction{{Name: Rotate3D, Args: vals}}, nil
	default:
		return nil, transformErr("rotate takes 1, 2, or 4 values")
	}
}

// parseIndividualScale parses the CSS `scale` property:
// none | [ <number> | <percentage> ]{1,3}.
func parseIndividualScale(value string) ([]TransformFunction, error) {
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "none") {
		return nil, nil
	}
	fields := strings.Fields(value)
	switch len(fields) {
	case oneArg:
		sx, err := transformScaleFactor(fields[0])
		if err != nil {
			return nil, err
		}
		return []TransformFunction{{Name: Scale, Args: []float64{sx, sx}}}, nil
	case twoArgs:
		sx, err := transformScaleFactor(fields[0])
		if err != nil {
			return nil, err
		}
		sy, err := transformScaleFactor(fields[1])
		if err != nil {
			return nil, err
		}
		return []TransformFunction{{Name: Scale, Args: []float64{sx, sy}}}, nil
	case threeArgs:
		sx, err := transformScaleFactor(fields[0])
		if err != nil {
			return nil, err
		}
		sy, err := transformScaleFactor(fields[1])
		if err != nil {
			return nil, err
		}
		sz, err := transformScaleFactor(fields[2])
		if err != nil {
			return nil, err
		}
		return []TransformFunction{{Name: Scale3D, Args: []float64{sx, sy, sz}}}, nil
	default:
		return nil, transformErr("scale takes 1 to 3 values")
	}
}

func transformScaleFactor(value string) (float64, error) {
	number, unit, err := transformValueUnit(value)
	if err != nil {
		return 0, err
	}
	switch unit {
	case "":
		return number, nil
	case "%":
		return number / percentScale, nil
	}
	return 0, transformErr("%q must be a number or percentage", value)
}

// parseTransformFunctions tokenizes the whitespace-separated function list.
func parseTransformFunctions(value string, size transformSize) ([]TransformFunction, error) {
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "none") {
		return nil, nil
	}
	if value == "" {
		return nil, transformErr("empty value")
	}
	var out []TransformFunction
	for i := 0; i < len(value); {
		for i < len(value) && isTransformSpace(value[i]) {
			i++
		}
		if i == len(value) {
			break
		}
		start := i
		for i < len(value) && isFunctionByte(value[i]) {
			i++
		}
		if i == start || i == len(value) || value[i] != '(' {
			return nil, transformErr("expected a function name at %q", value[start:])
		}
		name := strings.ToLower(value[start:i])
		canonical, ok := canonicalFunc(name)
		if !ok {
			return nil, transformErr("unknown function %q", name)
		}
		end := strings.IndexByte(value[i:], ')')
		if end < 0 {
			return nil, transformErr("%s is missing a closing parenthesis", name)
		}
		f, err := parseTransformFunction(canonical, value[i+1:i+end], size)
		if err != nil {
			return nil, err
		}
		i += end + 1
		out = append(out, f)
	}
	if len(out) == 0 {
		return nil, transformErr("no transform functions")
	}
	return out, nil
}

func isTransformSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f'
}

func isFunctionByte(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

// parseTransformFunction validates the argument count, resolves every unit,
// and normalizes optional arguments so matching lists compare equal for
// translate(10px) and translate(10px, 0px).
func parseTransformFunction(name TransformFunc, body string, size transformSize) (TransformFunction, error) {
	f := TransformFunction{Name: name}
	args := splitTransformArgs(body)
	switch f.Name {
	case Translate:
		values, err := transformLengths(args, size, oneArg, twoArgs, axisX, axisY)
		if err != nil {
			return f, err
		}
		f.Args = values
	case TranslateX:
		values, err := transformLengths(args, size, oneArg, oneArg, axisX)
		if err != nil {
			return f, err
		}
		f.Args = values
	case TranslateY:
		values, err := transformLengths(args, size, oneArg, oneArg, axisY)
		if err != nil {
			return f, err
		}
		f.Args = values
	case Translate3D:
		values, err := transformLengths(args, size, threeArgs, threeArgs, axisX, axisY, axisZ)
		if err != nil {
			return f, err
		}
		f.Args = values
	case TranslateZ:
		values, err := transformLengths(args, size, oneArg, oneArg, axisZ)
		if err != nil {
			return f, err
		}
		f.Args = values
	case Scale:
		f.Args = make([]float64, twoArgs)
		if len(args) < oneArg || len(args) > twoArgs {
			return f, transformErr("%s takes 1 or 2 arguments", name)
		}
		sx, err := transformNumber(args[0])
		if err != nil {
			return f, err
		}
		f.Args[0], f.Args[1] = sx, sx
		if len(args) == twoArgs {
			sy, err := transformNumber(args[1])
			if err != nil {
				return f, err
			}
			f.Args[1] = sy
		}
	case ScaleX, ScaleY, ScaleZ:
		x, err := transformSingleNumber(args, string(name))
		if err != nil {
			return f, err
		}
		f.Args = []float64{x}
	case Scale3D:
		if len(args) != scaleArgs {
			return f, transformErr("%s takes 3 arguments", name)
		}
		f.Args = make([]float64, scaleArgs)
		for i := range f.Args {
			x, err := transformNumber(args[i])
			if err != nil {
				return f, err
			}
			f.Args[i] = x
		}
	case Rotate, RotateX, RotateY, RotateZ, SkewX, SkewY:
		angle, err := transformAngle(args, string(name))
		if err != nil {
			return f, err
		}
		f.Args = []float64{angle}
	case Rotate3D:
		values, err := transformRotate3D(args)
		if err != nil {
			return f, err
		}
		f.Args = values
	case Skew:
		f.Args = make([]float64, twoArgs)
		if len(args) < oneArg || len(args) > twoArgs {
			return f, transformErr("%s takes 1 or 2 arguments", name)
		}
		ax, err := transformAngleArg(args[0])
		if err != nil {
			return f, err
		}
		f.Args[0] = ax
		if len(args) == twoArgs {
			ay, err := transformAngleArg(args[1])
			if err != nil {
				return f, err
			}
			f.Args[1] = ay
		}
	case Matrix:
		values, err := transformNumbers(args, matrixArgs, string(name))
		if err != nil {
			return f, err
		}
		values[matrixE] *= size.pointsPerPixel
		values[matrixF] *= size.pointsPerPixel
		f.Args = values
	case Matrix3D:
		values, err := transformNumbers(args, matrix3DArgs, string(name))
		if err != nil {
			return f, err
		}
		for _, i := range []int{matrix3DTx, matrix3DTy, matrix3DTz} {
			values[i] *= size.pointsPerPixel
		}
		f.Args = values
	case Perspective:
		values, err := transformLengths(args, size, oneArg, oneArg, axisZ)
		if err != nil {
			return f, err
		}
		if values[0] <= 0 {
			return f, transformErr("%s needs a positive length", name)
		}
		f.Args = values
	default:
		return f, transformErr("unknown function %q", name)
	}
	return f, nil
}

// transformLengths parses between min and max length arguments. A missing
// optional argument is zero.
func transformLengths(args []string, size transformSize, min, max int, axes ...lengthAxis) ([]float64, error) {
	if len(args) < min || len(args) > max || len(args) > len(axes) {
		return nil, transformErr("expected %d to %d length arguments", min, max)
	}
	out := make([]float64, len(axes))
	for i, axis := range axes {
		if i >= len(args) {
			continue
		}
		x, err := size.length(args[i], axis)
		if err != nil {
			return nil, err
		}
		out[i] = x
	}
	return out, nil
}

func transformSingleNumber(args []string, name string) (float64, error) {
	if len(args) != oneArg {
		return 0, transformErr("%s takes 1 argument", name)
	}
	return transformNumber(args[0])
}

func transformNumbers(args []string, want int, name string) ([]float64, error) {
	if len(args) != want {
		return nil, transformErr("%s takes %d arguments", name, want)
	}
	out := make([]float64, want)
	for i := range out {
		x, err := transformNumber(args[i])
		if err != nil {
			return nil, err
		}
		out[i] = x
	}
	return out, nil
}

func transformAngle(args []string, name string) (float64, error) {
	if len(args) != oneArg {
		return 0, transformErr("%s takes 1 argument", name)
	}
	return transformAngleArg(args[0])
}

func transformRotate3D(args []string) ([]float64, error) {
	if len(args) != rotate3DArgs {
		return nil, transformErr("rotate3d takes 4 arguments")
	}
	out := make([]float64, rotate3DArgs)
	for i := 0; i < threeArgs; i++ {
		x, err := transformNumber(args[i])
		if err != nil {
			return nil, err
		}
		out[i] = x
	}
	angle, err := transformAngleArg(args[threeArgs])
	if err != nil {
		return nil, err
	}
	out[threeArgs] = angle
	return out, nil
}

func splitTransformArgs(body string) []string {
	parts := strings.Split(body, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// length resolves a CSS length for one axis. Percentages need a box, and the
// z axis has no percentage reference.
func (size transformSize) length(value string, axis lengthAxis) (float64, error) {
	number, unit, err := transformValueUnit(value)
	if err != nil {
		return 0, err
	}
	switch unit {
	case "px":
		return number * size.pointsPerPixel, nil
	case "":
		if number == 0 {
			return 0, nil
		}
		return 0, transformErr("length %q needs px or a percentage", value)
	case "%":
		if axis == axisZ {
			return 0, transformErr("a percentage cannot resolve on the z axis")
		}
		if !size.hasBox {
			return 0, transformErr("percentage %q needs an element box", value)
		}
		reference := size.width
		if axis == axisY {
			reference = size.height
		}
		return number / percentScale * reference * size.pointsPerPixel, nil
	}
	return 0, transformErr("length %q has an unsupported unit %q", value, unit)
}

func transformNumber(value string) (float64, error) {
	number, unit, err := transformValueUnit(value)
	if err != nil {
		return 0, err
	}
	if unit != "" {
		return 0, transformErr("%q must be a unitless number", value)
	}
	return number, nil
}

func transformAngleArg(value string) (float64, error) {
	number, unit, err := transformValueUnit(value)
	if err != nil {
		return 0, err
	}
	switch unit {
	case "deg":
		return number * degToRad, nil
	case "grad":
		return number * gradToRad, nil
	case "rad":
		return number, nil
	case "turn":
		return number * turnToRad, nil
	case "":
		if number == 0 {
			return 0, nil
		}
		return 0, transformErr("angle %q needs deg, grad, rad, or turn", value)
	}
	return 0, transformErr("angle %q has an unsupported unit %q", value, unit)
}

// transformValueUnit splits a number from its unit suffix. A unitless value
// keeps an empty unit so the caller can decide whether zero is allowed.
func transformValueUnit(value string) (float64, string, error) {
	value = strings.TrimSpace(value)
	lower := strings.ToLower(value)
	unit := ""
	for _, candidate := range []string{"px", "deg", "grad", "rad", "turn", "%"} {
		if strings.HasSuffix(lower, candidate) {
			unit = candidate
			break
		}
	}
	number, err := strconv.ParseFloat(strings.TrimSpace(value[:len(value)-len(unit)]), 64)
	if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
		return 0, "", transformErr("%q is not a number", value)
	}
	return number, unit, nil
}

func transformErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrTransform, fmt.Sprintf(format, args...))
}

// is3DFunction reports whether a parsed transform function actually leaves the
// 2D plane. Planar translate3d(x, y, 0), translateZ(0), scale3d(sx, sy, 1),
// and scaleZ(1) stay on the 2D affine path so they can promote to compositor
// layers without forcing RestampTransforms every frame.
func is3DFunction(f TransformFunction) bool {
	switch f.Name {
	case Translate3D:
		return len(f.Args) >= threeArgs && f.Args[2] != 0
	case TranslateZ:
		return len(f.Args) >= oneArg && f.Args[0] != 0
	case Scale3D:
		return len(f.Args) >= scaleArgs && f.Args[2] != 1
	case ScaleZ:
		return len(f.Args) >= oneArg && f.Args[0] != 1
	default:
		return is3DFunc(f.Name)
	}
}

func is3DFunc(name TransformFunc) bool {
	switch name {
	case Translate3D, TranslateZ, Scale3D, ScaleZ,
		RotateX, RotateY, Rotate3D, Matrix3D,
		Perspective:
		return true
	}
	return false
}

// canonicalFunc maps the lower-cased CSS function name to its constant. CSS
// function names are case-insensitive.
func canonicalFunc(name string) (TransformFunc, bool) {
	switch name {
	case "translate":
		return Translate, true
	case "translatex":
		return TranslateX, true
	case "translatey":
		return TranslateY, true
	case "translate3d":
		return Translate3D, true
	case "translatez":
		return TranslateZ, true
	case "scale":
		return Scale, true
	case "scalex":
		return ScaleX, true
	case "scaley":
		return ScaleY, true
	case "scale3d":
		return Scale3D, true
	case "scalez":
		return ScaleZ, true
	case "rotate":
		return Rotate, true
	case "rotatex":
		return RotateX, true
	case "rotatey":
		return RotateY, true
	case "rotatez":
		return RotateZ, true
	case "rotate3d":
		return Rotate3D, true
	case "skew":
		return Skew, true
	case "skewx":
		return SkewX, true
	case "skewy":
		return SkewY, true
	case "matrix":
		return Matrix, true
	case "matrix3d":
		return Matrix3D, true
	case "perspective":
		return Perspective, true
	}
	return "", false
}

// ErrPerspective reports a perspective or perspective-origin value the parser
// rejects. NewComposite keeps the first one as the bind error for the element,
// so a malformed rule is counted once at bind rather than dropped per frame.
var ErrPerspective = errors.New("cssanimation: invalid perspective value")

// perspectiveHalf divides the border box for the default 50% 50% origin.
const perspectiveHalf = 2

// maxOriginArgs is the four-value edge-offset perspective-origin form.
const maxOriginArgs = 4

// perspectiveBaseFromTransform resolves the static camera an implicit keyframe
// endpoint holds. TransformInfo exports a resolved origin only alongside a
// declared perspective, so an element without one falls back to the CSS initial
// 50% 50% point against its own border box.
func perspectiveBaseFromTransform(size transformSize, info layout.TransformInfo) PerspectiveBase {
	base := PerspectiveBase{
		OriginX: size.width / perspectiveHalf * size.pointsPerPixel,
		OriginY: size.height / perspectiveHalf * size.pointsPerPixel,
		Set:     true,
	}
	if !info.HasPerspective {
		return base
	}
	base.Distance = info.Perspective
	base.HasDistance = info.Perspective > 0
	base.OriginX = info.PerspectiveOriginX
	base.OriginY = info.PerspectiveOriginY

	return base
}

// defaultPerspective returns the undeclared perspective value: no distance and
// the element's default 50% 50% vanishing point. keyframes stores the resolved
// point on every frame so an implicit endpoint can interpolate against a
// declared origin.
func defaultPerspective(size transformSize) PerspectiveValue {
	return PerspectiveValue{
		OriginX: size.width / perspectiveHalf * size.pointsPerPixel,
		OriginY: size.height / perspectiveHalf * size.pointsPerPixel,
	}
}

// framePerspective returns the origin every parsed frame starts with: the
// element's static vanishing point when the binder supplied one, otherwise the
// box default. A declared perspective-origin overwrites it.
func framePerspective(size transformSize, base PerspectiveBase) PerspectiveValue {
	value := defaultPerspective(size)
	if base.Set {
		value.OriginX, value.OriginY = base.OriginX, base.OriginY
	}

	return value
}

// parsePerspective parses perspective: none | <length>. The distance is
// returned in display points; zero means none. A negative distance is invalid
// CSS and fails the bind.
func parsePerspective(value string, size transformSize) (float64, error) {
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "none") {
		return 0, nil
	}
	if value == "" {
		return 0, perspectiveErr("empty value")
	}
	distance, err := size.length(value, axisZ)
	if err != nil {
		return 0, err
	}
	if distance < 0 {
		return 0, perspectiveErr("%q must not be negative", value)
	}

	return distance, nil
}

// parsePerspectiveOrigin parses perspective-origin: <position>. It returns the
// vanishing point in display points as an offset from the element's border-box
// origin. A percentage resolves against the element box: width for the
// horizontal component and height for the vertical one.
//
// The one- and two-value forms accept lengths and the left, right, top, bottom,
// and center keywords in either order. The three-value form is an edge keyword,
// its offset, and the opposite-axis edge keyword; the four-value form is two
// edge-offset pairs, either axis first. Center cannot take an offset and is
// rejected in those forms, as CSS does.
func parsePerspectiveOrigin(value string, size transformSize) (float64, float64, error) {
	fields := strings.Fields(strings.TrimSpace(value))
	if len(fields) == 0 {
		return 0, 0, perspectiveErr("empty value")
	}
	if len(fields) > maxOriginArgs {
		return 0, 0, perspectiveErr("expected one to four position values, got %d", len(fields))
	}
	if len(fields) >= threeArgs {
		return parseEdgeOrigin(fields, size)
	}
	terms := make([]originTerm, len(fields))
	for i, field := range fields {
		terms[i] = originTerm{keyword: originKeyword(field), raw: field}
	}
	xTerm, yTerm, err := assignOriginTerms(terms)
	if err != nil {
		return 0, 0, err
	}
	x, err := size.originCoordinate(xTerm, axisX)
	if err != nil {
		return 0, 0, err
	}
	y, err := size.originCoordinate(yTerm, axisY)
	if err != nil {
		return 0, 0, err
	}

	return x, y, nil
}

// parseEdgeOrigin resolves the three- and four-value edge-offset forms. Every
// keyword names an edge, the offset follows the edge it belongs to, and the two
// axes must differ. An offset is a length or percentage; unitless zero is the
// only unitless form the length parser accepts.
func parseEdgeOrigin(fields []string, size transformSize) (float64, float64, error) {
	firstAxis, err := edgeAxis(fields[0])
	if err != nil {
		return 0, 0, err
	}
	firstOffset, err := size.edgeOffset(fields[1], firstAxis)
	if err != nil {
		return 0, 0, err
	}
	secondAxis, err := edgeAxis(fields[2])
	if err != nil {
		return 0, 0, err
	}
	if secondAxis == firstAxis {
		return 0, 0, perspectiveErr("%q and %q name the same axis", fields[0], fields[2])
	}
	secondOffset := 0.0
	if len(fields) == maxOriginArgs {
		secondOffset, err = size.edgeOffset(fields[3], secondAxis)
		if err != nil {
			return 0, 0, err
		}
	}
	first, err := size.edgeCoordinate(fields[0], firstOffset, firstAxis)
	if err != nil {
		return 0, 0, err
	}
	second, err := size.edgeCoordinate(fields[2], secondOffset, secondAxis)
	if err != nil {
		return 0, 0, err
	}
	if firstAxis == axisHorizontal {
		return first, second, nil
	}

	return second, first, nil
}

// edgeAxis returns the axis an edge keyword constrains. Center and a length are
// not edges, so the edge-offset forms reject them.
func edgeAxis(value string) (positionAxis, error) {
	axis := keywordAxis(originKeyword(value))
	if axis == axisAny {
		return axisAny, perspectiveErr("%q must name a left, right, top, or bottom edge", value)
	}

	return axis, nil
}

// edgeOffset resolves one edge offset on its axis. A keyword in offset position
// is an error, not a zero.
func (size transformSize) edgeOffset(value string, axis positionAxis) (float64, error) {
	if originKeyword(value) != "" {
		return 0, perspectiveErr("%q is not an offset", value)
	}
	referenceAxis := axisX
	if axis == axisVertical {
		referenceAxis = axisY
	}

	return size.length(value, referenceAxis)
}

// edgeCoordinate places an edge-offset pair against the border box. The
// coordinate is measured from the left or top edge, so a right or bottom edge
// subtracts the offset from the box's width or height.
func (size transformSize) edgeCoordinate(keyword string, offset float64, axis positionAxis) (float64, error) {
	metric := size.width
	if axis == axisVertical {
		metric = size.height
	}
	metric *= size.pointsPerPixel
	switch keyword {
	case "left", "top":
		return offset, nil
	case "right", "bottom":
		return metric - offset, nil
	}

	return 0, perspectiveErr("unexpected edge %q", keyword)
}

// originTerm is one whitespace-separated perspective-origin component: either
// a keyword or a length-or-percentage token whose axis is decided after every
// term is classified.
type originTerm struct {
	keyword string
	raw     string
}

// positionAxis reports the axis a keyword constrains.
type positionAxis uint8

const (
	axisAny positionAxis = iota
	axisHorizontal
	axisVertical
)

// originKeyword returns the lower-cased keyword, or "" for a length token.
func originKeyword(value string) string {
	lower := strings.ToLower(value)
	switch lower {
	case "left", "right", "top", "bottom", "center":
		return lower
	}

	return ""
}

func keywordAxis(keyword string) positionAxis {
	switch keyword {
	case "left", "right":
		return axisHorizontal
	case "top", "bottom":
		return axisVertical
	}

	return axisAny
}

// assignOriginTerms maps one or two position components to the horizontal and
// vertical axis. It follows the CSS rule that a keyword naming a single axis
// determines that axis wherever it appears, and the remaining component takes
// the other one.
func assignOriginTerms(terms []originTerm) (originTerm, originTerm, error) {
	if len(terms) == oneArg {
		term := terms[0]
		if keywordAxis(term.keyword) == axisVertical {
			return originTerm{keyword: "center"}, term, nil
		}

		return term, originTerm{keyword: "center"}, nil
	}
	first, second := terms[0], terms[1]
	firstAxis, secondAxis := keywordAxis(first.keyword), keywordAxis(second.keyword)
	switch {
	case firstAxis == axisHorizontal && secondAxis == axisVertical:
		return first, second, nil
	case firstAxis == axisVertical && secondAxis == axisHorizontal:
		return second, first, nil
	case firstAxis == axisHorizontal:
		if secondAxis == axisHorizontal {
			return originTerm{}, originTerm{}, perspectiveErr("both values name a horizontal position")
		}

		return first, second, nil
	case firstAxis == axisVertical:
		if secondAxis == axisVertical {
			return originTerm{}, originTerm{}, perspectiveErr("both values name a vertical position")
		}

		return second, first, nil
	case secondAxis == axisHorizontal:
		return second, first, nil
	case secondAxis == axisVertical:
		return first, second, nil
	}

	return first, second, nil
}

// originCoordinate resolves one position component on its axis. Percentages
// and the keywords measure against the border box in display points.
func (size transformSize) originCoordinate(term originTerm, axis lengthAxis) (float64, error) {
	if axis == axisX && keywordAxis(term.keyword) == axisVertical {
		return 0, perspectiveErr("%s is not a horizontal position", term.keyword)
	}
	if axis == axisY && keywordAxis(term.keyword) == axisHorizontal {
		return 0, perspectiveErr("%s is not a vertical position", term.keyword)
	}
	reference := size.width
	if axis == axisY {
		reference = size.height
	}
	metric := reference * size.pointsPerPixel
	switch term.keyword {
	case "":
		return size.length(term.raw, axis)
	case "center":
		return metric / perspectiveHalf, nil
	case "left", "top":
		return 0, nil
	case "right", "bottom":
		return metric, nil
	}

	return 0, perspectiveErr("unexpected keyword %q", term.keyword)
}

func perspectiveErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrPerspective, fmt.Sprintf(format, args...))
}
