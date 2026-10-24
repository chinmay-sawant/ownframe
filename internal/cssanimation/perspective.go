package cssanimation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/animation"
)

// ErrPerspective reports a perspective or perspective-origin value the parser
// rejects. NewComposite keeps the first one as the bind error for the element,
// so a malformed rule is counted once at bind rather than dropped per frame.
var ErrPerspective = errors.New("cssanimation: invalid perspective value")

// perspectiveHalf divides the border box for the default 50% 50% origin.
const perspectiveHalf = 2

// defaultPerspective returns the undeclared perspective value: no distance and
// the element's default 50% 50% vanishing point. keyframes stores the resolved
// point on every frame so an implicit endpoint can interpolate against a
// declared origin.
func defaultPerspective(size transformSize) animation.PerspectiveValue {
	return animation.PerspectiveValue{
		OriginX: size.width / perspectiveHalf * size.pointsPerPixel,
		OriginY: size.height / perspectiveHalf * size.pointsPerPixel,
	}
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
// horizontal component and height for the vertical one. The two-value form
// accepts lengths and the left, right, top, bottom, and center keywords in
// either order; the three- and four-value edge-offset forms are not supported.
func parsePerspectiveOrigin(value string, size transformSize) (float64, float64, error) {
	fields := strings.Fields(strings.TrimSpace(value))
	if len(fields) == 0 {
		return 0, 0, perspectiveErr("empty value")
	}
	if len(fields) > twoArgs {
		return 0, 0, perspectiveErr("expected one or two position values, got %d", len(fields))
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
