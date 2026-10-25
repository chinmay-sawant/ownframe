package cssanimation

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/animation"
)

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
func parseTransform(value string, size transformSize) (animation.TransformValue, error) {
	functions, err := parseTransformFunctions(value, size)
	if err != nil {
		return animation.TransformValue{}, err
	}
	if len(functions) == 0 {
		return animation.TransformValue{Decomposed: animation.IdentityTransform()}, nil
	}
	threeD := false
	for _, f := range functions {
		threeD = threeD || is3DFunc(f.Name)
	}
	if threeD {
		// The 2D projection of a 3D list is not a matrix, but its translation
		// is still readable in X and Y until the restamp path lands.
		m := compose3D(functions)
		t := animation.IdentityTransform()
		t.Tx, t.Ty = m[matrix3DTx], m[matrix3DTy]
		return animation.TransformValue{Decomposed: t, Functions: functions, ThreeD: true}, nil
	}
	m := compose2D(functions)
	return animation.TransformValue{Decomposed: decompose2D(m), Functions: functions}, nil
}

// parseTransformFunctions tokenizes the whitespace-separated function list.
func parseTransformFunctions(value string, size transformSize) ([]animation.TransformFunction, error) {
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "none") {
		return nil, nil
	}
	if value == "" {
		return nil, transformErr("empty value")
	}
	var out []animation.TransformFunction
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
func parseTransformFunction(name animation.TransformFunc, body string, size transformSize) (animation.TransformFunction, error) {
	f := animation.TransformFunction{Name: name}
	args := splitTransformArgs(body)
	switch f.Name {
	case animation.Translate:
		values, err := transformLengths(args, size, oneArg, twoArgs, axisX, axisY)
		if err != nil {
			return f, err
		}
		f.Args = values
	case animation.TranslateX:
		values, err := transformLengths(args, size, oneArg, oneArg, axisX)
		if err != nil {
			return f, err
		}
		f.Args = values
	case animation.TranslateY:
		values, err := transformLengths(args, size, oneArg, oneArg, axisY)
		if err != nil {
			return f, err
		}
		f.Args = values
	case animation.Translate3D:
		values, err := transformLengths(args, size, threeArgs, threeArgs, axisX, axisY, axisZ)
		if err != nil {
			return f, err
		}
		f.Args = values
	case animation.TranslateZ:
		values, err := transformLengths(args, size, oneArg, oneArg, axisZ)
		if err != nil {
			return f, err
		}
		f.Args = values
	case animation.Scale:
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
	case animation.ScaleX, animation.ScaleY, animation.ScaleZ:
		x, err := transformSingleNumber(args, string(name))
		if err != nil {
			return f, err
		}
		f.Args = []float64{x}
	case animation.Scale3D:
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
	case animation.Rotate, animation.RotateX, animation.RotateY, animation.RotateZ, animation.SkewX, animation.SkewY:
		angle, err := transformAngle(args, string(name))
		if err != nil {
			return f, err
		}
		f.Args = []float64{angle}
	case animation.Rotate3D:
		values, err := transformRotate3D(args)
		if err != nil {
			return f, err
		}
		f.Args = values
	case animation.Skew:
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
	case animation.Matrix:
		values, err := transformNumbers(args, matrixArgs, string(name))
		if err != nil {
			return f, err
		}
		values[matrixE] *= size.pointsPerPixel
		values[matrixF] *= size.pointsPerPixel
		f.Args = values
	case animation.Matrix3D:
		values, err := transformNumbers(args, matrix3DArgs, string(name))
		if err != nil {
			return f, err
		}
		for _, i := range []int{matrix3DTx, matrix3DTy, matrix3DTz} {
			values[i] *= size.pointsPerPixel
		}
		f.Args = values
	case animation.Perspective:
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

func is3DFunc(name animation.TransformFunc) bool {
	switch name {
	case animation.Translate3D, animation.TranslateZ, animation.Scale3D, animation.ScaleZ,
		animation.RotateX, animation.RotateY, animation.Rotate3D, animation.Matrix3D,
		animation.Perspective:
		return true
	}
	return false
}

// canonicalFunc maps the lower-cased CSS function name to its constant. CSS
// function names are case-insensitive.
func canonicalFunc(name string) (animation.TransformFunc, bool) {
	switch name {
	case "translate":
		return animation.Translate, true
	case "translatex":
		return animation.TranslateX, true
	case "translatey":
		return animation.TranslateY, true
	case "translate3d":
		return animation.Translate3D, true
	case "translatez":
		return animation.TranslateZ, true
	case "scale":
		return animation.Scale, true
	case "scalex":
		return animation.ScaleX, true
	case "scaley":
		return animation.ScaleY, true
	case "scale3d":
		return animation.Scale3D, true
	case "scalez":
		return animation.ScaleZ, true
	case "rotate":
		return animation.Rotate, true
	case "rotatex":
		return animation.RotateX, true
	case "rotatey":
		return animation.RotateY, true
	case "rotatez":
		return animation.RotateZ, true
	case "rotate3d":
		return animation.Rotate3D, true
	case "skew":
		return animation.Skew, true
	case "skewx":
		return animation.SkewX, true
	case "skewy":
		return animation.SkewY, true
	case "matrix":
		return animation.Matrix, true
	case "matrix3d":
		return animation.Matrix3D, true
	case "perspective":
		return animation.Perspective, true
	}
	return "", false
}
