// Code generated from documentation/compatibility-matrix.md.
// DO NOT EDIT: run go generate ./ui after updating the matrix.
package ui

import (
	"fmt"
	"strings"
)

// SupportsProperty reports whether a CSS property is marked Implemented in
// the pinned Blinkless compatibility snapshot. This is name coverage only;
// individual value and layout semantics are enforced by Blinkless.
func SupportsProperty(property string) bool {
	switch strings.ToLower(strings.TrimSpace(property)) {
	case "-webkit-animation", "-webkit-animation-delay", "-webkit-animation-direction", "-webkit-animation-duration", "-webkit-animation-fill-mode",
		"-webkit-animation-iteration-count", "-webkit-animation-name", "-webkit-animation-play-state", "-webkit-animation-timing-function", "-webkit-appearance",
		"-webkit-backface-visibility", "-webkit-background-clip", "-webkit-background-origin", "-webkit-background-size", "-webkit-mask",
		"-webkit-mask-box-image", "-webkit-mask-box-image-outset", "-webkit-mask-box-image-repeat", "-webkit-mask-box-image-slice", "-webkit-mask-box-image-source",
		"-webkit-mask-box-image-width", "-webkit-mask-clip", "-webkit-mask-composite", "-webkit-mask-image", "-webkit-mask-origin",
		"-webkit-mask-position", "-webkit-mask-repeat", "-webkit-mask-size", "-webkit-perspective", "-webkit-perspective-origin",
		"-webkit-text-size-adjust", "-webkit-text-stroke", "-webkit-text-stroke-color", "-webkit-text-stroke-width", "-webkit-transform-style",
		"-webkit-transition", "-webkit-transition-delay", "-webkit-transition-duration", "-webkit-transition-property", "-webkit-transition-timing-function",
		"-webkit-user-select", "accent-color", "align-content", "align-items", "align-self",
		"alignment-baseline", "all", "anchor-name", "anchor-scope", "animation",
		"animation-composition", "animation-delay", "animation-delay-end", "animation-delay-start", "animation-direction",
		"animation-duration", "animation-fill-mode", "animation-iteration-count", "animation-name", "animation-play-state",
		"animation-range", "animation-range-center", "animation-range-end", "animation-range-start", "animation-timeline",
		"animation-timing-function", "animation-trigger", "aspect-ratio", "backdrop-filter", "backface-visibility",
		"background", "background-attachment", "background-blend-mode", "background-clip", "background-color",
		"background-image", "background-origin", "background-position", "background-position-block", "background-position-inline",
		"background-position-x", "background-position-y", "background-repeat", "background-repeat-block", "background-repeat-inline",
		"background-repeat-x", "background-repeat-y", "background-size", "background-tbd", "baseline-shift",
		"baseline-source", "block-ellipsis", "block-size", "bookmark-label", "bookmark-level",
		"bookmark-state", "border", "border-block", "border-block-clip", "border-block-color",
		"border-block-end", "border-block-end-clip", "border-block-end-color", "border-block-end-radius", "border-block-end-style",
		"border-block-end-width", "border-block-start", "border-block-start-clip", "border-block-start-color", "border-block-start-radius",
		"border-block-start-style", "border-block-start-width", "border-block-style", "border-block-width", "border-bottom",
		"border-bottom-clip", "border-bottom-color", "border-bottom-left-radius", "border-bottom-radius", "border-bottom-right-radius",
		"border-bottom-style", "border-bottom-width", "border-boundary", "border-clip", "border-collapse",
		"border-color", "border-end-end-radius", "border-end-start-radius", "border-image", "border-image-outset",
		"border-image-repeat", "border-image-slice", "border-image-source", "border-image-width", "border-inline",
		"border-inline-clip", "border-inline-color", "border-inline-end", "border-inline-end-clip", "border-inline-end-color",
		"border-inline-end-radius", "border-inline-end-style", "border-inline-end-width", "border-inline-start", "border-inline-start-clip",
		"border-inline-start-color", "border-inline-start-radius", "border-inline-start-style", "border-inline-start-width", "border-inline-style",
		"border-inline-width", "border-left", "border-left-clip", "border-left-color", "border-left-radius",
		"border-left-style", "border-left-width", "border-limit", "border-radius", "border-right",
		"border-right-clip", "border-right-color", "border-right-radius", "border-right-style", "border-right-width",
		"border-shape", "border-spacing", "border-start-end-radius", "border-start-start-radius", "border-style",
		"border-top", "border-top-clip", "border-top-color", "border-top-left-radius", "border-top-radius",
		"border-top-right-radius", "border-top-style", "border-top-width", "border-width", "bottom",
		"box-decoration-break", "box-shadow", "box-shadow-blur", "box-shadow-color", "box-shadow-inset",
		"box-shadow-offset", "box-shadow-position", "box-shadow-spread", "box-sizing", "break-after",
		"break-before", "break-inside", "caption-side", "caret", "caret-animation",
		"caret-shape", "clear", "clip", "clip-path", "clip-rule",
		"color", "color-adjust", "color-interpolation", "color-interpolation-filters", "color-scheme",
		"column-count", "column-fill", "column-gap", "column-height", "column-rule",
		"column-rule-break", "column-rule-color", "column-rule-inset", "column-rule-inset-cap", "column-rule-inset-cap-end",
		"column-rule-inset-cap-start", "column-rule-inset-end", "column-rule-inset-junction", "column-rule-inset-junction-end", "column-rule-inset-junction-start",
		"column-rule-inset-start", "column-rule-style", "column-rule-visibility-items", "column-rule-width", "column-span",
		"column-width", "column-wrap", "columns", "contain", "contain-intrinsic-block-size",
		"contain-intrinsic-height", "contain-intrinsic-inline-size", "contain-intrinsic-size", "contain-intrinsic-width", "container",
		"container-name", "container-type", "content", "content-visibility", "continue",
		"copy-into", "corner", "corner-block-end", "corner-block-end-shape", "corner-block-start",
		"corner-block-start-shape", "corner-bottom", "corner-bottom-left", "corner-bottom-left-shape", "corner-bottom-right",
		"corner-bottom-right-shape", "corner-bottom-shape", "corner-end-end", "corner-end-end-shape", "corner-end-start",
		"corner-end-start-shape", "corner-inline-end", "corner-inline-end-shape", "corner-inline-start", "corner-inline-start-shape",
		"corner-left", "corner-left-shape", "corner-right", "corner-right-shape", "corner-shape",
		"corner-start-end", "corner-start-end-shape", "corner-start-start", "corner-start-start-shape", "corner-top",
		"corner-top-left", "corner-top-left-shape", "corner-top-right", "corner-top-right-shape", "corner-top-shape",
		"counter-increment", "counter-reset", "counter-set", "cx", "cy",
		"d", "direction", "display", "dominant-baseline", "dynamic-range-limit",
		"empty-cells", "event-trigger", "event-trigger-name", "event-trigger-source", "field-sizing",
		"fill", "fill-opacity", "fill-rule", "filter", "flex",
		"flex-basis", "flex-direction", "flex-flow", "flex-grow", "flex-shrink",
		"flex-wrap", "float", "float-defer", "float-offset", "float-reference",
		"flood-color", "flood-opacity", "flow-from", "flow-into", "flow-tolerance",
		"font", "font-family", "font-feature-settings", "font-kerning", "font-language-override",
		"font-optical-sizing", "font-palette", "font-size", "font-size-adjust", "font-style",
		"font-synthesis", "font-synthesis-position", "font-synthesis-small-caps", "font-synthesis-style", "font-synthesis-weight",
		"font-variant", "font-variant-alternates", "font-variant-caps", "font-variant-east-asian", "font-variant-emoji",
		"font-variant-ligatures", "font-variant-numeric", "font-variant-position", "font-variation-settings", "font-weight",
		"font-width", "footnote-display", "footnote-policy", "forced-color-adjust", "gap",
		"glyph-orientation-vertical", "grid", "grid-area", "grid-auto-columns", "grid-auto-flow",
		"grid-auto-rows", "grid-column", "grid-column-end", "grid-column-start", "grid-row",
		"grid-row-end", "grid-row-start", "grid-template", "grid-template-areas", "grid-template-columns",
		"grid-template-rows", "hanging-punctuation", "height", "hyphenate-character", "hyphenate-limit-chars",
		"hyphenate-limit-last", "hyphenate-limit-lines", "hyphenate-limit-zone", "hyphens", "image-animation",
		"image-orientation", "image-rendering", "image-resolution", "initial-letter", "initial-letter-align",
		"initial-letter-wrap", "inline-size", "inline-sizing", "input-security", "inset",
		"inset-block", "inset-block-end", "inset-block-start", "inset-inline", "inset-inline-end",
		"inset-inline-start", "interactivity", "interest-delay", "interest-delay-end", "interest-delay-start",
		"isolation", "justify-content", "justify-items", "justify-self", "left",
		"letter-spacing", "lighting-color", "line-break", "line-clamp", "line-fit-edge",
		"line-height", "line-padding", "link-parameters", "list-style", "list-style-image",
		"list-style-position", "list-style-type", "margin", "margin-block", "margin-block-end",
		"margin-block-start", "margin-bottom", "margin-break", "margin-inline", "margin-inline-end",
		"margin-inline-start", "margin-left", "margin-right", "margin-top", "margin-trim",
		"marker", "marker-end", "marker-mid", "marker-side", "marker-start",
		"mask", "mask-border", "mask-border-mode", "mask-border-outset", "mask-border-repeat",
		"mask-border-slice", "mask-border-source", "mask-border-width", "mask-clip", "mask-composite",
		"mask-image", "mask-mode", "mask-origin", "mask-position", "mask-repeat",
		"mask-size", "mask-type", "max-block-size", "max-height", "max-inline-size",
		"max-lines", "max-size", "max-width", "min-block-size", "min-height",
		"min-inline-size", "min-size", "min-width", "mix-blend-mode", "nav-down",
		"nav-left", "nav-right", "nav-up", "object-fit", "object-position",
		"object-view-box", "offset", "offset-anchor", "offset-distance", "offset-path",
		"offset-position", "offset-rotate", "opacity", "order", "orphans",
		"outline", "outline-color", "outline-offset", "outline-style", "outline-width",
		"overflow", "overflow-block", "overflow-clip-margin", "overflow-clip-margin-block", "overflow-clip-margin-block-end",
		"overflow-clip-margin-block-start", "overflow-clip-margin-bottom", "overflow-clip-margin-inline", "overflow-clip-margin-inline-end", "overflow-clip-margin-inline-start",
		"overflow-clip-margin-left", "overflow-clip-margin-right", "overflow-clip-margin-top", "overflow-inline", "overflow-wrap",
		"overflow-x", "overflow-y", "overlay", "padding", "padding-block",
		"padding-block-end", "padding-block-start", "padding-bottom", "padding-inline", "padding-inline-end",
		"padding-inline-start", "padding-left", "padding-right", "padding-top", "paint-order",
		"path-length", "perspective", "perspective-origin", "place-content", "place-items",
		"place-self", "pointer-timeline", "pointer-timeline-axis", "pointer-timeline-name", "position",
		"position-anchor", "position-area", "position-try", "position-try-fallbacks", "position-try-order",
		"position-visibility", "print-color-adjust", "quotes", "r", "reading-flow",
		"reading-order", "region-fragment", "right", "rotate", "row-gap",
		"row-rule", "row-rule-break", "row-rule-color", "row-rule-inset", "row-rule-inset-cap",
		"row-rule-inset-cap-end", "row-rule-inset-cap-start", "row-rule-inset-end", "row-rule-inset-junction", "row-rule-inset-junction-end",
		"row-rule-inset-junction-start", "row-rule-inset-start", "row-rule-style", "row-rule-visibility-items", "row-rule-width",
		"ruby-align", "ruby-merge", "ruby-overhang", "ruby-position", "rule",
		"rule-break", "rule-color", "rule-inset", "rule-inset-cap", "rule-inset-end",
		"rule-inset-junction", "rule-inset-start", "rule-overlap", "rule-style", "rule-visibility-items",
		"rule-width", "rx", "ry", "scale", "scroll-margin",
		"scroll-margin-bottom", "scroll-margin-left", "scroll-margin-right", "scroll-margin-top", "scroll-timeline",
		"scroll-timeline-axis", "scroll-timeline-name", "shape-image-threshold", "shape-inside", "shape-margin",
		"shape-outside", "shape-padding", "shape-rendering", "size", "slider-orientation",
		"spatial-navigation-action", "spatial-navigation-contain", "spatial-navigation-function", "stop-color", "stop-opacity",
		"string-set", "stroke", "stroke-dasharray", "stroke-dashoffset", "stroke-linecap",
		"stroke-linejoin", "stroke-miterlimit", "stroke-opacity", "stroke-width", "tab-size",
		"table-layout", "text-align", "text-align-all", "text-align-last", "text-anchor",
		"text-autospace", "text-box", "text-box-edge", "text-box-trim", "text-combine-upright",
		"text-decoration", "text-decoration-color", "text-decoration-inset", "text-decoration-line", "text-decoration-skip",
		"text-decoration-skip-box", "text-decoration-skip-ink", "text-decoration-skip-self", "text-decoration-skip-spaces", "text-decoration-style",
		"text-decoration-thickness", "text-emphasis", "text-emphasis-color", "text-emphasis-position", "text-emphasis-skip",
		"text-emphasis-style", "text-fit", "text-group-align", "text-indent", "text-justify",
		"text-orientation", "text-overflow", "text-rendering", "text-shadow", "text-size-adjust",
		"text-spacing", "text-spacing-trim", "text-transform", "text-underline-offset", "text-underline-position",
		"text-wrap", "text-wrap-mode", "text-wrap-style", "timeline-scope", "timeline-trigger",
		"timeline-trigger-activation-range", "timeline-trigger-activation-range-end", "timeline-trigger-activation-range-start", "timeline-trigger-active-range", "timeline-trigger-active-range-end",
		"timeline-trigger-active-range-start", "timeline-trigger-name", "timeline-trigger-source", "top", "transform",
		"transform-box", "transform-origin", "transform-style", "transition", "transition-behavior",
		"transition-delay", "transition-duration", "transition-property", "transition-timing-function", "translate",
		"trigger-scope", "unicode-bidi", "vector-effect", "vertical-align", "view-timeline",
		"view-timeline-axis", "view-timeline-inset", "view-timeline-name", "view-transition-class", "view-transition-group",
		"view-transition-name", "view-transition-scope", "visibility", "white-space", "white-space-collapse",
		"white-space-trim", "widows", "width", "will-change", "window-drag",
		"word-break", "word-space-transform", "word-spacing", "wrap-after", "wrap-before",
		"wrap-flow", "wrap-inside", "wrap-through", "writing-mode", "x",
		"y", "z-index", "zoom":
		return true
	default:
		return false
	}
}

// SupportedProperties lists the implemented CSS property names, sorted.
func SupportedProperties() []string {
	return []string{
		"-webkit-animation", "-webkit-animation-delay", "-webkit-animation-direction", "-webkit-animation-duration", "-webkit-animation-fill-mode",
		"-webkit-animation-iteration-count", "-webkit-animation-name", "-webkit-animation-play-state", "-webkit-animation-timing-function", "-webkit-appearance",
		"-webkit-backface-visibility", "-webkit-background-clip", "-webkit-background-origin", "-webkit-background-size", "-webkit-mask",
		"-webkit-mask-box-image", "-webkit-mask-box-image-outset", "-webkit-mask-box-image-repeat", "-webkit-mask-box-image-slice", "-webkit-mask-box-image-source",
		"-webkit-mask-box-image-width", "-webkit-mask-clip", "-webkit-mask-composite", "-webkit-mask-image", "-webkit-mask-origin",
		"-webkit-mask-position", "-webkit-mask-repeat", "-webkit-mask-size", "-webkit-perspective", "-webkit-perspective-origin",
		"-webkit-text-size-adjust", "-webkit-text-stroke", "-webkit-text-stroke-color", "-webkit-text-stroke-width", "-webkit-transform-style",
		"-webkit-transition", "-webkit-transition-delay", "-webkit-transition-duration", "-webkit-transition-property", "-webkit-transition-timing-function",
		"-webkit-user-select", "accent-color", "align-content", "align-items", "align-self",
		"alignment-baseline", "all", "anchor-name", "anchor-scope", "animation",
		"animation-composition", "animation-delay", "animation-delay-end", "animation-delay-start", "animation-direction",
		"animation-duration", "animation-fill-mode", "animation-iteration-count", "animation-name", "animation-play-state",
		"animation-range", "animation-range-center", "animation-range-end", "animation-range-start", "animation-timeline",
		"animation-timing-function", "animation-trigger", "aspect-ratio", "backdrop-filter", "backface-visibility",
		"background", "background-attachment", "background-blend-mode", "background-clip", "background-color",
		"background-image", "background-origin", "background-position", "background-position-block", "background-position-inline",
		"background-position-x", "background-position-y", "background-repeat", "background-repeat-block", "background-repeat-inline",
		"background-repeat-x", "background-repeat-y", "background-size", "background-tbd", "baseline-shift",
		"baseline-source", "block-ellipsis", "block-size", "bookmark-label", "bookmark-level",
		"bookmark-state", "border", "border-block", "border-block-clip", "border-block-color",
		"border-block-end", "border-block-end-clip", "border-block-end-color", "border-block-end-radius", "border-block-end-style",
		"border-block-end-width", "border-block-start", "border-block-start-clip", "border-block-start-color", "border-block-start-radius",
		"border-block-start-style", "border-block-start-width", "border-block-style", "border-block-width", "border-bottom",
		"border-bottom-clip", "border-bottom-color", "border-bottom-left-radius", "border-bottom-radius", "border-bottom-right-radius",
		"border-bottom-style", "border-bottom-width", "border-boundary", "border-clip", "border-collapse",
		"border-color", "border-end-end-radius", "border-end-start-radius", "border-image", "border-image-outset",
		"border-image-repeat", "border-image-slice", "border-image-source", "border-image-width", "border-inline",
		"border-inline-clip", "border-inline-color", "border-inline-end", "border-inline-end-clip", "border-inline-end-color",
		"border-inline-end-radius", "border-inline-end-style", "border-inline-end-width", "border-inline-start", "border-inline-start-clip",
		"border-inline-start-color", "border-inline-start-radius", "border-inline-start-style", "border-inline-start-width", "border-inline-style",
		"border-inline-width", "border-left", "border-left-clip", "border-left-color", "border-left-radius",
		"border-left-style", "border-left-width", "border-limit", "border-radius", "border-right",
		"border-right-clip", "border-right-color", "border-right-radius", "border-right-style", "border-right-width",
		"border-shape", "border-spacing", "border-start-end-radius", "border-start-start-radius", "border-style",
		"border-top", "border-top-clip", "border-top-color", "border-top-left-radius", "border-top-radius",
		"border-top-right-radius", "border-top-style", "border-top-width", "border-width", "bottom",
		"box-decoration-break", "box-shadow", "box-shadow-blur", "box-shadow-color", "box-shadow-inset",
		"box-shadow-offset", "box-shadow-position", "box-shadow-spread", "box-sizing", "break-after",
		"break-before", "break-inside", "caption-side", "caret", "caret-animation",
		"caret-shape", "clear", "clip", "clip-path", "clip-rule",
		"color", "color-adjust", "color-interpolation", "color-interpolation-filters", "color-scheme",
		"column-count", "column-fill", "column-gap", "column-height", "column-rule",
		"column-rule-break", "column-rule-color", "column-rule-inset", "column-rule-inset-cap", "column-rule-inset-cap-end",
		"column-rule-inset-cap-start", "column-rule-inset-end", "column-rule-inset-junction", "column-rule-inset-junction-end", "column-rule-inset-junction-start",
		"column-rule-inset-start", "column-rule-style", "column-rule-visibility-items", "column-rule-width", "column-span",
		"column-width", "column-wrap", "columns", "contain", "contain-intrinsic-block-size",
		"contain-intrinsic-height", "contain-intrinsic-inline-size", "contain-intrinsic-size", "contain-intrinsic-width", "container",
		"container-name", "container-type", "content", "content-visibility", "continue",
		"copy-into", "corner", "corner-block-end", "corner-block-end-shape", "corner-block-start",
		"corner-block-start-shape", "corner-bottom", "corner-bottom-left", "corner-bottom-left-shape", "corner-bottom-right",
		"corner-bottom-right-shape", "corner-bottom-shape", "corner-end-end", "corner-end-end-shape", "corner-end-start",
		"corner-end-start-shape", "corner-inline-end", "corner-inline-end-shape", "corner-inline-start", "corner-inline-start-shape",
		"corner-left", "corner-left-shape", "corner-right", "corner-right-shape", "corner-shape",
		"corner-start-end", "corner-start-end-shape", "corner-start-start", "corner-start-start-shape", "corner-top",
		"corner-top-left", "corner-top-left-shape", "corner-top-right", "corner-top-right-shape", "corner-top-shape",
		"counter-increment", "counter-reset", "counter-set", "cx", "cy",
		"d", "direction", "display", "dominant-baseline", "dynamic-range-limit",
		"empty-cells", "event-trigger", "event-trigger-name", "event-trigger-source", "field-sizing",
		"fill", "fill-opacity", "fill-rule", "filter", "flex",
		"flex-basis", "flex-direction", "flex-flow", "flex-grow", "flex-shrink",
		"flex-wrap", "float", "float-defer", "float-offset", "float-reference",
		"flood-color", "flood-opacity", "flow-from", "flow-into", "flow-tolerance",
		"font", "font-family", "font-feature-settings", "font-kerning", "font-language-override",
		"font-optical-sizing", "font-palette", "font-size", "font-size-adjust", "font-style",
		"font-synthesis", "font-synthesis-position", "font-synthesis-small-caps", "font-synthesis-style", "font-synthesis-weight",
		"font-variant", "font-variant-alternates", "font-variant-caps", "font-variant-east-asian", "font-variant-emoji",
		"font-variant-ligatures", "font-variant-numeric", "font-variant-position", "font-variation-settings", "font-weight",
		"font-width", "footnote-display", "footnote-policy", "forced-color-adjust", "gap",
		"glyph-orientation-vertical", "grid", "grid-area", "grid-auto-columns", "grid-auto-flow",
		"grid-auto-rows", "grid-column", "grid-column-end", "grid-column-start", "grid-row",
		"grid-row-end", "grid-row-start", "grid-template", "grid-template-areas", "grid-template-columns",
		"grid-template-rows", "hanging-punctuation", "height", "hyphenate-character", "hyphenate-limit-chars",
		"hyphenate-limit-last", "hyphenate-limit-lines", "hyphenate-limit-zone", "hyphens", "image-animation",
		"image-orientation", "image-rendering", "image-resolution", "initial-letter", "initial-letter-align",
		"initial-letter-wrap", "inline-size", "inline-sizing", "input-security", "inset",
		"inset-block", "inset-block-end", "inset-block-start", "inset-inline", "inset-inline-end",
		"inset-inline-start", "interactivity", "interest-delay", "interest-delay-end", "interest-delay-start",
		"isolation", "justify-content", "justify-items", "justify-self", "left",
		"letter-spacing", "lighting-color", "line-break", "line-clamp", "line-fit-edge",
		"line-height", "line-padding", "link-parameters", "list-style", "list-style-image",
		"list-style-position", "list-style-type", "margin", "margin-block", "margin-block-end",
		"margin-block-start", "margin-bottom", "margin-break", "margin-inline", "margin-inline-end",
		"margin-inline-start", "margin-left", "margin-right", "margin-top", "margin-trim",
		"marker", "marker-end", "marker-mid", "marker-side", "marker-start",
		"mask", "mask-border", "mask-border-mode", "mask-border-outset", "mask-border-repeat",
		"mask-border-slice", "mask-border-source", "mask-border-width", "mask-clip", "mask-composite",
		"mask-image", "mask-mode", "mask-origin", "mask-position", "mask-repeat",
		"mask-size", "mask-type", "max-block-size", "max-height", "max-inline-size",
		"max-lines", "max-size", "max-width", "min-block-size", "min-height",
		"min-inline-size", "min-size", "min-width", "mix-blend-mode", "nav-down",
		"nav-left", "nav-right", "nav-up", "object-fit", "object-position",
		"object-view-box", "offset", "offset-anchor", "offset-distance", "offset-path",
		"offset-position", "offset-rotate", "opacity", "order", "orphans",
		"outline", "outline-color", "outline-offset", "outline-style", "outline-width",
		"overflow", "overflow-block", "overflow-clip-margin", "overflow-clip-margin-block", "overflow-clip-margin-block-end",
		"overflow-clip-margin-block-start", "overflow-clip-margin-bottom", "overflow-clip-margin-inline", "overflow-clip-margin-inline-end", "overflow-clip-margin-inline-start",
		"overflow-clip-margin-left", "overflow-clip-margin-right", "overflow-clip-margin-top", "overflow-inline", "overflow-wrap",
		"overflow-x", "overflow-y", "overlay", "padding", "padding-block",
		"padding-block-end", "padding-block-start", "padding-bottom", "padding-inline", "padding-inline-end",
		"padding-inline-start", "padding-left", "padding-right", "padding-top", "paint-order",
		"path-length", "perspective", "perspective-origin", "place-content", "place-items",
		"place-self", "pointer-timeline", "pointer-timeline-axis", "pointer-timeline-name", "position",
		"position-anchor", "position-area", "position-try", "position-try-fallbacks", "position-try-order",
		"position-visibility", "print-color-adjust", "quotes", "r", "reading-flow",
		"reading-order", "region-fragment", "right", "rotate", "row-gap",
		"row-rule", "row-rule-break", "row-rule-color", "row-rule-inset", "row-rule-inset-cap",
		"row-rule-inset-cap-end", "row-rule-inset-cap-start", "row-rule-inset-end", "row-rule-inset-junction", "row-rule-inset-junction-end",
		"row-rule-inset-junction-start", "row-rule-inset-start", "row-rule-style", "row-rule-visibility-items", "row-rule-width",
		"ruby-align", "ruby-merge", "ruby-overhang", "ruby-position", "rule",
		"rule-break", "rule-color", "rule-inset", "rule-inset-cap", "rule-inset-end",
		"rule-inset-junction", "rule-inset-start", "rule-overlap", "rule-style", "rule-visibility-items",
		"rule-width", "rx", "ry", "scale", "scroll-margin",
		"scroll-margin-bottom", "scroll-margin-left", "scroll-margin-right", "scroll-margin-top", "scroll-timeline",
		"scroll-timeline-axis", "scroll-timeline-name", "shape-image-threshold", "shape-inside", "shape-margin",
		"shape-outside", "shape-padding", "shape-rendering", "size", "slider-orientation",
		"spatial-navigation-action", "spatial-navigation-contain", "spatial-navigation-function", "stop-color", "stop-opacity",
		"string-set", "stroke", "stroke-dasharray", "stroke-dashoffset", "stroke-linecap",
		"stroke-linejoin", "stroke-miterlimit", "stroke-opacity", "stroke-width", "tab-size",
		"table-layout", "text-align", "text-align-all", "text-align-last", "text-anchor",
		"text-autospace", "text-box", "text-box-edge", "text-box-trim", "text-combine-upright",
		"text-decoration", "text-decoration-color", "text-decoration-inset", "text-decoration-line", "text-decoration-skip",
		"text-decoration-skip-box", "text-decoration-skip-ink", "text-decoration-skip-self", "text-decoration-skip-spaces", "text-decoration-style",
		"text-decoration-thickness", "text-emphasis", "text-emphasis-color", "text-emphasis-position", "text-emphasis-skip",
		"text-emphasis-style", "text-fit", "text-group-align", "text-indent", "text-justify",
		"text-orientation", "text-overflow", "text-rendering", "text-shadow", "text-size-adjust",
		"text-spacing", "text-spacing-trim", "text-transform", "text-underline-offset", "text-underline-position",
		"text-wrap", "text-wrap-mode", "text-wrap-style", "timeline-scope", "timeline-trigger",
		"timeline-trigger-activation-range", "timeline-trigger-activation-range-end", "timeline-trigger-activation-range-start", "timeline-trigger-active-range", "timeline-trigger-active-range-end",
		"timeline-trigger-active-range-start", "timeline-trigger-name", "timeline-trigger-source", "top", "transform",
		"transform-box", "transform-origin", "transform-style", "transition", "transition-behavior",
		"transition-delay", "transition-duration", "transition-property", "transition-timing-function", "translate",
		"trigger-scope", "unicode-bidi", "vector-effect", "vertical-align", "view-timeline",
		"view-timeline-axis", "view-timeline-inset", "view-timeline-name", "view-transition-class", "view-transition-group",
		"view-transition-name", "view-transition-scope", "visibility", "white-space", "white-space-collapse",
		"white-space-trim", "widows", "width", "will-change", "window-drag",
		"word-break", "word-space-transform", "word-spacing", "wrap-after", "wrap-before",
		"wrap-flow", "wrap-inside", "wrap-through", "writing-mode", "x",
		"y", "z-index", "zoom",
	}
}

// Value is a CSS value for properties without property-specific typed helpers.
type Value interface {
	cssValue
	isValue()
}

func (expression) isValue()  {}
func (wideKeyword) isValue() {}

type ColorValue interface {
	cssValue
	isColorValue()
}

func (expression) isColorValue()  {}
func (wideKeyword) isColorValue() {}
func (Color) isColorValue()       {}

type LengthPercentageValue interface {
	cssValue
	isLengthPercentageValue()
}

func (expression) isLengthPercentageValue()  {}
func (wideKeyword) isLengthPercentageValue() {}
func (Length) isLengthPercentageValue()      {}
func (Percentage) isLengthPercentageValue()  {}

type BorderStrokeValue interface {
	cssValue
	isBorderStrokeValue()
}

func (expression) isBorderStrokeValue()   {}
func (wideKeyword) isBorderStrokeValue()  {}
func (BorderStroke) isBorderStrokeValue() {}

type LengthValue interface {
	cssValue
	isLengthValue()
}

func (expression) isLengthValue()  {}
func (wideKeyword) isLengthValue() {}
func (Length) isLengthValue()      {}

type LengthPercentageSpacingValue interface {
	cssValue
	isLengthPercentageSpacingValue()
}

func (expression) isLengthPercentageSpacingValue()  {}
func (wideKeyword) isLengthPercentageSpacingValue() {}
func (Length) isLengthPercentageSpacingValue()      {}
func (Percentage) isLengthPercentageSpacingValue()  {}
func (Spacing) isLengthPercentageSpacingValue()     {}

type NumericValue interface {
	cssValue
	isNumericValue()
}

func (expression) isNumericValue()  {}
func (wideKeyword) isNumericValue() {}
func (Numeric) isNumericValue()     {}

type LengthNumericPercentageValue interface {
	cssValue
	isLengthNumericPercentageValue()
}

func (expression) isLengthNumericPercentageValue()  {}
func (wideKeyword) isLengthNumericPercentageValue() {}
func (Length) isLengthNumericPercentageValue()      {}
func (Numeric) isLengthNumericPercentageValue()     {}
func (Percentage) isLengthNumericPercentageValue()  {}

type TransformValue interface {
	cssValue
	isTransformValue()
}

func (expression) isTransformValue()  {}
func (wideKeyword) isTransformValue() {}
func (Transform) isTransformValue()   {}

type AlignContentValue interface {
	cssValue
	isAlignContent()
}
type alignContentKeyword uint16

const (
	AlignContentNormal alignContentKeyword = iota + 1
	AlignContentStart
	AlignContentEnd
	AlignContentFlexStart
	AlignContentFlexEnd
	AlignContentCenter
	AlignContentBaseline
	AlignContentStretch
	AlignContentSpaceBetween
	AlignContentSpaceAround
	AlignContentSpaceEvenly
)

func (expression) isAlignContent()          {}
func (wideKeyword) isAlignContent()         {}
func (alignContentKeyword) isAlignContent() {}
func (v alignContentKeyword) cssText() (string, error) {
	names := [...]string{"", "normal", "start", "end", "flex-start", "flex-end", "center", "baseline", "stretch", "space-between", "space-around", "space-evenly"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid align-content keyword %d", v)
	}
	return names[v], nil
}

type AlignItemsValue interface {
	cssValue
	isAlignItems()
}
type alignItemsKeyword uint16

const (
	AlignItemsNormal alignItemsKeyword = iota + 1
	AlignItemsStart
	AlignItemsEnd
	AlignItemsFlexStart
	AlignItemsFlexEnd
	AlignItemsCenter
	AlignItemsBaseline
	AlignItemsStretch
)

func (expression) isAlignItems()        {}
func (wideKeyword) isAlignItems()       {}
func (alignItemsKeyword) isAlignItems() {}
func (v alignItemsKeyword) cssText() (string, error) {
	names := [...]string{"", "normal", "start", "end", "flex-start", "flex-end", "center", "baseline", "stretch"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid align-items keyword %d", v)
	}
	return names[v], nil
}

type AlignSelfValue interface {
	cssValue
	isAlignSelf()
}
type alignSelfKeyword uint16

const (
	AlignSelfAuto alignSelfKeyword = iota + 1
	AlignSelfNormal
	AlignSelfStart
	AlignSelfEnd
	AlignSelfFlexStart
	AlignSelfFlexEnd
	AlignSelfCenter
	AlignSelfBaseline
	AlignSelfStretch
)

func (expression) isAlignSelf()       {}
func (wideKeyword) isAlignSelf()      {}
func (alignSelfKeyword) isAlignSelf() {}
func (v alignSelfKeyword) cssText() (string, error) {
	names := [...]string{"", "auto", "normal", "start", "end", "flex-start", "flex-end", "center", "baseline", "stretch"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid align-self keyword %d", v)
	}
	return names[v], nil
}

type BorderCollapseValue interface {
	cssValue
	isBorderCollapse()
}
type borderCollapseKeyword uint16

const (
	BorderCollapseCollapse borderCollapseKeyword = iota + 1
	BorderCollapseSeparate
)

func (expression) isBorderCollapse()            {}
func (wideKeyword) isBorderCollapse()           {}
func (borderCollapseKeyword) isBorderCollapse() {}
func (v borderCollapseKeyword) cssText() (string, error) {
	names := [...]string{"", "collapse", "separate"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid border-collapse keyword %d", v)
	}
	return names[v], nil
}

type BoxSizingValue interface {
	cssValue
	isBoxSizing()
}
type boxSizingKeyword uint16

const (
	BoxSizingBorderBox boxSizingKeyword = iota + 1
	BoxSizingContentBox
)

func (expression) isBoxSizing()       {}
func (wideKeyword) isBoxSizing()      {}
func (boxSizingKeyword) isBoxSizing() {}
func (v boxSizingKeyword) cssText() (string, error) {
	names := [...]string{"", "border-box", "content-box"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid box-sizing keyword %d", v)
	}
	return names[v], nil
}

type DisplayValue interface {
	cssValue
	isDisplay()
}
type displayKeyword uint16

const (
	DisplayNone displayKeyword = iota + 1
	DisplayBlock
	DisplayInline
	DisplayInlineBlock
	DisplayFlex
	DisplayInlineFlex
	DisplayGrid
	DisplayInlineGrid
	DisplayTable
	DisplayTableRow
	DisplayTableCell
	DisplayContents
)

func (expression) isDisplay()     {}
func (wideKeyword) isDisplay()    {}
func (displayKeyword) isDisplay() {}
func (v displayKeyword) cssText() (string, error) {
	names := [...]string{"", "none", "block", "inline", "inline-block", "flex", "inline-flex", "grid", "inline-grid", "table", "table-row", "table-cell", "contents"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid display keyword %d", v)
	}
	return names[v], nil
}

type FlexDirectionValue interface {
	cssValue
	isFlexDirection()
}
type flexDirectionKeyword uint16

const (
	FlexDirectionRow flexDirectionKeyword = iota + 1
	FlexDirectionRowReverse
	FlexDirectionColumn
	FlexDirectionColumnReverse
)

func (expression) isFlexDirection()           {}
func (wideKeyword) isFlexDirection()          {}
func (flexDirectionKeyword) isFlexDirection() {}
func (v flexDirectionKeyword) cssText() (string, error) {
	names := [...]string{"", "row", "row-reverse", "column", "column-reverse"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid flex-direction keyword %d", v)
	}
	return names[v], nil
}

type FlexWrapValue interface {
	cssValue
	isFlexWrap()
}
type flexWrapKeyword uint16

const (
	FlexWrapNowrap flexWrapKeyword = iota + 1
	FlexWrapWrap
	FlexWrapWrapReverse
)

func (expression) isFlexWrap()      {}
func (wideKeyword) isFlexWrap()     {}
func (flexWrapKeyword) isFlexWrap() {}
func (v flexWrapKeyword) cssText() (string, error) {
	names := [...]string{"", "nowrap", "wrap", "wrap-reverse"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid flex-wrap keyword %d", v)
	}
	return names[v], nil
}

type FontFamilyValue interface {
	cssValue
	isFontFamily()
}
type fontFamilyKeyword uint16

const (
	FontFamilySerif fontFamilyKeyword = iota + 1
	FontFamilySansSerif
	FontFamilyMonospace
)

func (expression) isFontFamily()        {}
func (wideKeyword) isFontFamily()       {}
func (fontFamilyKeyword) isFontFamily() {}
func (v fontFamilyKeyword) cssText() (string, error) {
	names := [...]string{"", "serif", "sans-serif", "monospace"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid font-family keyword %d", v)
	}
	return names[v], nil
}

type FontStyleValue interface {
	cssValue
	isFontStyle()
}
type fontStyleKeyword uint16

const (
	FontStyleNormal fontStyleKeyword = iota + 1
	FontStyleItalic
	FontStyleOblique
)

func (expression) isFontStyle()       {}
func (wideKeyword) isFontStyle()      {}
func (fontStyleKeyword) isFontStyle() {}
func (v fontStyleKeyword) cssText() (string, error) {
	names := [...]string{"", "normal", "italic", "oblique"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid font-style keyword %d", v)
	}
	return names[v], nil
}

type FontWeightValue interface {
	cssValue
	isFontWeight()
}
type fontWeightKeyword uint16

const (
	FontWeightNormal fontWeightKeyword = iota + 1
	FontWeightBold
	FontWeightBolder
	FontWeightLighter
)

func (expression) isFontWeight()        {}
func (wideKeyword) isFontWeight()       {}
func (fontWeightKeyword) isFontWeight() {}
func (Numeric) isFontWeight()           {}
func (v fontWeightKeyword) cssText() (string, error) {
	names := [...]string{"", "normal", "bold", "bolder", "lighter"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid font-weight keyword %d", v)
	}
	return names[v], nil
}

type HeightValue interface {
	cssValue
	isHeight()
}
type heightKeyword uint16

const (
	HeightAuto heightKeyword = iota + 1
	HeightMinContent
	HeightMaxContent
	HeightFitContent
)

func (expression) isHeight()    {}
func (wideKeyword) isHeight()   {}
func (heightKeyword) isHeight() {}
func (Length) isHeight()        {}
func (Percentage) isHeight()    {}
func (v heightKeyword) cssText() (string, error) {
	names := [...]string{"", "auto", "min-content", "max-content", "fit-content"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid height keyword %d", v)
	}
	return names[v], nil
}

type JustifyContentValue interface {
	cssValue
	isJustifyContent()
}
type justifyContentKeyword uint16

const (
	JustifyContentNormal justifyContentKeyword = iota + 1
	JustifyContentStart
	JustifyContentEnd
	JustifyContentFlexStart
	JustifyContentFlexEnd
	JustifyContentCenter
	JustifyContentSpaceBetween
	JustifyContentSpaceAround
	JustifyContentSpaceEvenly
	JustifyContentStretch
)

func (expression) isJustifyContent()            {}
func (wideKeyword) isJustifyContent()           {}
func (justifyContentKeyword) isJustifyContent() {}
func (v justifyContentKeyword) cssText() (string, error) {
	names := [...]string{"", "normal", "start", "end", "flex-start", "flex-end", "center", "space-between", "space-around", "space-evenly", "stretch"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid justify-content keyword %d", v)
	}
	return names[v], nil
}

type OverflowValue interface {
	cssValue
	isOverflow()
}
type overflowKeyword uint16

const (
	OverflowVisible overflowKeyword = iota + 1
	OverflowHidden
	OverflowClip
	OverflowScroll
	OverflowAuto
)

func (expression) isOverflow()      {}
func (wideKeyword) isOverflow()     {}
func (overflowKeyword) isOverflow() {}
func (v overflowKeyword) cssText() (string, error) {
	names := [...]string{"", "visible", "hidden", "clip", "scroll", "auto"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid overflow keyword %d", v)
	}
	return names[v], nil
}

type OverflowWrapValue interface {
	cssValue
	isOverflowWrap()
}
type overflowWrapKeyword uint16

const (
	OverflowWrapNormal overflowWrapKeyword = iota + 1
	OverflowWrapBreakWord
	OverflowWrapAnywhere
)

func (expression) isOverflowWrap()          {}
func (wideKeyword) isOverflowWrap()         {}
func (overflowWrapKeyword) isOverflowWrap() {}
func (v overflowWrapKeyword) cssText() (string, error) {
	names := [...]string{"", "normal", "break-word", "anywhere"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid overflow-wrap keyword %d", v)
	}
	return names[v], nil
}

type OverflowXValue interface {
	cssValue
	isOverflowX()
}
type overflowXKeyword uint16

const (
	OverflowXVisible overflowXKeyword = iota + 1
	OverflowXHidden
	OverflowXClip
	OverflowXScroll
	OverflowXAuto
)

func (expression) isOverflowX()       {}
func (wideKeyword) isOverflowX()      {}
func (overflowXKeyword) isOverflowX() {}
func (v overflowXKeyword) cssText() (string, error) {
	names := [...]string{"", "visible", "hidden", "clip", "scroll", "auto"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid overflow-x keyword %d", v)
	}
	return names[v], nil
}

type OverflowYValue interface {
	cssValue
	isOverflowY()
}
type overflowYKeyword uint16

const (
	OverflowYVisible overflowYKeyword = iota + 1
	OverflowYHidden
	OverflowYClip
	OverflowYScroll
	OverflowYAuto
)

func (expression) isOverflowY()       {}
func (wideKeyword) isOverflowY()      {}
func (overflowYKeyword) isOverflowY() {}
func (v overflowYKeyword) cssText() (string, error) {
	names := [...]string{"", "visible", "hidden", "clip", "scroll", "auto"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid overflow-y keyword %d", v)
	}
	return names[v], nil
}

type PositionValue interface {
	cssValue
	isPosition()
}
type positionKeyword uint16

const (
	PositionStatic positionKeyword = iota + 1
	PositionRelative
	PositionAbsolute
	PositionFixed
	PositionSticky
)

func (expression) isPosition()      {}
func (wideKeyword) isPosition()     {}
func (positionKeyword) isPosition() {}
func (v positionKeyword) cssText() (string, error) {
	names := [...]string{"", "static", "relative", "absolute", "fixed", "sticky"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid position keyword %d", v)
	}
	return names[v], nil
}

type TextAlignValue interface {
	cssValue
	isTextAlign()
}
type textAlignKeyword uint16

const (
	TextAlignStart textAlignKeyword = iota + 1
	TextAlignEnd
	TextAlignLeft
	TextAlignRight
	TextAlignCenter
	TextAlignJustify
)

func (expression) isTextAlign()       {}
func (wideKeyword) isTextAlign()      {}
func (textAlignKeyword) isTextAlign() {}
func (v textAlignKeyword) cssText() (string, error) {
	names := [...]string{"", "start", "end", "left", "right", "center", "justify"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid text-align keyword %d", v)
	}
	return names[v], nil
}

type TextDecorationValue interface {
	cssValue
	isTextDecoration()
}
type textDecorationKeyword uint16

const (
	TextDecorationNone textDecorationKeyword = iota + 1
	TextDecorationUnderline
	TextDecorationLineThrough
)

func (expression) isTextDecoration()            {}
func (wideKeyword) isTextDecoration()           {}
func (textDecorationKeyword) isTextDecoration() {}
func (v textDecorationKeyword) cssText() (string, error) {
	names := [...]string{"", "none", "underline", "line-through"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid text-decoration keyword %d", v)
	}
	return names[v], nil
}

type TextOverflowValue interface {
	cssValue
	isTextOverflow()
}
type textOverflowKeyword uint16

const (
	TextOverflowClip textOverflowKeyword = iota + 1
	TextOverflowEllipsis
)

func (expression) isTextOverflow()          {}
func (wideKeyword) isTextOverflow()         {}
func (textOverflowKeyword) isTextOverflow() {}
func (v textOverflowKeyword) cssText() (string, error) {
	names := [...]string{"", "clip", "ellipsis"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid text-overflow keyword %d", v)
	}
	return names[v], nil
}

type VisibilityValue interface {
	cssValue
	isVisibility()
}
type visibilityKeyword uint16

const (
	VisibilityVisible visibilityKeyword = iota + 1
	VisibilityHidden
	VisibilityCollapse
)

func (expression) isVisibility()        {}
func (wideKeyword) isVisibility()       {}
func (visibilityKeyword) isVisibility() {}
func (v visibilityKeyword) cssText() (string, error) {
	names := [...]string{"", "visible", "hidden", "collapse"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid visibility keyword %d", v)
	}
	return names[v], nil
}

type WhiteSpaceValue interface {
	cssValue
	isWhiteSpace()
}
type whiteSpaceKeyword uint16

const (
	WhiteSpaceNormal whiteSpaceKeyword = iota + 1
	WhiteSpaceNowrap
	WhiteSpacePre
	WhiteSpacePreWrap
	WhiteSpacePreLine
	WhiteSpaceBreakSpaces
)

func (expression) isWhiteSpace()        {}
func (wideKeyword) isWhiteSpace()       {}
func (whiteSpaceKeyword) isWhiteSpace() {}
func (v whiteSpaceKeyword) cssText() (string, error) {
	names := [...]string{"", "normal", "nowrap", "pre", "pre-wrap", "pre-line", "break-spaces"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid white-space keyword %d", v)
	}
	return names[v], nil
}

type WidthValue interface {
	cssValue
	isWidth()
}
type widthKeyword uint16

const (
	WidthAuto widthKeyword = iota + 1
	WidthMinContent
	WidthMaxContent
	WidthFitContent
)

func (expression) isWidth()   {}
func (wideKeyword) isWidth()  {}
func (widthKeyword) isWidth() {}
func (Length) isWidth()       {}
func (Percentage) isWidth()   {}
func (v widthKeyword) cssText() (string, error) {
	names := [...]string{"", "auto", "min-content", "max-content", "fit-content"}
	if v == 0 || int(v) >= len(names) {
		return "", fmt.Errorf("ui: invalid width keyword %d", v)
	}
	return names[v], nil
}

func (e *Element) WebkitAnimation(v Value) *Element { return e.setValue("-webkit-animation", v) }
func (e *Element) WebkitAnimationDelay(v Value) *Element {
	return e.setValue("-webkit-animation-delay", v)
}
func (e *Element) WebkitAnimationDirection(v Value) *Element {
	return e.setValue("-webkit-animation-direction", v)
}
func (e *Element) WebkitAnimationDuration(v Value) *Element {
	return e.setValue("-webkit-animation-duration", v)
}
func (e *Element) WebkitAnimationFillMode(v Value) *Element {
	return e.setValue("-webkit-animation-fill-mode", v)
}
func (e *Element) WebkitAnimationIterationCount(v Value) *Element {
	return e.setValue("-webkit-animation-iteration-count", v)
}
func (e *Element) WebkitAnimationName(v Value) *Element {
	return e.setValue("-webkit-animation-name", v)
}
func (e *Element) WebkitAnimationPlayState(v Value) *Element {
	return e.setValue("-webkit-animation-play-state", v)
}
func (e *Element) WebkitAnimationTimingFunction(v Value) *Element {
	return e.setValue("-webkit-animation-timing-function", v)
}
func (e *Element) WebkitAppearance(v Value) *Element { return e.setValue("-webkit-appearance", v) }
func (e *Element) WebkitBackfaceVisibility(v Value) *Element {
	return e.setValue("-webkit-backface-visibility", v)
}
func (e *Element) WebkitBackgroundClip(v Value) *Element {
	return e.setValue("-webkit-background-clip", v)
}
func (e *Element) WebkitBackgroundOrigin(v Value) *Element {
	return e.setValue("-webkit-background-origin", v)
}
func (e *Element) WebkitBackgroundSize(v Value) *Element {
	return e.setValue("-webkit-background-size", v)
}
func (e *Element) WebkitMask(v Value) *Element { return e.setValue("-webkit-mask", v) }
func (e *Element) WebkitMaskBoxImage(v Value) *Element {
	return e.setValue("-webkit-mask-box-image", v)
}
func (e *Element) WebkitMaskBoxImageOutset(v Value) *Element {
	return e.setValue("-webkit-mask-box-image-outset", v)
}
func (e *Element) WebkitMaskBoxImageRepeat(v Value) *Element {
	return e.setValue("-webkit-mask-box-image-repeat", v)
}
func (e *Element) WebkitMaskBoxImageSlice(v Value) *Element {
	return e.setValue("-webkit-mask-box-image-slice", v)
}
func (e *Element) WebkitMaskBoxImageSource(v Value) *Element {
	return e.setValue("-webkit-mask-box-image-source", v)
}
func (e *Element) WebkitMaskBoxImageWidth(v Value) *Element {
	return e.setValue("-webkit-mask-box-image-width", v)
}
func (e *Element) WebkitMaskClip(v Value) *Element { return e.setValue("-webkit-mask-clip", v) }
func (e *Element) WebkitMaskComposite(v Value) *Element {
	return e.setValue("-webkit-mask-composite", v)
}
func (e *Element) WebkitMaskImage(v Value) *Element    { return e.setValue("-webkit-mask-image", v) }
func (e *Element) WebkitMaskOrigin(v Value) *Element   { return e.setValue("-webkit-mask-origin", v) }
func (e *Element) WebkitMaskPosition(v Value) *Element { return e.setValue("-webkit-mask-position", v) }
func (e *Element) WebkitMaskRepeat(v Value) *Element   { return e.setValue("-webkit-mask-repeat", v) }
func (e *Element) WebkitMaskSize(v Value) *Element     { return e.setValue("-webkit-mask-size", v) }
func (e *Element) WebkitPerspective(v Value) *Element  { return e.setValue("-webkit-perspective", v) }
func (e *Element) WebkitPerspectiveOrigin(v Value) *Element {
	return e.setValue("-webkit-perspective-origin", v)
}
func (e *Element) WebkitTextSizeAdjust(v Value) *Element {
	return e.setValue("-webkit-text-size-adjust", v)
}
func (e *Element) WebkitTextStroke(v Value) *Element { return e.setValue("-webkit-text-stroke", v) }
func (e *Element) WebkitTextStrokeColor(v Value) *Element {
	return e.setValue("-webkit-text-stroke-color", v)
}
func (e *Element) WebkitTextStrokeWidth(v Value) *Element {
	return e.setValue("-webkit-text-stroke-width", v)
}
func (e *Element) WebkitTransformStyle(v Value) *Element {
	return e.setValue("-webkit-transform-style", v)
}
func (e *Element) WebkitTransition(v Value) *Element { return e.setValue("-webkit-transition", v) }
func (e *Element) WebkitTransitionDelay(v Value) *Element {
	return e.setValue("-webkit-transition-delay", v)
}
func (e *Element) WebkitTransitionDuration(v Value) *Element {
	return e.setValue("-webkit-transition-duration", v)
}
func (e *Element) WebkitTransitionProperty(v Value) *Element {
	return e.setValue("-webkit-transition-property", v)
}
func (e *Element) WebkitTransitionTimingFunction(v Value) *Element {
	return e.setValue("-webkit-transition-timing-function", v)
}
func (e *Element) WebkitUserSelect(v Value) *Element         { return e.setValue("-webkit-user-select", v) }
func (e *Element) AccentColor(v Value) *Element              { return e.setValue("accent-color", v) }
func (e *Element) AlignContent(v AlignContentValue) *Element { return e.setValue("align-content", v) }
func (e *Element) AlignItems(v AlignItemsValue) *Element     { return e.setValue("align-items", v) }
func (e *Element) AlignSelf(v AlignSelfValue) *Element       { return e.setValue("align-self", v) }
func (e *Element) AlignmentBaseline(v Value) *Element        { return e.setValue("alignment-baseline", v) }
func (e *Element) All(v Value) *Element                      { return e.setValue("all", v) }
func (e *Element) AnchorName(v Value) *Element               { return e.setValue("anchor-name", v) }
func (e *Element) AnchorScope(v Value) *Element              { return e.setValue("anchor-scope", v) }
func (e *Element) Animation(v Value) *Element                { return e.setValue("animation", v) }
func (e *Element) AnimationComposition(v Value) *Element {
	return e.setValue("animation-composition", v)
}
func (e *Element) AnimationDelay(v Value) *Element    { return e.setValue("animation-delay", v) }
func (e *Element) AnimationDelayEnd(v Value) *Element { return e.setValue("animation-delay-end", v) }
func (e *Element) AnimationDelayStart(v Value) *Element {
	return e.setValue("animation-delay-start", v)
}
func (e *Element) AnimationDirection(v Value) *Element { return e.setValue("animation-direction", v) }
func (e *Element) AnimationDuration(v Value) *Element  { return e.setValue("animation-duration", v) }
func (e *Element) AnimationFillMode(v Value) *Element  { return e.setValue("animation-fill-mode", v) }
func (e *Element) AnimationIterationCount(v Value) *Element {
	return e.setValue("animation-iteration-count", v)
}
func (e *Element) AnimationName(v Value) *Element      { return e.setValue("animation-name", v) }
func (e *Element) AnimationPlayState(v Value) *Element { return e.setValue("animation-play-state", v) }
func (e *Element) AnimationRange(v Value) *Element     { return e.setValue("animation-range", v) }
func (e *Element) AnimationRangeCenter(v Value) *Element {
	return e.setValue("animation-range-center", v)
}
func (e *Element) AnimationRangeEnd(v Value) *Element { return e.setValue("animation-range-end", v) }
func (e *Element) AnimationRangeStart(v Value) *Element {
	return e.setValue("animation-range-start", v)
}
func (e *Element) AnimationTimeline(v Value) *Element { return e.setValue("animation-timeline", v) }
func (e *Element) AnimationTimingFunction(v Value) *Element {
	return e.setValue("animation-timing-function", v)
}
func (e *Element) AnimationTrigger(v Value) *Element   { return e.setValue("animation-trigger", v) }
func (e *Element) AspectRatio(v Value) *Element        { return e.setValue("aspect-ratio", v) }
func (e *Element) BackdropFilter(v Value) *Element     { return e.setValue("backdrop-filter", v) }
func (e *Element) BackfaceVisibility(v Value) *Element { return e.setValue("backface-visibility", v) }
func (e *Element) Background(v Value) *Element         { return e.setValue("background", v) }
func (e *Element) BackgroundAttachment(v Value) *Element {
	return e.setValue("background-attachment", v)
}
func (e *Element) BackgroundBlendMode(v Value) *Element {
	return e.setValue("background-blend-mode", v)
}
func (e *Element) BackgroundClip(v Value) *Element       { return e.setValue("background-clip", v) }
func (e *Element) BackgroundColor(v ColorValue) *Element { return e.setValue("background-color", v) }
func (e *Element) BackgroundImage(v Value) *Element      { return e.setValue("background-image", v) }
func (e *Element) BackgroundOrigin(v Value) *Element     { return e.setValue("background-origin", v) }
func (e *Element) BackgroundPosition(v Value) *Element   { return e.setValue("background-position", v) }
func (e *Element) BackgroundPositionBlock(v Value) *Element {
	return e.setValue("background-position-block", v)
}
func (e *Element) BackgroundPositionInline(v Value) *Element {
	return e.setValue("background-position-inline", v)
}
func (e *Element) BackgroundPositionX(v Value) *Element {
	return e.setValue("background-position-x", v)
}
func (e *Element) BackgroundPositionY(v Value) *Element {
	return e.setValue("background-position-y", v)
}
func (e *Element) BackgroundRepeat(v Value) *Element { return e.setValue("background-repeat", v) }
func (e *Element) BackgroundRepeatBlock(v Value) *Element {
	return e.setValue("background-repeat-block", v)
}
func (e *Element) BackgroundRepeatInline(v Value) *Element {
	return e.setValue("background-repeat-inline", v)
}
func (e *Element) BackgroundRepeatX(v Value) *Element         { return e.setValue("background-repeat-x", v) }
func (e *Element) BackgroundRepeatY(v Value) *Element         { return e.setValue("background-repeat-y", v) }
func (e *Element) BackgroundSize(v Value) *Element            { return e.setValue("background-size", v) }
func (e *Element) BackgroundTbd(v Value) *Element             { return e.setValue("background-tbd", v) }
func (e *Element) BaselineShift(v Value) *Element             { return e.setValue("baseline-shift", v) }
func (e *Element) BaselineSource(v Value) *Element            { return e.setValue("baseline-source", v) }
func (e *Element) BlockEllipsis(v Value) *Element             { return e.setValue("block-ellipsis", v) }
func (e *Element) BlockSize(v LengthPercentageValue) *Element { return e.setValue("block-size", v) }
func (e *Element) BookmarkLabel(v Value) *Element             { return e.setValue("bookmark-label", v) }
func (e *Element) BookmarkLevel(v Value) *Element             { return e.setValue("bookmark-level", v) }
func (e *Element) BookmarkState(v Value) *Element             { return e.setValue("bookmark-state", v) }
func (e *Element) Border(v BorderStrokeValue) *Element        { return e.setValue("border", v) }
func (e *Element) BorderBlock(v Value) *Element               { return e.setValue("border-block", v) }
func (e *Element) BorderBlockClip(v Value) *Element           { return e.setValue("border-block-clip", v) }
func (e *Element) BorderBlockColor(v Value) *Element          { return e.setValue("border-block-color", v) }
func (e *Element) BorderBlockEnd(v Value) *Element            { return e.setValue("border-block-end", v) }
func (e *Element) BorderBlockEndClip(v Value) *Element        { return e.setValue("border-block-end-clip", v) }
func (e *Element) BorderBlockEndColor(v Value) *Element {
	return e.setValue("border-block-end-color", v)
}
func (e *Element) BorderBlockEndRadius(v Value) *Element {
	return e.setValue("border-block-end-radius", v)
}
func (e *Element) BorderBlockEndStyle(v Value) *Element {
	return e.setValue("border-block-end-style", v)
}
func (e *Element) BorderBlockEndWidth(v Value) *Element {
	return e.setValue("border-block-end-width", v)
}
func (e *Element) BorderBlockStart(v Value) *Element { return e.setValue("border-block-start", v) }
func (e *Element) BorderBlockStartClip(v Value) *Element {
	return e.setValue("border-block-start-clip", v)
}
func (e *Element) BorderBlockStartColor(v Value) *Element {
	return e.setValue("border-block-start-color", v)
}
func (e *Element) BorderBlockStartRadius(v Value) *Element {
	return e.setValue("border-block-start-radius", v)
}
func (e *Element) BorderBlockStartStyle(v Value) *Element {
	return e.setValue("border-block-start-style", v)
}
func (e *Element) BorderBlockStartWidth(v Value) *Element {
	return e.setValue("border-block-start-width", v)
}
func (e *Element) BorderBlockStyle(v Value) *Element         { return e.setValue("border-block-style", v) }
func (e *Element) BorderBlockWidth(v Value) *Element         { return e.setValue("border-block-width", v) }
func (e *Element) BorderBottom(v BorderStrokeValue) *Element { return e.setValue("border-bottom", v) }
func (e *Element) BorderBottomClip(v Value) *Element         { return e.setValue("border-bottom-clip", v) }
func (e *Element) BorderBottomColor(v ColorValue) *Element {
	return e.setValue("border-bottom-color", v)
}
func (e *Element) BorderBottomLeftRadius(v LengthPercentageValue) *Element {
	return e.setValue("border-bottom-left-radius", v)
}
func (e *Element) BorderBottomRadius(v Value) *Element { return e.setValue("border-bottom-radius", v) }
func (e *Element) BorderBottomRightRadius(v LengthPercentageValue) *Element {
	return e.setValue("border-bottom-right-radius", v)
}
func (e *Element) BorderBottomStyle(v Value) *Element { return e.setValue("border-bottom-style", v) }
func (e *Element) BorderBottomWidth(v LengthValue) *Element {
	return e.setValue("border-bottom-width", v)
}
func (e *Element) BorderBoundary(v Value) *Element { return e.setValue("border-boundary", v) }
func (e *Element) BorderClip(v Value) *Element     { return e.setValue("border-clip", v) }
func (e *Element) BorderCollapse(v BorderCollapseValue) *Element {
	return e.setValue("border-collapse", v)
}
func (e *Element) BorderColor(v ColorValue) *Element   { return e.setValue("border-color", v) }
func (e *Element) BorderEndEndRadius(v Value) *Element { return e.setValue("border-end-end-radius", v) }
func (e *Element) BorderEndStartRadius(v Value) *Element {
	return e.setValue("border-end-start-radius", v)
}
func (e *Element) BorderImage(v Value) *Element       { return e.setValue("border-image", v) }
func (e *Element) BorderImageOutset(v Value) *Element { return e.setValue("border-image-outset", v) }
func (e *Element) BorderImageRepeat(v Value) *Element { return e.setValue("border-image-repeat", v) }
func (e *Element) BorderImageSlice(v Value) *Element  { return e.setValue("border-image-slice", v) }
func (e *Element) BorderImageSource(v Value) *Element { return e.setValue("border-image-source", v) }
func (e *Element) BorderImageWidth(v Value) *Element  { return e.setValue("border-image-width", v) }
func (e *Element) BorderInline(v Value) *Element      { return e.setValue("border-inline", v) }
func (e *Element) BorderInlineClip(v Value) *Element  { return e.setValue("border-inline-clip", v) }
func (e *Element) BorderInlineColor(v Value) *Element { return e.setValue("border-inline-color", v) }
func (e *Element) BorderInlineEnd(v Value) *Element   { return e.setValue("border-inline-end", v) }
func (e *Element) BorderInlineEndClip(v Value) *Element {
	return e.setValue("border-inline-end-clip", v)
}
func (e *Element) BorderInlineEndColor(v Value) *Element {
	return e.setValue("border-inline-end-color", v)
}
func (e *Element) BorderInlineEndRadius(v Value) *Element {
	return e.setValue("border-inline-end-radius", v)
}
func (e *Element) BorderInlineEndStyle(v Value) *Element {
	return e.setValue("border-inline-end-style", v)
}
func (e *Element) BorderInlineEndWidth(v Value) *Element {
	return e.setValue("border-inline-end-width", v)
}
func (e *Element) BorderInlineStart(v Value) *Element { return e.setValue("border-inline-start", v) }
func (e *Element) BorderInlineStartClip(v Value) *Element {
	return e.setValue("border-inline-start-clip", v)
}
func (e *Element) BorderInlineStartColor(v Value) *Element {
	return e.setValue("border-inline-start-color", v)
}
func (e *Element) BorderInlineStartRadius(v Value) *Element {
	return e.setValue("border-inline-start-radius", v)
}
func (e *Element) BorderInlineStartStyle(v Value) *Element {
	return e.setValue("border-inline-start-style", v)
}
func (e *Element) BorderInlineStartWidth(v Value) *Element {
	return e.setValue("border-inline-start-width", v)
}
func (e *Element) BorderInlineStyle(v Value) *Element      { return e.setValue("border-inline-style", v) }
func (e *Element) BorderInlineWidth(v Value) *Element      { return e.setValue("border-inline-width", v) }
func (e *Element) BorderLeft(v BorderStrokeValue) *Element { return e.setValue("border-left", v) }
func (e *Element) BorderLeftClip(v Value) *Element         { return e.setValue("border-left-clip", v) }
func (e *Element) BorderLeftColor(v ColorValue) *Element   { return e.setValue("border-left-color", v) }
func (e *Element) BorderLeftRadius(v Value) *Element       { return e.setValue("border-left-radius", v) }
func (e *Element) BorderLeftStyle(v Value) *Element        { return e.setValue("border-left-style", v) }
func (e *Element) BorderLeftWidth(v LengthValue) *Element  { return e.setValue("border-left-width", v) }
func (e *Element) BorderLimit(v Value) *Element            { return e.setValue("border-limit", v) }
func (e *Element) BorderRadius(v LengthPercentageSpacingValue) *Element {
	return e.setValue("border-radius", v)
}
func (e *Element) BorderRight(v BorderStrokeValue) *Element { return e.setValue("border-right", v) }
func (e *Element) BorderRightClip(v Value) *Element         { return e.setValue("border-right-clip", v) }
func (e *Element) BorderRightColor(v ColorValue) *Element   { return e.setValue("border-right-color", v) }
func (e *Element) BorderRightRadius(v Value) *Element       { return e.setValue("border-right-radius", v) }
func (e *Element) BorderRightStyle(v Value) *Element        { return e.setValue("border-right-style", v) }
func (e *Element) BorderRightWidth(v LengthValue) *Element {
	return e.setValue("border-right-width", v)
}
func (e *Element) BorderShape(v Value) *Element   { return e.setValue("border-shape", v) }
func (e *Element) BorderSpacing(v Value) *Element { return e.setValue("border-spacing", v) }
func (e *Element) BorderStartEndRadius(v Value) *Element {
	return e.setValue("border-start-end-radius", v)
}
func (e *Element) BorderStartStartRadius(v Value) *Element {
	return e.setValue("border-start-start-radius", v)
}
func (e *Element) BorderStyle(v Value) *Element           { return e.setValue("border-style", v) }
func (e *Element) BorderTop(v BorderStrokeValue) *Element { return e.setValue("border-top", v) }
func (e *Element) BorderTopClip(v Value) *Element         { return e.setValue("border-top-clip", v) }
func (e *Element) BorderTopColor(v ColorValue) *Element   { return e.setValue("border-top-color", v) }
func (e *Element) BorderTopLeftRadius(v LengthPercentageValue) *Element {
	return e.setValue("border-top-left-radius", v)
}
func (e *Element) BorderTopRadius(v Value) *Element { return e.setValue("border-top-radius", v) }
func (e *Element) BorderTopRightRadius(v LengthPercentageValue) *Element {
	return e.setValue("border-top-right-radius", v)
}
func (e *Element) BorderTopStyle(v Value) *Element         { return e.setValue("border-top-style", v) }
func (e *Element) BorderTopWidth(v LengthValue) *Element   { return e.setValue("border-top-width", v) }
func (e *Element) BorderWidth(v LengthValue) *Element      { return e.setValue("border-width", v) }
func (e *Element) Bottom(v LengthPercentageValue) *Element { return e.setValue("bottom", v) }
func (e *Element) BoxDecorationBreak(v Value) *Element     { return e.setValue("box-decoration-break", v) }
func (e *Element) BoxShadow(v Value) *Element              { return e.setValue("box-shadow", v) }
func (e *Element) BoxShadowBlur(v Value) *Element          { return e.setValue("box-shadow-blur", v) }
func (e *Element) BoxShadowColor(v Value) *Element         { return e.setValue("box-shadow-color", v) }
func (e *Element) BoxShadowInset(v Value) *Element         { return e.setValue("box-shadow-inset", v) }
func (e *Element) BoxShadowOffset(v Value) *Element        { return e.setValue("box-shadow-offset", v) }
func (e *Element) BoxShadowPosition(v Value) *Element      { return e.setValue("box-shadow-position", v) }
func (e *Element) BoxShadowSpread(v Value) *Element        { return e.setValue("box-shadow-spread", v) }
func (e *Element) BoxSizing(v BoxSizingValue) *Element     { return e.setValue("box-sizing", v) }
func (e *Element) BreakAfter(v Value) *Element             { return e.setValue("break-after", v) }
func (e *Element) BreakBefore(v Value) *Element            { return e.setValue("break-before", v) }
func (e *Element) BreakInside(v Value) *Element            { return e.setValue("break-inside", v) }
func (e *Element) CaptionSide(v Value) *Element            { return e.setValue("caption-side", v) }
func (e *Element) Caret(v Value) *Element                  { return e.setValue("caret", v) }
func (e *Element) CaretAnimation(v Value) *Element         { return e.setValue("caret-animation", v) }
func (e *Element) CaretShape(v Value) *Element             { return e.setValue("caret-shape", v) }
func (e *Element) Clear(v Value) *Element                  { return e.setValue("clear", v) }
func (e *Element) Clip(v Value) *Element                   { return e.setValue("clip", v) }
func (e *Element) ClipPath(v Value) *Element               { return e.setValue("clip-path", v) }
func (e *Element) ClipRule(v Value) *Element               { return e.setValue("clip-rule", v) }
func (e *Element) Color(v ColorValue) *Element             { return e.setValue("color", v) }
func (e *Element) ColorAdjust(v Value) *Element            { return e.setValue("color-adjust", v) }
func (e *Element) ColorInterpolation(v Value) *Element     { return e.setValue("color-interpolation", v) }
func (e *Element) ColorInterpolationFilters(v Value) *Element {
	return e.setValue("color-interpolation-filters", v)
}
func (e *Element) ColorScheme(v Value) *Element               { return e.setValue("color-scheme", v) }
func (e *Element) ColumnCount(v NumericValue) *Element        { return e.setValue("column-count", v) }
func (e *Element) ColumnFill(v Value) *Element                { return e.setValue("column-fill", v) }
func (e *Element) ColumnGap(v LengthPercentageValue) *Element { return e.setValue("column-gap", v) }
func (e *Element) ColumnHeight(v Value) *Element              { return e.setValue("column-height", v) }
func (e *Element) ColumnRule(v Value) *Element                { return e.setValue("column-rule", v) }
func (e *Element) ColumnRuleBreak(v Value) *Element           { return e.setValue("column-rule-break", v) }
func (e *Element) ColumnRuleColor(v Value) *Element           { return e.setValue("column-rule-color", v) }
func (e *Element) ColumnRuleInset(v Value) *Element           { return e.setValue("column-rule-inset", v) }
func (e *Element) ColumnRuleInsetCap(v Value) *Element        { return e.setValue("column-rule-inset-cap", v) }
func (e *Element) ColumnRuleInsetCapEnd(v Value) *Element {
	return e.setValue("column-rule-inset-cap-end", v)
}
func (e *Element) ColumnRuleInsetCapStart(v Value) *Element {
	return e.setValue("column-rule-inset-cap-start", v)
}
func (e *Element) ColumnRuleInsetEnd(v Value) *Element { return e.setValue("column-rule-inset-end", v) }
func (e *Element) ColumnRuleInsetJunction(v Value) *Element {
	return e.setValue("column-rule-inset-junction", v)
}
func (e *Element) ColumnRuleInsetJunctionEnd(v Value) *Element {
	return e.setValue("column-rule-inset-junction-end", v)
}
func (e *Element) ColumnRuleInsetJunctionStart(v Value) *Element {
	return e.setValue("column-rule-inset-junction-start", v)
}
func (e *Element) ColumnRuleInsetStart(v Value) *Element {
	return e.setValue("column-rule-inset-start", v)
}
func (e *Element) ColumnRuleStyle(v Value) *Element { return e.setValue("column-rule-style", v) }
func (e *Element) ColumnRuleVisibilityItems(v Value) *Element {
	return e.setValue("column-rule-visibility-items", v)
}
func (e *Element) ColumnRuleWidth(v Value) *Element   { return e.setValue("column-rule-width", v) }
func (e *Element) ColumnSpan(v Value) *Element        { return e.setValue("column-span", v) }
func (e *Element) ColumnWidth(v LengthValue) *Element { return e.setValue("column-width", v) }
func (e *Element) ColumnWrap(v Value) *Element        { return e.setValue("column-wrap", v) }
func (e *Element) Columns(v Value) *Element           { return e.setValue("columns", v) }
func (e *Element) Contain(v Value) *Element           { return e.setValue("contain", v) }
func (e *Element) ContainIntrinsicBlockSize(v Value) *Element {
	return e.setValue("contain-intrinsic-block-size", v)
}
func (e *Element) ContainIntrinsicHeight(v Value) *Element {
	return e.setValue("contain-intrinsic-height", v)
}
func (e *Element) ContainIntrinsicInlineSize(v Value) *Element {
	return e.setValue("contain-intrinsic-inline-size", v)
}
func (e *Element) ContainIntrinsicSize(v Value) *Element {
	return e.setValue("contain-intrinsic-size", v)
}
func (e *Element) ContainIntrinsicWidth(v Value) *Element {
	return e.setValue("contain-intrinsic-width", v)
}
func (e *Element) Container(v Value) *Element         { return e.setValue("container", v) }
func (e *Element) ContainerName(v Value) *Element     { return e.setValue("container-name", v) }
func (e *Element) ContainerType(v Value) *Element     { return e.setValue("container-type", v) }
func (e *Element) Content(v Value) *Element           { return e.setValue("content", v) }
func (e *Element) ContentVisibility(v Value) *Element { return e.setValue("content-visibility", v) }
func (e *Element) Continue(v Value) *Element          { return e.setValue("continue", v) }
func (e *Element) CopyInto(v Value) *Element          { return e.setValue("copy-into", v) }
func (e *Element) Corner(v Value) *Element            { return e.setValue("corner", v) }
func (e *Element) CornerBlockEnd(v Value) *Element    { return e.setValue("corner-block-end", v) }
func (e *Element) CornerBlockEndShape(v Value) *Element {
	return e.setValue("corner-block-end-shape", v)
}
func (e *Element) CornerBlockStart(v Value) *Element { return e.setValue("corner-block-start", v) }
func (e *Element) CornerBlockStartShape(v Value) *Element {
	return e.setValue("corner-block-start-shape", v)
}
func (e *Element) CornerBottom(v Value) *Element     { return e.setValue("corner-bottom", v) }
func (e *Element) CornerBottomLeft(v Value) *Element { return e.setValue("corner-bottom-left", v) }
func (e *Element) CornerBottomLeftShape(v Value) *Element {
	return e.setValue("corner-bottom-left-shape", v)
}
func (e *Element) CornerBottomRight(v Value) *Element { return e.setValue("corner-bottom-right", v) }
func (e *Element) CornerBottomRightShape(v Value) *Element {
	return e.setValue("corner-bottom-right-shape", v)
}
func (e *Element) CornerBottomShape(v Value) *Element { return e.setValue("corner-bottom-shape", v) }
func (e *Element) CornerEndEnd(v Value) *Element      { return e.setValue("corner-end-end", v) }
func (e *Element) CornerEndEndShape(v Value) *Element { return e.setValue("corner-end-end-shape", v) }
func (e *Element) CornerEndStart(v Value) *Element    { return e.setValue("corner-end-start", v) }
func (e *Element) CornerEndStartShape(v Value) *Element {
	return e.setValue("corner-end-start-shape", v)
}
func (e *Element) CornerInlineEnd(v Value) *Element { return e.setValue("corner-inline-end", v) }
func (e *Element) CornerInlineEndShape(v Value) *Element {
	return e.setValue("corner-inline-end-shape", v)
}
func (e *Element) CornerInlineStart(v Value) *Element { return e.setValue("corner-inline-start", v) }
func (e *Element) CornerInlineStartShape(v Value) *Element {
	return e.setValue("corner-inline-start-shape", v)
}
func (e *Element) CornerLeft(v Value) *Element       { return e.setValue("corner-left", v) }
func (e *Element) CornerLeftShape(v Value) *Element  { return e.setValue("corner-left-shape", v) }
func (e *Element) CornerRight(v Value) *Element      { return e.setValue("corner-right", v) }
func (e *Element) CornerRightShape(v Value) *Element { return e.setValue("corner-right-shape", v) }
func (e *Element) CornerShape(v Value) *Element      { return e.setValue("corner-shape", v) }
func (e *Element) CornerStartEnd(v Value) *Element   { return e.setValue("corner-start-end", v) }
func (e *Element) CornerStartEndShape(v Value) *Element {
	return e.setValue("corner-start-end-shape", v)
}
func (e *Element) CornerStartStart(v Value) *Element { return e.setValue("corner-start-start", v) }
func (e *Element) CornerStartStartShape(v Value) *Element {
	return e.setValue("corner-start-start-shape", v)
}
func (e *Element) CornerTop(v Value) *Element          { return e.setValue("corner-top", v) }
func (e *Element) CornerTopLeft(v Value) *Element      { return e.setValue("corner-top-left", v) }
func (e *Element) CornerTopLeftShape(v Value) *Element { return e.setValue("corner-top-left-shape", v) }
func (e *Element) CornerTopRight(v Value) *Element     { return e.setValue("corner-top-right", v) }
func (e *Element) CornerTopRightShape(v Value) *Element {
	return e.setValue("corner-top-right-shape", v)
}
func (e *Element) CornerTopShape(v Value) *Element            { return e.setValue("corner-top-shape", v) }
func (e *Element) CounterIncrement(v Value) *Element          { return e.setValue("counter-increment", v) }
func (e *Element) CounterReset(v Value) *Element              { return e.setValue("counter-reset", v) }
func (e *Element) CounterSet(v Value) *Element                { return e.setValue("counter-set", v) }
func (e *Element) Cx(v Value) *Element                        { return e.setValue("cx", v) }
func (e *Element) Cy(v Value) *Element                        { return e.setValue("cy", v) }
func (e *Element) D(v Value) *Element                         { return e.setValue("d", v) }
func (e *Element) Direction(v Value) *Element                 { return e.setValue("direction", v) }
func (e *Element) Display(v DisplayValue) *Element            { return e.setValue("display", v) }
func (e *Element) DominantBaseline(v Value) *Element          { return e.setValue("dominant-baseline", v) }
func (e *Element) DynamicRangeLimit(v Value) *Element         { return e.setValue("dynamic-range-limit", v) }
func (e *Element) EmptyCells(v Value) *Element                { return e.setValue("empty-cells", v) }
func (e *Element) EventTrigger(v Value) *Element              { return e.setValue("event-trigger", v) }
func (e *Element) EventTriggerName(v Value) *Element          { return e.setValue("event-trigger-name", v) }
func (e *Element) EventTriggerSource(v Value) *Element        { return e.setValue("event-trigger-source", v) }
func (e *Element) FieldSizing(v Value) *Element               { return e.setValue("field-sizing", v) }
func (e *Element) Fill(v ColorValue) *Element                 { return e.setValue("fill", v) }
func (e *Element) FillOpacity(v NumericValue) *Element        { return e.setValue("fill-opacity", v) }
func (e *Element) FillRule(v Value) *Element                  { return e.setValue("fill-rule", v) }
func (e *Element) Filter(v Value) *Element                    { return e.setValue("filter", v) }
func (e *Element) Flex(v NumericValue) *Element               { return e.setValue("flex", v) }
func (e *Element) FlexBasis(v LengthPercentageValue) *Element { return e.setValue("flex-basis", v) }
func (e *Element) FlexDirection(v FlexDirectionValue) *Element {
	return e.setValue("flex-direction", v)
}
func (e *Element) FlexFlow(v Value) *Element             { return e.setValue("flex-flow", v) }
func (e *Element) FlexGrow(v NumericValue) *Element      { return e.setValue("flex-grow", v) }
func (e *Element) FlexShrink(v NumericValue) *Element    { return e.setValue("flex-shrink", v) }
func (e *Element) FlexWrap(v FlexWrapValue) *Element     { return e.setValue("flex-wrap", v) }
func (e *Element) Float(v Value) *Element                { return e.setValue("float", v) }
func (e *Element) FloatDefer(v Value) *Element           { return e.setValue("float-defer", v) }
func (e *Element) FloatOffset(v Value) *Element          { return e.setValue("float-offset", v) }
func (e *Element) FloatReference(v Value) *Element       { return e.setValue("float-reference", v) }
func (e *Element) FloodColor(v Value) *Element           { return e.setValue("flood-color", v) }
func (e *Element) FloodOpacity(v Value) *Element         { return e.setValue("flood-opacity", v) }
func (e *Element) FlowFrom(v Value) *Element             { return e.setValue("flow-from", v) }
func (e *Element) FlowInto(v Value) *Element             { return e.setValue("flow-into", v) }
func (e *Element) FlowTolerance(v Value) *Element        { return e.setValue("flow-tolerance", v) }
func (e *Element) Font(v Value) *Element                 { return e.setValue("font", v) }
func (e *Element) FontFamily(v FontFamilyValue) *Element { return e.setValue("font-family", v) }
func (e *Element) FontFeatureSettings(v Value) *Element {
	return e.setValue("font-feature-settings", v)
}
func (e *Element) FontKerning(v Value) *Element { return e.setValue("font-kerning", v) }
func (e *Element) FontLanguageOverride(v Value) *Element {
	return e.setValue("font-language-override", v)
}
func (e *Element) FontOpticalSizing(v Value) *Element        { return e.setValue("font-optical-sizing", v) }
func (e *Element) FontPalette(v Value) *Element              { return e.setValue("font-palette", v) }
func (e *Element) FontSize(v LengthPercentageValue) *Element { return e.setValue("font-size", v) }
func (e *Element) FontSizeAdjust(v Value) *Element           { return e.setValue("font-size-adjust", v) }
func (e *Element) FontStyle(v FontStyleValue) *Element       { return e.setValue("font-style", v) }
func (e *Element) FontSynthesis(v Value) *Element            { return e.setValue("font-synthesis", v) }
func (e *Element) FontSynthesisPosition(v Value) *Element {
	return e.setValue("font-synthesis-position", v)
}
func (e *Element) FontSynthesisSmallCaps(v Value) *Element {
	return e.setValue("font-synthesis-small-caps", v)
}
func (e *Element) FontSynthesisStyle(v Value) *Element { return e.setValue("font-synthesis-style", v) }
func (e *Element) FontSynthesisWeight(v Value) *Element {
	return e.setValue("font-synthesis-weight", v)
}
func (e *Element) FontVariant(v Value) *Element { return e.setValue("font-variant", v) }
func (e *Element) FontVariantAlternates(v Value) *Element {
	return e.setValue("font-variant-alternates", v)
}
func (e *Element) FontVariantCaps(v Value) *Element { return e.setValue("font-variant-caps", v) }
func (e *Element) FontVariantEastAsian(v Value) *Element {
	return e.setValue("font-variant-east-asian", v)
}
func (e *Element) FontVariantEmoji(v Value) *Element { return e.setValue("font-variant-emoji", v) }
func (e *Element) FontVariantLigatures(v Value) *Element {
	return e.setValue("font-variant-ligatures", v)
}
func (e *Element) FontVariantNumeric(v Value) *Element { return e.setValue("font-variant-numeric", v) }
func (e *Element) FontVariantPosition(v Value) *Element {
	return e.setValue("font-variant-position", v)
}
func (e *Element) FontVariationSettings(v Value) *Element {
	return e.setValue("font-variation-settings", v)
}
func (e *Element) FontWeight(v FontWeightValue) *Element       { return e.setValue("font-weight", v) }
func (e *Element) FontWidth(v Value) *Element                  { return e.setValue("font-width", v) }
func (e *Element) FootnoteDisplay(v Value) *Element            { return e.setValue("footnote-display", v) }
func (e *Element) FootnotePolicy(v Value) *Element             { return e.setValue("footnote-policy", v) }
func (e *Element) ForcedColorAdjust(v Value) *Element          { return e.setValue("forced-color-adjust", v) }
func (e *Element) Gap(v LengthPercentageSpacingValue) *Element { return e.setValue("gap", v) }
func (e *Element) GlyphOrientationVertical(v Value) *Element {
	return e.setValue("glyph-orientation-vertical", v)
}
func (e *Element) Grid(v Value) *Element              { return e.setValue("grid", v) }
func (e *Element) GridArea(v Value) *Element          { return e.setValue("grid-area", v) }
func (e *Element) GridAutoColumns(v Value) *Element   { return e.setValue("grid-auto-columns", v) }
func (e *Element) GridAutoFlow(v Value) *Element      { return e.setValue("grid-auto-flow", v) }
func (e *Element) GridAutoRows(v Value) *Element      { return e.setValue("grid-auto-rows", v) }
func (e *Element) GridColumn(v Value) *Element        { return e.setValue("grid-column", v) }
func (e *Element) GridColumnEnd(v Value) *Element     { return e.setValue("grid-column-end", v) }
func (e *Element) GridColumnStart(v Value) *Element   { return e.setValue("grid-column-start", v) }
func (e *Element) GridRow(v Value) *Element           { return e.setValue("grid-row", v) }
func (e *Element) GridRowEnd(v Value) *Element        { return e.setValue("grid-row-end", v) }
func (e *Element) GridRowStart(v Value) *Element      { return e.setValue("grid-row-start", v) }
func (e *Element) GridTemplate(v Value) *Element      { return e.setValue("grid-template", v) }
func (e *Element) GridTemplateAreas(v Value) *Element { return e.setValue("grid-template-areas", v) }
func (e *Element) GridTemplateColumns(v Value) *Element {
	return e.setValue("grid-template-columns", v)
}
func (e *Element) GridTemplateRows(v Value) *Element   { return e.setValue("grid-template-rows", v) }
func (e *Element) HangingPunctuation(v Value) *Element { return e.setValue("hanging-punctuation", v) }
func (e *Element) Height(v HeightValue) *Element       { return e.setValue("height", v) }
func (e *Element) HyphenateCharacter(v Value) *Element { return e.setValue("hyphenate-character", v) }
func (e *Element) HyphenateLimitChars(v Value) *Element {
	return e.setValue("hyphenate-limit-chars", v)
}
func (e *Element) HyphenateLimitLast(v Value) *Element { return e.setValue("hyphenate-limit-last", v) }
func (e *Element) HyphenateLimitLines(v Value) *Element {
	return e.setValue("hyphenate-limit-lines", v)
}
func (e *Element) HyphenateLimitZone(v Value) *Element         { return e.setValue("hyphenate-limit-zone", v) }
func (e *Element) Hyphens(v Value) *Element                    { return e.setValue("hyphens", v) }
func (e *Element) ImageAnimation(v Value) *Element             { return e.setValue("image-animation", v) }
func (e *Element) ImageOrientation(v Value) *Element           { return e.setValue("image-orientation", v) }
func (e *Element) ImageRendering(v Value) *Element             { return e.setValue("image-rendering", v) }
func (e *Element) ImageResolution(v Value) *Element            { return e.setValue("image-resolution", v) }
func (e *Element) InitialLetter(v Value) *Element              { return e.setValue("initial-letter", v) }
func (e *Element) InitialLetterAlign(v Value) *Element         { return e.setValue("initial-letter-align", v) }
func (e *Element) InitialLetterWrap(v Value) *Element          { return e.setValue("initial-letter-wrap", v) }
func (e *Element) InlineSize(v LengthPercentageValue) *Element { return e.setValue("inline-size", v) }
func (e *Element) InlineSizing(v Value) *Element               { return e.setValue("inline-sizing", v) }
func (e *Element) InputSecurity(v Value) *Element              { return e.setValue("input-security", v) }
func (e *Element) Inset(v LengthPercentageValue) *Element      { return e.setValue("inset", v) }
func (e *Element) InsetBlock(v LengthPercentageValue) *Element { return e.setValue("inset-block", v) }
func (e *Element) InsetBlockEnd(v LengthPercentageValue) *Element {
	return e.setValue("inset-block-end", v)
}
func (e *Element) InsetBlockStart(v LengthPercentageValue) *Element {
	return e.setValue("inset-block-start", v)
}
func (e *Element) InsetInline(v LengthPercentageValue) *Element { return e.setValue("inset-inline", v) }
func (e *Element) InsetInlineEnd(v LengthPercentageValue) *Element {
	return e.setValue("inset-inline-end", v)
}
func (e *Element) InsetInlineStart(v LengthPercentageValue) *Element {
	return e.setValue("inset-inline-start", v)
}
func (e *Element) Interactivity(v Value) *Element      { return e.setValue("interactivity", v) }
func (e *Element) InterestDelay(v Value) *Element      { return e.setValue("interest-delay", v) }
func (e *Element) InterestDelayEnd(v Value) *Element   { return e.setValue("interest-delay-end", v) }
func (e *Element) InterestDelayStart(v Value) *Element { return e.setValue("interest-delay-start", v) }
func (e *Element) Isolation(v Value) *Element          { return e.setValue("isolation", v) }
func (e *Element) JustifyContent(v JustifyContentValue) *Element {
	return e.setValue("justify-content", v)
}
func (e *Element) JustifyItems(v Value) *Element         { return e.setValue("justify-items", v) }
func (e *Element) JustifySelf(v Value) *Element          { return e.setValue("justify-self", v) }
func (e *Element) Left(v LengthPercentageValue) *Element { return e.setValue("left", v) }
func (e *Element) LetterSpacing(v LengthValue) *Element  { return e.setValue("letter-spacing", v) }
func (e *Element) LightingColor(v Value) *Element        { return e.setValue("lighting-color", v) }
func (e *Element) LineBreak(v Value) *Element            { return e.setValue("line-break", v) }
func (e *Element) LineClamp(v Value) *Element            { return e.setValue("line-clamp", v) }
func (e *Element) LineFitEdge(v Value) *Element          { return e.setValue("line-fit-edge", v) }
func (e *Element) LineHeight(v LengthNumericPercentageValue) *Element {
	return e.setValue("line-height", v)
}
func (e *Element) LinePadding(v Value) *Element                   { return e.setValue("line-padding", v) }
func (e *Element) LinkParameters(v Value) *Element                { return e.setValue("link-parameters", v) }
func (e *Element) ListStyle(v Value) *Element                     { return e.setValue("list-style", v) }
func (e *Element) ListStyleImage(v Value) *Element                { return e.setValue("list-style-image", v) }
func (e *Element) ListStylePosition(v Value) *Element             { return e.setValue("list-style-position", v) }
func (e *Element) ListStyleType(v Value) *Element                 { return e.setValue("list-style-type", v) }
func (e *Element) Margin(v LengthPercentageSpacingValue) *Element { return e.setValue("margin", v) }
func (e *Element) MarginBlock(v LengthPercentageSpacingValue) *Element {
	return e.setValue("margin-block", v)
}
func (e *Element) MarginBlockEnd(v LengthPercentageValue) *Element {
	return e.setValue("margin-block-end", v)
}
func (e *Element) MarginBlockStart(v LengthPercentageValue) *Element {
	return e.setValue("margin-block-start", v)
}
func (e *Element) MarginBottom(v LengthPercentageValue) *Element {
	return e.setValue("margin-bottom", v)
}
func (e *Element) MarginBreak(v Value) *Element { return e.setValue("margin-break", v) }
func (e *Element) MarginInline(v LengthPercentageSpacingValue) *Element {
	return e.setValue("margin-inline", v)
}
func (e *Element) MarginInlineEnd(v LengthPercentageValue) *Element {
	return e.setValue("margin-inline-end", v)
}
func (e *Element) MarginInlineStart(v LengthPercentageValue) *Element {
	return e.setValue("margin-inline-start", v)
}
func (e *Element) MarginLeft(v LengthPercentageValue) *Element  { return e.setValue("margin-left", v) }
func (e *Element) MarginRight(v LengthPercentageValue) *Element { return e.setValue("margin-right", v) }
func (e *Element) MarginTop(v LengthPercentageValue) *Element   { return e.setValue("margin-top", v) }
func (e *Element) MarginTrim(v Value) *Element                  { return e.setValue("margin-trim", v) }
func (e *Element) Marker(v Value) *Element                      { return e.setValue("marker", v) }
func (e *Element) MarkerEnd(v Value) *Element                   { return e.setValue("marker-end", v) }
func (e *Element) MarkerMid(v Value) *Element                   { return e.setValue("marker-mid", v) }
func (e *Element) MarkerSide(v Value) *Element                  { return e.setValue("marker-side", v) }
func (e *Element) MarkerStart(v Value) *Element                 { return e.setValue("marker-start", v) }
func (e *Element) Mask(v Value) *Element                        { return e.setValue("mask", v) }
func (e *Element) MaskBorder(v Value) *Element                  { return e.setValue("mask-border", v) }
func (e *Element) MaskBorderMode(v Value) *Element              { return e.setValue("mask-border-mode", v) }
func (e *Element) MaskBorderOutset(v Value) *Element            { return e.setValue("mask-border-outset", v) }
func (e *Element) MaskBorderRepeat(v Value) *Element            { return e.setValue("mask-border-repeat", v) }
func (e *Element) MaskBorderSlice(v Value) *Element             { return e.setValue("mask-border-slice", v) }
func (e *Element) MaskBorderSource(v Value) *Element            { return e.setValue("mask-border-source", v) }
func (e *Element) MaskBorderWidth(v Value) *Element             { return e.setValue("mask-border-width", v) }
func (e *Element) MaskClip(v Value) *Element                    { return e.setValue("mask-clip", v) }
func (e *Element) MaskComposite(v Value) *Element               { return e.setValue("mask-composite", v) }
func (e *Element) MaskImage(v Value) *Element                   { return e.setValue("mask-image", v) }
func (e *Element) MaskMode(v Value) *Element                    { return e.setValue("mask-mode", v) }
func (e *Element) MaskOrigin(v Value) *Element                  { return e.setValue("mask-origin", v) }
func (e *Element) MaskPosition(v Value) *Element                { return e.setValue("mask-position", v) }
func (e *Element) MaskRepeat(v Value) *Element                  { return e.setValue("mask-repeat", v) }
func (e *Element) MaskSize(v Value) *Element                    { return e.setValue("mask-size", v) }
func (e *Element) MaskType(v Value) *Element                    { return e.setValue("mask-type", v) }
func (e *Element) MaxBlockSize(v LengthPercentageValue) *Element {
	return e.setValue("max-block-size", v)
}
func (e *Element) MaxHeight(v LengthPercentageValue) *Element { return e.setValue("max-height", v) }
func (e *Element) MaxInlineSize(v LengthPercentageValue) *Element {
	return e.setValue("max-inline-size", v)
}
func (e *Element) MaxLines(v Value) *Element                 { return e.setValue("max-lines", v) }
func (e *Element) MaxSize(v Value) *Element                  { return e.setValue("max-size", v) }
func (e *Element) MaxWidth(v LengthPercentageValue) *Element { return e.setValue("max-width", v) }
func (e *Element) MinBlockSize(v LengthPercentageValue) *Element {
	return e.setValue("min-block-size", v)
}
func (e *Element) MinHeight(v LengthPercentageValue) *Element { return e.setValue("min-height", v) }
func (e *Element) MinInlineSize(v LengthPercentageValue) *Element {
	return e.setValue("min-inline-size", v)
}
func (e *Element) MinSize(v Value) *Element                  { return e.setValue("min-size", v) }
func (e *Element) MinWidth(v LengthPercentageValue) *Element { return e.setValue("min-width", v) }
func (e *Element) MixBlendMode(v Value) *Element             { return e.setValue("mix-blend-mode", v) }
func (e *Element) NavDown(v Value) *Element                  { return e.setValue("nav-down", v) }
func (e *Element) NavLeft(v Value) *Element                  { return e.setValue("nav-left", v) }
func (e *Element) NavRight(v Value) *Element                 { return e.setValue("nav-right", v) }
func (e *Element) NavUp(v Value) *Element                    { return e.setValue("nav-up", v) }
func (e *Element) ObjectFit(v Value) *Element                { return e.setValue("object-fit", v) }
func (e *Element) ObjectPosition(v Value) *Element           { return e.setValue("object-position", v) }
func (e *Element) ObjectViewBox(v Value) *Element            { return e.setValue("object-view-box", v) }
func (e *Element) Offset(v Value) *Element                   { return e.setValue("offset", v) }
func (e *Element) OffsetAnchor(v Value) *Element             { return e.setValue("offset-anchor", v) }
func (e *Element) OffsetDistance(v Value) *Element           { return e.setValue("offset-distance", v) }
func (e *Element) OffsetPath(v Value) *Element               { return e.setValue("offset-path", v) }
func (e *Element) OffsetPosition(v Value) *Element           { return e.setValue("offset-position", v) }
func (e *Element) OffsetRotate(v Value) *Element             { return e.setValue("offset-rotate", v) }
func (e *Element) Opacity(v NumericValue) *Element           { return e.setValue("opacity", v) }
func (e *Element) Order(v NumericValue) *Element             { return e.setValue("order", v) }
func (e *Element) Orphans(v NumericValue) *Element           { return e.setValue("orphans", v) }
func (e *Element) Outline(v Value) *Element                  { return e.setValue("outline", v) }
func (e *Element) OutlineColor(v ColorValue) *Element        { return e.setValue("outline-color", v) }
func (e *Element) OutlineOffset(v LengthValue) *Element      { return e.setValue("outline-offset", v) }
func (e *Element) OutlineStyle(v Value) *Element             { return e.setValue("outline-style", v) }
func (e *Element) OutlineWidth(v LengthValue) *Element       { return e.setValue("outline-width", v) }
func (e *Element) Overflow(v OverflowValue) *Element         { return e.setValue("overflow", v) }
func (e *Element) OverflowBlock(v Value) *Element            { return e.setValue("overflow-block", v) }
func (e *Element) OverflowClipMargin(v Value) *Element       { return e.setValue("overflow-clip-margin", v) }
func (e *Element) OverflowClipMarginBlock(v Value) *Element {
	return e.setValue("overflow-clip-margin-block", v)
}
func (e *Element) OverflowClipMarginBlockEnd(v Value) *Element {
	return e.setValue("overflow-clip-margin-block-end", v)
}
func (e *Element) OverflowClipMarginBlockStart(v Value) *Element {
	return e.setValue("overflow-clip-margin-block-start", v)
}
func (e *Element) OverflowClipMarginBottom(v Value) *Element {
	return e.setValue("overflow-clip-margin-bottom", v)
}
func (e *Element) OverflowClipMarginInline(v Value) *Element {
	return e.setValue("overflow-clip-margin-inline", v)
}
func (e *Element) OverflowClipMarginInlineEnd(v Value) *Element {
	return e.setValue("overflow-clip-margin-inline-end", v)
}
func (e *Element) OverflowClipMarginInlineStart(v Value) *Element {
	return e.setValue("overflow-clip-margin-inline-start", v)
}
func (e *Element) OverflowClipMarginLeft(v Value) *Element {
	return e.setValue("overflow-clip-margin-left", v)
}
func (e *Element) OverflowClipMarginRight(v Value) *Element {
	return e.setValue("overflow-clip-margin-right", v)
}
func (e *Element) OverflowClipMarginTop(v Value) *Element {
	return e.setValue("overflow-clip-margin-top", v)
}
func (e *Element) OverflowInline(v Value) *Element                 { return e.setValue("overflow-inline", v) }
func (e *Element) OverflowWrap(v OverflowWrapValue) *Element       { return e.setValue("overflow-wrap", v) }
func (e *Element) OverflowX(v OverflowXValue) *Element             { return e.setValue("overflow-x", v) }
func (e *Element) OverflowY(v OverflowYValue) *Element             { return e.setValue("overflow-y", v) }
func (e *Element) Overlay(v Value) *Element                        { return e.setValue("overlay", v) }
func (e *Element) Padding(v LengthPercentageSpacingValue) *Element { return e.setValue("padding", v) }
func (e *Element) PaddingBlock(v LengthPercentageSpacingValue) *Element {
	return e.setValue("padding-block", v)
}
func (e *Element) PaddingBlockEnd(v LengthPercentageValue) *Element {
	return e.setValue("padding-block-end", v)
}
func (e *Element) PaddingBlockStart(v LengthPercentageValue) *Element {
	return e.setValue("padding-block-start", v)
}
func (e *Element) PaddingBottom(v LengthPercentageValue) *Element {
	return e.setValue("padding-bottom", v)
}
func (e *Element) PaddingInline(v LengthPercentageSpacingValue) *Element {
	return e.setValue("padding-inline", v)
}
func (e *Element) PaddingInlineEnd(v LengthPercentageValue) *Element {
	return e.setValue("padding-inline-end", v)
}
func (e *Element) PaddingInlineStart(v LengthPercentageValue) *Element {
	return e.setValue("padding-inline-start", v)
}
func (e *Element) PaddingLeft(v LengthPercentageValue) *Element { return e.setValue("padding-left", v) }
func (e *Element) PaddingRight(v LengthPercentageValue) *Element {
	return e.setValue("padding-right", v)
}
func (e *Element) PaddingTop(v LengthPercentageValue) *Element { return e.setValue("padding-top", v) }
func (e *Element) PaintOrder(v Value) *Element                 { return e.setValue("paint-order", v) }
func (e *Element) PathLength(v Value) *Element                 { return e.setValue("path-length", v) }
func (e *Element) Perspective(v LengthValue) *Element          { return e.setValue("perspective", v) }
func (e *Element) PerspectiveOrigin(v Value) *Element          { return e.setValue("perspective-origin", v) }
func (e *Element) PlaceContent(v Value) *Element               { return e.setValue("place-content", v) }
func (e *Element) PlaceItems(v Value) *Element                 { return e.setValue("place-items", v) }
func (e *Element) PlaceSelf(v Value) *Element                  { return e.setValue("place-self", v) }
func (e *Element) PointerTimeline(v Value) *Element            { return e.setValue("pointer-timeline", v) }
func (e *Element) PointerTimelineAxis(v Value) *Element {
	return e.setValue("pointer-timeline-axis", v)
}
func (e *Element) PointerTimelineName(v Value) *Element {
	return e.setValue("pointer-timeline-name", v)
}
func (e *Element) Position(v PositionValue) *Element { return e.setValue("position", v) }
func (e *Element) PositionAnchor(v Value) *Element   { return e.setValue("position-anchor", v) }
func (e *Element) PositionArea(v Value) *Element     { return e.setValue("position-area", v) }
func (e *Element) PositionTry(v Value) *Element      { return e.setValue("position-try", v) }
func (e *Element) PositionTryFallbacks(v Value) *Element {
	return e.setValue("position-try-fallbacks", v)
}
func (e *Element) PositionTryOrder(v Value) *Element       { return e.setValue("position-try-order", v) }
func (e *Element) PositionVisibility(v Value) *Element     { return e.setValue("position-visibility", v) }
func (e *Element) PrintColorAdjust(v Value) *Element       { return e.setValue("print-color-adjust", v) }
func (e *Element) Quotes(v Value) *Element                 { return e.setValue("quotes", v) }
func (e *Element) R(v Value) *Element                      { return e.setValue("r", v) }
func (e *Element) ReadingFlow(v Value) *Element            { return e.setValue("reading-flow", v) }
func (e *Element) ReadingOrder(v Value) *Element           { return e.setValue("reading-order", v) }
func (e *Element) RegionFragment(v Value) *Element         { return e.setValue("region-fragment", v) }
func (e *Element) Right(v LengthPercentageValue) *Element  { return e.setValue("right", v) }
func (e *Element) Rotate(v Value) *Element                 { return e.setValue("rotate", v) }
func (e *Element) RowGap(v LengthPercentageValue) *Element { return e.setValue("row-gap", v) }
func (e *Element) RowRule(v Value) *Element                { return e.setValue("row-rule", v) }
func (e *Element) RowRuleBreak(v Value) *Element           { return e.setValue("row-rule-break", v) }
func (e *Element) RowRuleColor(v Value) *Element           { return e.setValue("row-rule-color", v) }
func (e *Element) RowRuleInset(v Value) *Element           { return e.setValue("row-rule-inset", v) }
func (e *Element) RowRuleInsetCap(v Value) *Element        { return e.setValue("row-rule-inset-cap", v) }
func (e *Element) RowRuleInsetCapEnd(v Value) *Element {
	return e.setValue("row-rule-inset-cap-end", v)
}
func (e *Element) RowRuleInsetCapStart(v Value) *Element {
	return e.setValue("row-rule-inset-cap-start", v)
}
func (e *Element) RowRuleInsetEnd(v Value) *Element { return e.setValue("row-rule-inset-end", v) }
func (e *Element) RowRuleInsetJunction(v Value) *Element {
	return e.setValue("row-rule-inset-junction", v)
}
func (e *Element) RowRuleInsetJunctionEnd(v Value) *Element {
	return e.setValue("row-rule-inset-junction-end", v)
}
func (e *Element) RowRuleInsetJunctionStart(v Value) *Element {
	return e.setValue("row-rule-inset-junction-start", v)
}
func (e *Element) RowRuleInsetStart(v Value) *Element { return e.setValue("row-rule-inset-start", v) }
func (e *Element) RowRuleStyle(v Value) *Element      { return e.setValue("row-rule-style", v) }
func (e *Element) RowRuleVisibilityItems(v Value) *Element {
	return e.setValue("row-rule-visibility-items", v)
}
func (e *Element) RowRuleWidth(v Value) *Element      { return e.setValue("row-rule-width", v) }
func (e *Element) RubyAlign(v Value) *Element         { return e.setValue("ruby-align", v) }
func (e *Element) RubyMerge(v Value) *Element         { return e.setValue("ruby-merge", v) }
func (e *Element) RubyOverhang(v Value) *Element      { return e.setValue("ruby-overhang", v) }
func (e *Element) RubyPosition(v Value) *Element      { return e.setValue("ruby-position", v) }
func (e *Element) Rule(v Value) *Element              { return e.setValue("rule", v) }
func (e *Element) RuleBreak(v Value) *Element         { return e.setValue("rule-break", v) }
func (e *Element) RuleColor(v Value) *Element         { return e.setValue("rule-color", v) }
func (e *Element) RuleInset(v Value) *Element         { return e.setValue("rule-inset", v) }
func (e *Element) RuleInsetCap(v Value) *Element      { return e.setValue("rule-inset-cap", v) }
func (e *Element) RuleInsetEnd(v Value) *Element      { return e.setValue("rule-inset-end", v) }
func (e *Element) RuleInsetJunction(v Value) *Element { return e.setValue("rule-inset-junction", v) }
func (e *Element) RuleInsetStart(v Value) *Element    { return e.setValue("rule-inset-start", v) }
func (e *Element) RuleOverlap(v Value) *Element       { return e.setValue("rule-overlap", v) }
func (e *Element) RuleStyle(v Value) *Element         { return e.setValue("rule-style", v) }
func (e *Element) RuleVisibilityItems(v Value) *Element {
	return e.setValue("rule-visibility-items", v)
}
func (e *Element) RuleWidth(v Value) *Element          { return e.setValue("rule-width", v) }
func (e *Element) Rx(v Value) *Element                 { return e.setValue("rx", v) }
func (e *Element) Ry(v Value) *Element                 { return e.setValue("ry", v) }
func (e *Element) Scale(v Value) *Element              { return e.setValue("scale", v) }
func (e *Element) ScrollMargin(v Value) *Element       { return e.setValue("scroll-margin", v) }
func (e *Element) ScrollMarginBottom(v Value) *Element { return e.setValue("scroll-margin-bottom", v) }
func (e *Element) ScrollMarginLeft(v Value) *Element   { return e.setValue("scroll-margin-left", v) }
func (e *Element) ScrollMarginRight(v Value) *Element  { return e.setValue("scroll-margin-right", v) }
func (e *Element) ScrollMarginTop(v Value) *Element    { return e.setValue("scroll-margin-top", v) }
func (e *Element) ScrollTimeline(v Value) *Element     { return e.setValue("scroll-timeline", v) }
func (e *Element) ScrollTimelineAxis(v Value) *Element { return e.setValue("scroll-timeline-axis", v) }
func (e *Element) ScrollTimelineName(v Value) *Element { return e.setValue("scroll-timeline-name", v) }
func (e *Element) ShapeImageThreshold(v Value) *Element {
	return e.setValue("shape-image-threshold", v)
}
func (e *Element) ShapeInside(v Value) *Element       { return e.setValue("shape-inside", v) }
func (e *Element) ShapeMargin(v Value) *Element       { return e.setValue("shape-margin", v) }
func (e *Element) ShapeOutside(v Value) *Element      { return e.setValue("shape-outside", v) }
func (e *Element) ShapePadding(v Value) *Element      { return e.setValue("shape-padding", v) }
func (e *Element) ShapeRendering(v Value) *Element    { return e.setValue("shape-rendering", v) }
func (e *Element) Size(v Value) *Element              { return e.setValue("size", v) }
func (e *Element) SliderOrientation(v Value) *Element { return e.setValue("slider-orientation", v) }
func (e *Element) SpatialNavigationAction(v Value) *Element {
	return e.setValue("spatial-navigation-action", v)
}
func (e *Element) SpatialNavigationContain(v Value) *Element {
	return e.setValue("spatial-navigation-contain", v)
}
func (e *Element) SpatialNavigationFunction(v Value) *Element {
	return e.setValue("spatial-navigation-function", v)
}
func (e *Element) StopColor(v Value) *Element            { return e.setValue("stop-color", v) }
func (e *Element) StopOpacity(v Value) *Element          { return e.setValue("stop-opacity", v) }
func (e *Element) StringSet(v Value) *Element            { return e.setValue("string-set", v) }
func (e *Element) Stroke(v ColorValue) *Element          { return e.setValue("stroke", v) }
func (e *Element) StrokeDasharray(v Value) *Element      { return e.setValue("stroke-dasharray", v) }
func (e *Element) StrokeDashoffset(v Value) *Element     { return e.setValue("stroke-dashoffset", v) }
func (e *Element) StrokeLinecap(v Value) *Element        { return e.setValue("stroke-linecap", v) }
func (e *Element) StrokeLinejoin(v Value) *Element       { return e.setValue("stroke-linejoin", v) }
func (e *Element) StrokeMiterlimit(v Value) *Element     { return e.setValue("stroke-miterlimit", v) }
func (e *Element) StrokeOpacity(v NumericValue) *Element { return e.setValue("stroke-opacity", v) }
func (e *Element) StrokeWidth(v LengthValue) *Element    { return e.setValue("stroke-width", v) }
func (e *Element) TabSize(v Value) *Element              { return e.setValue("tab-size", v) }
func (e *Element) TableLayout(v Value) *Element          { return e.setValue("table-layout", v) }
func (e *Element) TextAlign(v TextAlignValue) *Element   { return e.setValue("text-align", v) }
func (e *Element) TextAlignAll(v Value) *Element         { return e.setValue("text-align-all", v) }
func (e *Element) TextAlignLast(v Value) *Element        { return e.setValue("text-align-last", v) }
func (e *Element) TextAnchor(v Value) *Element           { return e.setValue("text-anchor", v) }
func (e *Element) TextAutospace(v Value) *Element        { return e.setValue("text-autospace", v) }
func (e *Element) TextBox(v Value) *Element              { return e.setValue("text-box", v) }
func (e *Element) TextBoxEdge(v Value) *Element          { return e.setValue("text-box-edge", v) }
func (e *Element) TextBoxTrim(v Value) *Element          { return e.setValue("text-box-trim", v) }
func (e *Element) TextCombineUpright(v Value) *Element   { return e.setValue("text-combine-upright", v) }
func (e *Element) TextDecoration(v TextDecorationValue) *Element {
	return e.setValue("text-decoration", v)
}
func (e *Element) TextDecorationColor(v ColorValue) *Element {
	return e.setValue("text-decoration-color", v)
}
func (e *Element) TextDecorationInset(v Value) *Element {
	return e.setValue("text-decoration-inset", v)
}
func (e *Element) TextDecorationLine(v Value) *Element { return e.setValue("text-decoration-line", v) }
func (e *Element) TextDecorationSkip(v Value) *Element { return e.setValue("text-decoration-skip", v) }
func (e *Element) TextDecorationSkipBox(v Value) *Element {
	return e.setValue("text-decoration-skip-box", v)
}
func (e *Element) TextDecorationSkipInk(v Value) *Element {
	return e.setValue("text-decoration-skip-ink", v)
}
func (e *Element) TextDecorationSkipSelf(v Value) *Element {
	return e.setValue("text-decoration-skip-self", v)
}
func (e *Element) TextDecorationSkipSpaces(v Value) *Element {
	return e.setValue("text-decoration-skip-spaces", v)
}
func (e *Element) TextDecorationStyle(v Value) *Element {
	return e.setValue("text-decoration-style", v)
}
func (e *Element) TextDecorationThickness(v Value) *Element {
	return e.setValue("text-decoration-thickness", v)
}
func (e *Element) TextEmphasis(v Value) *Element { return e.setValue("text-emphasis", v) }
func (e *Element) TextEmphasisColor(v ColorValue) *Element {
	return e.setValue("text-emphasis-color", v)
}
func (e *Element) TextEmphasisPosition(v Value) *Element {
	return e.setValue("text-emphasis-position", v)
}
func (e *Element) TextEmphasisSkip(v Value) *Element           { return e.setValue("text-emphasis-skip", v) }
func (e *Element) TextEmphasisStyle(v Value) *Element          { return e.setValue("text-emphasis-style", v) }
func (e *Element) TextFit(v Value) *Element                    { return e.setValue("text-fit", v) }
func (e *Element) TextGroupAlign(v Value) *Element             { return e.setValue("text-group-align", v) }
func (e *Element) TextIndent(v LengthPercentageValue) *Element { return e.setValue("text-indent", v) }
func (e *Element) TextJustify(v Value) *Element                { return e.setValue("text-justify", v) }
func (e *Element) TextOrientation(v Value) *Element            { return e.setValue("text-orientation", v) }
func (e *Element) TextOverflow(v TextOverflowValue) *Element   { return e.setValue("text-overflow", v) }
func (e *Element) TextRendering(v Value) *Element              { return e.setValue("text-rendering", v) }
func (e *Element) TextShadow(v Value) *Element                 { return e.setValue("text-shadow", v) }
func (e *Element) TextSizeAdjust(v Value) *Element             { return e.setValue("text-size-adjust", v) }
func (e *Element) TextSpacing(v Value) *Element                { return e.setValue("text-spacing", v) }
func (e *Element) TextSpacingTrim(v Value) *Element            { return e.setValue("text-spacing-trim", v) }
func (e *Element) TextTransform(v Value) *Element              { return e.setValue("text-transform", v) }
func (e *Element) TextUnderlineOffset(v Value) *Element {
	return e.setValue("text-underline-offset", v)
}
func (e *Element) TextUnderlinePosition(v Value) *Element {
	return e.setValue("text-underline-position", v)
}
func (e *Element) TextWrap(v Value) *Element        { return e.setValue("text-wrap", v) }
func (e *Element) TextWrapMode(v Value) *Element    { return e.setValue("text-wrap-mode", v) }
func (e *Element) TextWrapStyle(v Value) *Element   { return e.setValue("text-wrap-style", v) }
func (e *Element) TimelineScope(v Value) *Element   { return e.setValue("timeline-scope", v) }
func (e *Element) TimelineTrigger(v Value) *Element { return e.setValue("timeline-trigger", v) }
func (e *Element) TimelineTriggerActivationRange(v Value) *Element {
	return e.setValue("timeline-trigger-activation-range", v)
}
func (e *Element) TimelineTriggerActivationRangeEnd(v Value) *Element {
	return e.setValue("timeline-trigger-activation-range-end", v)
}
func (e *Element) TimelineTriggerActivationRangeStart(v Value) *Element {
	return e.setValue("timeline-trigger-activation-range-start", v)
}
func (e *Element) TimelineTriggerActiveRange(v Value) *Element {
	return e.setValue("timeline-trigger-active-range", v)
}
func (e *Element) TimelineTriggerActiveRangeEnd(v Value) *Element {
	return e.setValue("timeline-trigger-active-range-end", v)
}
func (e *Element) TimelineTriggerActiveRangeStart(v Value) *Element {
	return e.setValue("timeline-trigger-active-range-start", v)
}
func (e *Element) TimelineTriggerName(v Value) *Element {
	return e.setValue("timeline-trigger-name", v)
}
func (e *Element) TimelineTriggerSource(v Value) *Element {
	return e.setValue("timeline-trigger-source", v)
}
func (e *Element) Top(v LengthPercentageValue) *Element { return e.setValue("top", v) }
func (e *Element) Transform(v TransformValue) *Element  { return e.setValue("transform", v) }
func (e *Element) TransformBox(v Value) *Element        { return e.setValue("transform-box", v) }
func (e *Element) TransformOrigin(v Value) *Element     { return e.setValue("transform-origin", v) }
func (e *Element) TransformStyle(v Value) *Element      { return e.setValue("transform-style", v) }
func (e *Element) Transition(v Value) *Element          { return e.setValue("transition", v) }
func (e *Element) TransitionBehavior(v Value) *Element  { return e.setValue("transition-behavior", v) }
func (e *Element) TransitionDelay(v Value) *Element     { return e.setValue("transition-delay", v) }
func (e *Element) TransitionDuration(v Value) *Element  { return e.setValue("transition-duration", v) }
func (e *Element) TransitionProperty(v Value) *Element  { return e.setValue("transition-property", v) }
func (e *Element) TransitionTimingFunction(v Value) *Element {
	return e.setValue("transition-timing-function", v)
}
func (e *Element) Translate(v Value) *Element         { return e.setValue("translate", v) }
func (e *Element) TriggerScope(v Value) *Element      { return e.setValue("trigger-scope", v) }
func (e *Element) UnicodeBidi(v Value) *Element       { return e.setValue("unicode-bidi", v) }
func (e *Element) VectorEffect(v Value) *Element      { return e.setValue("vector-effect", v) }
func (e *Element) VerticalAlign(v Value) *Element     { return e.setValue("vertical-align", v) }
func (e *Element) ViewTimeline(v Value) *Element      { return e.setValue("view-timeline", v) }
func (e *Element) ViewTimelineAxis(v Value) *Element  { return e.setValue("view-timeline-axis", v) }
func (e *Element) ViewTimelineInset(v Value) *Element { return e.setValue("view-timeline-inset", v) }
func (e *Element) ViewTimelineName(v Value) *Element  { return e.setValue("view-timeline-name", v) }
func (e *Element) ViewTransitionClass(v Value) *Element {
	return e.setValue("view-transition-class", v)
}
func (e *Element) ViewTransitionGroup(v Value) *Element {
	return e.setValue("view-transition-group", v)
}
func (e *Element) ViewTransitionName(v Value) *Element { return e.setValue("view-transition-name", v) }
func (e *Element) ViewTransitionScope(v Value) *Element {
	return e.setValue("view-transition-scope", v)
}
func (e *Element) Visibility(v VisibilityValue) *Element { return e.setValue("visibility", v) }
func (e *Element) WhiteSpace(v WhiteSpaceValue) *Element { return e.setValue("white-space", v) }
func (e *Element) WhiteSpaceCollapse(v Value) *Element   { return e.setValue("white-space-collapse", v) }
func (e *Element) WhiteSpaceTrim(v Value) *Element       { return e.setValue("white-space-trim", v) }
func (e *Element) Widows(v NumericValue) *Element        { return e.setValue("widows", v) }
func (e *Element) Width(v WidthValue) *Element           { return e.setValue("width", v) }
func (e *Element) WillChange(v Value) *Element           { return e.setValue("will-change", v) }
func (e *Element) WindowDrag(v Value) *Element           { return e.setValue("window-drag", v) }
func (e *Element) WordBreak(v Value) *Element            { return e.setValue("word-break", v) }
func (e *Element) WordSpaceTransform(v Value) *Element   { return e.setValue("word-space-transform", v) }
func (e *Element) WordSpacing(v LengthValue) *Element    { return e.setValue("word-spacing", v) }
func (e *Element) WrapAfter(v Value) *Element            { return e.setValue("wrap-after", v) }
func (e *Element) WrapBefore(v Value) *Element           { return e.setValue("wrap-before", v) }
func (e *Element) WrapFlow(v Value) *Element             { return e.setValue("wrap-flow", v) }
func (e *Element) WrapInside(v Value) *Element           { return e.setValue("wrap-inside", v) }
func (e *Element) WrapThrough(v Value) *Element          { return e.setValue("wrap-through", v) }
func (e *Element) WritingMode(v Value) *Element          { return e.setValue("writing-mode", v) }
func (e *Element) X(v Value) *Element                    { return e.setValue("x", v) }
func (e *Element) Y(v Value) *Element                    { return e.setValue("y", v) }
func (e *Element) ZIndex(v NumericValue) *Element        { return e.setValue("z-index", v) }
func (e *Element) Zoom(v Value) *Element                 { return e.setValue("zoom", v) }
