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
		"-webkit-animation-iteration-count", "-webkit-animation-name", "-webkit-animation-play-state", "-webkit-animation-timing-function", "-webkit-backface-visibility",
		"-webkit-mask", "-webkit-mask-box-image", "-webkit-mask-box-image-outset", "-webkit-mask-box-image-repeat", "-webkit-mask-box-image-slice",
		"-webkit-mask-box-image-source", "-webkit-mask-box-image-width", "-webkit-mask-clip", "-webkit-mask-image", "-webkit-mask-origin",
		"-webkit-mask-position", "-webkit-mask-repeat", "-webkit-mask-size", "-webkit-perspective", "-webkit-perspective-origin",
		"-webkit-transform-style", "-webkit-transition", "-webkit-transition-delay", "-webkit-transition-duration", "-webkit-transition-property",
		"-webkit-transition-timing-function", "accent-color", "align-content", "align-items", "align-self",
		"alignment-baseline", "anchor-name", "anchor-scope", "animation", "animation-composition",
		"animation-delay", "animation-delay-end", "animation-delay-start", "animation-direction", "animation-duration",
		"animation-fill-mode", "animation-iteration-count", "animation-name", "animation-play-state", "animation-range",
		"animation-range-center", "animation-range-end", "animation-range-start", "animation-timeline", "animation-timing-function",
		"animation-trigger", "aspect-ratio", "backface-visibility", "background", "background-blend-mode",
		"background-clip", "background-color", "background-image", "background-origin", "background-position",
		"background-position-block", "background-position-inline", "background-position-x", "background-position-y", "background-repeat",
		"background-repeat-block", "background-repeat-inline", "background-repeat-x", "background-repeat-y", "background-size",
		"block-size", "border", "border-block", "border-block-color", "border-block-end",
		"border-block-end-color", "border-block-end-radius", "border-block-end-style", "border-block-end-width", "border-block-start",
		"border-block-start-color", "border-block-start-radius", "border-block-start-style", "border-block-start-width", "border-block-style",
		"border-block-width", "border-bottom", "border-bottom-color", "border-bottom-left-radius", "border-bottom-radius",
		"border-bottom-right-radius", "border-bottom-style", "border-bottom-width", "border-collapse", "border-color",
		"border-end-end-radius", "border-end-start-radius", "border-image", "border-image-outset", "border-image-repeat",
		"border-image-slice", "border-image-source", "border-image-width", "border-inline", "border-inline-color",
		"border-inline-end", "border-inline-end-color", "border-inline-end-radius", "border-inline-end-style", "border-inline-end-width",
		"border-inline-start", "border-inline-start-color", "border-inline-start-radius", "border-inline-start-style", "border-inline-start-width",
		"border-inline-style", "border-inline-width", "border-left", "border-left-color", "border-left-radius",
		"border-left-style", "border-left-width", "border-radius", "border-right", "border-right-color",
		"border-right-radius", "border-right-style", "border-right-width", "border-spacing", "border-start-end-radius",
		"border-start-start-radius", "border-style", "border-top", "border-top-color", "border-top-left-radius",
		"border-top-radius", "border-top-right-radius", "border-top-style", "border-top-width", "border-width",
		"bottom", "box-shadow", "box-shadow-blur", "box-shadow-color", "box-shadow-inset",
		"box-shadow-offset", "box-shadow-position", "box-shadow-spread", "box-sizing", "break-after",
		"break-before", "caption-side", "clear", "clip", "clip-path",
		"clip-rule", "color", "color-adjust", "color-interpolation", "color-interpolation-filters",
		"color-scheme", "column-count", "column-fill", "column-gap", "column-height",
		"column-rule", "column-rule-break", "column-rule-color", "column-rule-inset", "column-rule-inset-cap",
		"column-rule-inset-cap-end", "column-rule-inset-cap-start", "column-rule-inset-end", "column-rule-inset-junction", "column-rule-inset-junction-end",
		"column-rule-inset-junction-start", "column-rule-inset-start", "column-rule-style", "column-rule-width", "column-span",
		"column-width", "column-wrap", "columns", "contain", "contain-intrinsic-block-size",
		"contain-intrinsic-height", "contain-intrinsic-inline-size", "contain-intrinsic-size", "contain-intrinsic-width", "container",
		"container-name", "container-type", "content", "content-visibility", "corner",
		"corner-block-end", "corner-block-end-shape", "corner-block-start", "corner-block-start-shape", "corner-bottom",
		"corner-bottom-left", "corner-bottom-left-shape", "corner-bottom-right", "corner-bottom-right-shape", "corner-bottom-shape",
		"corner-end-end", "corner-end-end-shape", "corner-end-start", "corner-end-start-shape", "corner-inline-end",
		"corner-inline-end-shape", "corner-inline-start", "corner-inline-start-shape", "corner-left", "corner-left-shape",
		"corner-right", "corner-right-shape", "corner-shape", "corner-start-end", "corner-start-end-shape",
		"corner-start-start", "corner-start-start-shape", "corner-top", "corner-top-left", "corner-top-left-shape",
		"corner-top-right", "corner-top-right-shape", "corner-top-shape", "counter-increment", "counter-reset",
		"counter-set", "direction", "display", "dominant-baseline", "dynamic-range-limit",
		"empty-cells", "fill", "fill-opacity", "fill-rule", "filter",
		"flex", "flex-basis", "flex-direction", "flex-flow", "flex-grow",
		"flex-shrink", "flex-wrap", "float", "float-offset", "float-reference",
		"flow-from", "flow-into", "flow-tolerance", "font", "font-family",
		"font-feature-settings", "font-kerning", "font-language-override", "font-optical-sizing", "font-palette",
		"font-size", "font-size-adjust", "font-style", "font-synthesis", "font-synthesis-position",
		"font-synthesis-small-caps", "font-synthesis-style", "font-synthesis-weight", "font-variant", "font-variant-alternates",
		"font-variant-caps", "font-variant-east-asian", "font-variant-emoji", "font-variant-ligatures", "font-variant-numeric",
		"font-variant-position", "font-variation-settings", "font-weight", "font-width", "forced-color-adjust",
		"gap", "grid", "grid-area", "grid-auto-columns", "grid-auto-flow",
		"grid-auto-rows", "grid-column", "grid-column-end", "grid-column-start", "grid-row",
		"grid-row-end", "grid-row-start", "grid-template", "grid-template-areas", "grid-template-columns",
		"grid-template-rows", "hanging-punctuation", "height", "hyphenate-character", "hyphenate-limit-chars",
		"hyphenate-limit-last", "hyphenate-limit-lines", "hyphenate-limit-zone", "hyphens", "image-orientation",
		"image-resolution", "initial-letter", "initial-letter-align", "initial-letter-wrap", "inline-size",
		"inset", "inset-block", "inset-block-end", "inset-block-start", "inset-inline",
		"inset-inline-end", "inset-inline-start", "isolation", "justify-content", "justify-items",
		"justify-self", "left", "letter-spacing", "line-break", "line-clamp",
		"line-height", "list-style", "list-style-image", "list-style-position", "list-style-type",
		"margin", "margin-block", "margin-block-end", "margin-block-start", "margin-bottom",
		"margin-inline", "margin-inline-end", "margin-inline-start", "margin-left", "margin-right",
		"margin-top", "margin-trim", "mask", "mask-border", "mask-border-mode",
		"mask-border-outset", "mask-border-repeat", "mask-border-slice", "mask-border-source", "mask-border-width",
		"mask-clip", "mask-composite", "mask-image", "mask-mode", "mask-origin",
		"mask-position", "mask-repeat", "mask-size", "max-block-size", "max-height",
		"max-inline-size", "max-lines", "max-width", "min-block-size", "min-height",
		"min-inline-size", "min-width", "mix-blend-mode", "object-fit", "object-position",
		"object-view-box", "opacity", "order", "orphans", "outline",
		"outline-color", "outline-offset", "outline-style", "outline-width", "overflow",
		"overflow-block", "overflow-clip-margin", "overflow-clip-margin-block", "overflow-clip-margin-block-end", "overflow-clip-margin-block-start",
		"overflow-clip-margin-bottom", "overflow-clip-margin-inline", "overflow-clip-margin-inline-end", "overflow-clip-margin-inline-start", "overflow-clip-margin-left",
		"overflow-clip-margin-right", "overflow-clip-margin-top", "overflow-inline", "overflow-wrap", "overflow-x",
		"overflow-y", "padding", "padding-block", "padding-block-end", "padding-block-start",
		"padding-bottom", "padding-inline", "padding-inline-end", "padding-inline-start", "padding-left",
		"padding-right", "padding-top", "perspective", "perspective-origin", "place-content",
		"place-items", "place-self", "position", "position-anchor", "position-area",
		"position-try", "position-try-fallbacks", "position-try-order", "position-visibility", "print-color-adjust",
		"quotes", "region-fragment", "right", "rotate", "row-gap",
		"row-rule", "row-rule-break", "row-rule-color", "row-rule-inset", "row-rule-inset-cap",
		"row-rule-inset-cap-end", "row-rule-inset-cap-start", "row-rule-inset-end", "row-rule-inset-junction", "row-rule-inset-junction-end",
		"row-rule-inset-junction-start", "row-rule-inset-start", "row-rule-style", "row-rule-visibility-items", "row-rule-width",
		"ruby-align", "ruby-merge", "ruby-overhang", "ruby-position", "rule",
		"rule-break", "rule-color", "rule-inset", "rule-inset-cap", "rule-inset-end",
		"rule-inset-junction", "rule-inset-start", "rule-overlap", "rule-style", "rule-visibility-items",
		"rule-width", "scale", "scroll-margin", "scroll-margin-bottom", "scroll-margin-left",
		"scroll-margin-right", "scroll-margin-top", "shape-inside", "shape-margin", "shape-outside",
		"shape-padding", "shape-rendering", "stroke", "stroke-dasharray", "stroke-dashoffset",
		"stroke-linecap", "stroke-linejoin", "stroke-miterlimit", "stroke-opacity", "stroke-width",
		"tab-size", "table-layout", "text-align", "text-align-all", "text-align-last",
		"text-anchor", "text-autospace", "text-box", "text-box-edge", "text-box-trim",
		"text-combine-upright", "text-decoration", "text-decoration-color", "text-decoration-inset", "text-decoration-line",
		"text-decoration-skip", "text-decoration-skip-box", "text-decoration-skip-ink", "text-decoration-skip-self", "text-decoration-skip-spaces",
		"text-decoration-style", "text-decoration-thickness", "text-emphasis", "text-emphasis-color", "text-emphasis-position",
		"text-emphasis-skip", "text-emphasis-style", "text-group-align", "text-indent", "text-justify",
		"text-orientation", "text-overflow", "text-shadow", "text-spacing", "text-spacing-trim",
		"text-transform", "text-underline-offset", "text-underline-position", "text-wrap", "text-wrap-mode",
		"text-wrap-style", "top", "transform", "transform-box", "transform-origin",
		"transform-style", "transition", "transition-behavior", "transition-delay", "transition-duration",
		"transition-property", "transition-timing-function", "translate", "trigger-scope", "unicode-bidi",
		"vertical-align", "view-transition-class", "view-transition-group", "view-transition-name", "view-transition-scope",
		"visibility", "white-space", "white-space-collapse", "white-space-trim", "widows",
		"width", "word-break", "word-spacing", "wrap-after", "wrap-before",
		"wrap-flow", "wrap-inside", "wrap-through", "writing-mode", "z-index":
		return true
	default:
		return false
	}
}

// SupportedProperties lists the implemented CSS property names, sorted.
func SupportedProperties() []string {
	return []string{
		"-webkit-animation",
		"-webkit-animation-delay",
		"-webkit-animation-direction",
		"-webkit-animation-duration",
		"-webkit-animation-fill-mode",
		"-webkit-animation-iteration-count",
		"-webkit-animation-name",
		"-webkit-animation-play-state",
		"-webkit-animation-timing-function",
		"-webkit-backface-visibility",
		"-webkit-mask",
		"-webkit-mask-box-image",
		"-webkit-mask-box-image-outset",
		"-webkit-mask-box-image-repeat",
		"-webkit-mask-box-image-slice",
		"-webkit-mask-box-image-source",
		"-webkit-mask-box-image-width",
		"-webkit-mask-clip",
		"-webkit-mask-image",
		"-webkit-mask-origin",
		"-webkit-mask-position",
		"-webkit-mask-repeat",
		"-webkit-mask-size",
		"-webkit-perspective",
		"-webkit-perspective-origin",
		"-webkit-transform-style",
		"-webkit-transition",
		"-webkit-transition-delay",
		"-webkit-transition-duration",
		"-webkit-transition-property",
		"-webkit-transition-timing-function",
		"accent-color",
		"align-content",
		"align-items",
		"align-self",
		"alignment-baseline",
		"anchor-name",
		"anchor-scope",
		"animation",
		"animation-composition",
		"animation-delay",
		"animation-delay-end",
		"animation-delay-start",
		"animation-direction",
		"animation-duration",
		"animation-fill-mode",
		"animation-iteration-count",
		"animation-name",
		"animation-play-state",
		"animation-range",
		"animation-range-center",
		"animation-range-end",
		"animation-range-start",
		"animation-timeline",
		"animation-timing-function",
		"animation-trigger",
		"aspect-ratio",
		"backface-visibility",
		"background",
		"background-blend-mode",
		"background-clip",
		"background-color",
		"background-image",
		"background-origin",
		"background-position",
		"background-position-block",
		"background-position-inline",
		"background-position-x",
		"background-position-y",
		"background-repeat",
		"background-repeat-block",
		"background-repeat-inline",
		"background-repeat-x",
		"background-repeat-y",
		"background-size",
		"block-size",
		"border",
		"border-block",
		"border-block-color",
		"border-block-end",
		"border-block-end-color",
		"border-block-end-radius",
		"border-block-end-style",
		"border-block-end-width",
		"border-block-start",
		"border-block-start-color",
		"border-block-start-radius",
		"border-block-start-style",
		"border-block-start-width",
		"border-block-style",
		"border-block-width",
		"border-bottom",
		"border-bottom-color",
		"border-bottom-left-radius",
		"border-bottom-radius",
		"border-bottom-right-radius",
		"border-bottom-style",
		"border-bottom-width",
		"border-collapse",
		"border-color",
		"border-end-end-radius",
		"border-end-start-radius",
		"border-image",
		"border-image-outset",
		"border-image-repeat",
		"border-image-slice",
		"border-image-source",
		"border-image-width",
		"border-inline",
		"border-inline-color",
		"border-inline-end",
		"border-inline-end-color",
		"border-inline-end-radius",
		"border-inline-end-style",
		"border-inline-end-width",
		"border-inline-start",
		"border-inline-start-color",
		"border-inline-start-radius",
		"border-inline-start-style",
		"border-inline-start-width",
		"border-inline-style",
		"border-inline-width",
		"border-left",
		"border-left-color",
		"border-left-radius",
		"border-left-style",
		"border-left-width",
		"border-radius",
		"border-right",
		"border-right-color",
		"border-right-radius",
		"border-right-style",
		"border-right-width",
		"border-spacing",
		"border-start-end-radius",
		"border-start-start-radius",
		"border-style",
		"border-top",
		"border-top-color",
		"border-top-left-radius",
		"border-top-radius",
		"border-top-right-radius",
		"border-top-style",
		"border-top-width",
		"border-width",
		"bottom",
		"box-shadow",
		"box-shadow-blur",
		"box-shadow-color",
		"box-shadow-inset",
		"box-shadow-offset",
		"box-shadow-position",
		"box-shadow-spread",
		"box-sizing",
		"break-after",
		"break-before",
		"caption-side",
		"clear",
		"clip",
		"clip-path",
		"clip-rule",
		"color",
		"color-adjust",
		"color-interpolation",
		"color-interpolation-filters",
		"color-scheme",
		"column-count",
		"column-fill",
		"column-gap",
		"column-height",
		"column-rule",
		"column-rule-break",
		"column-rule-color",
		"column-rule-inset",
		"column-rule-inset-cap",
		"column-rule-inset-cap-end",
		"column-rule-inset-cap-start",
		"column-rule-inset-end",
		"column-rule-inset-junction",
		"column-rule-inset-junction-end",
		"column-rule-inset-junction-start",
		"column-rule-inset-start",
		"column-rule-style",
		"column-rule-width",
		"column-span",
		"column-width",
		"column-wrap",
		"columns",
		"contain",
		"contain-intrinsic-block-size",
		"contain-intrinsic-height",
		"contain-intrinsic-inline-size",
		"contain-intrinsic-size",
		"contain-intrinsic-width",
		"container",
		"container-name",
		"container-type",
		"content",
		"content-visibility",
		"corner",
		"corner-block-end",
		"corner-block-end-shape",
		"corner-block-start",
		"corner-block-start-shape",
		"corner-bottom",
		"corner-bottom-left",
		"corner-bottom-left-shape",
		"corner-bottom-right",
		"corner-bottom-right-shape",
		"corner-bottom-shape",
		"corner-end-end",
		"corner-end-end-shape",
		"corner-end-start",
		"corner-end-start-shape",
		"corner-inline-end",
		"corner-inline-end-shape",
		"corner-inline-start",
		"corner-inline-start-shape",
		"corner-left",
		"corner-left-shape",
		"corner-right",
		"corner-right-shape",
		"corner-shape",
		"corner-start-end",
		"corner-start-end-shape",
		"corner-start-start",
		"corner-start-start-shape",
		"corner-top",
		"corner-top-left",
		"corner-top-left-shape",
		"corner-top-right",
		"corner-top-right-shape",
		"corner-top-shape",
		"counter-increment",
		"counter-reset",
		"counter-set",
		"direction",
		"display",
		"dominant-baseline",
		"dynamic-range-limit",
		"empty-cells",
		"fill",
		"fill-opacity",
		"fill-rule",
		"filter",
		"flex",
		"flex-basis",
		"flex-direction",
		"flex-flow",
		"flex-grow",
		"flex-shrink",
		"flex-wrap",
		"float",
		"float-offset",
		"float-reference",
		"flow-from",
		"flow-into",
		"flow-tolerance",
		"font",
		"font-family",
		"font-feature-settings",
		"font-kerning",
		"font-language-override",
		"font-optical-sizing",
		"font-palette",
		"font-size",
		"font-size-adjust",
		"font-style",
		"font-synthesis",
		"font-synthesis-position",
		"font-synthesis-small-caps",
		"font-synthesis-style",
		"font-synthesis-weight",
		"font-variant",
		"font-variant-alternates",
		"font-variant-caps",
		"font-variant-east-asian",
		"font-variant-emoji",
		"font-variant-ligatures",
		"font-variant-numeric",
		"font-variant-position",
		"font-variation-settings",
		"font-weight",
		"font-width",
		"forced-color-adjust",
		"gap",
		"grid",
		"grid-area",
		"grid-auto-columns",
		"grid-auto-flow",
		"grid-auto-rows",
		"grid-column",
		"grid-column-end",
		"grid-column-start",
		"grid-row",
		"grid-row-end",
		"grid-row-start",
		"grid-template",
		"grid-template-areas",
		"grid-template-columns",
		"grid-template-rows",
		"hanging-punctuation",
		"height",
		"hyphenate-character",
		"hyphenate-limit-chars",
		"hyphenate-limit-last",
		"hyphenate-limit-lines",
		"hyphenate-limit-zone",
		"hyphens",
		"image-orientation",
		"image-resolution",
		"initial-letter",
		"initial-letter-align",
		"initial-letter-wrap",
		"inline-size",
		"inset",
		"inset-block",
		"inset-block-end",
		"inset-block-start",
		"inset-inline",
		"inset-inline-end",
		"inset-inline-start",
		"isolation",
		"justify-content",
		"justify-items",
		"justify-self",
		"left",
		"letter-spacing",
		"line-break",
		"line-clamp",
		"line-height",
		"list-style",
		"list-style-image",
		"list-style-position",
		"list-style-type",
		"margin",
		"margin-block",
		"margin-block-end",
		"margin-block-start",
		"margin-bottom",
		"margin-inline",
		"margin-inline-end",
		"margin-inline-start",
		"margin-left",
		"margin-right",
		"margin-top",
		"margin-trim",
		"mask",
		"mask-border",
		"mask-border-mode",
		"mask-border-outset",
		"mask-border-repeat",
		"mask-border-slice",
		"mask-border-source",
		"mask-border-width",
		"mask-clip",
		"mask-composite",
		"mask-image",
		"mask-mode",
		"mask-origin",
		"mask-position",
		"mask-repeat",
		"mask-size",
		"max-block-size",
		"max-height",
		"max-inline-size",
		"max-lines",
		"max-width",
		"min-block-size",
		"min-height",
		"min-inline-size",
		"min-width",
		"mix-blend-mode",
		"object-fit",
		"object-position",
		"object-view-box",
		"opacity",
		"order",
		"orphans",
		"outline",
		"outline-color",
		"outline-offset",
		"outline-style",
		"outline-width",
		"overflow",
		"overflow-block",
		"overflow-clip-margin",
		"overflow-clip-margin-block",
		"overflow-clip-margin-block-end",
		"overflow-clip-margin-block-start",
		"overflow-clip-margin-bottom",
		"overflow-clip-margin-inline",
		"overflow-clip-margin-inline-end",
		"overflow-clip-margin-inline-start",
		"overflow-clip-margin-left",
		"overflow-clip-margin-right",
		"overflow-clip-margin-top",
		"overflow-inline",
		"overflow-wrap",
		"overflow-x",
		"overflow-y",
		"padding",
		"padding-block",
		"padding-block-end",
		"padding-block-start",
		"padding-bottom",
		"padding-inline",
		"padding-inline-end",
		"padding-inline-start",
		"padding-left",
		"padding-right",
		"padding-top",
		"perspective",
		"perspective-origin",
		"place-content",
		"place-items",
		"place-self",
		"position",
		"position-anchor",
		"position-area",
		"position-try",
		"position-try-fallbacks",
		"position-try-order",
		"position-visibility",
		"print-color-adjust",
		"quotes",
		"region-fragment",
		"right",
		"rotate",
		"row-gap",
		"row-rule",
		"row-rule-break",
		"row-rule-color",
		"row-rule-inset",
		"row-rule-inset-cap",
		"row-rule-inset-cap-end",
		"row-rule-inset-cap-start",
		"row-rule-inset-end",
		"row-rule-inset-junction",
		"row-rule-inset-junction-end",
		"row-rule-inset-junction-start",
		"row-rule-inset-start",
		"row-rule-style",
		"row-rule-visibility-items",
		"row-rule-width",
		"ruby-align",
		"ruby-merge",
		"ruby-overhang",
		"ruby-position",
		"rule",
		"rule-break",
		"rule-color",
		"rule-inset",
		"rule-inset-cap",
		"rule-inset-end",
		"rule-inset-junction",
		"rule-inset-start",
		"rule-overlap",
		"rule-style",
		"rule-visibility-items",
		"rule-width",
		"scale",
		"scroll-margin",
		"scroll-margin-bottom",
		"scroll-margin-left",
		"scroll-margin-right",
		"scroll-margin-top",
		"shape-inside",
		"shape-margin",
		"shape-outside",
		"shape-padding",
		"shape-rendering",
		"stroke",
		"stroke-dasharray",
		"stroke-dashoffset",
		"stroke-linecap",
		"stroke-linejoin",
		"stroke-miterlimit",
		"stroke-opacity",
		"stroke-width",
		"tab-size",
		"table-layout",
		"text-align",
		"text-align-all",
		"text-align-last",
		"text-anchor",
		"text-autospace",
		"text-box",
		"text-box-edge",
		"text-box-trim",
		"text-combine-upright",
		"text-decoration",
		"text-decoration-color",
		"text-decoration-inset",
		"text-decoration-line",
		"text-decoration-skip",
		"text-decoration-skip-box",
		"text-decoration-skip-ink",
		"text-decoration-skip-self",
		"text-decoration-skip-spaces",
		"text-decoration-style",
		"text-decoration-thickness",
		"text-emphasis",
		"text-emphasis-color",
		"text-emphasis-position",
		"text-emphasis-skip",
		"text-emphasis-style",
		"text-group-align",
		"text-indent",
		"text-justify",
		"text-orientation",
		"text-overflow",
		"text-shadow",
		"text-spacing",
		"text-spacing-trim",
		"text-transform",
		"text-underline-offset",
		"text-underline-position",
		"text-wrap",
		"text-wrap-mode",
		"text-wrap-style",
		"top",
		"transform",
		"transform-box",
		"transform-origin",
		"transform-style",
		"transition",
		"transition-behavior",
		"transition-delay",
		"transition-duration",
		"transition-property",
		"transition-timing-function",
		"translate",
		"trigger-scope",
		"unicode-bidi",
		"vertical-align",
		"view-transition-class",
		"view-transition-group",
		"view-transition-name",
		"view-transition-scope",
		"visibility",
		"white-space",
		"white-space-collapse",
		"white-space-trim",
		"widows",
		"width",
		"word-break",
		"word-spacing",
		"wrap-after",
		"wrap-before",
		"wrap-flow",
		"wrap-inside",
		"wrap-through",
		"writing-mode",
		"z-index",
	}
}

// WebkitAnimationValue is a value for the CSS -webkit-animation property.
type WebkitAnimationValue interface {
	cssValue
	isWebkitAnimation()
}

// WebkitAnimation sets the CSS -webkit-animation property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitAnimation(v WebkitAnimationValue) *Element {
	return e.setValue("-webkit-animation", v)
}
func (expression) isWebkitAnimation()  {}
func (wideKeyword) isWebkitAnimation() {}

// WebkitAnimationDelayValue is a value for the CSS -webkit-animation-delay property.
type WebkitAnimationDelayValue interface {
	cssValue
	isWebkitAnimationDelay()
}

// WebkitAnimationDelay sets the CSS -webkit-animation-delay property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitAnimationDelay(v WebkitAnimationDelayValue) *Element {
	return e.setValue("-webkit-animation-delay", v)
}
func (expression) isWebkitAnimationDelay()  {}
func (wideKeyword) isWebkitAnimationDelay() {}

// WebkitAnimationDirectionValue is a value for the CSS -webkit-animation-direction property.
type WebkitAnimationDirectionValue interface {
	cssValue
	isWebkitAnimationDirection()
}

// WebkitAnimationDirection sets the CSS -webkit-animation-direction property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitAnimationDirection(v WebkitAnimationDirectionValue) *Element {
	return e.setValue("-webkit-animation-direction", v)
}
func (expression) isWebkitAnimationDirection()  {}
func (wideKeyword) isWebkitAnimationDirection() {}

// WebkitAnimationDurationValue is a value for the CSS -webkit-animation-duration property.
type WebkitAnimationDurationValue interface {
	cssValue
	isWebkitAnimationDuration()
}

// WebkitAnimationDuration sets the CSS -webkit-animation-duration property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitAnimationDuration(v WebkitAnimationDurationValue) *Element {
	return e.setValue("-webkit-animation-duration", v)
}
func (expression) isWebkitAnimationDuration()  {}
func (wideKeyword) isWebkitAnimationDuration() {}

// WebkitAnimationFillModeValue is a value for the CSS -webkit-animation-fill-mode property.
type WebkitAnimationFillModeValue interface {
	cssValue
	isWebkitAnimationFillMode()
}

// WebkitAnimationFillMode sets the CSS -webkit-animation-fill-mode property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitAnimationFillMode(v WebkitAnimationFillModeValue) *Element {
	return e.setValue("-webkit-animation-fill-mode", v)
}
func (expression) isWebkitAnimationFillMode()  {}
func (wideKeyword) isWebkitAnimationFillMode() {}

// WebkitAnimationIterationCountValue is a value for the CSS -webkit-animation-iteration-count property.
type WebkitAnimationIterationCountValue interface {
	cssValue
	isWebkitAnimationIterationCount()
}

// WebkitAnimationIterationCount sets the CSS -webkit-animation-iteration-count property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitAnimationIterationCount(v WebkitAnimationIterationCountValue) *Element {
	return e.setValue("-webkit-animation-iteration-count", v)
}
func (expression) isWebkitAnimationIterationCount()  {}
func (wideKeyword) isWebkitAnimationIterationCount() {}

// WebkitAnimationNameValue is a value for the CSS -webkit-animation-name property.
type WebkitAnimationNameValue interface {
	cssValue
	isWebkitAnimationName()
}

// WebkitAnimationName sets the CSS -webkit-animation-name property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitAnimationName(v WebkitAnimationNameValue) *Element {
	return e.setValue("-webkit-animation-name", v)
}
func (expression) isWebkitAnimationName()  {}
func (wideKeyword) isWebkitAnimationName() {}

// WebkitAnimationPlayStateValue is a value for the CSS -webkit-animation-play-state property.
type WebkitAnimationPlayStateValue interface {
	cssValue
	isWebkitAnimationPlayState()
}

// WebkitAnimationPlayState sets the CSS -webkit-animation-play-state property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitAnimationPlayState(v WebkitAnimationPlayStateValue) *Element {
	return e.setValue("-webkit-animation-play-state", v)
}
func (expression) isWebkitAnimationPlayState()  {}
func (wideKeyword) isWebkitAnimationPlayState() {}

// WebkitAnimationTimingFunctionValue is a value for the CSS -webkit-animation-timing-function property.
type WebkitAnimationTimingFunctionValue interface {
	cssValue
	isWebkitAnimationTimingFunction()
}

// WebkitAnimationTimingFunction sets the CSS -webkit-animation-timing-function property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitAnimationTimingFunction(v WebkitAnimationTimingFunctionValue) *Element {
	return e.setValue("-webkit-animation-timing-function", v)
}
func (expression) isWebkitAnimationTimingFunction()  {}
func (wideKeyword) isWebkitAnimationTimingFunction() {}

// WebkitBackfaceVisibilityValue is a value for the CSS -webkit-backface-visibility property.
type WebkitBackfaceVisibilityValue interface {
	cssValue
	isWebkitBackfaceVisibility()
}

// WebkitBackfaceVisibility sets the CSS -webkit-backface-visibility property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitBackfaceVisibility(v WebkitBackfaceVisibilityValue) *Element {
	return e.setValue("-webkit-backface-visibility", v)
}
func (expression) isWebkitBackfaceVisibility()  {}
func (wideKeyword) isWebkitBackfaceVisibility() {}

// WebkitMaskValue is a value for the CSS -webkit-mask property.
type WebkitMaskValue interface {
	cssValue
	isWebkitMask()
}

// WebkitMask sets the CSS -webkit-mask property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitMask(v WebkitMaskValue) *Element { return e.setValue("-webkit-mask", v) }
func (expression) isWebkitMask()                         {}
func (wideKeyword) isWebkitMask()                        {}

// WebkitMaskBoxImageValue is a value for the CSS -webkit-mask-box-image property.
type WebkitMaskBoxImageValue interface {
	cssValue
	isWebkitMaskBoxImage()
}

// WebkitMaskBoxImage sets the CSS -webkit-mask-box-image property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitMaskBoxImage(v WebkitMaskBoxImageValue) *Element {
	return e.setValue("-webkit-mask-box-image", v)
}
func (expression) isWebkitMaskBoxImage()  {}
func (wideKeyword) isWebkitMaskBoxImage() {}

// WebkitMaskBoxImageOutsetValue is a value for the CSS -webkit-mask-box-image-outset property.
type WebkitMaskBoxImageOutsetValue interface {
	cssValue
	isWebkitMaskBoxImageOutset()
}

// WebkitMaskBoxImageOutset sets the CSS -webkit-mask-box-image-outset property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitMaskBoxImageOutset(v WebkitMaskBoxImageOutsetValue) *Element {
	return e.setValue("-webkit-mask-box-image-outset", v)
}
func (expression) isWebkitMaskBoxImageOutset()  {}
func (wideKeyword) isWebkitMaskBoxImageOutset() {}

// WebkitMaskBoxImageRepeatValue is a value for the CSS -webkit-mask-box-image-repeat property.
type WebkitMaskBoxImageRepeatValue interface {
	cssValue
	isWebkitMaskBoxImageRepeat()
}

// WebkitMaskBoxImageRepeat sets the CSS -webkit-mask-box-image-repeat property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitMaskBoxImageRepeat(v WebkitMaskBoxImageRepeatValue) *Element {
	return e.setValue("-webkit-mask-box-image-repeat", v)
}
func (expression) isWebkitMaskBoxImageRepeat()  {}
func (wideKeyword) isWebkitMaskBoxImageRepeat() {}

// WebkitMaskBoxImageSliceValue is a value for the CSS -webkit-mask-box-image-slice property.
type WebkitMaskBoxImageSliceValue interface {
	cssValue
	isWebkitMaskBoxImageSlice()
}

// WebkitMaskBoxImageSlice sets the CSS -webkit-mask-box-image-slice property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitMaskBoxImageSlice(v WebkitMaskBoxImageSliceValue) *Element {
	return e.setValue("-webkit-mask-box-image-slice", v)
}
func (expression) isWebkitMaskBoxImageSlice()  {}
func (wideKeyword) isWebkitMaskBoxImageSlice() {}

// WebkitMaskBoxImageSourceValue is a value for the CSS -webkit-mask-box-image-source property.
type WebkitMaskBoxImageSourceValue interface {
	cssValue
	isWebkitMaskBoxImageSource()
}

// WebkitMaskBoxImageSource sets the CSS -webkit-mask-box-image-source property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitMaskBoxImageSource(v WebkitMaskBoxImageSourceValue) *Element {
	return e.setValue("-webkit-mask-box-image-source", v)
}
func (expression) isWebkitMaskBoxImageSource()  {}
func (wideKeyword) isWebkitMaskBoxImageSource() {}

// WebkitMaskBoxImageWidthValue is a value for the CSS -webkit-mask-box-image-width property.
type WebkitMaskBoxImageWidthValue interface {
	cssValue
	isWebkitMaskBoxImageWidth()
}

// WebkitMaskBoxImageWidth sets the CSS -webkit-mask-box-image-width property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitMaskBoxImageWidth(v WebkitMaskBoxImageWidthValue) *Element {
	return e.setValue("-webkit-mask-box-image-width", v)
}
func (expression) isWebkitMaskBoxImageWidth()  {}
func (wideKeyword) isWebkitMaskBoxImageWidth() {}

// WebkitMaskClipValue is a value for the CSS -webkit-mask-clip property.
type WebkitMaskClipValue interface {
	cssValue
	isWebkitMaskClip()
}

// WebkitMaskClip sets the CSS -webkit-mask-clip property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitMaskClip(v WebkitMaskClipValue) *Element {
	return e.setValue("-webkit-mask-clip", v)
}
func (expression) isWebkitMaskClip()  {}
func (wideKeyword) isWebkitMaskClip() {}

// WebkitMaskImageValue is a value for the CSS -webkit-mask-image property.
type WebkitMaskImageValue interface {
	cssValue
	isWebkitMaskImage()
}

// WebkitMaskImage sets the CSS -webkit-mask-image property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitMaskImage(v WebkitMaskImageValue) *Element {
	return e.setValue("-webkit-mask-image", v)
}
func (expression) isWebkitMaskImage()  {}
func (wideKeyword) isWebkitMaskImage() {}

// WebkitMaskOriginValue is a value for the CSS -webkit-mask-origin property.
type WebkitMaskOriginValue interface {
	cssValue
	isWebkitMaskOrigin()
}

// WebkitMaskOrigin sets the CSS -webkit-mask-origin property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitMaskOrigin(v WebkitMaskOriginValue) *Element {
	return e.setValue("-webkit-mask-origin", v)
}
func (expression) isWebkitMaskOrigin()  {}
func (wideKeyword) isWebkitMaskOrigin() {}

// WebkitMaskPositionValue is a value for the CSS -webkit-mask-position property.
type WebkitMaskPositionValue interface {
	cssValue
	isWebkitMaskPosition()
}

// WebkitMaskPosition sets the CSS -webkit-mask-position property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitMaskPosition(v WebkitMaskPositionValue) *Element {
	return e.setValue("-webkit-mask-position", v)
}
func (expression) isWebkitMaskPosition()  {}
func (wideKeyword) isWebkitMaskPosition() {}

// WebkitMaskRepeatValue is a value for the CSS -webkit-mask-repeat property.
type WebkitMaskRepeatValue interface {
	cssValue
	isWebkitMaskRepeat()
}

// WebkitMaskRepeat sets the CSS -webkit-mask-repeat property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitMaskRepeat(v WebkitMaskRepeatValue) *Element {
	return e.setValue("-webkit-mask-repeat", v)
}
func (expression) isWebkitMaskRepeat()  {}
func (wideKeyword) isWebkitMaskRepeat() {}

// WebkitMaskSizeValue is a value for the CSS -webkit-mask-size property.
type WebkitMaskSizeValue interface {
	cssValue
	isWebkitMaskSize()
}

// WebkitMaskSize sets the CSS -webkit-mask-size property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitMaskSize(v WebkitMaskSizeValue) *Element {
	return e.setValue("-webkit-mask-size", v)
}
func (expression) isWebkitMaskSize()  {}
func (wideKeyword) isWebkitMaskSize() {}

// WebkitPerspectiveValue is a value for the CSS -webkit-perspective property.
type WebkitPerspectiveValue interface {
	cssValue
	isWebkitPerspective()
}

// WebkitPerspective sets the CSS -webkit-perspective property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitPerspective(v WebkitPerspectiveValue) *Element {
	return e.setValue("-webkit-perspective", v)
}
func (expression) isWebkitPerspective()  {}
func (wideKeyword) isWebkitPerspective() {}

// WebkitPerspectiveOriginValue is a value for the CSS -webkit-perspective-origin property.
type WebkitPerspectiveOriginValue interface {
	cssValue
	isWebkitPerspectiveOrigin()
}

// WebkitPerspectiveOrigin sets the CSS -webkit-perspective-origin property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitPerspectiveOrigin(v WebkitPerspectiveOriginValue) *Element {
	return e.setValue("-webkit-perspective-origin", v)
}
func (expression) isWebkitPerspectiveOrigin()  {}
func (wideKeyword) isWebkitPerspectiveOrigin() {}

// WebkitTransformStyleValue is a value for the CSS -webkit-transform-style property.
type WebkitTransformStyleValue interface {
	cssValue
	isWebkitTransformStyle()
}

// WebkitTransformStyle sets the CSS -webkit-transform-style property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitTransformStyle(v WebkitTransformStyleValue) *Element {
	return e.setValue("-webkit-transform-style", v)
}
func (expression) isWebkitTransformStyle()  {}
func (wideKeyword) isWebkitTransformStyle() {}

// WebkitTransitionValue is a value for the CSS -webkit-transition property.
type WebkitTransitionValue interface {
	cssValue
	isWebkitTransition()
}

// WebkitTransition sets the CSS -webkit-transition property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitTransition(v WebkitTransitionValue) *Element {
	return e.setValue("-webkit-transition", v)
}
func (expression) isWebkitTransition()  {}
func (wideKeyword) isWebkitTransition() {}

// WebkitTransitionDelayValue is a value for the CSS -webkit-transition-delay property.
type WebkitTransitionDelayValue interface {
	cssValue
	isWebkitTransitionDelay()
}

// WebkitTransitionDelay sets the CSS -webkit-transition-delay property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitTransitionDelay(v WebkitTransitionDelayValue) *Element {
	return e.setValue("-webkit-transition-delay", v)
}
func (expression) isWebkitTransitionDelay()  {}
func (wideKeyword) isWebkitTransitionDelay() {}

// WebkitTransitionDurationValue is a value for the CSS -webkit-transition-duration property.
type WebkitTransitionDurationValue interface {
	cssValue
	isWebkitTransitionDuration()
}

// WebkitTransitionDuration sets the CSS -webkit-transition-duration property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitTransitionDuration(v WebkitTransitionDurationValue) *Element {
	return e.setValue("-webkit-transition-duration", v)
}
func (expression) isWebkitTransitionDuration()  {}
func (wideKeyword) isWebkitTransitionDuration() {}

// WebkitTransitionPropertyValue is a value for the CSS -webkit-transition-property property.
type WebkitTransitionPropertyValue interface {
	cssValue
	isWebkitTransitionProperty()
}

// WebkitTransitionProperty sets the CSS -webkit-transition-property property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitTransitionProperty(v WebkitTransitionPropertyValue) *Element {
	return e.setValue("-webkit-transition-property", v)
}
func (expression) isWebkitTransitionProperty()  {}
func (wideKeyword) isWebkitTransitionProperty() {}

// WebkitTransitionTimingFunctionValue is a value for the CSS -webkit-transition-timing-function property.
type WebkitTransitionTimingFunctionValue interface {
	cssValue
	isWebkitTransitionTimingFunction()
}

// WebkitTransitionTimingFunction sets the CSS -webkit-transition-timing-function property. Use RawValue for complex CSS syntax.
func (e *Element) WebkitTransitionTimingFunction(v WebkitTransitionTimingFunctionValue) *Element {
	return e.setValue("-webkit-transition-timing-function", v)
}
func (expression) isWebkitTransitionTimingFunction()  {}
func (wideKeyword) isWebkitTransitionTimingFunction() {}

// AccentColorValue is a value for the CSS accent-color property.
type AccentColorValue interface {
	cssValue
	isAccentColor()
}

// AccentColor sets the CSS accent-color property. Use RawValue for complex CSS syntax.
func (e *Element) AccentColor(v AccentColorValue) *Element { return e.setValue("accent-color", v) }
func (expression) isAccentColor()                          {}
func (wideKeyword) isAccentColor()                         {}

// AlignContentValue is a value for the CSS align-content property.
type AlignContentValue interface {
	cssValue
	isAlignContent()
}

// AlignContent sets the CSS align-content property. Use RawValue for complex CSS syntax.
func (e *Element) AlignContent(v AlignContentValue) *Element { return e.setValue("align-content", v) }
func (expression) isAlignContent()                           {}
func (wideKeyword) isAlignContent()                          {}

type alignContentKeyword uint16

const (
	// AlignContentNormal sets align-content to normal.
	AlignContentNormal alignContentKeyword = 1
	// AlignContentStart sets align-content to start.
	AlignContentStart alignContentKeyword = 2
	// AlignContentEnd sets align-content to end.
	AlignContentEnd alignContentKeyword = 3
	// AlignContentFlexStart sets align-content to flex-start.
	AlignContentFlexStart alignContentKeyword = 4
	// AlignContentFlexEnd sets align-content to flex-end.
	AlignContentFlexEnd alignContentKeyword = 5
	// AlignContentCenter sets align-content to center.
	AlignContentCenter alignContentKeyword = 6
	// AlignContentBaseline sets align-content to baseline.
	AlignContentBaseline alignContentKeyword = 7
	// AlignContentStretch sets align-content to stretch.
	AlignContentStretch alignContentKeyword = 8
	// AlignContentSpaceBetween sets align-content to space-between.
	AlignContentSpaceBetween alignContentKeyword = 9
	// AlignContentSpaceAround sets align-content to space-around.
	AlignContentSpaceAround alignContentKeyword = 10
	// AlignContentSpaceEvenly sets align-content to space-evenly.
	AlignContentSpaceEvenly alignContentKeyword = 11
)

func (alignContentKeyword) isAlignContent() {}
func (v alignContentKeyword) cssText() (string, error) {
	switch v {
	case AlignContentNormal:
		return "normal", nil
	case AlignContentStart:
		return "start", nil
	case AlignContentEnd:
		return "end", nil
	case AlignContentFlexStart:
		return "flex-start", nil
	case AlignContentFlexEnd:
		return "flex-end", nil
	case AlignContentCenter:
		return "center", nil
	case AlignContentBaseline:
		return "baseline", nil
	case AlignContentStretch:
		return "stretch", nil
	case AlignContentSpaceBetween:
		return "space-between", nil
	case AlignContentSpaceAround:
		return "space-around", nil
	case AlignContentSpaceEvenly:
		return "space-evenly", nil
	default:
		return "", fmt.Errorf("ui: invalid align-content keyword %d", v)
	}
}

// AlignItemsValue is a value for the CSS align-items property.
type AlignItemsValue interface {
	cssValue
	isAlignItems()
}

// AlignItems sets the CSS align-items property. Use RawValue for complex CSS syntax.
func (e *Element) AlignItems(v AlignItemsValue) *Element { return e.setValue("align-items", v) }
func (expression) isAlignItems()                         {}
func (wideKeyword) isAlignItems()                        {}

type alignItemsKeyword uint16

const (
	// AlignItemsNormal sets align-items to normal.
	AlignItemsNormal alignItemsKeyword = 1
	// AlignItemsStart sets align-items to start.
	AlignItemsStart alignItemsKeyword = 2
	// AlignItemsEnd sets align-items to end.
	AlignItemsEnd alignItemsKeyword = 3
	// AlignItemsFlexStart sets align-items to flex-start.
	AlignItemsFlexStart alignItemsKeyword = 4
	// AlignItemsFlexEnd sets align-items to flex-end.
	AlignItemsFlexEnd alignItemsKeyword = 5
	// AlignItemsCenter sets align-items to center.
	AlignItemsCenter alignItemsKeyword = 6
	// AlignItemsBaseline sets align-items to baseline.
	AlignItemsBaseline alignItemsKeyword = 7
	// AlignItemsStretch sets align-items to stretch.
	AlignItemsStretch alignItemsKeyword = 8
)

func (alignItemsKeyword) isAlignItems() {}
func (v alignItemsKeyword) cssText() (string, error) {
	switch v {
	case AlignItemsNormal:
		return "normal", nil
	case AlignItemsStart:
		return "start", nil
	case AlignItemsEnd:
		return "end", nil
	case AlignItemsFlexStart:
		return "flex-start", nil
	case AlignItemsFlexEnd:
		return "flex-end", nil
	case AlignItemsCenter:
		return "center", nil
	case AlignItemsBaseline:
		return "baseline", nil
	case AlignItemsStretch:
		return "stretch", nil
	default:
		return "", fmt.Errorf("ui: invalid align-items keyword %d", v)
	}
}

// AlignSelfValue is a value for the CSS align-self property.
type AlignSelfValue interface {
	cssValue
	isAlignSelf()
}

// AlignSelf sets the CSS align-self property. Use RawValue for complex CSS syntax.
func (e *Element) AlignSelf(v AlignSelfValue) *Element { return e.setValue("align-self", v) }
func (expression) isAlignSelf()                        {}
func (wideKeyword) isAlignSelf()                       {}

type alignSelfKeyword uint16

const (
	// AlignSelfAuto sets align-self to auto.
	AlignSelfAuto alignSelfKeyword = 1
	// AlignSelfNormal sets align-self to normal.
	AlignSelfNormal alignSelfKeyword = 2
	// AlignSelfStart sets align-self to start.
	AlignSelfStart alignSelfKeyword = 3
	// AlignSelfEnd sets align-self to end.
	AlignSelfEnd alignSelfKeyword = 4
	// AlignSelfFlexStart sets align-self to flex-start.
	AlignSelfFlexStart alignSelfKeyword = 5
	// AlignSelfFlexEnd sets align-self to flex-end.
	AlignSelfFlexEnd alignSelfKeyword = 6
	// AlignSelfCenter sets align-self to center.
	AlignSelfCenter alignSelfKeyword = 7
	// AlignSelfBaseline sets align-self to baseline.
	AlignSelfBaseline alignSelfKeyword = 8
	// AlignSelfStretch sets align-self to stretch.
	AlignSelfStretch alignSelfKeyword = 9
)

func (alignSelfKeyword) isAlignSelf() {}
func (v alignSelfKeyword) cssText() (string, error) {
	switch v {
	case AlignSelfAuto:
		return "auto", nil
	case AlignSelfNormal:
		return "normal", nil
	case AlignSelfStart:
		return "start", nil
	case AlignSelfEnd:
		return "end", nil
	case AlignSelfFlexStart:
		return "flex-start", nil
	case AlignSelfFlexEnd:
		return "flex-end", nil
	case AlignSelfCenter:
		return "center", nil
	case AlignSelfBaseline:
		return "baseline", nil
	case AlignSelfStretch:
		return "stretch", nil
	default:
		return "", fmt.Errorf("ui: invalid align-self keyword %d", v)
	}
}

// AlignmentBaselineValue is a value for the CSS alignment-baseline property.
type AlignmentBaselineValue interface {
	cssValue
	isAlignmentBaseline()
}

// AlignmentBaseline sets the CSS alignment-baseline property. Use RawValue for complex CSS syntax.
func (e *Element) AlignmentBaseline(v AlignmentBaselineValue) *Element {
	return e.setValue("alignment-baseline", v)
}
func (expression) isAlignmentBaseline()  {}
func (wideKeyword) isAlignmentBaseline() {}

// AnchorNameValue is a value for the CSS anchor-name property.
type AnchorNameValue interface {
	cssValue
	isAnchorName()
}

// AnchorName sets the CSS anchor-name property. Use RawValue for complex CSS syntax.
func (e *Element) AnchorName(v AnchorNameValue) *Element { return e.setValue("anchor-name", v) }
func (expression) isAnchorName()                         {}
func (wideKeyword) isAnchorName()                        {}

// AnchorScopeValue is a value for the CSS anchor-scope property.
type AnchorScopeValue interface {
	cssValue
	isAnchorScope()
}

// AnchorScope sets the CSS anchor-scope property. Use RawValue for complex CSS syntax.
func (e *Element) AnchorScope(v AnchorScopeValue) *Element { return e.setValue("anchor-scope", v) }
func (expression) isAnchorScope()                          {}
func (wideKeyword) isAnchorScope()                         {}

// AnimationValue is a value for the CSS animation property.
type AnimationValue interface {
	cssValue
	isAnimation()
}

// Animation sets the CSS animation property. Use RawValue for complex CSS syntax.
func (e *Element) Animation(v AnimationValue) *Element { return e.setValue("animation", v) }
func (expression) isAnimation()                        {}
func (wideKeyword) isAnimation()                       {}

// AnimationCompositionValue is a value for the CSS animation-composition property.
type AnimationCompositionValue interface {
	cssValue
	isAnimationComposition()
}

// AnimationComposition sets the CSS animation-composition property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationComposition(v AnimationCompositionValue) *Element {
	return e.setValue("animation-composition", v)
}
func (expression) isAnimationComposition()  {}
func (wideKeyword) isAnimationComposition() {}

// AnimationDelayValue is a value for the CSS animation-delay property.
type AnimationDelayValue interface {
	cssValue
	isAnimationDelay()
}

// AnimationDelay sets the CSS animation-delay property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationDelay(v AnimationDelayValue) *Element {
	return e.setValue("animation-delay", v)
}
func (expression) isAnimationDelay()  {}
func (wideKeyword) isAnimationDelay() {}

// AnimationDelayEndValue is a value for the CSS animation-delay-end property.
type AnimationDelayEndValue interface {
	cssValue
	isAnimationDelayEnd()
}

// AnimationDelayEnd sets the CSS animation-delay-end property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationDelayEnd(v AnimationDelayEndValue) *Element {
	return e.setValue("animation-delay-end", v)
}
func (expression) isAnimationDelayEnd()  {}
func (wideKeyword) isAnimationDelayEnd() {}

// AnimationDelayStartValue is a value for the CSS animation-delay-start property.
type AnimationDelayStartValue interface {
	cssValue
	isAnimationDelayStart()
}

// AnimationDelayStart sets the CSS animation-delay-start property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationDelayStart(v AnimationDelayStartValue) *Element {
	return e.setValue("animation-delay-start", v)
}
func (expression) isAnimationDelayStart()  {}
func (wideKeyword) isAnimationDelayStart() {}

// AnimationDirectionValue is a value for the CSS animation-direction property.
type AnimationDirectionValue interface {
	cssValue
	isAnimationDirection()
}

// AnimationDirection sets the CSS animation-direction property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationDirection(v AnimationDirectionValue) *Element {
	return e.setValue("animation-direction", v)
}
func (expression) isAnimationDirection()  {}
func (wideKeyword) isAnimationDirection() {}

// AnimationDurationValue is a value for the CSS animation-duration property.
type AnimationDurationValue interface {
	cssValue
	isAnimationDuration()
}

// AnimationDuration sets the CSS animation-duration property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationDuration(v AnimationDurationValue) *Element {
	return e.setValue("animation-duration", v)
}
func (expression) isAnimationDuration()  {}
func (wideKeyword) isAnimationDuration() {}

// AnimationFillModeValue is a value for the CSS animation-fill-mode property.
type AnimationFillModeValue interface {
	cssValue
	isAnimationFillMode()
}

// AnimationFillMode sets the CSS animation-fill-mode property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationFillMode(v AnimationFillModeValue) *Element {
	return e.setValue("animation-fill-mode", v)
}
func (expression) isAnimationFillMode()  {}
func (wideKeyword) isAnimationFillMode() {}

// AnimationIterationCountValue is a value for the CSS animation-iteration-count property.
type AnimationIterationCountValue interface {
	cssValue
	isAnimationIterationCount()
}

// AnimationIterationCount sets the CSS animation-iteration-count property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationIterationCount(v AnimationIterationCountValue) *Element {
	return e.setValue("animation-iteration-count", v)
}
func (expression) isAnimationIterationCount()  {}
func (wideKeyword) isAnimationIterationCount() {}

// AnimationNameValue is a value for the CSS animation-name property.
type AnimationNameValue interface {
	cssValue
	isAnimationName()
}

// AnimationName sets the CSS animation-name property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationName(v AnimationNameValue) *Element {
	return e.setValue("animation-name", v)
}
func (expression) isAnimationName()  {}
func (wideKeyword) isAnimationName() {}

// AnimationPlayStateValue is a value for the CSS animation-play-state property.
type AnimationPlayStateValue interface {
	cssValue
	isAnimationPlayState()
}

// AnimationPlayState sets the CSS animation-play-state property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationPlayState(v AnimationPlayStateValue) *Element {
	return e.setValue("animation-play-state", v)
}
func (expression) isAnimationPlayState()  {}
func (wideKeyword) isAnimationPlayState() {}

// AnimationRangeValue is a value for the CSS animation-range property.
type AnimationRangeValue interface {
	cssValue
	isAnimationRange()
}

// AnimationRange sets the CSS animation-range property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationRange(v AnimationRangeValue) *Element {
	return e.setValue("animation-range", v)
}
func (expression) isAnimationRange()  {}
func (wideKeyword) isAnimationRange() {}

// AnimationRangeCenterValue is a value for the CSS animation-range-center property.
type AnimationRangeCenterValue interface {
	cssValue
	isAnimationRangeCenter()
}

// AnimationRangeCenter sets the CSS animation-range-center property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationRangeCenter(v AnimationRangeCenterValue) *Element {
	return e.setValue("animation-range-center", v)
}
func (expression) isAnimationRangeCenter()  {}
func (wideKeyword) isAnimationRangeCenter() {}

// AnimationRangeEndValue is a value for the CSS animation-range-end property.
type AnimationRangeEndValue interface {
	cssValue
	isAnimationRangeEnd()
}

// AnimationRangeEnd sets the CSS animation-range-end property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationRangeEnd(v AnimationRangeEndValue) *Element {
	return e.setValue("animation-range-end", v)
}
func (expression) isAnimationRangeEnd()  {}
func (wideKeyword) isAnimationRangeEnd() {}

// AnimationRangeStartValue is a value for the CSS animation-range-start property.
type AnimationRangeStartValue interface {
	cssValue
	isAnimationRangeStart()
}

// AnimationRangeStart sets the CSS animation-range-start property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationRangeStart(v AnimationRangeStartValue) *Element {
	return e.setValue("animation-range-start", v)
}
func (expression) isAnimationRangeStart()  {}
func (wideKeyword) isAnimationRangeStart() {}

// AnimationTimelineValue is a value for the CSS animation-timeline property.
type AnimationTimelineValue interface {
	cssValue
	isAnimationTimeline()
}

// AnimationTimeline sets the CSS animation-timeline property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationTimeline(v AnimationTimelineValue) *Element {
	return e.setValue("animation-timeline", v)
}
func (expression) isAnimationTimeline()  {}
func (wideKeyword) isAnimationTimeline() {}

// AnimationTimingFunctionValue is a value for the CSS animation-timing-function property.
type AnimationTimingFunctionValue interface {
	cssValue
	isAnimationTimingFunction()
}

// AnimationTimingFunction sets the CSS animation-timing-function property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationTimingFunction(v AnimationTimingFunctionValue) *Element {
	return e.setValue("animation-timing-function", v)
}
func (expression) isAnimationTimingFunction()  {}
func (wideKeyword) isAnimationTimingFunction() {}

// AnimationTriggerValue is a value for the CSS animation-trigger property.
type AnimationTriggerValue interface {
	cssValue
	isAnimationTrigger()
}

// AnimationTrigger sets the CSS animation-trigger property. Use RawValue for complex CSS syntax.
func (e *Element) AnimationTrigger(v AnimationTriggerValue) *Element {
	return e.setValue("animation-trigger", v)
}
func (expression) isAnimationTrigger()  {}
func (wideKeyword) isAnimationTrigger() {}

// AspectRatioValue is a value for the CSS aspect-ratio property.
type AspectRatioValue interface {
	cssValue
	isAspectRatio()
}

// AspectRatio sets the CSS aspect-ratio property. Use RawValue for complex CSS syntax.
func (e *Element) AspectRatio(v AspectRatioValue) *Element { return e.setValue("aspect-ratio", v) }
func (expression) isAspectRatio()                          {}
func (wideKeyword) isAspectRatio()                         {}

// BackfaceVisibilityValue is a value for the CSS backface-visibility property.
type BackfaceVisibilityValue interface {
	cssValue
	isBackfaceVisibility()
}

// BackfaceVisibility sets the CSS backface-visibility property. Use RawValue for complex CSS syntax.
func (e *Element) BackfaceVisibility(v BackfaceVisibilityValue) *Element {
	return e.setValue("backface-visibility", v)
}
func (expression) isBackfaceVisibility()  {}
func (wideKeyword) isBackfaceVisibility() {}

// BackgroundValue is a value for the CSS background property.
type BackgroundValue interface {
	cssValue
	isBackground()
}

// Background sets the CSS background property. Use RawValue for complex CSS syntax.
func (e *Element) Background(v BackgroundValue) *Element { return e.setValue("background", v) }
func (expression) isBackground()                         {}
func (wideKeyword) isBackground()                        {}

// BackgroundBlendModeValue is a value for the CSS background-blend-mode property.
type BackgroundBlendModeValue interface {
	cssValue
	isBackgroundBlendMode()
}

// BackgroundBlendMode sets the CSS background-blend-mode property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundBlendMode(v BackgroundBlendModeValue) *Element {
	return e.setValue("background-blend-mode", v)
}
func (expression) isBackgroundBlendMode()  {}
func (wideKeyword) isBackgroundBlendMode() {}

// BackgroundClipValue is a value for the CSS background-clip property.
type BackgroundClipValue interface {
	cssValue
	isBackgroundClip()
}

// BackgroundClip sets the CSS background-clip property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundClip(v BackgroundClipValue) *Element {
	return e.setValue("background-clip", v)
}
func (expression) isBackgroundClip()  {}
func (wideKeyword) isBackgroundClip() {}

// BackgroundColorValue is a value for the CSS background-color property.
type BackgroundColorValue interface {
	cssValue
	isBackgroundColor()
}

// BackgroundColor sets the CSS background-color property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundColor(v BackgroundColorValue) *Element {
	return e.setValue("background-color", v)
}
func (expression) isBackgroundColor()  {}
func (wideKeyword) isBackgroundColor() {}
func (Color) isBackgroundColor()       {}

// BackgroundImageValue is a value for the CSS background-image property.
type BackgroundImageValue interface {
	cssValue
	isBackgroundImage()
}

// BackgroundImage sets the CSS background-image property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundImage(v BackgroundImageValue) *Element {
	return e.setValue("background-image", v)
}
func (expression) isBackgroundImage()  {}
func (wideKeyword) isBackgroundImage() {}

// BackgroundOriginValue is a value for the CSS background-origin property.
type BackgroundOriginValue interface {
	cssValue
	isBackgroundOrigin()
}

// BackgroundOrigin sets the CSS background-origin property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundOrigin(v BackgroundOriginValue) *Element {
	return e.setValue("background-origin", v)
}
func (expression) isBackgroundOrigin()  {}
func (wideKeyword) isBackgroundOrigin() {}

// BackgroundPositionValue is a value for the CSS background-position property.
type BackgroundPositionValue interface {
	cssValue
	isBackgroundPosition()
}

// BackgroundPosition sets the CSS background-position property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundPosition(v BackgroundPositionValue) *Element {
	return e.setValue("background-position", v)
}
func (expression) isBackgroundPosition()  {}
func (wideKeyword) isBackgroundPosition() {}

// BackgroundPositionBlockValue is a value for the CSS background-position-block property.
type BackgroundPositionBlockValue interface {
	cssValue
	isBackgroundPositionBlock()
}

// BackgroundPositionBlock sets the CSS background-position-block property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundPositionBlock(v BackgroundPositionBlockValue) *Element {
	return e.setValue("background-position-block", v)
}
func (expression) isBackgroundPositionBlock()  {}
func (wideKeyword) isBackgroundPositionBlock() {}

// BackgroundPositionInlineValue is a value for the CSS background-position-inline property.
type BackgroundPositionInlineValue interface {
	cssValue
	isBackgroundPositionInline()
}

// BackgroundPositionInline sets the CSS background-position-inline property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundPositionInline(v BackgroundPositionInlineValue) *Element {
	return e.setValue("background-position-inline", v)
}
func (expression) isBackgroundPositionInline()  {}
func (wideKeyword) isBackgroundPositionInline() {}

// BackgroundPositionXValue is a value for the CSS background-position-x property.
type BackgroundPositionXValue interface {
	cssValue
	isBackgroundPositionX()
}

// BackgroundPositionX sets the CSS background-position-x property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundPositionX(v BackgroundPositionXValue) *Element {
	return e.setValue("background-position-x", v)
}
func (expression) isBackgroundPositionX()  {}
func (wideKeyword) isBackgroundPositionX() {}

// BackgroundPositionYValue is a value for the CSS background-position-y property.
type BackgroundPositionYValue interface {
	cssValue
	isBackgroundPositionY()
}

// BackgroundPositionY sets the CSS background-position-y property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundPositionY(v BackgroundPositionYValue) *Element {
	return e.setValue("background-position-y", v)
}
func (expression) isBackgroundPositionY()  {}
func (wideKeyword) isBackgroundPositionY() {}

// BackgroundRepeatValue is a value for the CSS background-repeat property.
type BackgroundRepeatValue interface {
	cssValue
	isBackgroundRepeat()
}

// BackgroundRepeat sets the CSS background-repeat property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundRepeat(v BackgroundRepeatValue) *Element {
	return e.setValue("background-repeat", v)
}
func (expression) isBackgroundRepeat()  {}
func (wideKeyword) isBackgroundRepeat() {}

// BackgroundRepeatBlockValue is a value for the CSS background-repeat-block property.
type BackgroundRepeatBlockValue interface {
	cssValue
	isBackgroundRepeatBlock()
}

// BackgroundRepeatBlock sets the CSS background-repeat-block property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundRepeatBlock(v BackgroundRepeatBlockValue) *Element {
	return e.setValue("background-repeat-block", v)
}
func (expression) isBackgroundRepeatBlock()  {}
func (wideKeyword) isBackgroundRepeatBlock() {}

// BackgroundRepeatInlineValue is a value for the CSS background-repeat-inline property.
type BackgroundRepeatInlineValue interface {
	cssValue
	isBackgroundRepeatInline()
}

// BackgroundRepeatInline sets the CSS background-repeat-inline property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundRepeatInline(v BackgroundRepeatInlineValue) *Element {
	return e.setValue("background-repeat-inline", v)
}
func (expression) isBackgroundRepeatInline()  {}
func (wideKeyword) isBackgroundRepeatInline() {}

// BackgroundRepeatXValue is a value for the CSS background-repeat-x property.
type BackgroundRepeatXValue interface {
	cssValue
	isBackgroundRepeatX()
}

// BackgroundRepeatX sets the CSS background-repeat-x property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundRepeatX(v BackgroundRepeatXValue) *Element {
	return e.setValue("background-repeat-x", v)
}
func (expression) isBackgroundRepeatX()  {}
func (wideKeyword) isBackgroundRepeatX() {}

// BackgroundRepeatYValue is a value for the CSS background-repeat-y property.
type BackgroundRepeatYValue interface {
	cssValue
	isBackgroundRepeatY()
}

// BackgroundRepeatY sets the CSS background-repeat-y property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundRepeatY(v BackgroundRepeatYValue) *Element {
	return e.setValue("background-repeat-y", v)
}
func (expression) isBackgroundRepeatY()  {}
func (wideKeyword) isBackgroundRepeatY() {}

// BackgroundSizeValue is a value for the CSS background-size property.
type BackgroundSizeValue interface {
	cssValue
	isBackgroundSize()
}

// BackgroundSize sets the CSS background-size property. Use RawValue for complex CSS syntax.
func (e *Element) BackgroundSize(v BackgroundSizeValue) *Element {
	return e.setValue("background-size", v)
}
func (expression) isBackgroundSize()  {}
func (wideKeyword) isBackgroundSize() {}

// BlockSizeValue is a value for the CSS block-size property.
type BlockSizeValue interface {
	cssValue
	isBlockSize()
}

// BlockSize sets the CSS block-size property. Use RawValue for complex CSS syntax.
func (e *Element) BlockSize(v BlockSizeValue) *Element { return e.setValue("block-size", v) }
func (expression) isBlockSize()                        {}
func (wideKeyword) isBlockSize()                       {}
func (Length) isBlockSize()                            {}
func (Percentage) isBlockSize()                        {}

// BorderValue is a value for the CSS border property.
type BorderValue interface {
	cssValue
	isBorder()
}

// Border sets the CSS border property. Use RawValue for complex CSS syntax.
func (e *Element) Border(v BorderValue) *Element { return e.setValue("border", v) }
func (expression) isBorder()                     {}
func (wideKeyword) isBorder()                    {}
func (BorderStroke) isBorder()                   {}

// BorderBlockValue is a value for the CSS border-block property.
type BorderBlockValue interface {
	cssValue
	isBorderBlock()
}

// BorderBlock sets the CSS border-block property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlock(v BorderBlockValue) *Element { return e.setValue("border-block", v) }
func (expression) isBorderBlock()                          {}
func (wideKeyword) isBorderBlock()                         {}

// BorderBlockColorValue is a value for the CSS border-block-color property.
type BorderBlockColorValue interface {
	cssValue
	isBorderBlockColor()
}

// BorderBlockColor sets the CSS border-block-color property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlockColor(v BorderBlockColorValue) *Element {
	return e.setValue("border-block-color", v)
}
func (expression) isBorderBlockColor()  {}
func (wideKeyword) isBorderBlockColor() {}

// BorderBlockEndValue is a value for the CSS border-block-end property.
type BorderBlockEndValue interface {
	cssValue
	isBorderBlockEnd()
}

// BorderBlockEnd sets the CSS border-block-end property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlockEnd(v BorderBlockEndValue) *Element {
	return e.setValue("border-block-end", v)
}
func (expression) isBorderBlockEnd()  {}
func (wideKeyword) isBorderBlockEnd() {}

// BorderBlockEndColorValue is a value for the CSS border-block-end-color property.
type BorderBlockEndColorValue interface {
	cssValue
	isBorderBlockEndColor()
}

// BorderBlockEndColor sets the CSS border-block-end-color property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlockEndColor(v BorderBlockEndColorValue) *Element {
	return e.setValue("border-block-end-color", v)
}
func (expression) isBorderBlockEndColor()  {}
func (wideKeyword) isBorderBlockEndColor() {}

// BorderBlockEndRadiusValue is a value for the CSS border-block-end-radius property.
type BorderBlockEndRadiusValue interface {
	cssValue
	isBorderBlockEndRadius()
}

// BorderBlockEndRadius sets the CSS border-block-end-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlockEndRadius(v BorderBlockEndRadiusValue) *Element {
	return e.setValue("border-block-end-radius", v)
}
func (expression) isBorderBlockEndRadius()  {}
func (wideKeyword) isBorderBlockEndRadius() {}

// BorderBlockEndStyleValue is a value for the CSS border-block-end-style property.
type BorderBlockEndStyleValue interface {
	cssValue
	isBorderBlockEndStyle()
}

// BorderBlockEndStyle sets the CSS border-block-end-style property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlockEndStyle(v BorderBlockEndStyleValue) *Element {
	return e.setValue("border-block-end-style", v)
}
func (expression) isBorderBlockEndStyle()  {}
func (wideKeyword) isBorderBlockEndStyle() {}

// BorderBlockEndWidthValue is a value for the CSS border-block-end-width property.
type BorderBlockEndWidthValue interface {
	cssValue
	isBorderBlockEndWidth()
}

// BorderBlockEndWidth sets the CSS border-block-end-width property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlockEndWidth(v BorderBlockEndWidthValue) *Element {
	return e.setValue("border-block-end-width", v)
}
func (expression) isBorderBlockEndWidth()  {}
func (wideKeyword) isBorderBlockEndWidth() {}

// BorderBlockStartValue is a value for the CSS border-block-start property.
type BorderBlockStartValue interface {
	cssValue
	isBorderBlockStart()
}

// BorderBlockStart sets the CSS border-block-start property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlockStart(v BorderBlockStartValue) *Element {
	return e.setValue("border-block-start", v)
}
func (expression) isBorderBlockStart()  {}
func (wideKeyword) isBorderBlockStart() {}

// BorderBlockStartColorValue is a value for the CSS border-block-start-color property.
type BorderBlockStartColorValue interface {
	cssValue
	isBorderBlockStartColor()
}

// BorderBlockStartColor sets the CSS border-block-start-color property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlockStartColor(v BorderBlockStartColorValue) *Element {
	return e.setValue("border-block-start-color", v)
}
func (expression) isBorderBlockStartColor()  {}
func (wideKeyword) isBorderBlockStartColor() {}

// BorderBlockStartRadiusValue is a value for the CSS border-block-start-radius property.
type BorderBlockStartRadiusValue interface {
	cssValue
	isBorderBlockStartRadius()
}

// BorderBlockStartRadius sets the CSS border-block-start-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlockStartRadius(v BorderBlockStartRadiusValue) *Element {
	return e.setValue("border-block-start-radius", v)
}
func (expression) isBorderBlockStartRadius()  {}
func (wideKeyword) isBorderBlockStartRadius() {}

// BorderBlockStartStyleValue is a value for the CSS border-block-start-style property.
type BorderBlockStartStyleValue interface {
	cssValue
	isBorderBlockStartStyle()
}

// BorderBlockStartStyle sets the CSS border-block-start-style property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlockStartStyle(v BorderBlockStartStyleValue) *Element {
	return e.setValue("border-block-start-style", v)
}
func (expression) isBorderBlockStartStyle()  {}
func (wideKeyword) isBorderBlockStartStyle() {}

// BorderBlockStartWidthValue is a value for the CSS border-block-start-width property.
type BorderBlockStartWidthValue interface {
	cssValue
	isBorderBlockStartWidth()
}

// BorderBlockStartWidth sets the CSS border-block-start-width property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlockStartWidth(v BorderBlockStartWidthValue) *Element {
	return e.setValue("border-block-start-width", v)
}
func (expression) isBorderBlockStartWidth()  {}
func (wideKeyword) isBorderBlockStartWidth() {}

// BorderBlockStyleValue is a value for the CSS border-block-style property.
type BorderBlockStyleValue interface {
	cssValue
	isBorderBlockStyle()
}

// BorderBlockStyle sets the CSS border-block-style property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlockStyle(v BorderBlockStyleValue) *Element {
	return e.setValue("border-block-style", v)
}
func (expression) isBorderBlockStyle()  {}
func (wideKeyword) isBorderBlockStyle() {}

// BorderBlockWidthValue is a value for the CSS border-block-width property.
type BorderBlockWidthValue interface {
	cssValue
	isBorderBlockWidth()
}

// BorderBlockWidth sets the CSS border-block-width property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBlockWidth(v BorderBlockWidthValue) *Element {
	return e.setValue("border-block-width", v)
}
func (expression) isBorderBlockWidth()  {}
func (wideKeyword) isBorderBlockWidth() {}

// BorderBottomValue is a value for the CSS border-bottom property.
type BorderBottomValue interface {
	cssValue
	isBorderBottom()
}

// BorderBottom sets the CSS border-bottom property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBottom(v BorderBottomValue) *Element { return e.setValue("border-bottom", v) }
func (expression) isBorderBottom()                           {}
func (wideKeyword) isBorderBottom()                          {}
func (BorderStroke) isBorderBottom()                         {}

// BorderBottomColorValue is a value for the CSS border-bottom-color property.
type BorderBottomColorValue interface {
	cssValue
	isBorderBottomColor()
}

// BorderBottomColor sets the CSS border-bottom-color property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBottomColor(v BorderBottomColorValue) *Element {
	return e.setValue("border-bottom-color", v)
}
func (expression) isBorderBottomColor()  {}
func (wideKeyword) isBorderBottomColor() {}
func (Color) isBorderBottomColor()       {}

// BorderBottomLeftRadiusValue is a value for the CSS border-bottom-left-radius property.
type BorderBottomLeftRadiusValue interface {
	cssValue
	isBorderBottomLeftRadius()
}

// BorderBottomLeftRadius sets the CSS border-bottom-left-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBottomLeftRadius(v BorderBottomLeftRadiusValue) *Element {
	return e.setValue("border-bottom-left-radius", v)
}
func (expression) isBorderBottomLeftRadius()  {}
func (wideKeyword) isBorderBottomLeftRadius() {}
func (Length) isBorderBottomLeftRadius()      {}
func (Percentage) isBorderBottomLeftRadius()  {}

// BorderBottomRadiusValue is a value for the CSS border-bottom-radius property.
type BorderBottomRadiusValue interface {
	cssValue
	isBorderBottomRadius()
}

// BorderBottomRadius sets the CSS border-bottom-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBottomRadius(v BorderBottomRadiusValue) *Element {
	return e.setValue("border-bottom-radius", v)
}
func (expression) isBorderBottomRadius()  {}
func (wideKeyword) isBorderBottomRadius() {}

// BorderBottomRightRadiusValue is a value for the CSS border-bottom-right-radius property.
type BorderBottomRightRadiusValue interface {
	cssValue
	isBorderBottomRightRadius()
}

// BorderBottomRightRadius sets the CSS border-bottom-right-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBottomRightRadius(v BorderBottomRightRadiusValue) *Element {
	return e.setValue("border-bottom-right-radius", v)
}
func (expression) isBorderBottomRightRadius()  {}
func (wideKeyword) isBorderBottomRightRadius() {}
func (Length) isBorderBottomRightRadius()      {}
func (Percentage) isBorderBottomRightRadius()  {}

// BorderBottomStyleValue is a value for the CSS border-bottom-style property.
type BorderBottomStyleValue interface {
	cssValue
	isBorderBottomStyle()
}

// BorderBottomStyle sets the CSS border-bottom-style property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBottomStyle(v BorderBottomStyleValue) *Element {
	return e.setValue("border-bottom-style", v)
}
func (expression) isBorderBottomStyle()  {}
func (wideKeyword) isBorderBottomStyle() {}

// BorderBottomWidthValue is a value for the CSS border-bottom-width property.
type BorderBottomWidthValue interface {
	cssValue
	isBorderBottomWidth()
}

// BorderBottomWidth sets the CSS border-bottom-width property. Use RawValue for complex CSS syntax.
func (e *Element) BorderBottomWidth(v BorderBottomWidthValue) *Element {
	return e.setValue("border-bottom-width", v)
}
func (expression) isBorderBottomWidth()  {}
func (wideKeyword) isBorderBottomWidth() {}
func (Length) isBorderBottomWidth()      {}

// BorderCollapseValue is a value for the CSS border-collapse property.
type BorderCollapseValue interface {
	cssValue
	isBorderCollapse()
}

// BorderCollapse sets the CSS border-collapse property. Use RawValue for complex CSS syntax.
func (e *Element) BorderCollapse(v BorderCollapseValue) *Element {
	return e.setValue("border-collapse", v)
}
func (expression) isBorderCollapse()  {}
func (wideKeyword) isBorderCollapse() {}

type borderCollapseKeyword uint16

const (
	// BorderCollapseCollapse sets border-collapse to collapse.
	BorderCollapseCollapse borderCollapseKeyword = 1
	// BorderCollapseSeparate sets border-collapse to separate.
	BorderCollapseSeparate borderCollapseKeyword = 2
)

func (borderCollapseKeyword) isBorderCollapse() {}
func (v borderCollapseKeyword) cssText() (string, error) {
	switch v {
	case BorderCollapseCollapse:
		return "collapse", nil
	case BorderCollapseSeparate:
		return "separate", nil
	default:
		return "", fmt.Errorf("ui: invalid border-collapse keyword %d", v)
	}
}

// BorderColorValue is a value for the CSS border-color property.
type BorderColorValue interface {
	cssValue
	isBorderColor()
}

// BorderColor sets the CSS border-color property. Use RawValue for complex CSS syntax.
func (e *Element) BorderColor(v BorderColorValue) *Element { return e.setValue("border-color", v) }
func (expression) isBorderColor()                          {}
func (wideKeyword) isBorderColor()                         {}
func (Color) isBorderColor()                               {}

// BorderEndEndRadiusValue is a value for the CSS border-end-end-radius property.
type BorderEndEndRadiusValue interface {
	cssValue
	isBorderEndEndRadius()
}

// BorderEndEndRadius sets the CSS border-end-end-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderEndEndRadius(v BorderEndEndRadiusValue) *Element {
	return e.setValue("border-end-end-radius", v)
}
func (expression) isBorderEndEndRadius()  {}
func (wideKeyword) isBorderEndEndRadius() {}

// BorderEndStartRadiusValue is a value for the CSS border-end-start-radius property.
type BorderEndStartRadiusValue interface {
	cssValue
	isBorderEndStartRadius()
}

// BorderEndStartRadius sets the CSS border-end-start-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderEndStartRadius(v BorderEndStartRadiusValue) *Element {
	return e.setValue("border-end-start-radius", v)
}
func (expression) isBorderEndStartRadius()  {}
func (wideKeyword) isBorderEndStartRadius() {}

// BorderImageValue is a value for the CSS border-image property.
type BorderImageValue interface {
	cssValue
	isBorderImage()
}

// BorderImage sets the CSS border-image property. Use RawValue for complex CSS syntax.
func (e *Element) BorderImage(v BorderImageValue) *Element { return e.setValue("border-image", v) }
func (expression) isBorderImage()                          {}
func (wideKeyword) isBorderImage()                         {}

// BorderImageOutsetValue is a value for the CSS border-image-outset property.
type BorderImageOutsetValue interface {
	cssValue
	isBorderImageOutset()
}

// BorderImageOutset sets the CSS border-image-outset property. Use RawValue for complex CSS syntax.
func (e *Element) BorderImageOutset(v BorderImageOutsetValue) *Element {
	return e.setValue("border-image-outset", v)
}
func (expression) isBorderImageOutset()  {}
func (wideKeyword) isBorderImageOutset() {}

// BorderImageRepeatValue is a value for the CSS border-image-repeat property.
type BorderImageRepeatValue interface {
	cssValue
	isBorderImageRepeat()
}

// BorderImageRepeat sets the CSS border-image-repeat property. Use RawValue for complex CSS syntax.
func (e *Element) BorderImageRepeat(v BorderImageRepeatValue) *Element {
	return e.setValue("border-image-repeat", v)
}
func (expression) isBorderImageRepeat()  {}
func (wideKeyword) isBorderImageRepeat() {}

// BorderImageSliceValue is a value for the CSS border-image-slice property.
type BorderImageSliceValue interface {
	cssValue
	isBorderImageSlice()
}

// BorderImageSlice sets the CSS border-image-slice property. Use RawValue for complex CSS syntax.
func (e *Element) BorderImageSlice(v BorderImageSliceValue) *Element {
	return e.setValue("border-image-slice", v)
}
func (expression) isBorderImageSlice()  {}
func (wideKeyword) isBorderImageSlice() {}

// BorderImageSourceValue is a value for the CSS border-image-source property.
type BorderImageSourceValue interface {
	cssValue
	isBorderImageSource()
}

// BorderImageSource sets the CSS border-image-source property. Use RawValue for complex CSS syntax.
func (e *Element) BorderImageSource(v BorderImageSourceValue) *Element {
	return e.setValue("border-image-source", v)
}
func (expression) isBorderImageSource()  {}
func (wideKeyword) isBorderImageSource() {}

// BorderImageWidthValue is a value for the CSS border-image-width property.
type BorderImageWidthValue interface {
	cssValue
	isBorderImageWidth()
}

// BorderImageWidth sets the CSS border-image-width property. Use RawValue for complex CSS syntax.
func (e *Element) BorderImageWidth(v BorderImageWidthValue) *Element {
	return e.setValue("border-image-width", v)
}
func (expression) isBorderImageWidth()  {}
func (wideKeyword) isBorderImageWidth() {}

// BorderInlineValue is a value for the CSS border-inline property.
type BorderInlineValue interface {
	cssValue
	isBorderInline()
}

// BorderInline sets the CSS border-inline property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInline(v BorderInlineValue) *Element { return e.setValue("border-inline", v) }
func (expression) isBorderInline()                           {}
func (wideKeyword) isBorderInline()                          {}

// BorderInlineColorValue is a value for the CSS border-inline-color property.
type BorderInlineColorValue interface {
	cssValue
	isBorderInlineColor()
}

// BorderInlineColor sets the CSS border-inline-color property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInlineColor(v BorderInlineColorValue) *Element {
	return e.setValue("border-inline-color", v)
}
func (expression) isBorderInlineColor()  {}
func (wideKeyword) isBorderInlineColor() {}

// BorderInlineEndValue is a value for the CSS border-inline-end property.
type BorderInlineEndValue interface {
	cssValue
	isBorderInlineEnd()
}

// BorderInlineEnd sets the CSS border-inline-end property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInlineEnd(v BorderInlineEndValue) *Element {
	return e.setValue("border-inline-end", v)
}
func (expression) isBorderInlineEnd()  {}
func (wideKeyword) isBorderInlineEnd() {}

// BorderInlineEndColorValue is a value for the CSS border-inline-end-color property.
type BorderInlineEndColorValue interface {
	cssValue
	isBorderInlineEndColor()
}

// BorderInlineEndColor sets the CSS border-inline-end-color property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInlineEndColor(v BorderInlineEndColorValue) *Element {
	return e.setValue("border-inline-end-color", v)
}
func (expression) isBorderInlineEndColor()  {}
func (wideKeyword) isBorderInlineEndColor() {}

// BorderInlineEndRadiusValue is a value for the CSS border-inline-end-radius property.
type BorderInlineEndRadiusValue interface {
	cssValue
	isBorderInlineEndRadius()
}

// BorderInlineEndRadius sets the CSS border-inline-end-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInlineEndRadius(v BorderInlineEndRadiusValue) *Element {
	return e.setValue("border-inline-end-radius", v)
}
func (expression) isBorderInlineEndRadius()  {}
func (wideKeyword) isBorderInlineEndRadius() {}

// BorderInlineEndStyleValue is a value for the CSS border-inline-end-style property.
type BorderInlineEndStyleValue interface {
	cssValue
	isBorderInlineEndStyle()
}

// BorderInlineEndStyle sets the CSS border-inline-end-style property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInlineEndStyle(v BorderInlineEndStyleValue) *Element {
	return e.setValue("border-inline-end-style", v)
}
func (expression) isBorderInlineEndStyle()  {}
func (wideKeyword) isBorderInlineEndStyle() {}

// BorderInlineEndWidthValue is a value for the CSS border-inline-end-width property.
type BorderInlineEndWidthValue interface {
	cssValue
	isBorderInlineEndWidth()
}

// BorderInlineEndWidth sets the CSS border-inline-end-width property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInlineEndWidth(v BorderInlineEndWidthValue) *Element {
	return e.setValue("border-inline-end-width", v)
}
func (expression) isBorderInlineEndWidth()  {}
func (wideKeyword) isBorderInlineEndWidth() {}

// BorderInlineStartValue is a value for the CSS border-inline-start property.
type BorderInlineStartValue interface {
	cssValue
	isBorderInlineStart()
}

// BorderInlineStart sets the CSS border-inline-start property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInlineStart(v BorderInlineStartValue) *Element {
	return e.setValue("border-inline-start", v)
}
func (expression) isBorderInlineStart()  {}
func (wideKeyword) isBorderInlineStart() {}

// BorderInlineStartColorValue is a value for the CSS border-inline-start-color property.
type BorderInlineStartColorValue interface {
	cssValue
	isBorderInlineStartColor()
}

// BorderInlineStartColor sets the CSS border-inline-start-color property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInlineStartColor(v BorderInlineStartColorValue) *Element {
	return e.setValue("border-inline-start-color", v)
}
func (expression) isBorderInlineStartColor()  {}
func (wideKeyword) isBorderInlineStartColor() {}

// BorderInlineStartRadiusValue is a value for the CSS border-inline-start-radius property.
type BorderInlineStartRadiusValue interface {
	cssValue
	isBorderInlineStartRadius()
}

// BorderInlineStartRadius sets the CSS border-inline-start-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInlineStartRadius(v BorderInlineStartRadiusValue) *Element {
	return e.setValue("border-inline-start-radius", v)
}
func (expression) isBorderInlineStartRadius()  {}
func (wideKeyword) isBorderInlineStartRadius() {}

// BorderInlineStartStyleValue is a value for the CSS border-inline-start-style property.
type BorderInlineStartStyleValue interface {
	cssValue
	isBorderInlineStartStyle()
}

// BorderInlineStartStyle sets the CSS border-inline-start-style property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInlineStartStyle(v BorderInlineStartStyleValue) *Element {
	return e.setValue("border-inline-start-style", v)
}
func (expression) isBorderInlineStartStyle()  {}
func (wideKeyword) isBorderInlineStartStyle() {}

// BorderInlineStartWidthValue is a value for the CSS border-inline-start-width property.
type BorderInlineStartWidthValue interface {
	cssValue
	isBorderInlineStartWidth()
}

// BorderInlineStartWidth sets the CSS border-inline-start-width property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInlineStartWidth(v BorderInlineStartWidthValue) *Element {
	return e.setValue("border-inline-start-width", v)
}
func (expression) isBorderInlineStartWidth()  {}
func (wideKeyword) isBorderInlineStartWidth() {}

// BorderInlineStyleValue is a value for the CSS border-inline-style property.
type BorderInlineStyleValue interface {
	cssValue
	isBorderInlineStyle()
}

// BorderInlineStyle sets the CSS border-inline-style property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInlineStyle(v BorderInlineStyleValue) *Element {
	return e.setValue("border-inline-style", v)
}
func (expression) isBorderInlineStyle()  {}
func (wideKeyword) isBorderInlineStyle() {}

// BorderInlineWidthValue is a value for the CSS border-inline-width property.
type BorderInlineWidthValue interface {
	cssValue
	isBorderInlineWidth()
}

// BorderInlineWidth sets the CSS border-inline-width property. Use RawValue for complex CSS syntax.
func (e *Element) BorderInlineWidth(v BorderInlineWidthValue) *Element {
	return e.setValue("border-inline-width", v)
}
func (expression) isBorderInlineWidth()  {}
func (wideKeyword) isBorderInlineWidth() {}

// BorderLeftValue is a value for the CSS border-left property.
type BorderLeftValue interface {
	cssValue
	isBorderLeft()
}

// BorderLeft sets the CSS border-left property. Use RawValue for complex CSS syntax.
func (e *Element) BorderLeft(v BorderLeftValue) *Element { return e.setValue("border-left", v) }
func (expression) isBorderLeft()                         {}
func (wideKeyword) isBorderLeft()                        {}
func (BorderStroke) isBorderLeft()                       {}

// BorderLeftColorValue is a value for the CSS border-left-color property.
type BorderLeftColorValue interface {
	cssValue
	isBorderLeftColor()
}

// BorderLeftColor sets the CSS border-left-color property. Use RawValue for complex CSS syntax.
func (e *Element) BorderLeftColor(v BorderLeftColorValue) *Element {
	return e.setValue("border-left-color", v)
}
func (expression) isBorderLeftColor()  {}
func (wideKeyword) isBorderLeftColor() {}
func (Color) isBorderLeftColor()       {}

// BorderLeftRadiusValue is a value for the CSS border-left-radius property.
type BorderLeftRadiusValue interface {
	cssValue
	isBorderLeftRadius()
}

// BorderLeftRadius sets the CSS border-left-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderLeftRadius(v BorderLeftRadiusValue) *Element {
	return e.setValue("border-left-radius", v)
}
func (expression) isBorderLeftRadius()  {}
func (wideKeyword) isBorderLeftRadius() {}

// BorderLeftStyleValue is a value for the CSS border-left-style property.
type BorderLeftStyleValue interface {
	cssValue
	isBorderLeftStyle()
}

// BorderLeftStyle sets the CSS border-left-style property. Use RawValue for complex CSS syntax.
func (e *Element) BorderLeftStyle(v BorderLeftStyleValue) *Element {
	return e.setValue("border-left-style", v)
}
func (expression) isBorderLeftStyle()  {}
func (wideKeyword) isBorderLeftStyle() {}

// BorderLeftWidthValue is a value for the CSS border-left-width property.
type BorderLeftWidthValue interface {
	cssValue
	isBorderLeftWidth()
}

// BorderLeftWidth sets the CSS border-left-width property. Use RawValue for complex CSS syntax.
func (e *Element) BorderLeftWidth(v BorderLeftWidthValue) *Element {
	return e.setValue("border-left-width", v)
}
func (expression) isBorderLeftWidth()  {}
func (wideKeyword) isBorderLeftWidth() {}
func (Length) isBorderLeftWidth()      {}

// BorderRadiusValue is a value for the CSS border-radius property.
type BorderRadiusValue interface {
	cssValue
	isBorderRadius()
}

// BorderRadius sets the CSS border-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderRadius(v BorderRadiusValue) *Element { return e.setValue("border-radius", v) }
func (expression) isBorderRadius()                           {}
func (wideKeyword) isBorderRadius()                          {}
func (Length) isBorderRadius()                               {}
func (Percentage) isBorderRadius()                           {}
func (Spacing) isBorderRadius()                              {}

// BorderRightValue is a value for the CSS border-right property.
type BorderRightValue interface {
	cssValue
	isBorderRight()
}

// BorderRight sets the CSS border-right property. Use RawValue for complex CSS syntax.
func (e *Element) BorderRight(v BorderRightValue) *Element { return e.setValue("border-right", v) }
func (expression) isBorderRight()                          {}
func (wideKeyword) isBorderRight()                         {}
func (BorderStroke) isBorderRight()                        {}

// BorderRightColorValue is a value for the CSS border-right-color property.
type BorderRightColorValue interface {
	cssValue
	isBorderRightColor()
}

// BorderRightColor sets the CSS border-right-color property. Use RawValue for complex CSS syntax.
func (e *Element) BorderRightColor(v BorderRightColorValue) *Element {
	return e.setValue("border-right-color", v)
}
func (expression) isBorderRightColor()  {}
func (wideKeyword) isBorderRightColor() {}
func (Color) isBorderRightColor()       {}

// BorderRightRadiusValue is a value for the CSS border-right-radius property.
type BorderRightRadiusValue interface {
	cssValue
	isBorderRightRadius()
}

// BorderRightRadius sets the CSS border-right-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderRightRadius(v BorderRightRadiusValue) *Element {
	return e.setValue("border-right-radius", v)
}
func (expression) isBorderRightRadius()  {}
func (wideKeyword) isBorderRightRadius() {}

// BorderRightStyleValue is a value for the CSS border-right-style property.
type BorderRightStyleValue interface {
	cssValue
	isBorderRightStyle()
}

// BorderRightStyle sets the CSS border-right-style property. Use RawValue for complex CSS syntax.
func (e *Element) BorderRightStyle(v BorderRightStyleValue) *Element {
	return e.setValue("border-right-style", v)
}
func (expression) isBorderRightStyle()  {}
func (wideKeyword) isBorderRightStyle() {}

// BorderRightWidthValue is a value for the CSS border-right-width property.
type BorderRightWidthValue interface {
	cssValue
	isBorderRightWidth()
}

// BorderRightWidth sets the CSS border-right-width property. Use RawValue for complex CSS syntax.
func (e *Element) BorderRightWidth(v BorderRightWidthValue) *Element {
	return e.setValue("border-right-width", v)
}
func (expression) isBorderRightWidth()  {}
func (wideKeyword) isBorderRightWidth() {}
func (Length) isBorderRightWidth()      {}

// BorderSpacingValue is a value for the CSS border-spacing property.
type BorderSpacingValue interface {
	cssValue
	isBorderSpacing()
}

// BorderSpacing sets the CSS border-spacing property. Use RawValue for complex CSS syntax.
func (e *Element) BorderSpacing(v BorderSpacingValue) *Element {
	return e.setValue("border-spacing", v)
}
func (expression) isBorderSpacing()  {}
func (wideKeyword) isBorderSpacing() {}

// BorderStartEndRadiusValue is a value for the CSS border-start-end-radius property.
type BorderStartEndRadiusValue interface {
	cssValue
	isBorderStartEndRadius()
}

// BorderStartEndRadius sets the CSS border-start-end-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderStartEndRadius(v BorderStartEndRadiusValue) *Element {
	return e.setValue("border-start-end-radius", v)
}
func (expression) isBorderStartEndRadius()  {}
func (wideKeyword) isBorderStartEndRadius() {}

// BorderStartStartRadiusValue is a value for the CSS border-start-start-radius property.
type BorderStartStartRadiusValue interface {
	cssValue
	isBorderStartStartRadius()
}

// BorderStartStartRadius sets the CSS border-start-start-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderStartStartRadius(v BorderStartStartRadiusValue) *Element {
	return e.setValue("border-start-start-radius", v)
}
func (expression) isBorderStartStartRadius()  {}
func (wideKeyword) isBorderStartStartRadius() {}

// BorderStyleValue is a value for the CSS border-style property.
type BorderStyleValue interface {
	cssValue
	isBorderStyle()
}

// BorderStyle sets the CSS border-style property. Use RawValue for complex CSS syntax.
func (e *Element) BorderStyle(v BorderStyleValue) *Element { return e.setValue("border-style", v) }
func (expression) isBorderStyle()                          {}
func (wideKeyword) isBorderStyle()                         {}

// BorderTopValue is a value for the CSS border-top property.
type BorderTopValue interface {
	cssValue
	isBorderTop()
}

// BorderTop sets the CSS border-top property. Use RawValue for complex CSS syntax.
func (e *Element) BorderTop(v BorderTopValue) *Element { return e.setValue("border-top", v) }
func (expression) isBorderTop()                        {}
func (wideKeyword) isBorderTop()                       {}
func (BorderStroke) isBorderTop()                      {}

// BorderTopColorValue is a value for the CSS border-top-color property.
type BorderTopColorValue interface {
	cssValue
	isBorderTopColor()
}

// BorderTopColor sets the CSS border-top-color property. Use RawValue for complex CSS syntax.
func (e *Element) BorderTopColor(v BorderTopColorValue) *Element {
	return e.setValue("border-top-color", v)
}
func (expression) isBorderTopColor()  {}
func (wideKeyword) isBorderTopColor() {}
func (Color) isBorderTopColor()       {}

// BorderTopLeftRadiusValue is a value for the CSS border-top-left-radius property.
type BorderTopLeftRadiusValue interface {
	cssValue
	isBorderTopLeftRadius()
}

// BorderTopLeftRadius sets the CSS border-top-left-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderTopLeftRadius(v BorderTopLeftRadiusValue) *Element {
	return e.setValue("border-top-left-radius", v)
}
func (expression) isBorderTopLeftRadius()  {}
func (wideKeyword) isBorderTopLeftRadius() {}
func (Length) isBorderTopLeftRadius()      {}
func (Percentage) isBorderTopLeftRadius()  {}

// BorderTopRadiusValue is a value for the CSS border-top-radius property.
type BorderTopRadiusValue interface {
	cssValue
	isBorderTopRadius()
}

// BorderTopRadius sets the CSS border-top-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderTopRadius(v BorderTopRadiusValue) *Element {
	return e.setValue("border-top-radius", v)
}
func (expression) isBorderTopRadius()  {}
func (wideKeyword) isBorderTopRadius() {}

// BorderTopRightRadiusValue is a value for the CSS border-top-right-radius property.
type BorderTopRightRadiusValue interface {
	cssValue
	isBorderTopRightRadius()
}

// BorderTopRightRadius sets the CSS border-top-right-radius property. Use RawValue for complex CSS syntax.
func (e *Element) BorderTopRightRadius(v BorderTopRightRadiusValue) *Element {
	return e.setValue("border-top-right-radius", v)
}
func (expression) isBorderTopRightRadius()  {}
func (wideKeyword) isBorderTopRightRadius() {}
func (Length) isBorderTopRightRadius()      {}
func (Percentage) isBorderTopRightRadius()  {}

// BorderTopStyleValue is a value for the CSS border-top-style property.
type BorderTopStyleValue interface {
	cssValue
	isBorderTopStyle()
}

// BorderTopStyle sets the CSS border-top-style property. Use RawValue for complex CSS syntax.
func (e *Element) BorderTopStyle(v BorderTopStyleValue) *Element {
	return e.setValue("border-top-style", v)
}
func (expression) isBorderTopStyle()  {}
func (wideKeyword) isBorderTopStyle() {}

// BorderTopWidthValue is a value for the CSS border-top-width property.
type BorderTopWidthValue interface {
	cssValue
	isBorderTopWidth()
}

// BorderTopWidth sets the CSS border-top-width property. Use RawValue for complex CSS syntax.
func (e *Element) BorderTopWidth(v BorderTopWidthValue) *Element {
	return e.setValue("border-top-width", v)
}
func (expression) isBorderTopWidth()  {}
func (wideKeyword) isBorderTopWidth() {}
func (Length) isBorderTopWidth()      {}

// BorderWidthValue is a value for the CSS border-width property.
type BorderWidthValue interface {
	cssValue
	isBorderWidth()
}

// BorderWidth sets the CSS border-width property. Use RawValue for complex CSS syntax.
func (e *Element) BorderWidth(v BorderWidthValue) *Element { return e.setValue("border-width", v) }
func (expression) isBorderWidth()                          {}
func (wideKeyword) isBorderWidth()                         {}
func (Length) isBorderWidth()                              {}

// BottomValue is a value for the CSS bottom property.
type BottomValue interface {
	cssValue
	isBottom()
}

// Bottom sets the CSS bottom property. Use RawValue for complex CSS syntax.
func (e *Element) Bottom(v BottomValue) *Element { return e.setValue("bottom", v) }
func (expression) isBottom()                     {}
func (wideKeyword) isBottom()                    {}
func (Length) isBottom()                         {}
func (Percentage) isBottom()                     {}

// BoxShadowValue is a value for the CSS box-shadow property.
type BoxShadowValue interface {
	cssValue
	isBoxShadow()
}

// BoxShadow sets the CSS box-shadow property. Use RawValue for complex CSS syntax.
func (e *Element) BoxShadow(v BoxShadowValue) *Element { return e.setValue("box-shadow", v) }
func (expression) isBoxShadow()                        {}
func (wideKeyword) isBoxShadow()                       {}

// BoxShadowBlurValue is a value for the CSS box-shadow-blur property.
type BoxShadowBlurValue interface {
	cssValue
	isBoxShadowBlur()
}

// BoxShadowBlur sets the CSS box-shadow-blur property. Use RawValue for complex CSS syntax.
func (e *Element) BoxShadowBlur(v BoxShadowBlurValue) *Element {
	return e.setValue("box-shadow-blur", v)
}
func (expression) isBoxShadowBlur()  {}
func (wideKeyword) isBoxShadowBlur() {}

// BoxShadowColorValue is a value for the CSS box-shadow-color property.
type BoxShadowColorValue interface {
	cssValue
	isBoxShadowColor()
}

// BoxShadowColor sets the CSS box-shadow-color property. Use RawValue for complex CSS syntax.
func (e *Element) BoxShadowColor(v BoxShadowColorValue) *Element {
	return e.setValue("box-shadow-color", v)
}
func (expression) isBoxShadowColor()  {}
func (wideKeyword) isBoxShadowColor() {}

// BoxShadowInsetValue is a value for the CSS box-shadow-inset property.
type BoxShadowInsetValue interface {
	cssValue
	isBoxShadowInset()
}

// BoxShadowInset sets the CSS box-shadow-inset property. Use RawValue for complex CSS syntax.
func (e *Element) BoxShadowInset(v BoxShadowInsetValue) *Element {
	return e.setValue("box-shadow-inset", v)
}
func (expression) isBoxShadowInset()  {}
func (wideKeyword) isBoxShadowInset() {}

// BoxShadowOffsetValue is a value for the CSS box-shadow-offset property.
type BoxShadowOffsetValue interface {
	cssValue
	isBoxShadowOffset()
}

// BoxShadowOffset sets the CSS box-shadow-offset property. Use RawValue for complex CSS syntax.
func (e *Element) BoxShadowOffset(v BoxShadowOffsetValue) *Element {
	return e.setValue("box-shadow-offset", v)
}
func (expression) isBoxShadowOffset()  {}
func (wideKeyword) isBoxShadowOffset() {}

// BoxShadowPositionValue is a value for the CSS box-shadow-position property.
type BoxShadowPositionValue interface {
	cssValue
	isBoxShadowPosition()
}

// BoxShadowPosition sets the CSS box-shadow-position property. Use RawValue for complex CSS syntax.
func (e *Element) BoxShadowPosition(v BoxShadowPositionValue) *Element {
	return e.setValue("box-shadow-position", v)
}
func (expression) isBoxShadowPosition()  {}
func (wideKeyword) isBoxShadowPosition() {}

// BoxShadowSpreadValue is a value for the CSS box-shadow-spread property.
type BoxShadowSpreadValue interface {
	cssValue
	isBoxShadowSpread()
}

// BoxShadowSpread sets the CSS box-shadow-spread property. Use RawValue for complex CSS syntax.
func (e *Element) BoxShadowSpread(v BoxShadowSpreadValue) *Element {
	return e.setValue("box-shadow-spread", v)
}
func (expression) isBoxShadowSpread()  {}
func (wideKeyword) isBoxShadowSpread() {}

// BoxSizingValue is a value for the CSS box-sizing property.
type BoxSizingValue interface {
	cssValue
	isBoxSizing()
}

// BoxSizing sets the CSS box-sizing property. Use RawValue for complex CSS syntax.
func (e *Element) BoxSizing(v BoxSizingValue) *Element { return e.setValue("box-sizing", v) }
func (expression) isBoxSizing()                        {}
func (wideKeyword) isBoxSizing()                       {}

type boxSizingKeyword uint16

const (
	// BoxSizingBorderBox sets box-sizing to border-box.
	BoxSizingBorderBox boxSizingKeyword = 1
	// BoxSizingContentBox sets box-sizing to content-box.
	BoxSizingContentBox boxSizingKeyword = 2
)

func (boxSizingKeyword) isBoxSizing() {}
func (v boxSizingKeyword) cssText() (string, error) {
	switch v {
	case BoxSizingBorderBox:
		return "border-box", nil
	case BoxSizingContentBox:
		return "content-box", nil
	default:
		return "", fmt.Errorf("ui: invalid box-sizing keyword %d", v)
	}
}

// BreakAfterValue is a value for the CSS break-after property.
type BreakAfterValue interface {
	cssValue
	isBreakAfter()
}

// BreakAfter sets the CSS break-after property. Use RawValue for complex CSS syntax.
func (e *Element) BreakAfter(v BreakAfterValue) *Element { return e.setValue("break-after", v) }
func (expression) isBreakAfter()                         {}
func (wideKeyword) isBreakAfter()                        {}

// BreakBeforeValue is a value for the CSS break-before property.
type BreakBeforeValue interface {
	cssValue
	isBreakBefore()
}

// BreakBefore sets the CSS break-before property. Use RawValue for complex CSS syntax.
func (e *Element) BreakBefore(v BreakBeforeValue) *Element { return e.setValue("break-before", v) }
func (expression) isBreakBefore()                          {}
func (wideKeyword) isBreakBefore()                         {}

// CaptionSideValue is a value for the CSS caption-side property.
type CaptionSideValue interface {
	cssValue
	isCaptionSide()
}

// CaptionSide sets the CSS caption-side property. Use RawValue for complex CSS syntax.
func (e *Element) CaptionSide(v CaptionSideValue) *Element { return e.setValue("caption-side", v) }
func (expression) isCaptionSide()                          {}
func (wideKeyword) isCaptionSide()                         {}

// ClearValue is a value for the CSS clear property.
type ClearValue interface {
	cssValue
	isClear()
}

// Clear sets the CSS clear property. Use RawValue for complex CSS syntax.
func (e *Element) Clear(v ClearValue) *Element { return e.setValue("clear", v) }
func (expression) isClear()                    {}
func (wideKeyword) isClear()                   {}

// ClipValue is a value for the CSS clip property.
type ClipValue interface {
	cssValue
	isClip()
}

// Clip sets the CSS clip property. Use RawValue for complex CSS syntax.
func (e *Element) Clip(v ClipValue) *Element { return e.setValue("clip", v) }
func (expression) isClip()                   {}
func (wideKeyword) isClip()                  {}

// ClipPathValue is a value for the CSS clip-path property.
type ClipPathValue interface {
	cssValue
	isClipPath()
}

// ClipPath sets the CSS clip-path property. Use RawValue for complex CSS syntax.
func (e *Element) ClipPath(v ClipPathValue) *Element { return e.setValue("clip-path", v) }
func (expression) isClipPath()                       {}
func (wideKeyword) isClipPath()                      {}

// ClipRuleValue is a value for the CSS clip-rule property.
type ClipRuleValue interface {
	cssValue
	isClipRule()
}

// ClipRule sets the CSS clip-rule property. Use RawValue for complex CSS syntax.
func (e *Element) ClipRule(v ClipRuleValue) *Element { return e.setValue("clip-rule", v) }
func (expression) isClipRule()                       {}
func (wideKeyword) isClipRule()                      {}

// ColorValue is a value for the CSS color property.
type ColorValue interface {
	cssValue
	isColor()
}

// Color sets the CSS color property. Use RawValue for complex CSS syntax.
func (e *Element) Color(v ColorValue) *Element { return e.setValue("color", v) }
func (expression) isColor()                    {}
func (wideKeyword) isColor()                   {}
func (Color) isColor()                         {}

// ColorAdjustValue is a value for the CSS color-adjust property.
type ColorAdjustValue interface {
	cssValue
	isColorAdjust()
}

// ColorAdjust sets the CSS color-adjust property. Use RawValue for complex CSS syntax.
func (e *Element) ColorAdjust(v ColorAdjustValue) *Element { return e.setValue("color-adjust", v) }
func (expression) isColorAdjust()                          {}
func (wideKeyword) isColorAdjust()                         {}

// ColorInterpolationValue is a value for the CSS color-interpolation property.
type ColorInterpolationValue interface {
	cssValue
	isColorInterpolation()
}

// ColorInterpolation sets the CSS color-interpolation property. Use RawValue for complex CSS syntax.
func (e *Element) ColorInterpolation(v ColorInterpolationValue) *Element {
	return e.setValue("color-interpolation", v)
}
func (expression) isColorInterpolation()  {}
func (wideKeyword) isColorInterpolation() {}

// ColorInterpolationFiltersValue is a value for the CSS color-interpolation-filters property.
type ColorInterpolationFiltersValue interface {
	cssValue
	isColorInterpolationFilters()
}

// ColorInterpolationFilters sets the CSS color-interpolation-filters property. Use RawValue for complex CSS syntax.
func (e *Element) ColorInterpolationFilters(v ColorInterpolationFiltersValue) *Element {
	return e.setValue("color-interpolation-filters", v)
}
func (expression) isColorInterpolationFilters()  {}
func (wideKeyword) isColorInterpolationFilters() {}

// ColorSchemeValue is a value for the CSS color-scheme property.
type ColorSchemeValue interface {
	cssValue
	isColorScheme()
}

// ColorScheme sets the CSS color-scheme property. Use RawValue for complex CSS syntax.
func (e *Element) ColorScheme(v ColorSchemeValue) *Element { return e.setValue("color-scheme", v) }
func (expression) isColorScheme()                          {}
func (wideKeyword) isColorScheme()                         {}

// ColumnCountValue is a value for the CSS column-count property.
type ColumnCountValue interface {
	cssValue
	isColumnCount()
}

// ColumnCount sets the CSS column-count property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnCount(v ColumnCountValue) *Element { return e.setValue("column-count", v) }
func (expression) isColumnCount()                          {}
func (wideKeyword) isColumnCount()                         {}
func (Numeric) isColumnCount()                             {}

// ColumnFillValue is a value for the CSS column-fill property.
type ColumnFillValue interface {
	cssValue
	isColumnFill()
}

// ColumnFill sets the CSS column-fill property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnFill(v ColumnFillValue) *Element { return e.setValue("column-fill", v) }
func (expression) isColumnFill()                         {}
func (wideKeyword) isColumnFill()                        {}

// ColumnGapValue is a value for the CSS column-gap property.
type ColumnGapValue interface {
	cssValue
	isColumnGap()
}

// ColumnGap sets the CSS column-gap property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnGap(v ColumnGapValue) *Element { return e.setValue("column-gap", v) }
func (expression) isColumnGap()                        {}
func (wideKeyword) isColumnGap()                       {}
func (Length) isColumnGap()                            {}
func (Percentage) isColumnGap()                        {}

// ColumnHeightValue is a value for the CSS column-height property.
type ColumnHeightValue interface {
	cssValue
	isColumnHeight()
}

// ColumnHeight sets the CSS column-height property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnHeight(v ColumnHeightValue) *Element { return e.setValue("column-height", v) }
func (expression) isColumnHeight()                           {}
func (wideKeyword) isColumnHeight()                          {}

// ColumnRuleValue is a value for the CSS column-rule property.
type ColumnRuleValue interface {
	cssValue
	isColumnRule()
}

// ColumnRule sets the CSS column-rule property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRule(v ColumnRuleValue) *Element { return e.setValue("column-rule", v) }
func (expression) isColumnRule()                         {}
func (wideKeyword) isColumnRule()                        {}

// ColumnRuleBreakValue is a value for the CSS column-rule-break property.
type ColumnRuleBreakValue interface {
	cssValue
	isColumnRuleBreak()
}

// ColumnRuleBreak sets the CSS column-rule-break property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRuleBreak(v ColumnRuleBreakValue) *Element {
	return e.setValue("column-rule-break", v)
}
func (expression) isColumnRuleBreak()  {}
func (wideKeyword) isColumnRuleBreak() {}

// ColumnRuleColorValue is a value for the CSS column-rule-color property.
type ColumnRuleColorValue interface {
	cssValue
	isColumnRuleColor()
}

// ColumnRuleColor sets the CSS column-rule-color property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRuleColor(v ColumnRuleColorValue) *Element {
	return e.setValue("column-rule-color", v)
}
func (expression) isColumnRuleColor()  {}
func (wideKeyword) isColumnRuleColor() {}

// ColumnRuleInsetValue is a value for the CSS column-rule-inset property.
type ColumnRuleInsetValue interface {
	cssValue
	isColumnRuleInset()
}

// ColumnRuleInset sets the CSS column-rule-inset property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRuleInset(v ColumnRuleInsetValue) *Element {
	return e.setValue("column-rule-inset", v)
}
func (expression) isColumnRuleInset()  {}
func (wideKeyword) isColumnRuleInset() {}

// ColumnRuleInsetCapValue is a value for the CSS column-rule-inset-cap property.
type ColumnRuleInsetCapValue interface {
	cssValue
	isColumnRuleInsetCap()
}

// ColumnRuleInsetCap sets the CSS column-rule-inset-cap property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRuleInsetCap(v ColumnRuleInsetCapValue) *Element {
	return e.setValue("column-rule-inset-cap", v)
}
func (expression) isColumnRuleInsetCap()  {}
func (wideKeyword) isColumnRuleInsetCap() {}

// ColumnRuleInsetCapEndValue is a value for the CSS column-rule-inset-cap-end property.
type ColumnRuleInsetCapEndValue interface {
	cssValue
	isColumnRuleInsetCapEnd()
}

// ColumnRuleInsetCapEnd sets the CSS column-rule-inset-cap-end property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRuleInsetCapEnd(v ColumnRuleInsetCapEndValue) *Element {
	return e.setValue("column-rule-inset-cap-end", v)
}
func (expression) isColumnRuleInsetCapEnd()  {}
func (wideKeyword) isColumnRuleInsetCapEnd() {}

// ColumnRuleInsetCapStartValue is a value for the CSS column-rule-inset-cap-start property.
type ColumnRuleInsetCapStartValue interface {
	cssValue
	isColumnRuleInsetCapStart()
}

// ColumnRuleInsetCapStart sets the CSS column-rule-inset-cap-start property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRuleInsetCapStart(v ColumnRuleInsetCapStartValue) *Element {
	return e.setValue("column-rule-inset-cap-start", v)
}
func (expression) isColumnRuleInsetCapStart()  {}
func (wideKeyword) isColumnRuleInsetCapStart() {}

// ColumnRuleInsetEndValue is a value for the CSS column-rule-inset-end property.
type ColumnRuleInsetEndValue interface {
	cssValue
	isColumnRuleInsetEnd()
}

// ColumnRuleInsetEnd sets the CSS column-rule-inset-end property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRuleInsetEnd(v ColumnRuleInsetEndValue) *Element {
	return e.setValue("column-rule-inset-end", v)
}
func (expression) isColumnRuleInsetEnd()  {}
func (wideKeyword) isColumnRuleInsetEnd() {}

// ColumnRuleInsetJunctionValue is a value for the CSS column-rule-inset-junction property.
type ColumnRuleInsetJunctionValue interface {
	cssValue
	isColumnRuleInsetJunction()
}

// ColumnRuleInsetJunction sets the CSS column-rule-inset-junction property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRuleInsetJunction(v ColumnRuleInsetJunctionValue) *Element {
	return e.setValue("column-rule-inset-junction", v)
}
func (expression) isColumnRuleInsetJunction()  {}
func (wideKeyword) isColumnRuleInsetJunction() {}

// ColumnRuleInsetJunctionEndValue is a value for the CSS column-rule-inset-junction-end property.
type ColumnRuleInsetJunctionEndValue interface {
	cssValue
	isColumnRuleInsetJunctionEnd()
}

// ColumnRuleInsetJunctionEnd sets the CSS column-rule-inset-junction-end property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRuleInsetJunctionEnd(v ColumnRuleInsetJunctionEndValue) *Element {
	return e.setValue("column-rule-inset-junction-end", v)
}
func (expression) isColumnRuleInsetJunctionEnd()  {}
func (wideKeyword) isColumnRuleInsetJunctionEnd() {}

// ColumnRuleInsetJunctionStartValue is a value for the CSS column-rule-inset-junction-start property.
type ColumnRuleInsetJunctionStartValue interface {
	cssValue
	isColumnRuleInsetJunctionStart()
}

// ColumnRuleInsetJunctionStart sets the CSS column-rule-inset-junction-start property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRuleInsetJunctionStart(v ColumnRuleInsetJunctionStartValue) *Element {
	return e.setValue("column-rule-inset-junction-start", v)
}
func (expression) isColumnRuleInsetJunctionStart()  {}
func (wideKeyword) isColumnRuleInsetJunctionStart() {}

// ColumnRuleInsetStartValue is a value for the CSS column-rule-inset-start property.
type ColumnRuleInsetStartValue interface {
	cssValue
	isColumnRuleInsetStart()
}

// ColumnRuleInsetStart sets the CSS column-rule-inset-start property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRuleInsetStart(v ColumnRuleInsetStartValue) *Element {
	return e.setValue("column-rule-inset-start", v)
}
func (expression) isColumnRuleInsetStart()  {}
func (wideKeyword) isColumnRuleInsetStart() {}

// ColumnRuleStyleValue is a value for the CSS column-rule-style property.
type ColumnRuleStyleValue interface {
	cssValue
	isColumnRuleStyle()
}

// ColumnRuleStyle sets the CSS column-rule-style property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRuleStyle(v ColumnRuleStyleValue) *Element {
	return e.setValue("column-rule-style", v)
}
func (expression) isColumnRuleStyle()  {}
func (wideKeyword) isColumnRuleStyle() {}

// ColumnRuleWidthValue is a value for the CSS column-rule-width property.
type ColumnRuleWidthValue interface {
	cssValue
	isColumnRuleWidth()
}

// ColumnRuleWidth sets the CSS column-rule-width property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnRuleWidth(v ColumnRuleWidthValue) *Element {
	return e.setValue("column-rule-width", v)
}
func (expression) isColumnRuleWidth()  {}
func (wideKeyword) isColumnRuleWidth() {}

// ColumnSpanValue is a value for the CSS column-span property.
type ColumnSpanValue interface {
	cssValue
	isColumnSpan()
}

// ColumnSpan sets the CSS column-span property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnSpan(v ColumnSpanValue) *Element { return e.setValue("column-span", v) }
func (expression) isColumnSpan()                         {}
func (wideKeyword) isColumnSpan()                        {}

// ColumnWidthValue is a value for the CSS column-width property.
type ColumnWidthValue interface {
	cssValue
	isColumnWidth()
}

// ColumnWidth sets the CSS column-width property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnWidth(v ColumnWidthValue) *Element { return e.setValue("column-width", v) }
func (expression) isColumnWidth()                          {}
func (wideKeyword) isColumnWidth()                         {}
func (Length) isColumnWidth()                              {}

// ColumnWrapValue is a value for the CSS column-wrap property.
type ColumnWrapValue interface {
	cssValue
	isColumnWrap()
}

// ColumnWrap sets the CSS column-wrap property. Use RawValue for complex CSS syntax.
func (e *Element) ColumnWrap(v ColumnWrapValue) *Element { return e.setValue("column-wrap", v) }
func (expression) isColumnWrap()                         {}
func (wideKeyword) isColumnWrap()                        {}

// ColumnsValue is a value for the CSS columns property.
type ColumnsValue interface {
	cssValue
	isColumns()
}

// Columns sets the CSS columns property. Use RawValue for complex CSS syntax.
func (e *Element) Columns(v ColumnsValue) *Element { return e.setValue("columns", v) }
func (expression) isColumns()                      {}
func (wideKeyword) isColumns()                     {}

// ContainValue is a value for the CSS contain property.
type ContainValue interface {
	cssValue
	isContain()
}

// Contain sets the CSS contain property. Use RawValue for complex CSS syntax.
func (e *Element) Contain(v ContainValue) *Element { return e.setValue("contain", v) }
func (expression) isContain()                      {}
func (wideKeyword) isContain()                     {}

// ContainIntrinsicBlockSizeValue is a value for the CSS contain-intrinsic-block-size property.
type ContainIntrinsicBlockSizeValue interface {
	cssValue
	isContainIntrinsicBlockSize()
}

// ContainIntrinsicBlockSize sets the CSS contain-intrinsic-block-size property. Use RawValue for complex CSS syntax.
func (e *Element) ContainIntrinsicBlockSize(v ContainIntrinsicBlockSizeValue) *Element {
	return e.setValue("contain-intrinsic-block-size", v)
}
func (expression) isContainIntrinsicBlockSize()  {}
func (wideKeyword) isContainIntrinsicBlockSize() {}

// ContainIntrinsicHeightValue is a value for the CSS contain-intrinsic-height property.
type ContainIntrinsicHeightValue interface {
	cssValue
	isContainIntrinsicHeight()
}

// ContainIntrinsicHeight sets the CSS contain-intrinsic-height property. Use RawValue for complex CSS syntax.
func (e *Element) ContainIntrinsicHeight(v ContainIntrinsicHeightValue) *Element {
	return e.setValue("contain-intrinsic-height", v)
}
func (expression) isContainIntrinsicHeight()  {}
func (wideKeyword) isContainIntrinsicHeight() {}

// ContainIntrinsicInlineSizeValue is a value for the CSS contain-intrinsic-inline-size property.
type ContainIntrinsicInlineSizeValue interface {
	cssValue
	isContainIntrinsicInlineSize()
}

// ContainIntrinsicInlineSize sets the CSS contain-intrinsic-inline-size property. Use RawValue for complex CSS syntax.
func (e *Element) ContainIntrinsicInlineSize(v ContainIntrinsicInlineSizeValue) *Element {
	return e.setValue("contain-intrinsic-inline-size", v)
}
func (expression) isContainIntrinsicInlineSize()  {}
func (wideKeyword) isContainIntrinsicInlineSize() {}

// ContainIntrinsicSizeValue is a value for the CSS contain-intrinsic-size property.
type ContainIntrinsicSizeValue interface {
	cssValue
	isContainIntrinsicSize()
}

// ContainIntrinsicSize sets the CSS contain-intrinsic-size property. Use RawValue for complex CSS syntax.
func (e *Element) ContainIntrinsicSize(v ContainIntrinsicSizeValue) *Element {
	return e.setValue("contain-intrinsic-size", v)
}
func (expression) isContainIntrinsicSize()  {}
func (wideKeyword) isContainIntrinsicSize() {}

// ContainIntrinsicWidthValue is a value for the CSS contain-intrinsic-width property.
type ContainIntrinsicWidthValue interface {
	cssValue
	isContainIntrinsicWidth()
}

// ContainIntrinsicWidth sets the CSS contain-intrinsic-width property. Use RawValue for complex CSS syntax.
func (e *Element) ContainIntrinsicWidth(v ContainIntrinsicWidthValue) *Element {
	return e.setValue("contain-intrinsic-width", v)
}
func (expression) isContainIntrinsicWidth()  {}
func (wideKeyword) isContainIntrinsicWidth() {}

// ContainerValue is a value for the CSS container property.
type ContainerValue interface {
	cssValue
	isContainer()
}

// Container sets the CSS container property. Use RawValue for complex CSS syntax.
func (e *Element) Container(v ContainerValue) *Element { return e.setValue("container", v) }
func (expression) isContainer()                        {}
func (wideKeyword) isContainer()                       {}

// ContainerNameValue is a value for the CSS container-name property.
type ContainerNameValue interface {
	cssValue
	isContainerName()
}

// ContainerName sets the CSS container-name property. Use RawValue for complex CSS syntax.
func (e *Element) ContainerName(v ContainerNameValue) *Element {
	return e.setValue("container-name", v)
}
func (expression) isContainerName()  {}
func (wideKeyword) isContainerName() {}

// ContainerTypeValue is a value for the CSS container-type property.
type ContainerTypeValue interface {
	cssValue
	isContainerType()
}

// ContainerType sets the CSS container-type property. Use RawValue for complex CSS syntax.
func (e *Element) ContainerType(v ContainerTypeValue) *Element {
	return e.setValue("container-type", v)
}
func (expression) isContainerType()  {}
func (wideKeyword) isContainerType() {}

// ContentValue is a value for the CSS content property.
type ContentValue interface {
	cssValue
	isContent()
}

// Content sets the CSS content property. Use RawValue for complex CSS syntax.
func (e *Element) Content(v ContentValue) *Element { return e.setValue("content", v) }
func (expression) isContent()                      {}
func (wideKeyword) isContent()                     {}

// ContentVisibilityValue is a value for the CSS content-visibility property.
type ContentVisibilityValue interface {
	cssValue
	isContentVisibility()
}

// ContentVisibility sets the CSS content-visibility property. Use RawValue for complex CSS syntax.
func (e *Element) ContentVisibility(v ContentVisibilityValue) *Element {
	return e.setValue("content-visibility", v)
}
func (expression) isContentVisibility()  {}
func (wideKeyword) isContentVisibility() {}

// CornerValue is a value for the CSS corner property.
type CornerValue interface {
	cssValue
	isCorner()
}

// Corner sets the CSS corner property. Use RawValue for complex CSS syntax.
func (e *Element) Corner(v CornerValue) *Element { return e.setValue("corner", v) }
func (expression) isCorner()                     {}
func (wideKeyword) isCorner()                    {}

// CornerBlockEndValue is a value for the CSS corner-block-end property.
type CornerBlockEndValue interface {
	cssValue
	isCornerBlockEnd()
}

// CornerBlockEnd sets the CSS corner-block-end property. Use RawValue for complex CSS syntax.
func (e *Element) CornerBlockEnd(v CornerBlockEndValue) *Element {
	return e.setValue("corner-block-end", v)
}
func (expression) isCornerBlockEnd()  {}
func (wideKeyword) isCornerBlockEnd() {}

// CornerBlockEndShapeValue is a value for the CSS corner-block-end-shape property.
type CornerBlockEndShapeValue interface {
	cssValue
	isCornerBlockEndShape()
}

// CornerBlockEndShape sets the CSS corner-block-end-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerBlockEndShape(v CornerBlockEndShapeValue) *Element {
	return e.setValue("corner-block-end-shape", v)
}
func (expression) isCornerBlockEndShape()  {}
func (wideKeyword) isCornerBlockEndShape() {}

// CornerBlockStartValue is a value for the CSS corner-block-start property.
type CornerBlockStartValue interface {
	cssValue
	isCornerBlockStart()
}

// CornerBlockStart sets the CSS corner-block-start property. Use RawValue for complex CSS syntax.
func (e *Element) CornerBlockStart(v CornerBlockStartValue) *Element {
	return e.setValue("corner-block-start", v)
}
func (expression) isCornerBlockStart()  {}
func (wideKeyword) isCornerBlockStart() {}

// CornerBlockStartShapeValue is a value for the CSS corner-block-start-shape property.
type CornerBlockStartShapeValue interface {
	cssValue
	isCornerBlockStartShape()
}

// CornerBlockStartShape sets the CSS corner-block-start-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerBlockStartShape(v CornerBlockStartShapeValue) *Element {
	return e.setValue("corner-block-start-shape", v)
}
func (expression) isCornerBlockStartShape()  {}
func (wideKeyword) isCornerBlockStartShape() {}

// CornerBottomValue is a value for the CSS corner-bottom property.
type CornerBottomValue interface {
	cssValue
	isCornerBottom()
}

// CornerBottom sets the CSS corner-bottom property. Use RawValue for complex CSS syntax.
func (e *Element) CornerBottom(v CornerBottomValue) *Element { return e.setValue("corner-bottom", v) }
func (expression) isCornerBottom()                           {}
func (wideKeyword) isCornerBottom()                          {}

// CornerBottomLeftValue is a value for the CSS corner-bottom-left property.
type CornerBottomLeftValue interface {
	cssValue
	isCornerBottomLeft()
}

// CornerBottomLeft sets the CSS corner-bottom-left property. Use RawValue for complex CSS syntax.
func (e *Element) CornerBottomLeft(v CornerBottomLeftValue) *Element {
	return e.setValue("corner-bottom-left", v)
}
func (expression) isCornerBottomLeft()  {}
func (wideKeyword) isCornerBottomLeft() {}

// CornerBottomLeftShapeValue is a value for the CSS corner-bottom-left-shape property.
type CornerBottomLeftShapeValue interface {
	cssValue
	isCornerBottomLeftShape()
}

// CornerBottomLeftShape sets the CSS corner-bottom-left-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerBottomLeftShape(v CornerBottomLeftShapeValue) *Element {
	return e.setValue("corner-bottom-left-shape", v)
}
func (expression) isCornerBottomLeftShape()  {}
func (wideKeyword) isCornerBottomLeftShape() {}

// CornerBottomRightValue is a value for the CSS corner-bottom-right property.
type CornerBottomRightValue interface {
	cssValue
	isCornerBottomRight()
}

// CornerBottomRight sets the CSS corner-bottom-right property. Use RawValue for complex CSS syntax.
func (e *Element) CornerBottomRight(v CornerBottomRightValue) *Element {
	return e.setValue("corner-bottom-right", v)
}
func (expression) isCornerBottomRight()  {}
func (wideKeyword) isCornerBottomRight() {}

// CornerBottomRightShapeValue is a value for the CSS corner-bottom-right-shape property.
type CornerBottomRightShapeValue interface {
	cssValue
	isCornerBottomRightShape()
}

// CornerBottomRightShape sets the CSS corner-bottom-right-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerBottomRightShape(v CornerBottomRightShapeValue) *Element {
	return e.setValue("corner-bottom-right-shape", v)
}
func (expression) isCornerBottomRightShape()  {}
func (wideKeyword) isCornerBottomRightShape() {}

// CornerBottomShapeValue is a value for the CSS corner-bottom-shape property.
type CornerBottomShapeValue interface {
	cssValue
	isCornerBottomShape()
}

// CornerBottomShape sets the CSS corner-bottom-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerBottomShape(v CornerBottomShapeValue) *Element {
	return e.setValue("corner-bottom-shape", v)
}
func (expression) isCornerBottomShape()  {}
func (wideKeyword) isCornerBottomShape() {}

// CornerEndEndValue is a value for the CSS corner-end-end property.
type CornerEndEndValue interface {
	cssValue
	isCornerEndEnd()
}

// CornerEndEnd sets the CSS corner-end-end property. Use RawValue for complex CSS syntax.
func (e *Element) CornerEndEnd(v CornerEndEndValue) *Element { return e.setValue("corner-end-end", v) }
func (expression) isCornerEndEnd()                           {}
func (wideKeyword) isCornerEndEnd()                          {}

// CornerEndEndShapeValue is a value for the CSS corner-end-end-shape property.
type CornerEndEndShapeValue interface {
	cssValue
	isCornerEndEndShape()
}

// CornerEndEndShape sets the CSS corner-end-end-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerEndEndShape(v CornerEndEndShapeValue) *Element {
	return e.setValue("corner-end-end-shape", v)
}
func (expression) isCornerEndEndShape()  {}
func (wideKeyword) isCornerEndEndShape() {}

// CornerEndStartValue is a value for the CSS corner-end-start property.
type CornerEndStartValue interface {
	cssValue
	isCornerEndStart()
}

// CornerEndStart sets the CSS corner-end-start property. Use RawValue for complex CSS syntax.
func (e *Element) CornerEndStart(v CornerEndStartValue) *Element {
	return e.setValue("corner-end-start", v)
}
func (expression) isCornerEndStart()  {}
func (wideKeyword) isCornerEndStart() {}

// CornerEndStartShapeValue is a value for the CSS corner-end-start-shape property.
type CornerEndStartShapeValue interface {
	cssValue
	isCornerEndStartShape()
}

// CornerEndStartShape sets the CSS corner-end-start-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerEndStartShape(v CornerEndStartShapeValue) *Element {
	return e.setValue("corner-end-start-shape", v)
}
func (expression) isCornerEndStartShape()  {}
func (wideKeyword) isCornerEndStartShape() {}

// CornerInlineEndValue is a value for the CSS corner-inline-end property.
type CornerInlineEndValue interface {
	cssValue
	isCornerInlineEnd()
}

// CornerInlineEnd sets the CSS corner-inline-end property. Use RawValue for complex CSS syntax.
func (e *Element) CornerInlineEnd(v CornerInlineEndValue) *Element {
	return e.setValue("corner-inline-end", v)
}
func (expression) isCornerInlineEnd()  {}
func (wideKeyword) isCornerInlineEnd() {}

// CornerInlineEndShapeValue is a value for the CSS corner-inline-end-shape property.
type CornerInlineEndShapeValue interface {
	cssValue
	isCornerInlineEndShape()
}

// CornerInlineEndShape sets the CSS corner-inline-end-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerInlineEndShape(v CornerInlineEndShapeValue) *Element {
	return e.setValue("corner-inline-end-shape", v)
}
func (expression) isCornerInlineEndShape()  {}
func (wideKeyword) isCornerInlineEndShape() {}

// CornerInlineStartValue is a value for the CSS corner-inline-start property.
type CornerInlineStartValue interface {
	cssValue
	isCornerInlineStart()
}

// CornerInlineStart sets the CSS corner-inline-start property. Use RawValue for complex CSS syntax.
func (e *Element) CornerInlineStart(v CornerInlineStartValue) *Element {
	return e.setValue("corner-inline-start", v)
}
func (expression) isCornerInlineStart()  {}
func (wideKeyword) isCornerInlineStart() {}

// CornerInlineStartShapeValue is a value for the CSS corner-inline-start-shape property.
type CornerInlineStartShapeValue interface {
	cssValue
	isCornerInlineStartShape()
}

// CornerInlineStartShape sets the CSS corner-inline-start-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerInlineStartShape(v CornerInlineStartShapeValue) *Element {
	return e.setValue("corner-inline-start-shape", v)
}
func (expression) isCornerInlineStartShape()  {}
func (wideKeyword) isCornerInlineStartShape() {}

// CornerLeftValue is a value for the CSS corner-left property.
type CornerLeftValue interface {
	cssValue
	isCornerLeft()
}

// CornerLeft sets the CSS corner-left property. Use RawValue for complex CSS syntax.
func (e *Element) CornerLeft(v CornerLeftValue) *Element { return e.setValue("corner-left", v) }
func (expression) isCornerLeft()                         {}
func (wideKeyword) isCornerLeft()                        {}

// CornerLeftShapeValue is a value for the CSS corner-left-shape property.
type CornerLeftShapeValue interface {
	cssValue
	isCornerLeftShape()
}

// CornerLeftShape sets the CSS corner-left-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerLeftShape(v CornerLeftShapeValue) *Element {
	return e.setValue("corner-left-shape", v)
}
func (expression) isCornerLeftShape()  {}
func (wideKeyword) isCornerLeftShape() {}

// CornerRightValue is a value for the CSS corner-right property.
type CornerRightValue interface {
	cssValue
	isCornerRight()
}

// CornerRight sets the CSS corner-right property. Use RawValue for complex CSS syntax.
func (e *Element) CornerRight(v CornerRightValue) *Element { return e.setValue("corner-right", v) }
func (expression) isCornerRight()                          {}
func (wideKeyword) isCornerRight()                         {}

// CornerRightShapeValue is a value for the CSS corner-right-shape property.
type CornerRightShapeValue interface {
	cssValue
	isCornerRightShape()
}

// CornerRightShape sets the CSS corner-right-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerRightShape(v CornerRightShapeValue) *Element {
	return e.setValue("corner-right-shape", v)
}
func (expression) isCornerRightShape()  {}
func (wideKeyword) isCornerRightShape() {}

// CornerShapeValue is a value for the CSS corner-shape property.
type CornerShapeValue interface {
	cssValue
	isCornerShape()
}

// CornerShape sets the CSS corner-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerShape(v CornerShapeValue) *Element { return e.setValue("corner-shape", v) }
func (expression) isCornerShape()                          {}
func (wideKeyword) isCornerShape()                         {}

// CornerStartEndValue is a value for the CSS corner-start-end property.
type CornerStartEndValue interface {
	cssValue
	isCornerStartEnd()
}

// CornerStartEnd sets the CSS corner-start-end property. Use RawValue for complex CSS syntax.
func (e *Element) CornerStartEnd(v CornerStartEndValue) *Element {
	return e.setValue("corner-start-end", v)
}
func (expression) isCornerStartEnd()  {}
func (wideKeyword) isCornerStartEnd() {}

// CornerStartEndShapeValue is a value for the CSS corner-start-end-shape property.
type CornerStartEndShapeValue interface {
	cssValue
	isCornerStartEndShape()
}

// CornerStartEndShape sets the CSS corner-start-end-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerStartEndShape(v CornerStartEndShapeValue) *Element {
	return e.setValue("corner-start-end-shape", v)
}
func (expression) isCornerStartEndShape()  {}
func (wideKeyword) isCornerStartEndShape() {}

// CornerStartStartValue is a value for the CSS corner-start-start property.
type CornerStartStartValue interface {
	cssValue
	isCornerStartStart()
}

// CornerStartStart sets the CSS corner-start-start property. Use RawValue for complex CSS syntax.
func (e *Element) CornerStartStart(v CornerStartStartValue) *Element {
	return e.setValue("corner-start-start", v)
}
func (expression) isCornerStartStart()  {}
func (wideKeyword) isCornerStartStart() {}

// CornerStartStartShapeValue is a value for the CSS corner-start-start-shape property.
type CornerStartStartShapeValue interface {
	cssValue
	isCornerStartStartShape()
}

// CornerStartStartShape sets the CSS corner-start-start-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerStartStartShape(v CornerStartStartShapeValue) *Element {
	return e.setValue("corner-start-start-shape", v)
}
func (expression) isCornerStartStartShape()  {}
func (wideKeyword) isCornerStartStartShape() {}

// CornerTopValue is a value for the CSS corner-top property.
type CornerTopValue interface {
	cssValue
	isCornerTop()
}

// CornerTop sets the CSS corner-top property. Use RawValue for complex CSS syntax.
func (e *Element) CornerTop(v CornerTopValue) *Element { return e.setValue("corner-top", v) }
func (expression) isCornerTop()                        {}
func (wideKeyword) isCornerTop()                       {}

// CornerTopLeftValue is a value for the CSS corner-top-left property.
type CornerTopLeftValue interface {
	cssValue
	isCornerTopLeft()
}

// CornerTopLeft sets the CSS corner-top-left property. Use RawValue for complex CSS syntax.
func (e *Element) CornerTopLeft(v CornerTopLeftValue) *Element {
	return e.setValue("corner-top-left", v)
}
func (expression) isCornerTopLeft()  {}
func (wideKeyword) isCornerTopLeft() {}

// CornerTopLeftShapeValue is a value for the CSS corner-top-left-shape property.
type CornerTopLeftShapeValue interface {
	cssValue
	isCornerTopLeftShape()
}

// CornerTopLeftShape sets the CSS corner-top-left-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerTopLeftShape(v CornerTopLeftShapeValue) *Element {
	return e.setValue("corner-top-left-shape", v)
}
func (expression) isCornerTopLeftShape()  {}
func (wideKeyword) isCornerTopLeftShape() {}

// CornerTopRightValue is a value for the CSS corner-top-right property.
type CornerTopRightValue interface {
	cssValue
	isCornerTopRight()
}

// CornerTopRight sets the CSS corner-top-right property. Use RawValue for complex CSS syntax.
func (e *Element) CornerTopRight(v CornerTopRightValue) *Element {
	return e.setValue("corner-top-right", v)
}
func (expression) isCornerTopRight()  {}
func (wideKeyword) isCornerTopRight() {}

// CornerTopRightShapeValue is a value for the CSS corner-top-right-shape property.
type CornerTopRightShapeValue interface {
	cssValue
	isCornerTopRightShape()
}

// CornerTopRightShape sets the CSS corner-top-right-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerTopRightShape(v CornerTopRightShapeValue) *Element {
	return e.setValue("corner-top-right-shape", v)
}
func (expression) isCornerTopRightShape()  {}
func (wideKeyword) isCornerTopRightShape() {}

// CornerTopShapeValue is a value for the CSS corner-top-shape property.
type CornerTopShapeValue interface {
	cssValue
	isCornerTopShape()
}

// CornerTopShape sets the CSS corner-top-shape property. Use RawValue for complex CSS syntax.
func (e *Element) CornerTopShape(v CornerTopShapeValue) *Element {
	return e.setValue("corner-top-shape", v)
}
func (expression) isCornerTopShape()  {}
func (wideKeyword) isCornerTopShape() {}

// CounterIncrementValue is a value for the CSS counter-increment property.
type CounterIncrementValue interface {
	cssValue
	isCounterIncrement()
}

// CounterIncrement sets the CSS counter-increment property. Use RawValue for complex CSS syntax.
func (e *Element) CounterIncrement(v CounterIncrementValue) *Element {
	return e.setValue("counter-increment", v)
}
func (expression) isCounterIncrement()  {}
func (wideKeyword) isCounterIncrement() {}

// CounterResetValue is a value for the CSS counter-reset property.
type CounterResetValue interface {
	cssValue
	isCounterReset()
}

// CounterReset sets the CSS counter-reset property. Use RawValue for complex CSS syntax.
func (e *Element) CounterReset(v CounterResetValue) *Element { return e.setValue("counter-reset", v) }
func (expression) isCounterReset()                           {}
func (wideKeyword) isCounterReset()                          {}

// CounterSetValue is a value for the CSS counter-set property.
type CounterSetValue interface {
	cssValue
	isCounterSet()
}

// CounterSet sets the CSS counter-set property. Use RawValue for complex CSS syntax.
func (e *Element) CounterSet(v CounterSetValue) *Element { return e.setValue("counter-set", v) }
func (expression) isCounterSet()                         {}
func (wideKeyword) isCounterSet()                        {}

// DirectionValue is a value for the CSS direction property.
type DirectionValue interface {
	cssValue
	isDirection()
}

// Direction sets the CSS direction property. Use RawValue for complex CSS syntax.
func (e *Element) Direction(v DirectionValue) *Element { return e.setValue("direction", v) }
func (expression) isDirection()                        {}
func (wideKeyword) isDirection()                       {}

// DisplayValue is a value for the CSS display property.
type DisplayValue interface {
	cssValue
	isDisplay()
}

// Display sets the CSS display property. Use RawValue for complex CSS syntax.
func (e *Element) Display(v DisplayValue) *Element { return e.setValue("display", v) }
func (expression) isDisplay()                      {}
func (wideKeyword) isDisplay()                     {}

type displayKeyword uint16

const (
	// DisplayNone sets display to none.
	DisplayNone displayKeyword = 1
	// DisplayBlock sets display to block.
	DisplayBlock displayKeyword = 2
	// DisplayInline sets display to inline.
	DisplayInline displayKeyword = 3
	// DisplayInlineBlock sets display to inline-block.
	DisplayInlineBlock displayKeyword = 4
	// DisplayFlex sets display to flex.
	DisplayFlex displayKeyword = 5
	// DisplayInlineFlex sets display to inline-flex.
	DisplayInlineFlex displayKeyword = 6
	// DisplayGrid sets display to grid.
	DisplayGrid displayKeyword = 7
	// DisplayInlineGrid sets display to inline-grid.
	DisplayInlineGrid displayKeyword = 8
	// DisplayTable sets display to table.
	DisplayTable displayKeyword = 9
	// DisplayTableRow sets display to table-row.
	DisplayTableRow displayKeyword = 10
	// DisplayTableCell sets display to table-cell.
	DisplayTableCell displayKeyword = 11
	// DisplayContents sets display to contents.
	DisplayContents displayKeyword = 12
)

func (displayKeyword) isDisplay() {}
func (v displayKeyword) cssText() (string, error) {
	switch v {
	case DisplayNone:
		return "none", nil
	case DisplayBlock:
		return "block", nil
	case DisplayInline:
		return "inline", nil
	case DisplayInlineBlock:
		return "inline-block", nil
	case DisplayFlex:
		return "flex", nil
	case DisplayInlineFlex:
		return "inline-flex", nil
	case DisplayGrid:
		return "grid", nil
	case DisplayInlineGrid:
		return "inline-grid", nil
	case DisplayTable:
		return "table", nil
	case DisplayTableRow:
		return "table-row", nil
	case DisplayTableCell:
		return "table-cell", nil
	case DisplayContents:
		return "contents", nil
	default:
		return "", fmt.Errorf("ui: invalid display keyword %d", v)
	}
}

// DominantBaselineValue is a value for the CSS dominant-baseline property.
type DominantBaselineValue interface {
	cssValue
	isDominantBaseline()
}

// DominantBaseline sets the CSS dominant-baseline property. Use RawValue for complex CSS syntax.
func (e *Element) DominantBaseline(v DominantBaselineValue) *Element {
	return e.setValue("dominant-baseline", v)
}
func (expression) isDominantBaseline()  {}
func (wideKeyword) isDominantBaseline() {}

// DynamicRangeLimitValue is a value for the CSS dynamic-range-limit property.
type DynamicRangeLimitValue interface {
	cssValue
	isDynamicRangeLimit()
}

// DynamicRangeLimit sets the CSS dynamic-range-limit property. Use RawValue for complex CSS syntax.
func (e *Element) DynamicRangeLimit(v DynamicRangeLimitValue) *Element {
	return e.setValue("dynamic-range-limit", v)
}
func (expression) isDynamicRangeLimit()  {}
func (wideKeyword) isDynamicRangeLimit() {}

// EmptyCellsValue is a value for the CSS empty-cells property.
type EmptyCellsValue interface {
	cssValue
	isEmptyCells()
}

// EmptyCells sets the CSS empty-cells property. Use RawValue for complex CSS syntax.
func (e *Element) EmptyCells(v EmptyCellsValue) *Element { return e.setValue("empty-cells", v) }
func (expression) isEmptyCells()                         {}
func (wideKeyword) isEmptyCells()                        {}

// FillValue is a value for the CSS fill property.
type FillValue interface {
	cssValue
	isFill()
}

// Fill sets the CSS fill property. Use RawValue for complex CSS syntax.
func (e *Element) Fill(v FillValue) *Element { return e.setValue("fill", v) }
func (expression) isFill()                   {}
func (wideKeyword) isFill()                  {}
func (Color) isFill()                        {}

// FillOpacityValue is a value for the CSS fill-opacity property.
type FillOpacityValue interface {
	cssValue
	isFillOpacity()
}

// FillOpacity sets the CSS fill-opacity property. Use RawValue for complex CSS syntax.
func (e *Element) FillOpacity(v FillOpacityValue) *Element { return e.setValue("fill-opacity", v) }
func (expression) isFillOpacity()                          {}
func (wideKeyword) isFillOpacity()                         {}
func (Numeric) isFillOpacity()                             {}

// FillRuleValue is a value for the CSS fill-rule property.
type FillRuleValue interface {
	cssValue
	isFillRule()
}

// FillRule sets the CSS fill-rule property. Use RawValue for complex CSS syntax.
func (e *Element) FillRule(v FillRuleValue) *Element { return e.setValue("fill-rule", v) }
func (expression) isFillRule()                       {}
func (wideKeyword) isFillRule()                      {}

// FilterValue is a value for the CSS filter property.
type FilterValue interface {
	cssValue
	isFilter()
}

// Filter sets the CSS filter property. Use RawValue for complex CSS syntax.
func (e *Element) Filter(v FilterValue) *Element { return e.setValue("filter", v) }
func (expression) isFilter()                     {}
func (wideKeyword) isFilter()                    {}

// FlexValue is a value for the CSS flex property.
type FlexValue interface {
	cssValue
	isFlex()
}

// Flex sets the CSS flex property. Use RawValue for complex CSS syntax.
func (e *Element) Flex(v FlexValue) *Element { return e.setValue("flex", v) }
func (expression) isFlex()                   {}
func (wideKeyword) isFlex()                  {}
func (Numeric) isFlex()                      {}

// FlexBasisValue is a value for the CSS flex-basis property.
type FlexBasisValue interface {
	cssValue
	isFlexBasis()
}

// FlexBasis sets the CSS flex-basis property. Use RawValue for complex CSS syntax.
func (e *Element) FlexBasis(v FlexBasisValue) *Element { return e.setValue("flex-basis", v) }
func (expression) isFlexBasis()                        {}
func (wideKeyword) isFlexBasis()                       {}
func (Length) isFlexBasis()                            {}
func (Percentage) isFlexBasis()                        {}

// FlexDirectionValue is a value for the CSS flex-direction property.
type FlexDirectionValue interface {
	cssValue
	isFlexDirection()
}

// FlexDirection sets the CSS flex-direction property. Use RawValue for complex CSS syntax.
func (e *Element) FlexDirection(v FlexDirectionValue) *Element {
	return e.setValue("flex-direction", v)
}
func (expression) isFlexDirection()  {}
func (wideKeyword) isFlexDirection() {}

type flexDirectionKeyword uint16

const (
	// FlexDirectionRow sets flex-direction to row.
	FlexDirectionRow flexDirectionKeyword = 1
	// FlexDirectionRowReverse sets flex-direction to row-reverse.
	FlexDirectionRowReverse flexDirectionKeyword = 2
	// FlexDirectionColumn sets flex-direction to column.
	FlexDirectionColumn flexDirectionKeyword = 3
	// FlexDirectionColumnReverse sets flex-direction to column-reverse.
	FlexDirectionColumnReverse flexDirectionKeyword = 4
)

func (flexDirectionKeyword) isFlexDirection() {}
func (v flexDirectionKeyword) cssText() (string, error) {
	switch v {
	case FlexDirectionRow:
		return "row", nil
	case FlexDirectionRowReverse:
		return "row-reverse", nil
	case FlexDirectionColumn:
		return "column", nil
	case FlexDirectionColumnReverse:
		return "column-reverse", nil
	default:
		return "", fmt.Errorf("ui: invalid flex-direction keyword %d", v)
	}
}

// FlexFlowValue is a value for the CSS flex-flow property.
type FlexFlowValue interface {
	cssValue
	isFlexFlow()
}

// FlexFlow sets the CSS flex-flow property. Use RawValue for complex CSS syntax.
func (e *Element) FlexFlow(v FlexFlowValue) *Element { return e.setValue("flex-flow", v) }
func (expression) isFlexFlow()                       {}
func (wideKeyword) isFlexFlow()                      {}

// FlexGrowValue is a value for the CSS flex-grow property.
type FlexGrowValue interface {
	cssValue
	isFlexGrow()
}

// FlexGrow sets the CSS flex-grow property. Use RawValue for complex CSS syntax.
func (e *Element) FlexGrow(v FlexGrowValue) *Element { return e.setValue("flex-grow", v) }
func (expression) isFlexGrow()                       {}
func (wideKeyword) isFlexGrow()                      {}
func (Numeric) isFlexGrow()                          {}

// FlexShrinkValue is a value for the CSS flex-shrink property.
type FlexShrinkValue interface {
	cssValue
	isFlexShrink()
}

// FlexShrink sets the CSS flex-shrink property. Use RawValue for complex CSS syntax.
func (e *Element) FlexShrink(v FlexShrinkValue) *Element { return e.setValue("flex-shrink", v) }
func (expression) isFlexShrink()                         {}
func (wideKeyword) isFlexShrink()                        {}
func (Numeric) isFlexShrink()                            {}

// FlexWrapValue is a value for the CSS flex-wrap property.
type FlexWrapValue interface {
	cssValue
	isFlexWrap()
}

// FlexWrap sets the CSS flex-wrap property. Use RawValue for complex CSS syntax.
func (e *Element) FlexWrap(v FlexWrapValue) *Element { return e.setValue("flex-wrap", v) }
func (expression) isFlexWrap()                       {}
func (wideKeyword) isFlexWrap()                      {}

type flexWrapKeyword uint16

const (
	// FlexWrapNowrap sets flex-wrap to nowrap.
	FlexWrapNowrap flexWrapKeyword = 1
	// FlexWrapWrap sets flex-wrap to wrap.
	FlexWrapWrap flexWrapKeyword = 2
	// FlexWrapWrapReverse sets flex-wrap to wrap-reverse.
	FlexWrapWrapReverse flexWrapKeyword = 3
)

func (flexWrapKeyword) isFlexWrap() {}
func (v flexWrapKeyword) cssText() (string, error) {
	switch v {
	case FlexWrapNowrap:
		return "nowrap", nil
	case FlexWrapWrap:
		return "wrap", nil
	case FlexWrapWrapReverse:
		return "wrap-reverse", nil
	default:
		return "", fmt.Errorf("ui: invalid flex-wrap keyword %d", v)
	}
}

// FloatValue is a value for the CSS float property.
type FloatValue interface {
	cssValue
	isFloat()
}

// Float sets the CSS float property. Use RawValue for complex CSS syntax.
func (e *Element) Float(v FloatValue) *Element { return e.setValue("float", v) }
func (expression) isFloat()                    {}
func (wideKeyword) isFloat()                   {}

// FloatOffsetValue is a value for the CSS float-offset property.
type FloatOffsetValue interface {
	cssValue
	isFloatOffset()
}

// FloatOffset sets the CSS float-offset property. Use RawValue for complex CSS syntax.
func (e *Element) FloatOffset(v FloatOffsetValue) *Element { return e.setValue("float-offset", v) }
func (expression) isFloatOffset()                          {}
func (wideKeyword) isFloatOffset()                         {}

// FloatReferenceValue is a value for the CSS float-reference property.
type FloatReferenceValue interface {
	cssValue
	isFloatReference()
}

// FloatReference sets the CSS float-reference property. Use RawValue for complex CSS syntax.
func (e *Element) FloatReference(v FloatReferenceValue) *Element {
	return e.setValue("float-reference", v)
}
func (expression) isFloatReference()  {}
func (wideKeyword) isFloatReference() {}

// FlowFromValue is a value for the CSS flow-from property.
type FlowFromValue interface {
	cssValue
	isFlowFrom()
}

// FlowFrom sets the CSS flow-from property. Use RawValue for complex CSS syntax.
func (e *Element) FlowFrom(v FlowFromValue) *Element { return e.setValue("flow-from", v) }
func (expression) isFlowFrom()                       {}
func (wideKeyword) isFlowFrom()                      {}

// FlowIntoValue is a value for the CSS flow-into property.
type FlowIntoValue interface {
	cssValue
	isFlowInto()
}

// FlowInto sets the CSS flow-into property. Use RawValue for complex CSS syntax.
func (e *Element) FlowInto(v FlowIntoValue) *Element { return e.setValue("flow-into", v) }
func (expression) isFlowInto()                       {}
func (wideKeyword) isFlowInto()                      {}

// FlowToleranceValue is a value for the CSS flow-tolerance property.
type FlowToleranceValue interface {
	cssValue
	isFlowTolerance()
}

// FlowTolerance sets the CSS flow-tolerance property. Use RawValue for complex CSS syntax.
func (e *Element) FlowTolerance(v FlowToleranceValue) *Element {
	return e.setValue("flow-tolerance", v)
}
func (expression) isFlowTolerance()  {}
func (wideKeyword) isFlowTolerance() {}

// FontValue is a value for the CSS font property.
type FontValue interface {
	cssValue
	isFont()
}

// Font sets the CSS font property. Use RawValue for complex CSS syntax.
func (e *Element) Font(v FontValue) *Element { return e.setValue("font", v) }
func (expression) isFont()                   {}
func (wideKeyword) isFont()                  {}

// FontFamilyValue is a value for the CSS font-family property.
type FontFamilyValue interface {
	cssValue
	isFontFamily()
}

// FontFamily sets the CSS font-family property. Use RawValue for complex CSS syntax.
func (e *Element) FontFamily(v FontFamilyValue) *Element { return e.setValue("font-family", v) }
func (expression) isFontFamily()                         {}
func (wideKeyword) isFontFamily()                        {}

type fontFamilyKeyword uint16

const (
	// FontFamilySerif sets font-family to serif.
	FontFamilySerif fontFamilyKeyword = 1
	// FontFamilySansSerif sets font-family to sans-serif.
	FontFamilySansSerif fontFamilyKeyword = 2
	// FontFamilyMonospace sets font-family to monospace.
	FontFamilyMonospace fontFamilyKeyword = 3
)

func (fontFamilyKeyword) isFontFamily() {}
func (v fontFamilyKeyword) cssText() (string, error) {
	switch v {
	case FontFamilySerif:
		return "serif", nil
	case FontFamilySansSerif:
		return "sans-serif", nil
	case FontFamilyMonospace:
		return "monospace", nil
	default:
		return "", fmt.Errorf("ui: invalid font-family keyword %d", v)
	}
}

// FontFeatureSettingsValue is a value for the CSS font-feature-settings property.
type FontFeatureSettingsValue interface {
	cssValue
	isFontFeatureSettings()
}

// FontFeatureSettings sets the CSS font-feature-settings property. Use RawValue for complex CSS syntax.
func (e *Element) FontFeatureSettings(v FontFeatureSettingsValue) *Element {
	return e.setValue("font-feature-settings", v)
}
func (expression) isFontFeatureSettings()  {}
func (wideKeyword) isFontFeatureSettings() {}

// FontKerningValue is a value for the CSS font-kerning property.
type FontKerningValue interface {
	cssValue
	isFontKerning()
}

// FontKerning sets the CSS font-kerning property. Use RawValue for complex CSS syntax.
func (e *Element) FontKerning(v FontKerningValue) *Element { return e.setValue("font-kerning", v) }
func (expression) isFontKerning()                          {}
func (wideKeyword) isFontKerning()                         {}

// FontLanguageOverrideValue is a value for the CSS font-language-override property.
type FontLanguageOverrideValue interface {
	cssValue
	isFontLanguageOverride()
}

// FontLanguageOverride sets the CSS font-language-override property. Use RawValue for complex CSS syntax.
func (e *Element) FontLanguageOverride(v FontLanguageOverrideValue) *Element {
	return e.setValue("font-language-override", v)
}
func (expression) isFontLanguageOverride()  {}
func (wideKeyword) isFontLanguageOverride() {}

// FontOpticalSizingValue is a value for the CSS font-optical-sizing property.
type FontOpticalSizingValue interface {
	cssValue
	isFontOpticalSizing()
}

// FontOpticalSizing sets the CSS font-optical-sizing property. Use RawValue for complex CSS syntax.
func (e *Element) FontOpticalSizing(v FontOpticalSizingValue) *Element {
	return e.setValue("font-optical-sizing", v)
}
func (expression) isFontOpticalSizing()  {}
func (wideKeyword) isFontOpticalSizing() {}

// FontPaletteValue is a value for the CSS font-palette property.
type FontPaletteValue interface {
	cssValue
	isFontPalette()
}

// FontPalette sets the CSS font-palette property. Use RawValue for complex CSS syntax.
func (e *Element) FontPalette(v FontPaletteValue) *Element { return e.setValue("font-palette", v) }
func (expression) isFontPalette()                          {}
func (wideKeyword) isFontPalette()                         {}

// FontSizeValue is a value for the CSS font-size property.
type FontSizeValue interface {
	cssValue
	isFontSize()
}

// FontSize sets the CSS font-size property. Use RawValue for complex CSS syntax.
func (e *Element) FontSize(v FontSizeValue) *Element { return e.setValue("font-size", v) }
func (expression) isFontSize()                       {}
func (wideKeyword) isFontSize()                      {}
func (Length) isFontSize()                           {}
func (Percentage) isFontSize()                       {}

// FontSizeAdjustValue is a value for the CSS font-size-adjust property.
type FontSizeAdjustValue interface {
	cssValue
	isFontSizeAdjust()
}

// FontSizeAdjust sets the CSS font-size-adjust property. Use RawValue for complex CSS syntax.
func (e *Element) FontSizeAdjust(v FontSizeAdjustValue) *Element {
	return e.setValue("font-size-adjust", v)
}
func (expression) isFontSizeAdjust()  {}
func (wideKeyword) isFontSizeAdjust() {}

// FontStyleValue is a value for the CSS font-style property.
type FontStyleValue interface {
	cssValue
	isFontStyle()
}

// FontStyle sets the CSS font-style property. Use RawValue for complex CSS syntax.
func (e *Element) FontStyle(v FontStyleValue) *Element { return e.setValue("font-style", v) }
func (expression) isFontStyle()                        {}
func (wideKeyword) isFontStyle()                       {}

type fontStyleKeyword uint16

const (
	// FontStyleNormal sets font-style to normal.
	FontStyleNormal fontStyleKeyword = 1
	// FontStyleItalic sets font-style to italic.
	FontStyleItalic fontStyleKeyword = 2
	// FontStyleOblique sets font-style to oblique.
	FontStyleOblique fontStyleKeyword = 3
)

func (fontStyleKeyword) isFontStyle() {}
func (v fontStyleKeyword) cssText() (string, error) {
	switch v {
	case FontStyleNormal:
		return "normal", nil
	case FontStyleItalic:
		return "italic", nil
	case FontStyleOblique:
		return "oblique", nil
	default:
		return "", fmt.Errorf("ui: invalid font-style keyword %d", v)
	}
}

// FontSynthesisValue is a value for the CSS font-synthesis property.
type FontSynthesisValue interface {
	cssValue
	isFontSynthesis()
}

// FontSynthesis sets the CSS font-synthesis property. Use RawValue for complex CSS syntax.
func (e *Element) FontSynthesis(v FontSynthesisValue) *Element {
	return e.setValue("font-synthesis", v)
}
func (expression) isFontSynthesis()  {}
func (wideKeyword) isFontSynthesis() {}

// FontSynthesisPositionValue is a value for the CSS font-synthesis-position property.
type FontSynthesisPositionValue interface {
	cssValue
	isFontSynthesisPosition()
}

// FontSynthesisPosition sets the CSS font-synthesis-position property. Use RawValue for complex CSS syntax.
func (e *Element) FontSynthesisPosition(v FontSynthesisPositionValue) *Element {
	return e.setValue("font-synthesis-position", v)
}
func (expression) isFontSynthesisPosition()  {}
func (wideKeyword) isFontSynthesisPosition() {}

// FontSynthesisSmallCapsValue is a value for the CSS font-synthesis-small-caps property.
type FontSynthesisSmallCapsValue interface {
	cssValue
	isFontSynthesisSmallCaps()
}

// FontSynthesisSmallCaps sets the CSS font-synthesis-small-caps property. Use RawValue for complex CSS syntax.
func (e *Element) FontSynthesisSmallCaps(v FontSynthesisSmallCapsValue) *Element {
	return e.setValue("font-synthesis-small-caps", v)
}
func (expression) isFontSynthesisSmallCaps()  {}
func (wideKeyword) isFontSynthesisSmallCaps() {}

// FontSynthesisStyleValue is a value for the CSS font-synthesis-style property.
type FontSynthesisStyleValue interface {
	cssValue
	isFontSynthesisStyle()
}

// FontSynthesisStyle sets the CSS font-synthesis-style property. Use RawValue for complex CSS syntax.
func (e *Element) FontSynthesisStyle(v FontSynthesisStyleValue) *Element {
	return e.setValue("font-synthesis-style", v)
}
func (expression) isFontSynthesisStyle()  {}
func (wideKeyword) isFontSynthesisStyle() {}

// FontSynthesisWeightValue is a value for the CSS font-synthesis-weight property.
type FontSynthesisWeightValue interface {
	cssValue
	isFontSynthesisWeight()
}

// FontSynthesisWeight sets the CSS font-synthesis-weight property. Use RawValue for complex CSS syntax.
func (e *Element) FontSynthesisWeight(v FontSynthesisWeightValue) *Element {
	return e.setValue("font-synthesis-weight", v)
}
func (expression) isFontSynthesisWeight()  {}
func (wideKeyword) isFontSynthesisWeight() {}

// FontVariantValue is a value for the CSS font-variant property.
type FontVariantValue interface {
	cssValue
	isFontVariant()
}

// FontVariant sets the CSS font-variant property. Use RawValue for complex CSS syntax.
func (e *Element) FontVariant(v FontVariantValue) *Element { return e.setValue("font-variant", v) }
func (expression) isFontVariant()                          {}
func (wideKeyword) isFontVariant()                         {}

// FontVariantAlternatesValue is a value for the CSS font-variant-alternates property.
type FontVariantAlternatesValue interface {
	cssValue
	isFontVariantAlternates()
}

// FontVariantAlternates sets the CSS font-variant-alternates property. Use RawValue for complex CSS syntax.
func (e *Element) FontVariantAlternates(v FontVariantAlternatesValue) *Element {
	return e.setValue("font-variant-alternates", v)
}
func (expression) isFontVariantAlternates()  {}
func (wideKeyword) isFontVariantAlternates() {}

// FontVariantCapsValue is a value for the CSS font-variant-caps property.
type FontVariantCapsValue interface {
	cssValue
	isFontVariantCaps()
}

// FontVariantCaps sets the CSS font-variant-caps property. Use RawValue for complex CSS syntax.
func (e *Element) FontVariantCaps(v FontVariantCapsValue) *Element {
	return e.setValue("font-variant-caps", v)
}
func (expression) isFontVariantCaps()  {}
func (wideKeyword) isFontVariantCaps() {}

// FontVariantEastAsianValue is a value for the CSS font-variant-east-asian property.
type FontVariantEastAsianValue interface {
	cssValue
	isFontVariantEastAsian()
}

// FontVariantEastAsian sets the CSS font-variant-east-asian property. Use RawValue for complex CSS syntax.
func (e *Element) FontVariantEastAsian(v FontVariantEastAsianValue) *Element {
	return e.setValue("font-variant-east-asian", v)
}
func (expression) isFontVariantEastAsian()  {}
func (wideKeyword) isFontVariantEastAsian() {}

// FontVariantEmojiValue is a value for the CSS font-variant-emoji property.
type FontVariantEmojiValue interface {
	cssValue
	isFontVariantEmoji()
}

// FontVariantEmoji sets the CSS font-variant-emoji property. Use RawValue for complex CSS syntax.
func (e *Element) FontVariantEmoji(v FontVariantEmojiValue) *Element {
	return e.setValue("font-variant-emoji", v)
}
func (expression) isFontVariantEmoji()  {}
func (wideKeyword) isFontVariantEmoji() {}

// FontVariantLigaturesValue is a value for the CSS font-variant-ligatures property.
type FontVariantLigaturesValue interface {
	cssValue
	isFontVariantLigatures()
}

// FontVariantLigatures sets the CSS font-variant-ligatures property. Use RawValue for complex CSS syntax.
func (e *Element) FontVariantLigatures(v FontVariantLigaturesValue) *Element {
	return e.setValue("font-variant-ligatures", v)
}
func (expression) isFontVariantLigatures()  {}
func (wideKeyword) isFontVariantLigatures() {}

// FontVariantNumericValue is a value for the CSS font-variant-numeric property.
type FontVariantNumericValue interface {
	cssValue
	isFontVariantNumeric()
}

// FontVariantNumeric sets the CSS font-variant-numeric property. Use RawValue for complex CSS syntax.
func (e *Element) FontVariantNumeric(v FontVariantNumericValue) *Element {
	return e.setValue("font-variant-numeric", v)
}
func (expression) isFontVariantNumeric()  {}
func (wideKeyword) isFontVariantNumeric() {}

// FontVariantPositionValue is a value for the CSS font-variant-position property.
type FontVariantPositionValue interface {
	cssValue
	isFontVariantPosition()
}

// FontVariantPosition sets the CSS font-variant-position property. Use RawValue for complex CSS syntax.
func (e *Element) FontVariantPosition(v FontVariantPositionValue) *Element {
	return e.setValue("font-variant-position", v)
}
func (expression) isFontVariantPosition()  {}
func (wideKeyword) isFontVariantPosition() {}

// FontVariationSettingsValue is a value for the CSS font-variation-settings property.
type FontVariationSettingsValue interface {
	cssValue
	isFontVariationSettings()
}

// FontVariationSettings sets the CSS font-variation-settings property. Use RawValue for complex CSS syntax.
func (e *Element) FontVariationSettings(v FontVariationSettingsValue) *Element {
	return e.setValue("font-variation-settings", v)
}
func (expression) isFontVariationSettings()  {}
func (wideKeyword) isFontVariationSettings() {}

// FontWeightValue is a value for the CSS font-weight property.
type FontWeightValue interface {
	cssValue
	isFontWeight()
}

// FontWeight sets the CSS font-weight property. Use RawValue for complex CSS syntax.
func (e *Element) FontWeight(v FontWeightValue) *Element { return e.setValue("font-weight", v) }
func (expression) isFontWeight()                         {}
func (wideKeyword) isFontWeight()                        {}
func (Numeric) isFontWeight()                            {}

type fontWeightKeyword uint16

const (
	// FontWeightNormal sets font-weight to normal.
	FontWeightNormal fontWeightKeyword = 1
	// FontWeightBold sets font-weight to bold.
	FontWeightBold fontWeightKeyword = 2
	// FontWeightBolder sets font-weight to bolder.
	FontWeightBolder fontWeightKeyword = 3
	// FontWeightLighter sets font-weight to lighter.
	FontWeightLighter fontWeightKeyword = 4
)

func (fontWeightKeyword) isFontWeight() {}
func (v fontWeightKeyword) cssText() (string, error) {
	switch v {
	case FontWeightNormal:
		return "normal", nil
	case FontWeightBold:
		return "bold", nil
	case FontWeightBolder:
		return "bolder", nil
	case FontWeightLighter:
		return "lighter", nil
	default:
		return "", fmt.Errorf("ui: invalid font-weight keyword %d", v)
	}
}

// FontWidthValue is a value for the CSS font-width property.
type FontWidthValue interface {
	cssValue
	isFontWidth()
}

// FontWidth sets the CSS font-width property. Use RawValue for complex CSS syntax.
func (e *Element) FontWidth(v FontWidthValue) *Element { return e.setValue("font-width", v) }
func (expression) isFontWidth()                        {}
func (wideKeyword) isFontWidth()                       {}

// ForcedColorAdjustValue is a value for the CSS forced-color-adjust property.
type ForcedColorAdjustValue interface {
	cssValue
	isForcedColorAdjust()
}

// ForcedColorAdjust sets the CSS forced-color-adjust property. Use RawValue for complex CSS syntax.
func (e *Element) ForcedColorAdjust(v ForcedColorAdjustValue) *Element {
	return e.setValue("forced-color-adjust", v)
}
func (expression) isForcedColorAdjust()  {}
func (wideKeyword) isForcedColorAdjust() {}

// GapValue is a value for the CSS gap property.
type GapValue interface {
	cssValue
	isGap()
}

// Gap sets the CSS gap property. Use RawValue for complex CSS syntax.
func (e *Element) Gap(v GapValue) *Element { return e.setValue("gap", v) }
func (expression) isGap()                  {}
func (wideKeyword) isGap()                 {}
func (Length) isGap()                      {}
func (Percentage) isGap()                  {}
func (Spacing) isGap()                     {}

// GridValue is a value for the CSS grid property.
type GridValue interface {
	cssValue
	isGrid()
}

// Grid sets the CSS grid property. Use RawValue for complex CSS syntax.
func (e *Element) Grid(v GridValue) *Element { return e.setValue("grid", v) }
func (expression) isGrid()                   {}
func (wideKeyword) isGrid()                  {}

// GridAreaValue is a value for the CSS grid-area property.
type GridAreaValue interface {
	cssValue
	isGridArea()
}

// GridArea sets the CSS grid-area property. Use RawValue for complex CSS syntax.
func (e *Element) GridArea(v GridAreaValue) *Element { return e.setValue("grid-area", v) }
func (expression) isGridArea()                       {}
func (wideKeyword) isGridArea()                      {}

// GridAutoColumnsValue is a value for the CSS grid-auto-columns property.
type GridAutoColumnsValue interface {
	cssValue
	isGridAutoColumns()
}

// GridAutoColumns sets the CSS grid-auto-columns property. Use RawValue for complex CSS syntax.
func (e *Element) GridAutoColumns(v GridAutoColumnsValue) *Element {
	return e.setValue("grid-auto-columns", v)
}
func (expression) isGridAutoColumns()  {}
func (wideKeyword) isGridAutoColumns() {}

// GridAutoFlowValue is a value for the CSS grid-auto-flow property.
type GridAutoFlowValue interface {
	cssValue
	isGridAutoFlow()
}

// GridAutoFlow sets the CSS grid-auto-flow property. Use RawValue for complex CSS syntax.
func (e *Element) GridAutoFlow(v GridAutoFlowValue) *Element { return e.setValue("grid-auto-flow", v) }
func (expression) isGridAutoFlow()                           {}
func (wideKeyword) isGridAutoFlow()                          {}

// GridAutoRowsValue is a value for the CSS grid-auto-rows property.
type GridAutoRowsValue interface {
	cssValue
	isGridAutoRows()
}

// GridAutoRows sets the CSS grid-auto-rows property. Use RawValue for complex CSS syntax.
func (e *Element) GridAutoRows(v GridAutoRowsValue) *Element { return e.setValue("grid-auto-rows", v) }
func (expression) isGridAutoRows()                           {}
func (wideKeyword) isGridAutoRows()                          {}

// GridColumnValue is a value for the CSS grid-column property.
type GridColumnValue interface {
	cssValue
	isGridColumn()
}

// GridColumn sets the CSS grid-column property. Use RawValue for complex CSS syntax.
func (e *Element) GridColumn(v GridColumnValue) *Element { return e.setValue("grid-column", v) }
func (expression) isGridColumn()                         {}
func (wideKeyword) isGridColumn()                        {}

// GridColumnEndValue is a value for the CSS grid-column-end property.
type GridColumnEndValue interface {
	cssValue
	isGridColumnEnd()
}

// GridColumnEnd sets the CSS grid-column-end property. Use RawValue for complex CSS syntax.
func (e *Element) GridColumnEnd(v GridColumnEndValue) *Element {
	return e.setValue("grid-column-end", v)
}
func (expression) isGridColumnEnd()  {}
func (wideKeyword) isGridColumnEnd() {}

// GridColumnStartValue is a value for the CSS grid-column-start property.
type GridColumnStartValue interface {
	cssValue
	isGridColumnStart()
}

// GridColumnStart sets the CSS grid-column-start property. Use RawValue for complex CSS syntax.
func (e *Element) GridColumnStart(v GridColumnStartValue) *Element {
	return e.setValue("grid-column-start", v)
}
func (expression) isGridColumnStart()  {}
func (wideKeyword) isGridColumnStart() {}

// GridRowValue is a value for the CSS grid-row property.
type GridRowValue interface {
	cssValue
	isGridRow()
}

// GridRow sets the CSS grid-row property. Use RawValue for complex CSS syntax.
func (e *Element) GridRow(v GridRowValue) *Element { return e.setValue("grid-row", v) }
func (expression) isGridRow()                      {}
func (wideKeyword) isGridRow()                     {}

// GridRowEndValue is a value for the CSS grid-row-end property.
type GridRowEndValue interface {
	cssValue
	isGridRowEnd()
}

// GridRowEnd sets the CSS grid-row-end property. Use RawValue for complex CSS syntax.
func (e *Element) GridRowEnd(v GridRowEndValue) *Element { return e.setValue("grid-row-end", v) }
func (expression) isGridRowEnd()                         {}
func (wideKeyword) isGridRowEnd()                        {}

// GridRowStartValue is a value for the CSS grid-row-start property.
type GridRowStartValue interface {
	cssValue
	isGridRowStart()
}

// GridRowStart sets the CSS grid-row-start property. Use RawValue for complex CSS syntax.
func (e *Element) GridRowStart(v GridRowStartValue) *Element { return e.setValue("grid-row-start", v) }
func (expression) isGridRowStart()                           {}
func (wideKeyword) isGridRowStart()                          {}

// GridTemplateValue is a value for the CSS grid-template property.
type GridTemplateValue interface {
	cssValue
	isGridTemplate()
}

// GridTemplate sets the CSS grid-template property. Use RawValue for complex CSS syntax.
func (e *Element) GridTemplate(v GridTemplateValue) *Element { return e.setValue("grid-template", v) }
func (expression) isGridTemplate()                           {}
func (wideKeyword) isGridTemplate()                          {}

// GridTemplateAreasValue is a value for the CSS grid-template-areas property.
type GridTemplateAreasValue interface {
	cssValue
	isGridTemplateAreas()
}

// GridTemplateAreas sets the CSS grid-template-areas property. Use RawValue for complex CSS syntax.
func (e *Element) GridTemplateAreas(v GridTemplateAreasValue) *Element {
	return e.setValue("grid-template-areas", v)
}
func (expression) isGridTemplateAreas()  {}
func (wideKeyword) isGridTemplateAreas() {}

// GridTemplateColumnsValue is a value for the CSS grid-template-columns property.
type GridTemplateColumnsValue interface {
	cssValue
	isGridTemplateColumns()
}

// GridTemplateColumns sets the CSS grid-template-columns property. Use RawValue for complex CSS syntax.
func (e *Element) GridTemplateColumns(v GridTemplateColumnsValue) *Element {
	return e.setValue("grid-template-columns", v)
}
func (expression) isGridTemplateColumns()  {}
func (wideKeyword) isGridTemplateColumns() {}

// GridTemplateRowsValue is a value for the CSS grid-template-rows property.
type GridTemplateRowsValue interface {
	cssValue
	isGridTemplateRows()
}

// GridTemplateRows sets the CSS grid-template-rows property. Use RawValue for complex CSS syntax.
func (e *Element) GridTemplateRows(v GridTemplateRowsValue) *Element {
	return e.setValue("grid-template-rows", v)
}
func (expression) isGridTemplateRows()  {}
func (wideKeyword) isGridTemplateRows() {}

// HangingPunctuationValue is a value for the CSS hanging-punctuation property.
type HangingPunctuationValue interface {
	cssValue
	isHangingPunctuation()
}

// HangingPunctuation sets the CSS hanging-punctuation property. Use RawValue for complex CSS syntax.
func (e *Element) HangingPunctuation(v HangingPunctuationValue) *Element {
	return e.setValue("hanging-punctuation", v)
}
func (expression) isHangingPunctuation()  {}
func (wideKeyword) isHangingPunctuation() {}

// HeightValue is a value for the CSS height property.
type HeightValue interface {
	cssValue
	isHeight()
}

// Height sets the CSS height property. Use RawValue for complex CSS syntax.
func (e *Element) Height(v HeightValue) *Element { return e.setValue("height", v) }
func (expression) isHeight()                     {}
func (wideKeyword) isHeight()                    {}
func (Length) isHeight()                         {}
func (Percentage) isHeight()                     {}

type heightKeyword uint16

const (
	// HeightAuto sets height to auto.
	HeightAuto heightKeyword = 1
	// HeightMinContent sets height to min-content.
	HeightMinContent heightKeyword = 2
	// HeightMaxContent sets height to max-content.
	HeightMaxContent heightKeyword = 3
	// HeightFitContent sets height to fit-content.
	HeightFitContent heightKeyword = 4
)

func (heightKeyword) isHeight() {}
func (v heightKeyword) cssText() (string, error) {
	switch v {
	case HeightAuto:
		return "auto", nil
	case HeightMinContent:
		return "min-content", nil
	case HeightMaxContent:
		return "max-content", nil
	case HeightFitContent:
		return "fit-content", nil
	default:
		return "", fmt.Errorf("ui: invalid height keyword %d", v)
	}
}

// HyphenateCharacterValue is a value for the CSS hyphenate-character property.
type HyphenateCharacterValue interface {
	cssValue
	isHyphenateCharacter()
}

// HyphenateCharacter sets the CSS hyphenate-character property. Use RawValue for complex CSS syntax.
func (e *Element) HyphenateCharacter(v HyphenateCharacterValue) *Element {
	return e.setValue("hyphenate-character", v)
}
func (expression) isHyphenateCharacter()  {}
func (wideKeyword) isHyphenateCharacter() {}

// HyphenateLimitCharsValue is a value for the CSS hyphenate-limit-chars property.
type HyphenateLimitCharsValue interface {
	cssValue
	isHyphenateLimitChars()
}

// HyphenateLimitChars sets the CSS hyphenate-limit-chars property. Use RawValue for complex CSS syntax.
func (e *Element) HyphenateLimitChars(v HyphenateLimitCharsValue) *Element {
	return e.setValue("hyphenate-limit-chars", v)
}
func (expression) isHyphenateLimitChars()  {}
func (wideKeyword) isHyphenateLimitChars() {}

// HyphenateLimitLastValue is a value for the CSS hyphenate-limit-last property.
type HyphenateLimitLastValue interface {
	cssValue
	isHyphenateLimitLast()
}

// HyphenateLimitLast sets the CSS hyphenate-limit-last property. Use RawValue for complex CSS syntax.
func (e *Element) HyphenateLimitLast(v HyphenateLimitLastValue) *Element {
	return e.setValue("hyphenate-limit-last", v)
}
func (expression) isHyphenateLimitLast()  {}
func (wideKeyword) isHyphenateLimitLast() {}

// HyphenateLimitLinesValue is a value for the CSS hyphenate-limit-lines property.
type HyphenateLimitLinesValue interface {
	cssValue
	isHyphenateLimitLines()
}

// HyphenateLimitLines sets the CSS hyphenate-limit-lines property. Use RawValue for complex CSS syntax.
func (e *Element) HyphenateLimitLines(v HyphenateLimitLinesValue) *Element {
	return e.setValue("hyphenate-limit-lines", v)
}
func (expression) isHyphenateLimitLines()  {}
func (wideKeyword) isHyphenateLimitLines() {}

// HyphenateLimitZoneValue is a value for the CSS hyphenate-limit-zone property.
type HyphenateLimitZoneValue interface {
	cssValue
	isHyphenateLimitZone()
}

// HyphenateLimitZone sets the CSS hyphenate-limit-zone property. Use RawValue for complex CSS syntax.
func (e *Element) HyphenateLimitZone(v HyphenateLimitZoneValue) *Element {
	return e.setValue("hyphenate-limit-zone", v)
}
func (expression) isHyphenateLimitZone()  {}
func (wideKeyword) isHyphenateLimitZone() {}

// HyphensValue is a value for the CSS hyphens property.
type HyphensValue interface {
	cssValue
	isHyphens()
}

// Hyphens sets the CSS hyphens property. Use RawValue for complex CSS syntax.
func (e *Element) Hyphens(v HyphensValue) *Element { return e.setValue("hyphens", v) }
func (expression) isHyphens()                      {}
func (wideKeyword) isHyphens()                     {}

// ImageOrientationValue is a value for the CSS image-orientation property.
type ImageOrientationValue interface {
	cssValue
	isImageOrientation()
}

// ImageOrientation sets the CSS image-orientation property. Use RawValue for complex CSS syntax.
func (e *Element) ImageOrientation(v ImageOrientationValue) *Element {
	return e.setValue("image-orientation", v)
}
func (expression) isImageOrientation()  {}
func (wideKeyword) isImageOrientation() {}

// ImageResolutionValue is a value for the CSS image-resolution property.
type ImageResolutionValue interface {
	cssValue
	isImageResolution()
}

// ImageResolution sets the CSS image-resolution property. Use RawValue for complex CSS syntax.
func (e *Element) ImageResolution(v ImageResolutionValue) *Element {
	return e.setValue("image-resolution", v)
}
func (expression) isImageResolution()  {}
func (wideKeyword) isImageResolution() {}

// InitialLetterValue is a value for the CSS initial-letter property.
type InitialLetterValue interface {
	cssValue
	isInitialLetter()
}

// InitialLetter sets the CSS initial-letter property. Use RawValue for complex CSS syntax.
func (e *Element) InitialLetter(v InitialLetterValue) *Element {
	return e.setValue("initial-letter", v)
}
func (expression) isInitialLetter()  {}
func (wideKeyword) isInitialLetter() {}

// InitialLetterAlignValue is a value for the CSS initial-letter-align property.
type InitialLetterAlignValue interface {
	cssValue
	isInitialLetterAlign()
}

// InitialLetterAlign sets the CSS initial-letter-align property. Use RawValue for complex CSS syntax.
func (e *Element) InitialLetterAlign(v InitialLetterAlignValue) *Element {
	return e.setValue("initial-letter-align", v)
}
func (expression) isInitialLetterAlign()  {}
func (wideKeyword) isInitialLetterAlign() {}

// InitialLetterWrapValue is a value for the CSS initial-letter-wrap property.
type InitialLetterWrapValue interface {
	cssValue
	isInitialLetterWrap()
}

// InitialLetterWrap sets the CSS initial-letter-wrap property. Use RawValue for complex CSS syntax.
func (e *Element) InitialLetterWrap(v InitialLetterWrapValue) *Element {
	return e.setValue("initial-letter-wrap", v)
}
func (expression) isInitialLetterWrap()  {}
func (wideKeyword) isInitialLetterWrap() {}

// InlineSizeValue is a value for the CSS inline-size property.
type InlineSizeValue interface {
	cssValue
	isInlineSize()
}

// InlineSize sets the CSS inline-size property. Use RawValue for complex CSS syntax.
func (e *Element) InlineSize(v InlineSizeValue) *Element { return e.setValue("inline-size", v) }
func (expression) isInlineSize()                         {}
func (wideKeyword) isInlineSize()                        {}
func (Length) isInlineSize()                             {}
func (Percentage) isInlineSize()                         {}

// InsetValue is a value for the CSS inset property.
type InsetValue interface {
	cssValue
	isInset()
}

// Inset sets the CSS inset property. Use RawValue for complex CSS syntax.
func (e *Element) Inset(v InsetValue) *Element { return e.setValue("inset", v) }
func (expression) isInset()                    {}
func (wideKeyword) isInset()                   {}
func (Length) isInset()                        {}
func (Percentage) isInset()                    {}

// InsetBlockValue is a value for the CSS inset-block property.
type InsetBlockValue interface {
	cssValue
	isInsetBlock()
}

// InsetBlock sets the CSS inset-block property. Use RawValue for complex CSS syntax.
func (e *Element) InsetBlock(v InsetBlockValue) *Element { return e.setValue("inset-block", v) }
func (expression) isInsetBlock()                         {}
func (wideKeyword) isInsetBlock()                        {}
func (Length) isInsetBlock()                             {}
func (Percentage) isInsetBlock()                         {}

// InsetBlockEndValue is a value for the CSS inset-block-end property.
type InsetBlockEndValue interface {
	cssValue
	isInsetBlockEnd()
}

// InsetBlockEnd sets the CSS inset-block-end property. Use RawValue for complex CSS syntax.
func (e *Element) InsetBlockEnd(v InsetBlockEndValue) *Element {
	return e.setValue("inset-block-end", v)
}
func (expression) isInsetBlockEnd()  {}
func (wideKeyword) isInsetBlockEnd() {}
func (Length) isInsetBlockEnd()      {}
func (Percentage) isInsetBlockEnd()  {}

// InsetBlockStartValue is a value for the CSS inset-block-start property.
type InsetBlockStartValue interface {
	cssValue
	isInsetBlockStart()
}

// InsetBlockStart sets the CSS inset-block-start property. Use RawValue for complex CSS syntax.
func (e *Element) InsetBlockStart(v InsetBlockStartValue) *Element {
	return e.setValue("inset-block-start", v)
}
func (expression) isInsetBlockStart()  {}
func (wideKeyword) isInsetBlockStart() {}
func (Length) isInsetBlockStart()      {}
func (Percentage) isInsetBlockStart()  {}

// InsetInlineValue is a value for the CSS inset-inline property.
type InsetInlineValue interface {
	cssValue
	isInsetInline()
}

// InsetInline sets the CSS inset-inline property. Use RawValue for complex CSS syntax.
func (e *Element) InsetInline(v InsetInlineValue) *Element { return e.setValue("inset-inline", v) }
func (expression) isInsetInline()                          {}
func (wideKeyword) isInsetInline()                         {}
func (Length) isInsetInline()                              {}
func (Percentage) isInsetInline()                          {}

// InsetInlineEndValue is a value for the CSS inset-inline-end property.
type InsetInlineEndValue interface {
	cssValue
	isInsetInlineEnd()
}

// InsetInlineEnd sets the CSS inset-inline-end property. Use RawValue for complex CSS syntax.
func (e *Element) InsetInlineEnd(v InsetInlineEndValue) *Element {
	return e.setValue("inset-inline-end", v)
}
func (expression) isInsetInlineEnd()  {}
func (wideKeyword) isInsetInlineEnd() {}
func (Length) isInsetInlineEnd()      {}
func (Percentage) isInsetInlineEnd()  {}

// InsetInlineStartValue is a value for the CSS inset-inline-start property.
type InsetInlineStartValue interface {
	cssValue
	isInsetInlineStart()
}

// InsetInlineStart sets the CSS inset-inline-start property. Use RawValue for complex CSS syntax.
func (e *Element) InsetInlineStart(v InsetInlineStartValue) *Element {
	return e.setValue("inset-inline-start", v)
}
func (expression) isInsetInlineStart()  {}
func (wideKeyword) isInsetInlineStart() {}
func (Length) isInsetInlineStart()      {}
func (Percentage) isInsetInlineStart()  {}

// IsolationValue is a value for the CSS isolation property.
type IsolationValue interface {
	cssValue
	isIsolation()
}

// Isolation sets the CSS isolation property. Use RawValue for complex CSS syntax.
func (e *Element) Isolation(v IsolationValue) *Element { return e.setValue("isolation", v) }
func (expression) isIsolation()                        {}
func (wideKeyword) isIsolation()                       {}

// JustifyContentValue is a value for the CSS justify-content property.
type JustifyContentValue interface {
	cssValue
	isJustifyContent()
}

// JustifyContent sets the CSS justify-content property. Use RawValue for complex CSS syntax.
func (e *Element) JustifyContent(v JustifyContentValue) *Element {
	return e.setValue("justify-content", v)
}
func (expression) isJustifyContent()  {}
func (wideKeyword) isJustifyContent() {}

type justifyContentKeyword uint16

const (
	// JustifyContentNormal sets justify-content to normal.
	JustifyContentNormal justifyContentKeyword = 1
	// JustifyContentStart sets justify-content to start.
	JustifyContentStart justifyContentKeyword = 2
	// JustifyContentEnd sets justify-content to end.
	JustifyContentEnd justifyContentKeyword = 3
	// JustifyContentFlexStart sets justify-content to flex-start.
	JustifyContentFlexStart justifyContentKeyword = 4
	// JustifyContentFlexEnd sets justify-content to flex-end.
	JustifyContentFlexEnd justifyContentKeyword = 5
	// JustifyContentCenter sets justify-content to center.
	JustifyContentCenter justifyContentKeyword = 6
	// JustifyContentSpaceBetween sets justify-content to space-between.
	JustifyContentSpaceBetween justifyContentKeyword = 7
	// JustifyContentSpaceAround sets justify-content to space-around.
	JustifyContentSpaceAround justifyContentKeyword = 8
	// JustifyContentSpaceEvenly sets justify-content to space-evenly.
	JustifyContentSpaceEvenly justifyContentKeyword = 9
	// JustifyContentStretch sets justify-content to stretch.
	JustifyContentStretch justifyContentKeyword = 10
)

func (justifyContentKeyword) isJustifyContent() {}
func (v justifyContentKeyword) cssText() (string, error) {
	switch v {
	case JustifyContentNormal:
		return "normal", nil
	case JustifyContentStart:
		return "start", nil
	case JustifyContentEnd:
		return "end", nil
	case JustifyContentFlexStart:
		return "flex-start", nil
	case JustifyContentFlexEnd:
		return "flex-end", nil
	case JustifyContentCenter:
		return "center", nil
	case JustifyContentSpaceBetween:
		return "space-between", nil
	case JustifyContentSpaceAround:
		return "space-around", nil
	case JustifyContentSpaceEvenly:
		return "space-evenly", nil
	case JustifyContentStretch:
		return "stretch", nil
	default:
		return "", fmt.Errorf("ui: invalid justify-content keyword %d", v)
	}
}

// JustifyItemsValue is a value for the CSS justify-items property.
type JustifyItemsValue interface {
	cssValue
	isJustifyItems()
}

// JustifyItems sets the CSS justify-items property. Use RawValue for complex CSS syntax.
func (e *Element) JustifyItems(v JustifyItemsValue) *Element { return e.setValue("justify-items", v) }
func (expression) isJustifyItems()                           {}
func (wideKeyword) isJustifyItems()                          {}

// JustifySelfValue is a value for the CSS justify-self property.
type JustifySelfValue interface {
	cssValue
	isJustifySelf()
}

// JustifySelf sets the CSS justify-self property. Use RawValue for complex CSS syntax.
func (e *Element) JustifySelf(v JustifySelfValue) *Element { return e.setValue("justify-self", v) }
func (expression) isJustifySelf()                          {}
func (wideKeyword) isJustifySelf()                         {}

// LeftValue is a value for the CSS left property.
type LeftValue interface {
	cssValue
	isLeft()
}

// Left sets the CSS left property. Use RawValue for complex CSS syntax.
func (e *Element) Left(v LeftValue) *Element { return e.setValue("left", v) }
func (expression) isLeft()                   {}
func (wideKeyword) isLeft()                  {}
func (Length) isLeft()                       {}
func (Percentage) isLeft()                   {}

// LetterSpacingValue is a value for the CSS letter-spacing property.
type LetterSpacingValue interface {
	cssValue
	isLetterSpacing()
}

// LetterSpacing sets the CSS letter-spacing property. Use RawValue for complex CSS syntax.
func (e *Element) LetterSpacing(v LetterSpacingValue) *Element {
	return e.setValue("letter-spacing", v)
}
func (expression) isLetterSpacing()  {}
func (wideKeyword) isLetterSpacing() {}
func (Length) isLetterSpacing()      {}

// LineBreakValue is a value for the CSS line-break property.
type LineBreakValue interface {
	cssValue
	isLineBreak()
}

// LineBreak sets the CSS line-break property. Use RawValue for complex CSS syntax.
func (e *Element) LineBreak(v LineBreakValue) *Element { return e.setValue("line-break", v) }
func (expression) isLineBreak()                        {}
func (wideKeyword) isLineBreak()                       {}

// LineClampValue is a value for the CSS line-clamp property.
type LineClampValue interface {
	cssValue
	isLineClamp()
}

// LineClamp sets the CSS line-clamp property. Use RawValue for complex CSS syntax.
func (e *Element) LineClamp(v LineClampValue) *Element { return e.setValue("line-clamp", v) }
func (expression) isLineClamp()                        {}
func (wideKeyword) isLineClamp()                       {}

// LineHeightValue is a value for the CSS line-height property.
type LineHeightValue interface {
	cssValue
	isLineHeight()
}

// LineHeight sets the CSS line-height property. Use RawValue for complex CSS syntax.
func (e *Element) LineHeight(v LineHeightValue) *Element { return e.setValue("line-height", v) }
func (expression) isLineHeight()                         {}
func (wideKeyword) isLineHeight()                        {}
func (Length) isLineHeight()                             {}
func (Numeric) isLineHeight()                            {}
func (Percentage) isLineHeight()                         {}

// ListStyleValue is a value for the CSS list-style property.
type ListStyleValue interface {
	cssValue
	isListStyle()
}

// ListStyle sets the CSS list-style property. Use RawValue for complex CSS syntax.
func (e *Element) ListStyle(v ListStyleValue) *Element { return e.setValue("list-style", v) }
func (expression) isListStyle()                        {}
func (wideKeyword) isListStyle()                       {}

// ListStyleImageValue is a value for the CSS list-style-image property.
type ListStyleImageValue interface {
	cssValue
	isListStyleImage()
}

// ListStyleImage sets the CSS list-style-image property. Use RawValue for complex CSS syntax.
func (e *Element) ListStyleImage(v ListStyleImageValue) *Element {
	return e.setValue("list-style-image", v)
}
func (expression) isListStyleImage()  {}
func (wideKeyword) isListStyleImage() {}

// ListStylePositionValue is a value for the CSS list-style-position property.
type ListStylePositionValue interface {
	cssValue
	isListStylePosition()
}

// ListStylePosition sets the CSS list-style-position property. Use RawValue for complex CSS syntax.
func (e *Element) ListStylePosition(v ListStylePositionValue) *Element {
	return e.setValue("list-style-position", v)
}
func (expression) isListStylePosition()  {}
func (wideKeyword) isListStylePosition() {}

// ListStyleTypeValue is a value for the CSS list-style-type property.
type ListStyleTypeValue interface {
	cssValue
	isListStyleType()
}

// ListStyleType sets the CSS list-style-type property. Use RawValue for complex CSS syntax.
func (e *Element) ListStyleType(v ListStyleTypeValue) *Element {
	return e.setValue("list-style-type", v)
}
func (expression) isListStyleType()  {}
func (wideKeyword) isListStyleType() {}

// MarginValue is a value for the CSS margin property.
type MarginValue interface {
	cssValue
	isMargin()
}

// Margin sets the CSS margin property. Use RawValue for complex CSS syntax.
func (e *Element) Margin(v MarginValue) *Element { return e.setValue("margin", v) }
func (expression) isMargin()                     {}
func (wideKeyword) isMargin()                    {}
func (Length) isMargin()                         {}
func (Percentage) isMargin()                     {}
func (Spacing) isMargin()                        {}

// MarginBlockValue is a value for the CSS margin-block property.
type MarginBlockValue interface {
	cssValue
	isMarginBlock()
}

// MarginBlock sets the CSS margin-block property. Use RawValue for complex CSS syntax.
func (e *Element) MarginBlock(v MarginBlockValue) *Element { return e.setValue("margin-block", v) }
func (expression) isMarginBlock()                          {}
func (wideKeyword) isMarginBlock()                         {}
func (Length) isMarginBlock()                              {}
func (Percentage) isMarginBlock()                          {}
func (Spacing) isMarginBlock()                             {}

// MarginBlockEndValue is a value for the CSS margin-block-end property.
type MarginBlockEndValue interface {
	cssValue
	isMarginBlockEnd()
}

// MarginBlockEnd sets the CSS margin-block-end property. Use RawValue for complex CSS syntax.
func (e *Element) MarginBlockEnd(v MarginBlockEndValue) *Element {
	return e.setValue("margin-block-end", v)
}
func (expression) isMarginBlockEnd()  {}
func (wideKeyword) isMarginBlockEnd() {}
func (Length) isMarginBlockEnd()      {}
func (Percentage) isMarginBlockEnd()  {}

// MarginBlockStartValue is a value for the CSS margin-block-start property.
type MarginBlockStartValue interface {
	cssValue
	isMarginBlockStart()
}

// MarginBlockStart sets the CSS margin-block-start property. Use RawValue for complex CSS syntax.
func (e *Element) MarginBlockStart(v MarginBlockStartValue) *Element {
	return e.setValue("margin-block-start", v)
}
func (expression) isMarginBlockStart()  {}
func (wideKeyword) isMarginBlockStart() {}
func (Length) isMarginBlockStart()      {}
func (Percentage) isMarginBlockStart()  {}

// MarginBottomValue is a value for the CSS margin-bottom property.
type MarginBottomValue interface {
	cssValue
	isMarginBottom()
}

// MarginBottom sets the CSS margin-bottom property. Use RawValue for complex CSS syntax.
func (e *Element) MarginBottom(v MarginBottomValue) *Element { return e.setValue("margin-bottom", v) }
func (expression) isMarginBottom()                           {}
func (wideKeyword) isMarginBottom()                          {}
func (Length) isMarginBottom()                               {}
func (Percentage) isMarginBottom()                           {}

// MarginInlineValue is a value for the CSS margin-inline property.
type MarginInlineValue interface {
	cssValue
	isMarginInline()
}

// MarginInline sets the CSS margin-inline property. Use RawValue for complex CSS syntax.
func (e *Element) MarginInline(v MarginInlineValue) *Element { return e.setValue("margin-inline", v) }
func (expression) isMarginInline()                           {}
func (wideKeyword) isMarginInline()                          {}
func (Length) isMarginInline()                               {}
func (Percentage) isMarginInline()                           {}
func (Spacing) isMarginInline()                              {}

// MarginInlineEndValue is a value for the CSS margin-inline-end property.
type MarginInlineEndValue interface {
	cssValue
	isMarginInlineEnd()
}

// MarginInlineEnd sets the CSS margin-inline-end property. Use RawValue for complex CSS syntax.
func (e *Element) MarginInlineEnd(v MarginInlineEndValue) *Element {
	return e.setValue("margin-inline-end", v)
}
func (expression) isMarginInlineEnd()  {}
func (wideKeyword) isMarginInlineEnd() {}
func (Length) isMarginInlineEnd()      {}
func (Percentage) isMarginInlineEnd()  {}

// MarginInlineStartValue is a value for the CSS margin-inline-start property.
type MarginInlineStartValue interface {
	cssValue
	isMarginInlineStart()
}

// MarginInlineStart sets the CSS margin-inline-start property. Use RawValue for complex CSS syntax.
func (e *Element) MarginInlineStart(v MarginInlineStartValue) *Element {
	return e.setValue("margin-inline-start", v)
}
func (expression) isMarginInlineStart()  {}
func (wideKeyword) isMarginInlineStart() {}
func (Length) isMarginInlineStart()      {}
func (Percentage) isMarginInlineStart()  {}

// MarginLeftValue is a value for the CSS margin-left property.
type MarginLeftValue interface {
	cssValue
	isMarginLeft()
}

// MarginLeft sets the CSS margin-left property. Use RawValue for complex CSS syntax.
func (e *Element) MarginLeft(v MarginLeftValue) *Element { return e.setValue("margin-left", v) }
func (expression) isMarginLeft()                         {}
func (wideKeyword) isMarginLeft()                        {}
func (Length) isMarginLeft()                             {}
func (Percentage) isMarginLeft()                         {}

// MarginRightValue is a value for the CSS margin-right property.
type MarginRightValue interface {
	cssValue
	isMarginRight()
}

// MarginRight sets the CSS margin-right property. Use RawValue for complex CSS syntax.
func (e *Element) MarginRight(v MarginRightValue) *Element { return e.setValue("margin-right", v) }
func (expression) isMarginRight()                          {}
func (wideKeyword) isMarginRight()                         {}
func (Length) isMarginRight()                              {}
func (Percentage) isMarginRight()                          {}

// MarginTopValue is a value for the CSS margin-top property.
type MarginTopValue interface {
	cssValue
	isMarginTop()
}

// MarginTop sets the CSS margin-top property. Use RawValue for complex CSS syntax.
func (e *Element) MarginTop(v MarginTopValue) *Element { return e.setValue("margin-top", v) }
func (expression) isMarginTop()                        {}
func (wideKeyword) isMarginTop()                       {}
func (Length) isMarginTop()                            {}
func (Percentage) isMarginTop()                        {}

// MarginTrimValue is a value for the CSS margin-trim property.
type MarginTrimValue interface {
	cssValue
	isMarginTrim()
}

// MarginTrim sets the CSS margin-trim property. Use RawValue for complex CSS syntax.
func (e *Element) MarginTrim(v MarginTrimValue) *Element { return e.setValue("margin-trim", v) }
func (expression) isMarginTrim()                         {}
func (wideKeyword) isMarginTrim()                        {}

// MaskValue is a value for the CSS mask property.
type MaskValue interface {
	cssValue
	isMask()
}

// Mask sets the CSS mask property. Use RawValue for complex CSS syntax.
func (e *Element) Mask(v MaskValue) *Element { return e.setValue("mask", v) }
func (expression) isMask()                   {}
func (wideKeyword) isMask()                  {}

// MaskBorderValue is a value for the CSS mask-border property.
type MaskBorderValue interface {
	cssValue
	isMaskBorder()
}

// MaskBorder sets the CSS mask-border property. Use RawValue for complex CSS syntax.
func (e *Element) MaskBorder(v MaskBorderValue) *Element { return e.setValue("mask-border", v) }
func (expression) isMaskBorder()                         {}
func (wideKeyword) isMaskBorder()                        {}

// MaskBorderModeValue is a value for the CSS mask-border-mode property.
type MaskBorderModeValue interface {
	cssValue
	isMaskBorderMode()
}

// MaskBorderMode sets the CSS mask-border-mode property. Use RawValue for complex CSS syntax.
func (e *Element) MaskBorderMode(v MaskBorderModeValue) *Element {
	return e.setValue("mask-border-mode", v)
}
func (expression) isMaskBorderMode()  {}
func (wideKeyword) isMaskBorderMode() {}

// MaskBorderOutsetValue is a value for the CSS mask-border-outset property.
type MaskBorderOutsetValue interface {
	cssValue
	isMaskBorderOutset()
}

// MaskBorderOutset sets the CSS mask-border-outset property. Use RawValue for complex CSS syntax.
func (e *Element) MaskBorderOutset(v MaskBorderOutsetValue) *Element {
	return e.setValue("mask-border-outset", v)
}
func (expression) isMaskBorderOutset()  {}
func (wideKeyword) isMaskBorderOutset() {}

// MaskBorderRepeatValue is a value for the CSS mask-border-repeat property.
type MaskBorderRepeatValue interface {
	cssValue
	isMaskBorderRepeat()
}

// MaskBorderRepeat sets the CSS mask-border-repeat property. Use RawValue for complex CSS syntax.
func (e *Element) MaskBorderRepeat(v MaskBorderRepeatValue) *Element {
	return e.setValue("mask-border-repeat", v)
}
func (expression) isMaskBorderRepeat()  {}
func (wideKeyword) isMaskBorderRepeat() {}

// MaskBorderSliceValue is a value for the CSS mask-border-slice property.
type MaskBorderSliceValue interface {
	cssValue
	isMaskBorderSlice()
}

// MaskBorderSlice sets the CSS mask-border-slice property. Use RawValue for complex CSS syntax.
func (e *Element) MaskBorderSlice(v MaskBorderSliceValue) *Element {
	return e.setValue("mask-border-slice", v)
}
func (expression) isMaskBorderSlice()  {}
func (wideKeyword) isMaskBorderSlice() {}

// MaskBorderSourceValue is a value for the CSS mask-border-source property.
type MaskBorderSourceValue interface {
	cssValue
	isMaskBorderSource()
}

// MaskBorderSource sets the CSS mask-border-source property. Use RawValue for complex CSS syntax.
func (e *Element) MaskBorderSource(v MaskBorderSourceValue) *Element {
	return e.setValue("mask-border-source", v)
}
func (expression) isMaskBorderSource()  {}
func (wideKeyword) isMaskBorderSource() {}

// MaskBorderWidthValue is a value for the CSS mask-border-width property.
type MaskBorderWidthValue interface {
	cssValue
	isMaskBorderWidth()
}

// MaskBorderWidth sets the CSS mask-border-width property. Use RawValue for complex CSS syntax.
func (e *Element) MaskBorderWidth(v MaskBorderWidthValue) *Element {
	return e.setValue("mask-border-width", v)
}
func (expression) isMaskBorderWidth()  {}
func (wideKeyword) isMaskBorderWidth() {}

// MaskClipValue is a value for the CSS mask-clip property.
type MaskClipValue interface {
	cssValue
	isMaskClip()
}

// MaskClip sets the CSS mask-clip property. Use RawValue for complex CSS syntax.
func (e *Element) MaskClip(v MaskClipValue) *Element { return e.setValue("mask-clip", v) }
func (expression) isMaskClip()                       {}
func (wideKeyword) isMaskClip()                      {}

// MaskCompositeValue is a value for the CSS mask-composite property.
type MaskCompositeValue interface {
	cssValue
	isMaskComposite()
}

// MaskComposite sets the CSS mask-composite property. Use RawValue for complex CSS syntax.
func (e *Element) MaskComposite(v MaskCompositeValue) *Element {
	return e.setValue("mask-composite", v)
}
func (expression) isMaskComposite()  {}
func (wideKeyword) isMaskComposite() {}

// MaskImageValue is a value for the CSS mask-image property.
type MaskImageValue interface {
	cssValue
	isMaskImage()
}

// MaskImage sets the CSS mask-image property. Use RawValue for complex CSS syntax.
func (e *Element) MaskImage(v MaskImageValue) *Element { return e.setValue("mask-image", v) }
func (expression) isMaskImage()                        {}
func (wideKeyword) isMaskImage()                       {}

// MaskModeValue is a value for the CSS mask-mode property.
type MaskModeValue interface {
	cssValue
	isMaskMode()
}

// MaskMode sets the CSS mask-mode property. Use RawValue for complex CSS syntax.
func (e *Element) MaskMode(v MaskModeValue) *Element { return e.setValue("mask-mode", v) }
func (expression) isMaskMode()                       {}
func (wideKeyword) isMaskMode()                      {}

// MaskOriginValue is a value for the CSS mask-origin property.
type MaskOriginValue interface {
	cssValue
	isMaskOrigin()
}

// MaskOrigin sets the CSS mask-origin property. Use RawValue for complex CSS syntax.
func (e *Element) MaskOrigin(v MaskOriginValue) *Element { return e.setValue("mask-origin", v) }
func (expression) isMaskOrigin()                         {}
func (wideKeyword) isMaskOrigin()                        {}

// MaskPositionValue is a value for the CSS mask-position property.
type MaskPositionValue interface {
	cssValue
	isMaskPosition()
}

// MaskPosition sets the CSS mask-position property. Use RawValue for complex CSS syntax.
func (e *Element) MaskPosition(v MaskPositionValue) *Element { return e.setValue("mask-position", v) }
func (expression) isMaskPosition()                           {}
func (wideKeyword) isMaskPosition()                          {}

// MaskRepeatValue is a value for the CSS mask-repeat property.
type MaskRepeatValue interface {
	cssValue
	isMaskRepeat()
}

// MaskRepeat sets the CSS mask-repeat property. Use RawValue for complex CSS syntax.
func (e *Element) MaskRepeat(v MaskRepeatValue) *Element { return e.setValue("mask-repeat", v) }
func (expression) isMaskRepeat()                         {}
func (wideKeyword) isMaskRepeat()                        {}

// MaskSizeValue is a value for the CSS mask-size property.
type MaskSizeValue interface {
	cssValue
	isMaskSize()
}

// MaskSize sets the CSS mask-size property. Use RawValue for complex CSS syntax.
func (e *Element) MaskSize(v MaskSizeValue) *Element { return e.setValue("mask-size", v) }
func (expression) isMaskSize()                       {}
func (wideKeyword) isMaskSize()                      {}

// MaxBlockSizeValue is a value for the CSS max-block-size property.
type MaxBlockSizeValue interface {
	cssValue
	isMaxBlockSize()
}

// MaxBlockSize sets the CSS max-block-size property. Use RawValue for complex CSS syntax.
func (e *Element) MaxBlockSize(v MaxBlockSizeValue) *Element { return e.setValue("max-block-size", v) }
func (expression) isMaxBlockSize()                           {}
func (wideKeyword) isMaxBlockSize()                          {}
func (Length) isMaxBlockSize()                               {}
func (Percentage) isMaxBlockSize()                           {}

// MaxHeightValue is a value for the CSS max-height property.
type MaxHeightValue interface {
	cssValue
	isMaxHeight()
}

// MaxHeight sets the CSS max-height property. Use RawValue for complex CSS syntax.
func (e *Element) MaxHeight(v MaxHeightValue) *Element { return e.setValue("max-height", v) }
func (expression) isMaxHeight()                        {}
func (wideKeyword) isMaxHeight()                       {}
func (Length) isMaxHeight()                            {}
func (Percentage) isMaxHeight()                        {}

// MaxInlineSizeValue is a value for the CSS max-inline-size property.
type MaxInlineSizeValue interface {
	cssValue
	isMaxInlineSize()
}

// MaxInlineSize sets the CSS max-inline-size property. Use RawValue for complex CSS syntax.
func (e *Element) MaxInlineSize(v MaxInlineSizeValue) *Element {
	return e.setValue("max-inline-size", v)
}
func (expression) isMaxInlineSize()  {}
func (wideKeyword) isMaxInlineSize() {}
func (Length) isMaxInlineSize()      {}
func (Percentage) isMaxInlineSize()  {}

// MaxLinesValue is a value for the CSS max-lines property.
type MaxLinesValue interface {
	cssValue
	isMaxLines()
}

// MaxLines sets the CSS max-lines property. Use RawValue for complex CSS syntax.
func (e *Element) MaxLines(v MaxLinesValue) *Element { return e.setValue("max-lines", v) }
func (expression) isMaxLines()                       {}
func (wideKeyword) isMaxLines()                      {}

// MaxWidthValue is a value for the CSS max-width property.
type MaxWidthValue interface {
	cssValue
	isMaxWidth()
}

// MaxWidth sets the CSS max-width property. Use RawValue for complex CSS syntax.
func (e *Element) MaxWidth(v MaxWidthValue) *Element { return e.setValue("max-width", v) }
func (expression) isMaxWidth()                       {}
func (wideKeyword) isMaxWidth()                      {}
func (Length) isMaxWidth()                           {}
func (Percentage) isMaxWidth()                       {}

// MinBlockSizeValue is a value for the CSS min-block-size property.
type MinBlockSizeValue interface {
	cssValue
	isMinBlockSize()
}

// MinBlockSize sets the CSS min-block-size property. Use RawValue for complex CSS syntax.
func (e *Element) MinBlockSize(v MinBlockSizeValue) *Element { return e.setValue("min-block-size", v) }
func (expression) isMinBlockSize()                           {}
func (wideKeyword) isMinBlockSize()                          {}
func (Length) isMinBlockSize()                               {}
func (Percentage) isMinBlockSize()                           {}

// MinHeightValue is a value for the CSS min-height property.
type MinHeightValue interface {
	cssValue
	isMinHeight()
}

// MinHeight sets the CSS min-height property. Use RawValue for complex CSS syntax.
func (e *Element) MinHeight(v MinHeightValue) *Element { return e.setValue("min-height", v) }
func (expression) isMinHeight()                        {}
func (wideKeyword) isMinHeight()                       {}
func (Length) isMinHeight()                            {}
func (Percentage) isMinHeight()                        {}

// MinInlineSizeValue is a value for the CSS min-inline-size property.
type MinInlineSizeValue interface {
	cssValue
	isMinInlineSize()
}

// MinInlineSize sets the CSS min-inline-size property. Use RawValue for complex CSS syntax.
func (e *Element) MinInlineSize(v MinInlineSizeValue) *Element {
	return e.setValue("min-inline-size", v)
}
func (expression) isMinInlineSize()  {}
func (wideKeyword) isMinInlineSize() {}
func (Length) isMinInlineSize()      {}
func (Percentage) isMinInlineSize()  {}

// MinWidthValue is a value for the CSS min-width property.
type MinWidthValue interface {
	cssValue
	isMinWidth()
}

// MinWidth sets the CSS min-width property. Use RawValue for complex CSS syntax.
func (e *Element) MinWidth(v MinWidthValue) *Element { return e.setValue("min-width", v) }
func (expression) isMinWidth()                       {}
func (wideKeyword) isMinWidth()                      {}
func (Length) isMinWidth()                           {}
func (Percentage) isMinWidth()                       {}

// MixBlendModeValue is a value for the CSS mix-blend-mode property.
type MixBlendModeValue interface {
	cssValue
	isMixBlendMode()
}

// MixBlendMode sets the CSS mix-blend-mode property. Use RawValue for complex CSS syntax.
func (e *Element) MixBlendMode(v MixBlendModeValue) *Element { return e.setValue("mix-blend-mode", v) }
func (expression) isMixBlendMode()                           {}
func (wideKeyword) isMixBlendMode()                          {}

// ObjectFitValue is a value for the CSS object-fit property.
type ObjectFitValue interface {
	cssValue
	isObjectFit()
}

// ObjectFit sets the CSS object-fit property. Use RawValue for complex CSS syntax.
func (e *Element) ObjectFit(v ObjectFitValue) *Element { return e.setValue("object-fit", v) }
func (expression) isObjectFit()                        {}
func (wideKeyword) isObjectFit()                       {}

// ObjectPositionValue is a value for the CSS object-position property.
type ObjectPositionValue interface {
	cssValue
	isObjectPosition()
}

// ObjectPosition sets the CSS object-position property. Use RawValue for complex CSS syntax.
func (e *Element) ObjectPosition(v ObjectPositionValue) *Element {
	return e.setValue("object-position", v)
}
func (expression) isObjectPosition()  {}
func (wideKeyword) isObjectPosition() {}

// ObjectViewBoxValue is a value for the CSS object-view-box property.
type ObjectViewBoxValue interface {
	cssValue
	isObjectViewBox()
}

// ObjectViewBox sets the CSS object-view-box property. Use RawValue for complex CSS syntax.
func (e *Element) ObjectViewBox(v ObjectViewBoxValue) *Element {
	return e.setValue("object-view-box", v)
}
func (expression) isObjectViewBox()  {}
func (wideKeyword) isObjectViewBox() {}

// OpacityValue is a value for the CSS opacity property.
type OpacityValue interface {
	cssValue
	isOpacity()
}

// Opacity sets the CSS opacity property. Use RawValue for complex CSS syntax.
func (e *Element) Opacity(v OpacityValue) *Element { return e.setValue("opacity", v) }
func (expression) isOpacity()                      {}
func (wideKeyword) isOpacity()                     {}
func (Numeric) isOpacity()                         {}

// OrderValue is a value for the CSS order property.
type OrderValue interface {
	cssValue
	isOrder()
}

// Order sets the CSS order property. Use RawValue for complex CSS syntax.
func (e *Element) Order(v OrderValue) *Element { return e.setValue("order", v) }
func (expression) isOrder()                    {}
func (wideKeyword) isOrder()                   {}
func (Numeric) isOrder()                       {}

// OrphansValue is a value for the CSS orphans property.
type OrphansValue interface {
	cssValue
	isOrphans()
}

// Orphans sets the CSS orphans property. Use RawValue for complex CSS syntax.
func (e *Element) Orphans(v OrphansValue) *Element { return e.setValue("orphans", v) }
func (expression) isOrphans()                      {}
func (wideKeyword) isOrphans()                     {}
func (Numeric) isOrphans()                         {}

// OutlineValue is a value for the CSS outline property.
type OutlineValue interface {
	cssValue
	isOutline()
}

// Outline sets the CSS outline property. Use RawValue for complex CSS syntax.
func (e *Element) Outline(v OutlineValue) *Element { return e.setValue("outline", v) }
func (expression) isOutline()                      {}
func (wideKeyword) isOutline()                     {}

// OutlineColorValue is a value for the CSS outline-color property.
type OutlineColorValue interface {
	cssValue
	isOutlineColor()
}

// OutlineColor sets the CSS outline-color property. Use RawValue for complex CSS syntax.
func (e *Element) OutlineColor(v OutlineColorValue) *Element { return e.setValue("outline-color", v) }
func (expression) isOutlineColor()                           {}
func (wideKeyword) isOutlineColor()                          {}
func (Color) isOutlineColor()                                {}

// OutlineOffsetValue is a value for the CSS outline-offset property.
type OutlineOffsetValue interface {
	cssValue
	isOutlineOffset()
}

// OutlineOffset sets the CSS outline-offset property. Use RawValue for complex CSS syntax.
func (e *Element) OutlineOffset(v OutlineOffsetValue) *Element {
	return e.setValue("outline-offset", v)
}
func (expression) isOutlineOffset()  {}
func (wideKeyword) isOutlineOffset() {}
func (Length) isOutlineOffset()      {}

// OutlineStyleValue is a value for the CSS outline-style property.
type OutlineStyleValue interface {
	cssValue
	isOutlineStyle()
}

// OutlineStyle sets the CSS outline-style property. Use RawValue for complex CSS syntax.
func (e *Element) OutlineStyle(v OutlineStyleValue) *Element { return e.setValue("outline-style", v) }
func (expression) isOutlineStyle()                           {}
func (wideKeyword) isOutlineStyle()                          {}

// OutlineWidthValue is a value for the CSS outline-width property.
type OutlineWidthValue interface {
	cssValue
	isOutlineWidth()
}

// OutlineWidth sets the CSS outline-width property. Use RawValue for complex CSS syntax.
func (e *Element) OutlineWidth(v OutlineWidthValue) *Element { return e.setValue("outline-width", v) }
func (expression) isOutlineWidth()                           {}
func (wideKeyword) isOutlineWidth()                          {}
func (Length) isOutlineWidth()                               {}

// OverflowValue is a value for the CSS overflow property.
type OverflowValue interface {
	cssValue
	isOverflow()
}

// Overflow sets the CSS overflow property. Use RawValue for complex CSS syntax.
func (e *Element) Overflow(v OverflowValue) *Element { return e.setValue("overflow", v) }
func (expression) isOverflow()                       {}
func (wideKeyword) isOverflow()                      {}

type overflowKeyword uint16

const (
	// OverflowVisible sets overflow to visible.
	OverflowVisible overflowKeyword = 1
	// OverflowHidden sets overflow to hidden.
	OverflowHidden overflowKeyword = 2
	// OverflowClip sets overflow to clip.
	OverflowClip overflowKeyword = 3
	// OverflowScroll sets overflow to scroll.
	OverflowScroll overflowKeyword = 4
	// OverflowAuto sets overflow to auto.
	OverflowAuto overflowKeyword = 5
)

func (overflowKeyword) isOverflow() {}
func (v overflowKeyword) cssText() (string, error) {
	switch v {
	case OverflowVisible:
		return "visible", nil
	case OverflowHidden:
		return "hidden", nil
	case OverflowClip:
		return "clip", nil
	case OverflowScroll:
		return "scroll", nil
	case OverflowAuto:
		return "auto", nil
	default:
		return "", fmt.Errorf("ui: invalid overflow keyword %d", v)
	}
}

// OverflowBlockValue is a value for the CSS overflow-block property.
type OverflowBlockValue interface {
	cssValue
	isOverflowBlock()
}

// OverflowBlock sets the CSS overflow-block property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowBlock(v OverflowBlockValue) *Element {
	return e.setValue("overflow-block", v)
}
func (expression) isOverflowBlock()  {}
func (wideKeyword) isOverflowBlock() {}

// OverflowClipMarginValue is a value for the CSS overflow-clip-margin property.
type OverflowClipMarginValue interface {
	cssValue
	isOverflowClipMargin()
}

// OverflowClipMargin sets the CSS overflow-clip-margin property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowClipMargin(v OverflowClipMarginValue) *Element {
	return e.setValue("overflow-clip-margin", v)
}
func (expression) isOverflowClipMargin()  {}
func (wideKeyword) isOverflowClipMargin() {}

// OverflowClipMarginBlockValue is a value for the CSS overflow-clip-margin-block property.
type OverflowClipMarginBlockValue interface {
	cssValue
	isOverflowClipMarginBlock()
}

// OverflowClipMarginBlock sets the CSS overflow-clip-margin-block property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowClipMarginBlock(v OverflowClipMarginBlockValue) *Element {
	return e.setValue("overflow-clip-margin-block", v)
}
func (expression) isOverflowClipMarginBlock()  {}
func (wideKeyword) isOverflowClipMarginBlock() {}

// OverflowClipMarginBlockEndValue is a value for the CSS overflow-clip-margin-block-end property.
type OverflowClipMarginBlockEndValue interface {
	cssValue
	isOverflowClipMarginBlockEnd()
}

// OverflowClipMarginBlockEnd sets the CSS overflow-clip-margin-block-end property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowClipMarginBlockEnd(v OverflowClipMarginBlockEndValue) *Element {
	return e.setValue("overflow-clip-margin-block-end", v)
}
func (expression) isOverflowClipMarginBlockEnd()  {}
func (wideKeyword) isOverflowClipMarginBlockEnd() {}

// OverflowClipMarginBlockStartValue is a value for the CSS overflow-clip-margin-block-start property.
type OverflowClipMarginBlockStartValue interface {
	cssValue
	isOverflowClipMarginBlockStart()
}

// OverflowClipMarginBlockStart sets the CSS overflow-clip-margin-block-start property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowClipMarginBlockStart(v OverflowClipMarginBlockStartValue) *Element {
	return e.setValue("overflow-clip-margin-block-start", v)
}
func (expression) isOverflowClipMarginBlockStart()  {}
func (wideKeyword) isOverflowClipMarginBlockStart() {}

// OverflowClipMarginBottomValue is a value for the CSS overflow-clip-margin-bottom property.
type OverflowClipMarginBottomValue interface {
	cssValue
	isOverflowClipMarginBottom()
}

// OverflowClipMarginBottom sets the CSS overflow-clip-margin-bottom property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowClipMarginBottom(v OverflowClipMarginBottomValue) *Element {
	return e.setValue("overflow-clip-margin-bottom", v)
}
func (expression) isOverflowClipMarginBottom()  {}
func (wideKeyword) isOverflowClipMarginBottom() {}

// OverflowClipMarginInlineValue is a value for the CSS overflow-clip-margin-inline property.
type OverflowClipMarginInlineValue interface {
	cssValue
	isOverflowClipMarginInline()
}

// OverflowClipMarginInline sets the CSS overflow-clip-margin-inline property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowClipMarginInline(v OverflowClipMarginInlineValue) *Element {
	return e.setValue("overflow-clip-margin-inline", v)
}
func (expression) isOverflowClipMarginInline()  {}
func (wideKeyword) isOverflowClipMarginInline() {}

// OverflowClipMarginInlineEndValue is a value for the CSS overflow-clip-margin-inline-end property.
type OverflowClipMarginInlineEndValue interface {
	cssValue
	isOverflowClipMarginInlineEnd()
}

// OverflowClipMarginInlineEnd sets the CSS overflow-clip-margin-inline-end property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowClipMarginInlineEnd(v OverflowClipMarginInlineEndValue) *Element {
	return e.setValue("overflow-clip-margin-inline-end", v)
}
func (expression) isOverflowClipMarginInlineEnd()  {}
func (wideKeyword) isOverflowClipMarginInlineEnd() {}

// OverflowClipMarginInlineStartValue is a value for the CSS overflow-clip-margin-inline-start property.
type OverflowClipMarginInlineStartValue interface {
	cssValue
	isOverflowClipMarginInlineStart()
}

// OverflowClipMarginInlineStart sets the CSS overflow-clip-margin-inline-start property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowClipMarginInlineStart(v OverflowClipMarginInlineStartValue) *Element {
	return e.setValue("overflow-clip-margin-inline-start", v)
}
func (expression) isOverflowClipMarginInlineStart()  {}
func (wideKeyword) isOverflowClipMarginInlineStart() {}

// OverflowClipMarginLeftValue is a value for the CSS overflow-clip-margin-left property.
type OverflowClipMarginLeftValue interface {
	cssValue
	isOverflowClipMarginLeft()
}

// OverflowClipMarginLeft sets the CSS overflow-clip-margin-left property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowClipMarginLeft(v OverflowClipMarginLeftValue) *Element {
	return e.setValue("overflow-clip-margin-left", v)
}
func (expression) isOverflowClipMarginLeft()  {}
func (wideKeyword) isOverflowClipMarginLeft() {}

// OverflowClipMarginRightValue is a value for the CSS overflow-clip-margin-right property.
type OverflowClipMarginRightValue interface {
	cssValue
	isOverflowClipMarginRight()
}

// OverflowClipMarginRight sets the CSS overflow-clip-margin-right property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowClipMarginRight(v OverflowClipMarginRightValue) *Element {
	return e.setValue("overflow-clip-margin-right", v)
}
func (expression) isOverflowClipMarginRight()  {}
func (wideKeyword) isOverflowClipMarginRight() {}

// OverflowClipMarginTopValue is a value for the CSS overflow-clip-margin-top property.
type OverflowClipMarginTopValue interface {
	cssValue
	isOverflowClipMarginTop()
}

// OverflowClipMarginTop sets the CSS overflow-clip-margin-top property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowClipMarginTop(v OverflowClipMarginTopValue) *Element {
	return e.setValue("overflow-clip-margin-top", v)
}
func (expression) isOverflowClipMarginTop()  {}
func (wideKeyword) isOverflowClipMarginTop() {}

// OverflowInlineValue is a value for the CSS overflow-inline property.
type OverflowInlineValue interface {
	cssValue
	isOverflowInline()
}

// OverflowInline sets the CSS overflow-inline property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowInline(v OverflowInlineValue) *Element {
	return e.setValue("overflow-inline", v)
}
func (expression) isOverflowInline()  {}
func (wideKeyword) isOverflowInline() {}

// OverflowWrapValue is a value for the CSS overflow-wrap property.
type OverflowWrapValue interface {
	cssValue
	isOverflowWrap()
}

// OverflowWrap sets the CSS overflow-wrap property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowWrap(v OverflowWrapValue) *Element { return e.setValue("overflow-wrap", v) }
func (expression) isOverflowWrap()                           {}
func (wideKeyword) isOverflowWrap()                          {}

type overflowWrapKeyword uint16

const (
	// OverflowWrapNormal sets overflow-wrap to normal.
	OverflowWrapNormal overflowWrapKeyword = 1
	// OverflowWrapBreakWord sets overflow-wrap to break-word.
	OverflowWrapBreakWord overflowWrapKeyword = 2
	// OverflowWrapAnywhere sets overflow-wrap to anywhere.
	OverflowWrapAnywhere overflowWrapKeyword = 3
)

func (overflowWrapKeyword) isOverflowWrap() {}
func (v overflowWrapKeyword) cssText() (string, error) {
	switch v {
	case OverflowWrapNormal:
		return "normal", nil
	case OverflowWrapBreakWord:
		return "break-word", nil
	case OverflowWrapAnywhere:
		return "anywhere", nil
	default:
		return "", fmt.Errorf("ui: invalid overflow-wrap keyword %d", v)
	}
}

// OverflowXValue is a value for the CSS overflow-x property.
type OverflowXValue interface {
	cssValue
	isOverflowX()
}

// OverflowX sets the CSS overflow-x property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowX(v OverflowXValue) *Element { return e.setValue("overflow-x", v) }
func (expression) isOverflowX()                        {}
func (wideKeyword) isOverflowX()                       {}

type overflowXKeyword uint16

const (
	// OverflowXVisible sets overflow-x to visible.
	OverflowXVisible overflowXKeyword = 1
	// OverflowXHidden sets overflow-x to hidden.
	OverflowXHidden overflowXKeyword = 2
	// OverflowXClip sets overflow-x to clip.
	OverflowXClip overflowXKeyword = 3
	// OverflowXScroll sets overflow-x to scroll.
	OverflowXScroll overflowXKeyword = 4
	// OverflowXAuto sets overflow-x to auto.
	OverflowXAuto overflowXKeyword = 5
)

func (overflowXKeyword) isOverflowX() {}
func (v overflowXKeyword) cssText() (string, error) {
	switch v {
	case OverflowXVisible:
		return "visible", nil
	case OverflowXHidden:
		return "hidden", nil
	case OverflowXClip:
		return "clip", nil
	case OverflowXScroll:
		return "scroll", nil
	case OverflowXAuto:
		return "auto", nil
	default:
		return "", fmt.Errorf("ui: invalid overflow-x keyword %d", v)
	}
}

// OverflowYValue is a value for the CSS overflow-y property.
type OverflowYValue interface {
	cssValue
	isOverflowY()
}

// OverflowY sets the CSS overflow-y property. Use RawValue for complex CSS syntax.
func (e *Element) OverflowY(v OverflowYValue) *Element { return e.setValue("overflow-y", v) }
func (expression) isOverflowY()                        {}
func (wideKeyword) isOverflowY()                       {}

type overflowYKeyword uint16

const (
	// OverflowYVisible sets overflow-y to visible.
	OverflowYVisible overflowYKeyword = 1
	// OverflowYHidden sets overflow-y to hidden.
	OverflowYHidden overflowYKeyword = 2
	// OverflowYClip sets overflow-y to clip.
	OverflowYClip overflowYKeyword = 3
	// OverflowYScroll sets overflow-y to scroll.
	OverflowYScroll overflowYKeyword = 4
	// OverflowYAuto sets overflow-y to auto.
	OverflowYAuto overflowYKeyword = 5
)

func (overflowYKeyword) isOverflowY() {}
func (v overflowYKeyword) cssText() (string, error) {
	switch v {
	case OverflowYVisible:
		return "visible", nil
	case OverflowYHidden:
		return "hidden", nil
	case OverflowYClip:
		return "clip", nil
	case OverflowYScroll:
		return "scroll", nil
	case OverflowYAuto:
		return "auto", nil
	default:
		return "", fmt.Errorf("ui: invalid overflow-y keyword %d", v)
	}
}

// PaddingValue is a value for the CSS padding property.
type PaddingValue interface {
	cssValue
	isPadding()
}

// Padding sets the CSS padding property. Use RawValue for complex CSS syntax.
func (e *Element) Padding(v PaddingValue) *Element { return e.setValue("padding", v) }
func (expression) isPadding()                      {}
func (wideKeyword) isPadding()                     {}
func (Length) isPadding()                          {}
func (Percentage) isPadding()                      {}
func (Spacing) isPadding()                         {}

// PaddingBlockValue is a value for the CSS padding-block property.
type PaddingBlockValue interface {
	cssValue
	isPaddingBlock()
}

// PaddingBlock sets the CSS padding-block property. Use RawValue for complex CSS syntax.
func (e *Element) PaddingBlock(v PaddingBlockValue) *Element { return e.setValue("padding-block", v) }
func (expression) isPaddingBlock()                           {}
func (wideKeyword) isPaddingBlock()                          {}
func (Length) isPaddingBlock()                               {}
func (Percentage) isPaddingBlock()                           {}
func (Spacing) isPaddingBlock()                              {}

// PaddingBlockEndValue is a value for the CSS padding-block-end property.
type PaddingBlockEndValue interface {
	cssValue
	isPaddingBlockEnd()
}

// PaddingBlockEnd sets the CSS padding-block-end property. Use RawValue for complex CSS syntax.
func (e *Element) PaddingBlockEnd(v PaddingBlockEndValue) *Element {
	return e.setValue("padding-block-end", v)
}
func (expression) isPaddingBlockEnd()  {}
func (wideKeyword) isPaddingBlockEnd() {}
func (Length) isPaddingBlockEnd()      {}
func (Percentage) isPaddingBlockEnd()  {}

// PaddingBlockStartValue is a value for the CSS padding-block-start property.
type PaddingBlockStartValue interface {
	cssValue
	isPaddingBlockStart()
}

// PaddingBlockStart sets the CSS padding-block-start property. Use RawValue for complex CSS syntax.
func (e *Element) PaddingBlockStart(v PaddingBlockStartValue) *Element {
	return e.setValue("padding-block-start", v)
}
func (expression) isPaddingBlockStart()  {}
func (wideKeyword) isPaddingBlockStart() {}
func (Length) isPaddingBlockStart()      {}
func (Percentage) isPaddingBlockStart()  {}

// PaddingBottomValue is a value for the CSS padding-bottom property.
type PaddingBottomValue interface {
	cssValue
	isPaddingBottom()
}

// PaddingBottom sets the CSS padding-bottom property. Use RawValue for complex CSS syntax.
func (e *Element) PaddingBottom(v PaddingBottomValue) *Element {
	return e.setValue("padding-bottom", v)
}
func (expression) isPaddingBottom()  {}
func (wideKeyword) isPaddingBottom() {}
func (Length) isPaddingBottom()      {}
func (Percentage) isPaddingBottom()  {}

// PaddingInlineValue is a value for the CSS padding-inline property.
type PaddingInlineValue interface {
	cssValue
	isPaddingInline()
}

// PaddingInline sets the CSS padding-inline property. Use RawValue for complex CSS syntax.
func (e *Element) PaddingInline(v PaddingInlineValue) *Element {
	return e.setValue("padding-inline", v)
}
func (expression) isPaddingInline()  {}
func (wideKeyword) isPaddingInline() {}
func (Length) isPaddingInline()      {}
func (Percentage) isPaddingInline()  {}
func (Spacing) isPaddingInline()     {}

// PaddingInlineEndValue is a value for the CSS padding-inline-end property.
type PaddingInlineEndValue interface {
	cssValue
	isPaddingInlineEnd()
}

// PaddingInlineEnd sets the CSS padding-inline-end property. Use RawValue for complex CSS syntax.
func (e *Element) PaddingInlineEnd(v PaddingInlineEndValue) *Element {
	return e.setValue("padding-inline-end", v)
}
func (expression) isPaddingInlineEnd()  {}
func (wideKeyword) isPaddingInlineEnd() {}
func (Length) isPaddingInlineEnd()      {}
func (Percentage) isPaddingInlineEnd()  {}

// PaddingInlineStartValue is a value for the CSS padding-inline-start property.
type PaddingInlineStartValue interface {
	cssValue
	isPaddingInlineStart()
}

// PaddingInlineStart sets the CSS padding-inline-start property. Use RawValue for complex CSS syntax.
func (e *Element) PaddingInlineStart(v PaddingInlineStartValue) *Element {
	return e.setValue("padding-inline-start", v)
}
func (expression) isPaddingInlineStart()  {}
func (wideKeyword) isPaddingInlineStart() {}
func (Length) isPaddingInlineStart()      {}
func (Percentage) isPaddingInlineStart()  {}

// PaddingLeftValue is a value for the CSS padding-left property.
type PaddingLeftValue interface {
	cssValue
	isPaddingLeft()
}

// PaddingLeft sets the CSS padding-left property. Use RawValue for complex CSS syntax.
func (e *Element) PaddingLeft(v PaddingLeftValue) *Element { return e.setValue("padding-left", v) }
func (expression) isPaddingLeft()                          {}
func (wideKeyword) isPaddingLeft()                         {}
func (Length) isPaddingLeft()                              {}
func (Percentage) isPaddingLeft()                          {}

// PaddingRightValue is a value for the CSS padding-right property.
type PaddingRightValue interface {
	cssValue
	isPaddingRight()
}

// PaddingRight sets the CSS padding-right property. Use RawValue for complex CSS syntax.
func (e *Element) PaddingRight(v PaddingRightValue) *Element { return e.setValue("padding-right", v) }
func (expression) isPaddingRight()                           {}
func (wideKeyword) isPaddingRight()                          {}
func (Length) isPaddingRight()                               {}
func (Percentage) isPaddingRight()                           {}

// PaddingTopValue is a value for the CSS padding-top property.
type PaddingTopValue interface {
	cssValue
	isPaddingTop()
}

// PaddingTop sets the CSS padding-top property. Use RawValue for complex CSS syntax.
func (e *Element) PaddingTop(v PaddingTopValue) *Element { return e.setValue("padding-top", v) }
func (expression) isPaddingTop()                         {}
func (wideKeyword) isPaddingTop()                        {}
func (Length) isPaddingTop()                             {}
func (Percentage) isPaddingTop()                         {}

// PerspectiveValue is a value for the CSS perspective property.
type PerspectiveValue interface {
	cssValue
	isPerspective()
}

// Perspective sets the CSS perspective property. Use RawValue for complex CSS syntax.
func (e *Element) Perspective(v PerspectiveValue) *Element { return e.setValue("perspective", v) }
func (expression) isPerspective()                          {}
func (wideKeyword) isPerspective()                         {}
func (Length) isPerspective()                              {}

// PerspectiveOriginValue is a value for the CSS perspective-origin property.
type PerspectiveOriginValue interface {
	cssValue
	isPerspectiveOrigin()
}

// PerspectiveOrigin sets the CSS perspective-origin property. Use RawValue for complex CSS syntax.
func (e *Element) PerspectiveOrigin(v PerspectiveOriginValue) *Element {
	return e.setValue("perspective-origin", v)
}
func (expression) isPerspectiveOrigin()  {}
func (wideKeyword) isPerspectiveOrigin() {}

// PlaceContentValue is a value for the CSS place-content property.
type PlaceContentValue interface {
	cssValue
	isPlaceContent()
}

// PlaceContent sets the CSS place-content property. Use RawValue for complex CSS syntax.
func (e *Element) PlaceContent(v PlaceContentValue) *Element { return e.setValue("place-content", v) }
func (expression) isPlaceContent()                           {}
func (wideKeyword) isPlaceContent()                          {}

// PlaceItemsValue is a value for the CSS place-items property.
type PlaceItemsValue interface {
	cssValue
	isPlaceItems()
}

// PlaceItems sets the CSS place-items property. Use RawValue for complex CSS syntax.
func (e *Element) PlaceItems(v PlaceItemsValue) *Element { return e.setValue("place-items", v) }
func (expression) isPlaceItems()                         {}
func (wideKeyword) isPlaceItems()                        {}

// PlaceSelfValue is a value for the CSS place-self property.
type PlaceSelfValue interface {
	cssValue
	isPlaceSelf()
}

// PlaceSelf sets the CSS place-self property. Use RawValue for complex CSS syntax.
func (e *Element) PlaceSelf(v PlaceSelfValue) *Element { return e.setValue("place-self", v) }
func (expression) isPlaceSelf()                        {}
func (wideKeyword) isPlaceSelf()                       {}

// PositionValue is a value for the CSS position property.
type PositionValue interface {
	cssValue
	isPosition()
}

// Position sets the CSS position property. Use RawValue for complex CSS syntax.
func (e *Element) Position(v PositionValue) *Element { return e.setValue("position", v) }
func (expression) isPosition()                       {}
func (wideKeyword) isPosition()                      {}

type positionKeyword uint16

const (
	// PositionStatic sets position to static.
	PositionStatic positionKeyword = 1
	// PositionRelative sets position to relative.
	PositionRelative positionKeyword = 2
	// PositionAbsolute sets position to absolute.
	PositionAbsolute positionKeyword = 3
	// PositionFixed sets position to fixed.
	PositionFixed positionKeyword = 4
	// PositionSticky sets position to sticky.
	PositionSticky positionKeyword = 5
)

func (positionKeyword) isPosition() {}
func (v positionKeyword) cssText() (string, error) {
	switch v {
	case PositionStatic:
		return "static", nil
	case PositionRelative:
		return "relative", nil
	case PositionAbsolute:
		return "absolute", nil
	case PositionFixed:
		return "fixed", nil
	case PositionSticky:
		return "sticky", nil
	default:
		return "", fmt.Errorf("ui: invalid position keyword %d", v)
	}
}

// PositionAnchorValue is a value for the CSS position-anchor property.
type PositionAnchorValue interface {
	cssValue
	isPositionAnchor()
}

// PositionAnchor sets the CSS position-anchor property. Use RawValue for complex CSS syntax.
func (e *Element) PositionAnchor(v PositionAnchorValue) *Element {
	return e.setValue("position-anchor", v)
}
func (expression) isPositionAnchor()  {}
func (wideKeyword) isPositionAnchor() {}

// PositionAreaValue is a value for the CSS position-area property.
type PositionAreaValue interface {
	cssValue
	isPositionArea()
}

// PositionArea sets the CSS position-area property. Use RawValue for complex CSS syntax.
func (e *Element) PositionArea(v PositionAreaValue) *Element { return e.setValue("position-area", v) }
func (expression) isPositionArea()                           {}
func (wideKeyword) isPositionArea()                          {}

// PositionTryValue is a value for the CSS position-try property.
type PositionTryValue interface {
	cssValue
	isPositionTry()
}

// PositionTry sets the CSS position-try property. Use RawValue for complex CSS syntax.
func (e *Element) PositionTry(v PositionTryValue) *Element { return e.setValue("position-try", v) }
func (expression) isPositionTry()                          {}
func (wideKeyword) isPositionTry()                         {}

// PositionTryFallbacksValue is a value for the CSS position-try-fallbacks property.
type PositionTryFallbacksValue interface {
	cssValue
	isPositionTryFallbacks()
}

// PositionTryFallbacks sets the CSS position-try-fallbacks property. Use RawValue for complex CSS syntax.
func (e *Element) PositionTryFallbacks(v PositionTryFallbacksValue) *Element {
	return e.setValue("position-try-fallbacks", v)
}
func (expression) isPositionTryFallbacks()  {}
func (wideKeyword) isPositionTryFallbacks() {}

// PositionTryOrderValue is a value for the CSS position-try-order property.
type PositionTryOrderValue interface {
	cssValue
	isPositionTryOrder()
}

// PositionTryOrder sets the CSS position-try-order property. Use RawValue for complex CSS syntax.
func (e *Element) PositionTryOrder(v PositionTryOrderValue) *Element {
	return e.setValue("position-try-order", v)
}
func (expression) isPositionTryOrder()  {}
func (wideKeyword) isPositionTryOrder() {}

// PositionVisibilityValue is a value for the CSS position-visibility property.
type PositionVisibilityValue interface {
	cssValue
	isPositionVisibility()
}

// PositionVisibility sets the CSS position-visibility property. Use RawValue for complex CSS syntax.
func (e *Element) PositionVisibility(v PositionVisibilityValue) *Element {
	return e.setValue("position-visibility", v)
}
func (expression) isPositionVisibility()  {}
func (wideKeyword) isPositionVisibility() {}

// PrintColorAdjustValue is a value for the CSS print-color-adjust property.
type PrintColorAdjustValue interface {
	cssValue
	isPrintColorAdjust()
}

// PrintColorAdjust sets the CSS print-color-adjust property. Use RawValue for complex CSS syntax.
func (e *Element) PrintColorAdjust(v PrintColorAdjustValue) *Element {
	return e.setValue("print-color-adjust", v)
}
func (expression) isPrintColorAdjust()  {}
func (wideKeyword) isPrintColorAdjust() {}

// QuotesValue is a value for the CSS quotes property.
type QuotesValue interface {
	cssValue
	isQuotes()
}

// Quotes sets the CSS quotes property. Use RawValue for complex CSS syntax.
func (e *Element) Quotes(v QuotesValue) *Element { return e.setValue("quotes", v) }
func (expression) isQuotes()                     {}
func (wideKeyword) isQuotes()                    {}

// RegionFragmentValue is a value for the CSS region-fragment property.
type RegionFragmentValue interface {
	cssValue
	isRegionFragment()
}

// RegionFragment sets the CSS region-fragment property. Use RawValue for complex CSS syntax.
func (e *Element) RegionFragment(v RegionFragmentValue) *Element {
	return e.setValue("region-fragment", v)
}
func (expression) isRegionFragment()  {}
func (wideKeyword) isRegionFragment() {}

// RightValue is a value for the CSS right property.
type RightValue interface {
	cssValue
	isRight()
}

// Right sets the CSS right property. Use RawValue for complex CSS syntax.
func (e *Element) Right(v RightValue) *Element { return e.setValue("right", v) }
func (expression) isRight()                    {}
func (wideKeyword) isRight()                   {}
func (Length) isRight()                        {}
func (Percentage) isRight()                    {}

// RotateValue is a value for the CSS rotate property.
type RotateValue interface {
	cssValue
	isRotate()
}

// Rotate sets the CSS rotate property. Use RawValue for complex CSS syntax.
func (e *Element) Rotate(v RotateValue) *Element { return e.setValue("rotate", v) }
func (expression) isRotate()                     {}
func (wideKeyword) isRotate()                    {}

// RowGapValue is a value for the CSS row-gap property.
type RowGapValue interface {
	cssValue
	isRowGap()
}

// RowGap sets the CSS row-gap property. Use RawValue for complex CSS syntax.
func (e *Element) RowGap(v RowGapValue) *Element { return e.setValue("row-gap", v) }
func (expression) isRowGap()                     {}
func (wideKeyword) isRowGap()                    {}
func (Length) isRowGap()                         {}
func (Percentage) isRowGap()                     {}

// RowRuleValue is a value for the CSS row-rule property.
type RowRuleValue interface {
	cssValue
	isRowRule()
}

// RowRule sets the CSS row-rule property. Use RawValue for complex CSS syntax.
func (e *Element) RowRule(v RowRuleValue) *Element { return e.setValue("row-rule", v) }
func (expression) isRowRule()                      {}
func (wideKeyword) isRowRule()                     {}

// RowRuleBreakValue is a value for the CSS row-rule-break property.
type RowRuleBreakValue interface {
	cssValue
	isRowRuleBreak()
}

// RowRuleBreak sets the CSS row-rule-break property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleBreak(v RowRuleBreakValue) *Element { return e.setValue("row-rule-break", v) }
func (expression) isRowRuleBreak()                           {}
func (wideKeyword) isRowRuleBreak()                          {}

// RowRuleColorValue is a value for the CSS row-rule-color property.
type RowRuleColorValue interface {
	cssValue
	isRowRuleColor()
}

// RowRuleColor sets the CSS row-rule-color property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleColor(v RowRuleColorValue) *Element { return e.setValue("row-rule-color", v) }
func (expression) isRowRuleColor()                           {}
func (wideKeyword) isRowRuleColor()                          {}

// RowRuleInsetValue is a value for the CSS row-rule-inset property.
type RowRuleInsetValue interface {
	cssValue
	isRowRuleInset()
}

// RowRuleInset sets the CSS row-rule-inset property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleInset(v RowRuleInsetValue) *Element { return e.setValue("row-rule-inset", v) }
func (expression) isRowRuleInset()                           {}
func (wideKeyword) isRowRuleInset()                          {}

// RowRuleInsetCapValue is a value for the CSS row-rule-inset-cap property.
type RowRuleInsetCapValue interface {
	cssValue
	isRowRuleInsetCap()
}

// RowRuleInsetCap sets the CSS row-rule-inset-cap property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleInsetCap(v RowRuleInsetCapValue) *Element {
	return e.setValue("row-rule-inset-cap", v)
}
func (expression) isRowRuleInsetCap()  {}
func (wideKeyword) isRowRuleInsetCap() {}

// RowRuleInsetCapEndValue is a value for the CSS row-rule-inset-cap-end property.
type RowRuleInsetCapEndValue interface {
	cssValue
	isRowRuleInsetCapEnd()
}

// RowRuleInsetCapEnd sets the CSS row-rule-inset-cap-end property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleInsetCapEnd(v RowRuleInsetCapEndValue) *Element {
	return e.setValue("row-rule-inset-cap-end", v)
}
func (expression) isRowRuleInsetCapEnd()  {}
func (wideKeyword) isRowRuleInsetCapEnd() {}

// RowRuleInsetCapStartValue is a value for the CSS row-rule-inset-cap-start property.
type RowRuleInsetCapStartValue interface {
	cssValue
	isRowRuleInsetCapStart()
}

// RowRuleInsetCapStart sets the CSS row-rule-inset-cap-start property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleInsetCapStart(v RowRuleInsetCapStartValue) *Element {
	return e.setValue("row-rule-inset-cap-start", v)
}
func (expression) isRowRuleInsetCapStart()  {}
func (wideKeyword) isRowRuleInsetCapStart() {}

// RowRuleInsetEndValue is a value for the CSS row-rule-inset-end property.
type RowRuleInsetEndValue interface {
	cssValue
	isRowRuleInsetEnd()
}

// RowRuleInsetEnd sets the CSS row-rule-inset-end property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleInsetEnd(v RowRuleInsetEndValue) *Element {
	return e.setValue("row-rule-inset-end", v)
}
func (expression) isRowRuleInsetEnd()  {}
func (wideKeyword) isRowRuleInsetEnd() {}

// RowRuleInsetJunctionValue is a value for the CSS row-rule-inset-junction property.
type RowRuleInsetJunctionValue interface {
	cssValue
	isRowRuleInsetJunction()
}

// RowRuleInsetJunction sets the CSS row-rule-inset-junction property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleInsetJunction(v RowRuleInsetJunctionValue) *Element {
	return e.setValue("row-rule-inset-junction", v)
}
func (expression) isRowRuleInsetJunction()  {}
func (wideKeyword) isRowRuleInsetJunction() {}

// RowRuleInsetJunctionEndValue is a value for the CSS row-rule-inset-junction-end property.
type RowRuleInsetJunctionEndValue interface {
	cssValue
	isRowRuleInsetJunctionEnd()
}

// RowRuleInsetJunctionEnd sets the CSS row-rule-inset-junction-end property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleInsetJunctionEnd(v RowRuleInsetJunctionEndValue) *Element {
	return e.setValue("row-rule-inset-junction-end", v)
}
func (expression) isRowRuleInsetJunctionEnd()  {}
func (wideKeyword) isRowRuleInsetJunctionEnd() {}

// RowRuleInsetJunctionStartValue is a value for the CSS row-rule-inset-junction-start property.
type RowRuleInsetJunctionStartValue interface {
	cssValue
	isRowRuleInsetJunctionStart()
}

// RowRuleInsetJunctionStart sets the CSS row-rule-inset-junction-start property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleInsetJunctionStart(v RowRuleInsetJunctionStartValue) *Element {
	return e.setValue("row-rule-inset-junction-start", v)
}
func (expression) isRowRuleInsetJunctionStart()  {}
func (wideKeyword) isRowRuleInsetJunctionStart() {}

// RowRuleInsetStartValue is a value for the CSS row-rule-inset-start property.
type RowRuleInsetStartValue interface {
	cssValue
	isRowRuleInsetStart()
}

// RowRuleInsetStart sets the CSS row-rule-inset-start property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleInsetStart(v RowRuleInsetStartValue) *Element {
	return e.setValue("row-rule-inset-start", v)
}
func (expression) isRowRuleInsetStart()  {}
func (wideKeyword) isRowRuleInsetStart() {}

// RowRuleStyleValue is a value for the CSS row-rule-style property.
type RowRuleStyleValue interface {
	cssValue
	isRowRuleStyle()
}

// RowRuleStyle sets the CSS row-rule-style property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleStyle(v RowRuleStyleValue) *Element { return e.setValue("row-rule-style", v) }
func (expression) isRowRuleStyle()                           {}
func (wideKeyword) isRowRuleStyle()                          {}

// RowRuleVisibilityItemsValue is a value for the CSS row-rule-visibility-items property.
type RowRuleVisibilityItemsValue interface {
	cssValue
	isRowRuleVisibilityItems()
}

// RowRuleVisibilityItems sets the CSS row-rule-visibility-items property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleVisibilityItems(v RowRuleVisibilityItemsValue) *Element {
	return e.setValue("row-rule-visibility-items", v)
}
func (expression) isRowRuleVisibilityItems()  {}
func (wideKeyword) isRowRuleVisibilityItems() {}

// RowRuleWidthValue is a value for the CSS row-rule-width property.
type RowRuleWidthValue interface {
	cssValue
	isRowRuleWidth()
}

// RowRuleWidth sets the CSS row-rule-width property. Use RawValue for complex CSS syntax.
func (e *Element) RowRuleWidth(v RowRuleWidthValue) *Element { return e.setValue("row-rule-width", v) }
func (expression) isRowRuleWidth()                           {}
func (wideKeyword) isRowRuleWidth()                          {}

// RubyAlignValue is a value for the CSS ruby-align property.
type RubyAlignValue interface {
	cssValue
	isRubyAlign()
}

// RubyAlign sets the CSS ruby-align property. Use RawValue for complex CSS syntax.
func (e *Element) RubyAlign(v RubyAlignValue) *Element { return e.setValue("ruby-align", v) }
func (expression) isRubyAlign()                        {}
func (wideKeyword) isRubyAlign()                       {}

// RubyMergeValue is a value for the CSS ruby-merge property.
type RubyMergeValue interface {
	cssValue
	isRubyMerge()
}

// RubyMerge sets the CSS ruby-merge property. Use RawValue for complex CSS syntax.
func (e *Element) RubyMerge(v RubyMergeValue) *Element { return e.setValue("ruby-merge", v) }
func (expression) isRubyMerge()                        {}
func (wideKeyword) isRubyMerge()                       {}

// RubyOverhangValue is a value for the CSS ruby-overhang property.
type RubyOverhangValue interface {
	cssValue
	isRubyOverhang()
}

// RubyOverhang sets the CSS ruby-overhang property. Use RawValue for complex CSS syntax.
func (e *Element) RubyOverhang(v RubyOverhangValue) *Element { return e.setValue("ruby-overhang", v) }
func (expression) isRubyOverhang()                           {}
func (wideKeyword) isRubyOverhang()                          {}

// RubyPositionValue is a value for the CSS ruby-position property.
type RubyPositionValue interface {
	cssValue
	isRubyPosition()
}

// RubyPosition sets the CSS ruby-position property. Use RawValue for complex CSS syntax.
func (e *Element) RubyPosition(v RubyPositionValue) *Element { return e.setValue("ruby-position", v) }
func (expression) isRubyPosition()                           {}
func (wideKeyword) isRubyPosition()                          {}

// RuleValue is a value for the CSS rule property.
type RuleValue interface {
	cssValue
	isRule()
}

// Rule sets the CSS rule property. Use RawValue for complex CSS syntax.
func (e *Element) Rule(v RuleValue) *Element { return e.setValue("rule", v) }
func (expression) isRule()                   {}
func (wideKeyword) isRule()                  {}

// RuleBreakValue is a value for the CSS rule-break property.
type RuleBreakValue interface {
	cssValue
	isRuleBreak()
}

// RuleBreak sets the CSS rule-break property. Use RawValue for complex CSS syntax.
func (e *Element) RuleBreak(v RuleBreakValue) *Element { return e.setValue("rule-break", v) }
func (expression) isRuleBreak()                        {}
func (wideKeyword) isRuleBreak()                       {}

// RuleColorValue is a value for the CSS rule-color property.
type RuleColorValue interface {
	cssValue
	isRuleColor()
}

// RuleColor sets the CSS rule-color property. Use RawValue for complex CSS syntax.
func (e *Element) RuleColor(v RuleColorValue) *Element { return e.setValue("rule-color", v) }
func (expression) isRuleColor()                        {}
func (wideKeyword) isRuleColor()                       {}

// RuleInsetValue is a value for the CSS rule-inset property.
type RuleInsetValue interface {
	cssValue
	isRuleInset()
}

// RuleInset sets the CSS rule-inset property. Use RawValue for complex CSS syntax.
func (e *Element) RuleInset(v RuleInsetValue) *Element { return e.setValue("rule-inset", v) }
func (expression) isRuleInset()                        {}
func (wideKeyword) isRuleInset()                       {}

// RuleInsetCapValue is a value for the CSS rule-inset-cap property.
type RuleInsetCapValue interface {
	cssValue
	isRuleInsetCap()
}

// RuleInsetCap sets the CSS rule-inset-cap property. Use RawValue for complex CSS syntax.
func (e *Element) RuleInsetCap(v RuleInsetCapValue) *Element { return e.setValue("rule-inset-cap", v) }
func (expression) isRuleInsetCap()                           {}
func (wideKeyword) isRuleInsetCap()                          {}

// RuleInsetEndValue is a value for the CSS rule-inset-end property.
type RuleInsetEndValue interface {
	cssValue
	isRuleInsetEnd()
}

// RuleInsetEnd sets the CSS rule-inset-end property. Use RawValue for complex CSS syntax.
func (e *Element) RuleInsetEnd(v RuleInsetEndValue) *Element { return e.setValue("rule-inset-end", v) }
func (expression) isRuleInsetEnd()                           {}
func (wideKeyword) isRuleInsetEnd()                          {}

// RuleInsetJunctionValue is a value for the CSS rule-inset-junction property.
type RuleInsetJunctionValue interface {
	cssValue
	isRuleInsetJunction()
}

// RuleInsetJunction sets the CSS rule-inset-junction property. Use RawValue for complex CSS syntax.
func (e *Element) RuleInsetJunction(v RuleInsetJunctionValue) *Element {
	return e.setValue("rule-inset-junction", v)
}
func (expression) isRuleInsetJunction()  {}
func (wideKeyword) isRuleInsetJunction() {}

// RuleInsetStartValue is a value for the CSS rule-inset-start property.
type RuleInsetStartValue interface {
	cssValue
	isRuleInsetStart()
}

// RuleInsetStart sets the CSS rule-inset-start property. Use RawValue for complex CSS syntax.
func (e *Element) RuleInsetStart(v RuleInsetStartValue) *Element {
	return e.setValue("rule-inset-start", v)
}
func (expression) isRuleInsetStart()  {}
func (wideKeyword) isRuleInsetStart() {}

// RuleOverlapValue is a value for the CSS rule-overlap property.
type RuleOverlapValue interface {
	cssValue
	isRuleOverlap()
}

// RuleOverlap sets the CSS rule-overlap property. Use RawValue for complex CSS syntax.
func (e *Element) RuleOverlap(v RuleOverlapValue) *Element { return e.setValue("rule-overlap", v) }
func (expression) isRuleOverlap()                          {}
func (wideKeyword) isRuleOverlap()                         {}

// RuleStyleValue is a value for the CSS rule-style property.
type RuleStyleValue interface {
	cssValue
	isRuleStyle()
}

// RuleStyle sets the CSS rule-style property. Use RawValue for complex CSS syntax.
func (e *Element) RuleStyle(v RuleStyleValue) *Element { return e.setValue("rule-style", v) }
func (expression) isRuleStyle()                        {}
func (wideKeyword) isRuleStyle()                       {}

// RuleVisibilityItemsValue is a value for the CSS rule-visibility-items property.
type RuleVisibilityItemsValue interface {
	cssValue
	isRuleVisibilityItems()
}

// RuleVisibilityItems sets the CSS rule-visibility-items property. Use RawValue for complex CSS syntax.
func (e *Element) RuleVisibilityItems(v RuleVisibilityItemsValue) *Element {
	return e.setValue("rule-visibility-items", v)
}
func (expression) isRuleVisibilityItems()  {}
func (wideKeyword) isRuleVisibilityItems() {}

// RuleWidthValue is a value for the CSS rule-width property.
type RuleWidthValue interface {
	cssValue
	isRuleWidth()
}

// RuleWidth sets the CSS rule-width property. Use RawValue for complex CSS syntax.
func (e *Element) RuleWidth(v RuleWidthValue) *Element { return e.setValue("rule-width", v) }
func (expression) isRuleWidth()                        {}
func (wideKeyword) isRuleWidth()                       {}

// ScaleValue is a value for the CSS scale property.
type ScaleValue interface {
	cssValue
	isScale()
}

// Scale sets the CSS scale property. Use RawValue for complex CSS syntax.
func (e *Element) Scale(v ScaleValue) *Element { return e.setValue("scale", v) }
func (expression) isScale()                    {}
func (wideKeyword) isScale()                   {}

// ScrollMarginValue is a value for the CSS scroll-margin property.
type ScrollMarginValue interface {
	cssValue
	isScrollMargin()
}

// ScrollMargin sets the CSS scroll-margin property. Use RawValue for complex CSS syntax.
func (e *Element) ScrollMargin(v ScrollMarginValue) *Element { return e.setValue("scroll-margin", v) }
func (expression) isScrollMargin()                           {}
func (wideKeyword) isScrollMargin()                          {}

// ScrollMarginBottomValue is a value for the CSS scroll-margin-bottom property.
type ScrollMarginBottomValue interface {
	cssValue
	isScrollMarginBottom()
}

// ScrollMarginBottom sets the CSS scroll-margin-bottom property. Use RawValue for complex CSS syntax.
func (e *Element) ScrollMarginBottom(v ScrollMarginBottomValue) *Element {
	return e.setValue("scroll-margin-bottom", v)
}
func (expression) isScrollMarginBottom()  {}
func (wideKeyword) isScrollMarginBottom() {}

// ScrollMarginLeftValue is a value for the CSS scroll-margin-left property.
type ScrollMarginLeftValue interface {
	cssValue
	isScrollMarginLeft()
}

// ScrollMarginLeft sets the CSS scroll-margin-left property. Use RawValue for complex CSS syntax.
func (e *Element) ScrollMarginLeft(v ScrollMarginLeftValue) *Element {
	return e.setValue("scroll-margin-left", v)
}
func (expression) isScrollMarginLeft()  {}
func (wideKeyword) isScrollMarginLeft() {}

// ScrollMarginRightValue is a value for the CSS scroll-margin-right property.
type ScrollMarginRightValue interface {
	cssValue
	isScrollMarginRight()
}

// ScrollMarginRight sets the CSS scroll-margin-right property. Use RawValue for complex CSS syntax.
func (e *Element) ScrollMarginRight(v ScrollMarginRightValue) *Element {
	return e.setValue("scroll-margin-right", v)
}
func (expression) isScrollMarginRight()  {}
func (wideKeyword) isScrollMarginRight() {}

// ScrollMarginTopValue is a value for the CSS scroll-margin-top property.
type ScrollMarginTopValue interface {
	cssValue
	isScrollMarginTop()
}

// ScrollMarginTop sets the CSS scroll-margin-top property. Use RawValue for complex CSS syntax.
func (e *Element) ScrollMarginTop(v ScrollMarginTopValue) *Element {
	return e.setValue("scroll-margin-top", v)
}
func (expression) isScrollMarginTop()  {}
func (wideKeyword) isScrollMarginTop() {}

// ShapeInsideValue is a value for the CSS shape-inside property.
type ShapeInsideValue interface {
	cssValue
	isShapeInside()
}

// ShapeInside sets the CSS shape-inside property. Use RawValue for complex CSS syntax.
func (e *Element) ShapeInside(v ShapeInsideValue) *Element { return e.setValue("shape-inside", v) }
func (expression) isShapeInside()                          {}
func (wideKeyword) isShapeInside()                         {}

// ShapeMarginValue is a value for the CSS shape-margin property.
type ShapeMarginValue interface {
	cssValue
	isShapeMargin()
}

// ShapeMargin sets the CSS shape-margin property. Use RawValue for complex CSS syntax.
func (e *Element) ShapeMargin(v ShapeMarginValue) *Element { return e.setValue("shape-margin", v) }
func (expression) isShapeMargin()                          {}
func (wideKeyword) isShapeMargin()                         {}

// ShapeOutsideValue is a value for the CSS shape-outside property.
type ShapeOutsideValue interface {
	cssValue
	isShapeOutside()
}

// ShapeOutside sets the CSS shape-outside property. Use RawValue for complex CSS syntax.
func (e *Element) ShapeOutside(v ShapeOutsideValue) *Element { return e.setValue("shape-outside", v) }
func (expression) isShapeOutside()                           {}
func (wideKeyword) isShapeOutside()                          {}

// ShapePaddingValue is a value for the CSS shape-padding property.
type ShapePaddingValue interface {
	cssValue
	isShapePadding()
}

// ShapePadding sets the CSS shape-padding property. Use RawValue for complex CSS syntax.
func (e *Element) ShapePadding(v ShapePaddingValue) *Element { return e.setValue("shape-padding", v) }
func (expression) isShapePadding()                           {}
func (wideKeyword) isShapePadding()                          {}

// ShapeRenderingValue is a value for the CSS shape-rendering property.
type ShapeRenderingValue interface {
	cssValue
	isShapeRendering()
}

// ShapeRendering sets the CSS shape-rendering property. Use RawValue for complex CSS syntax.
func (e *Element) ShapeRendering(v ShapeRenderingValue) *Element {
	return e.setValue("shape-rendering", v)
}
func (expression) isShapeRendering()  {}
func (wideKeyword) isShapeRendering() {}

// StrokeValue is a value for the CSS stroke property.
type StrokeValue interface {
	cssValue
	isStroke()
}

// Stroke sets the CSS stroke property. Use RawValue for complex CSS syntax.
func (e *Element) Stroke(v StrokeValue) *Element { return e.setValue("stroke", v) }
func (expression) isStroke()                     {}
func (wideKeyword) isStroke()                    {}
func (Color) isStroke()                          {}

// StrokeDasharrayValue is a value for the CSS stroke-dasharray property.
type StrokeDasharrayValue interface {
	cssValue
	isStrokeDasharray()
}

// StrokeDasharray sets the CSS stroke-dasharray property. Use RawValue for complex CSS syntax.
func (e *Element) StrokeDasharray(v StrokeDasharrayValue) *Element {
	return e.setValue("stroke-dasharray", v)
}
func (expression) isStrokeDasharray()  {}
func (wideKeyword) isStrokeDasharray() {}

// StrokeDashoffsetValue is a value for the CSS stroke-dashoffset property.
type StrokeDashoffsetValue interface {
	cssValue
	isStrokeDashoffset()
}

// StrokeDashoffset sets the CSS stroke-dashoffset property. Use RawValue for complex CSS syntax.
func (e *Element) StrokeDashoffset(v StrokeDashoffsetValue) *Element {
	return e.setValue("stroke-dashoffset", v)
}
func (expression) isStrokeDashoffset()  {}
func (wideKeyword) isStrokeDashoffset() {}

// StrokeLinecapValue is a value for the CSS stroke-linecap property.
type StrokeLinecapValue interface {
	cssValue
	isStrokeLinecap()
}

// StrokeLinecap sets the CSS stroke-linecap property. Use RawValue for complex CSS syntax.
func (e *Element) StrokeLinecap(v StrokeLinecapValue) *Element {
	return e.setValue("stroke-linecap", v)
}
func (expression) isStrokeLinecap()  {}
func (wideKeyword) isStrokeLinecap() {}

// StrokeLinejoinValue is a value for the CSS stroke-linejoin property.
type StrokeLinejoinValue interface {
	cssValue
	isStrokeLinejoin()
}

// StrokeLinejoin sets the CSS stroke-linejoin property. Use RawValue for complex CSS syntax.
func (e *Element) StrokeLinejoin(v StrokeLinejoinValue) *Element {
	return e.setValue("stroke-linejoin", v)
}
func (expression) isStrokeLinejoin()  {}
func (wideKeyword) isStrokeLinejoin() {}

// StrokeMiterlimitValue is a value for the CSS stroke-miterlimit property.
type StrokeMiterlimitValue interface {
	cssValue
	isStrokeMiterlimit()
}

// StrokeMiterlimit sets the CSS stroke-miterlimit property. Use RawValue for complex CSS syntax.
func (e *Element) StrokeMiterlimit(v StrokeMiterlimitValue) *Element {
	return e.setValue("stroke-miterlimit", v)
}
func (expression) isStrokeMiterlimit()  {}
func (wideKeyword) isStrokeMiterlimit() {}

// StrokeOpacityValue is a value for the CSS stroke-opacity property.
type StrokeOpacityValue interface {
	cssValue
	isStrokeOpacity()
}

// StrokeOpacity sets the CSS stroke-opacity property. Use RawValue for complex CSS syntax.
func (e *Element) StrokeOpacity(v StrokeOpacityValue) *Element {
	return e.setValue("stroke-opacity", v)
}
func (expression) isStrokeOpacity()  {}
func (wideKeyword) isStrokeOpacity() {}
func (Numeric) isStrokeOpacity()     {}

// StrokeWidthValue is a value for the CSS stroke-width property.
type StrokeWidthValue interface {
	cssValue
	isStrokeWidth()
}

// StrokeWidth sets the CSS stroke-width property. Use RawValue for complex CSS syntax.
func (e *Element) StrokeWidth(v StrokeWidthValue) *Element { return e.setValue("stroke-width", v) }
func (expression) isStrokeWidth()                          {}
func (wideKeyword) isStrokeWidth()                         {}
func (Length) isStrokeWidth()                              {}

// TabSizeValue is a value for the CSS tab-size property.
type TabSizeValue interface {
	cssValue
	isTabSize()
}

// TabSize sets the CSS tab-size property. Use RawValue for complex CSS syntax.
func (e *Element) TabSize(v TabSizeValue) *Element { return e.setValue("tab-size", v) }
func (expression) isTabSize()                      {}
func (wideKeyword) isTabSize()                     {}

// TableLayoutValue is a value for the CSS table-layout property.
type TableLayoutValue interface {
	cssValue
	isTableLayout()
}

// TableLayout sets the CSS table-layout property. Use RawValue for complex CSS syntax.
func (e *Element) TableLayout(v TableLayoutValue) *Element { return e.setValue("table-layout", v) }
func (expression) isTableLayout()                          {}
func (wideKeyword) isTableLayout()                         {}

// TextAlignValue is a value for the CSS text-align property.
type TextAlignValue interface {
	cssValue
	isTextAlign()
}

// TextAlign sets the CSS text-align property. Use RawValue for complex CSS syntax.
func (e *Element) TextAlign(v TextAlignValue) *Element { return e.setValue("text-align", v) }
func (expression) isTextAlign()                        {}
func (wideKeyword) isTextAlign()                       {}

type textAlignKeyword uint16

const (
	// TextAlignStart sets text-align to start.
	TextAlignStart textAlignKeyword = 1
	// TextAlignEnd sets text-align to end.
	TextAlignEnd textAlignKeyword = 2
	// TextAlignLeft sets text-align to left.
	TextAlignLeft textAlignKeyword = 3
	// TextAlignRight sets text-align to right.
	TextAlignRight textAlignKeyword = 4
	// TextAlignCenter sets text-align to center.
	TextAlignCenter textAlignKeyword = 5
	// TextAlignJustify sets text-align to justify.
	TextAlignJustify textAlignKeyword = 6
)

func (textAlignKeyword) isTextAlign() {}
func (v textAlignKeyword) cssText() (string, error) {
	switch v {
	case TextAlignStart:
		return "start", nil
	case TextAlignEnd:
		return "end", nil
	case TextAlignLeft:
		return "left", nil
	case TextAlignRight:
		return "right", nil
	case TextAlignCenter:
		return "center", nil
	case TextAlignJustify:
		return "justify", nil
	default:
		return "", fmt.Errorf("ui: invalid text-align keyword %d", v)
	}
}

// TextAlignAllValue is a value for the CSS text-align-all property.
type TextAlignAllValue interface {
	cssValue
	isTextAlignAll()
}

// TextAlignAll sets the CSS text-align-all property. Use RawValue for complex CSS syntax.
func (e *Element) TextAlignAll(v TextAlignAllValue) *Element { return e.setValue("text-align-all", v) }
func (expression) isTextAlignAll()                           {}
func (wideKeyword) isTextAlignAll()                          {}

// TextAlignLastValue is a value for the CSS text-align-last property.
type TextAlignLastValue interface {
	cssValue
	isTextAlignLast()
}

// TextAlignLast sets the CSS text-align-last property. Use RawValue for complex CSS syntax.
func (e *Element) TextAlignLast(v TextAlignLastValue) *Element {
	return e.setValue("text-align-last", v)
}
func (expression) isTextAlignLast()  {}
func (wideKeyword) isTextAlignLast() {}

// TextAnchorValue is a value for the CSS text-anchor property.
type TextAnchorValue interface {
	cssValue
	isTextAnchor()
}

// TextAnchor sets the CSS text-anchor property. Use RawValue for complex CSS syntax.
func (e *Element) TextAnchor(v TextAnchorValue) *Element { return e.setValue("text-anchor", v) }
func (expression) isTextAnchor()                         {}
func (wideKeyword) isTextAnchor()                        {}

// TextAutospaceValue is a value for the CSS text-autospace property.
type TextAutospaceValue interface {
	cssValue
	isTextAutospace()
}

// TextAutospace sets the CSS text-autospace property. Use RawValue for complex CSS syntax.
func (e *Element) TextAutospace(v TextAutospaceValue) *Element {
	return e.setValue("text-autospace", v)
}
func (expression) isTextAutospace()  {}
func (wideKeyword) isTextAutospace() {}

// TextBoxValue is a value for the CSS text-box property.
type TextBoxValue interface {
	cssValue
	isTextBox()
}

// TextBox sets the CSS text-box property. Use RawValue for complex CSS syntax.
func (e *Element) TextBox(v TextBoxValue) *Element { return e.setValue("text-box", v) }
func (expression) isTextBox()                      {}
func (wideKeyword) isTextBox()                     {}

// TextBoxEdgeValue is a value for the CSS text-box-edge property.
type TextBoxEdgeValue interface {
	cssValue
	isTextBoxEdge()
}

// TextBoxEdge sets the CSS text-box-edge property. Use RawValue for complex CSS syntax.
func (e *Element) TextBoxEdge(v TextBoxEdgeValue) *Element { return e.setValue("text-box-edge", v) }
func (expression) isTextBoxEdge()                          {}
func (wideKeyword) isTextBoxEdge()                         {}

// TextBoxTrimValue is a value for the CSS text-box-trim property.
type TextBoxTrimValue interface {
	cssValue
	isTextBoxTrim()
}

// TextBoxTrim sets the CSS text-box-trim property. Use RawValue for complex CSS syntax.
func (e *Element) TextBoxTrim(v TextBoxTrimValue) *Element { return e.setValue("text-box-trim", v) }
func (expression) isTextBoxTrim()                          {}
func (wideKeyword) isTextBoxTrim()                         {}

// TextCombineUprightValue is a value for the CSS text-combine-upright property.
type TextCombineUprightValue interface {
	cssValue
	isTextCombineUpright()
}

// TextCombineUpright sets the CSS text-combine-upright property. Use RawValue for complex CSS syntax.
func (e *Element) TextCombineUpright(v TextCombineUprightValue) *Element {
	return e.setValue("text-combine-upright", v)
}
func (expression) isTextCombineUpright()  {}
func (wideKeyword) isTextCombineUpright() {}

// TextDecorationValue is a value for the CSS text-decoration property.
type TextDecorationValue interface {
	cssValue
	isTextDecoration()
}

// TextDecoration sets the CSS text-decoration property. Use RawValue for complex CSS syntax.
func (e *Element) TextDecoration(v TextDecorationValue) *Element {
	return e.setValue("text-decoration", v)
}
func (expression) isTextDecoration()  {}
func (wideKeyword) isTextDecoration() {}

type textDecorationKeyword uint16

const (
	// TextDecorationNone sets text-decoration to none.
	TextDecorationNone textDecorationKeyword = 1
	// TextDecorationUnderline sets text-decoration to underline.
	TextDecorationUnderline textDecorationKeyword = 2
	// TextDecorationLineThrough sets text-decoration to line-through.
	TextDecorationLineThrough textDecorationKeyword = 3
)

func (textDecorationKeyword) isTextDecoration() {}
func (v textDecorationKeyword) cssText() (string, error) {
	switch v {
	case TextDecorationNone:
		return "none", nil
	case TextDecorationUnderline:
		return "underline", nil
	case TextDecorationLineThrough:
		return "line-through", nil
	default:
		return "", fmt.Errorf("ui: invalid text-decoration keyword %d", v)
	}
}

// TextDecorationColorValue is a value for the CSS text-decoration-color property.
type TextDecorationColorValue interface {
	cssValue
	isTextDecorationColor()
}

// TextDecorationColor sets the CSS text-decoration-color property. Use RawValue for complex CSS syntax.
func (e *Element) TextDecorationColor(v TextDecorationColorValue) *Element {
	return e.setValue("text-decoration-color", v)
}
func (expression) isTextDecorationColor()  {}
func (wideKeyword) isTextDecorationColor() {}
func (Color) isTextDecorationColor()       {}

// TextDecorationInsetValue is a value for the CSS text-decoration-inset property.
type TextDecorationInsetValue interface {
	cssValue
	isTextDecorationInset()
}

// TextDecorationInset sets the CSS text-decoration-inset property. Use RawValue for complex CSS syntax.
func (e *Element) TextDecorationInset(v TextDecorationInsetValue) *Element {
	return e.setValue("text-decoration-inset", v)
}
func (expression) isTextDecorationInset()  {}
func (wideKeyword) isTextDecorationInset() {}

// TextDecorationLineValue is a value for the CSS text-decoration-line property.
type TextDecorationLineValue interface {
	cssValue
	isTextDecorationLine()
}

// TextDecorationLine sets the CSS text-decoration-line property. Use RawValue for complex CSS syntax.
func (e *Element) TextDecorationLine(v TextDecorationLineValue) *Element {
	return e.setValue("text-decoration-line", v)
}
func (expression) isTextDecorationLine()  {}
func (wideKeyword) isTextDecorationLine() {}

// TextDecorationSkipValue is a value for the CSS text-decoration-skip property.
type TextDecorationSkipValue interface {
	cssValue
	isTextDecorationSkip()
}

// TextDecorationSkip sets the CSS text-decoration-skip property. Use RawValue for complex CSS syntax.
func (e *Element) TextDecorationSkip(v TextDecorationSkipValue) *Element {
	return e.setValue("text-decoration-skip", v)
}
func (expression) isTextDecorationSkip()  {}
func (wideKeyword) isTextDecorationSkip() {}

// TextDecorationSkipBoxValue is a value for the CSS text-decoration-skip-box property.
type TextDecorationSkipBoxValue interface {
	cssValue
	isTextDecorationSkipBox()
}

// TextDecorationSkipBox sets the CSS text-decoration-skip-box property. Use RawValue for complex CSS syntax.
func (e *Element) TextDecorationSkipBox(v TextDecorationSkipBoxValue) *Element {
	return e.setValue("text-decoration-skip-box", v)
}
func (expression) isTextDecorationSkipBox()  {}
func (wideKeyword) isTextDecorationSkipBox() {}

// TextDecorationSkipInkValue is a value for the CSS text-decoration-skip-ink property.
type TextDecorationSkipInkValue interface {
	cssValue
	isTextDecorationSkipInk()
}

// TextDecorationSkipInk sets the CSS text-decoration-skip-ink property. Use RawValue for complex CSS syntax.
func (e *Element) TextDecorationSkipInk(v TextDecorationSkipInkValue) *Element {
	return e.setValue("text-decoration-skip-ink", v)
}
func (expression) isTextDecorationSkipInk()  {}
func (wideKeyword) isTextDecorationSkipInk() {}

// TextDecorationSkipSelfValue is a value for the CSS text-decoration-skip-self property.
type TextDecorationSkipSelfValue interface {
	cssValue
	isTextDecorationSkipSelf()
}

// TextDecorationSkipSelf sets the CSS text-decoration-skip-self property. Use RawValue for complex CSS syntax.
func (e *Element) TextDecorationSkipSelf(v TextDecorationSkipSelfValue) *Element {
	return e.setValue("text-decoration-skip-self", v)
}
func (expression) isTextDecorationSkipSelf()  {}
func (wideKeyword) isTextDecorationSkipSelf() {}

// TextDecorationSkipSpacesValue is a value for the CSS text-decoration-skip-spaces property.
type TextDecorationSkipSpacesValue interface {
	cssValue
	isTextDecorationSkipSpaces()
}

// TextDecorationSkipSpaces sets the CSS text-decoration-skip-spaces property. Use RawValue for complex CSS syntax.
func (e *Element) TextDecorationSkipSpaces(v TextDecorationSkipSpacesValue) *Element {
	return e.setValue("text-decoration-skip-spaces", v)
}
func (expression) isTextDecorationSkipSpaces()  {}
func (wideKeyword) isTextDecorationSkipSpaces() {}

// TextDecorationStyleValue is a value for the CSS text-decoration-style property.
type TextDecorationStyleValue interface {
	cssValue
	isTextDecorationStyle()
}

// TextDecorationStyle sets the CSS text-decoration-style property. Use RawValue for complex CSS syntax.
func (e *Element) TextDecorationStyle(v TextDecorationStyleValue) *Element {
	return e.setValue("text-decoration-style", v)
}
func (expression) isTextDecorationStyle()  {}
func (wideKeyword) isTextDecorationStyle() {}

// TextDecorationThicknessValue is a value for the CSS text-decoration-thickness property.
type TextDecorationThicknessValue interface {
	cssValue
	isTextDecorationThickness()
}

// TextDecorationThickness sets the CSS text-decoration-thickness property. Use RawValue for complex CSS syntax.
func (e *Element) TextDecorationThickness(v TextDecorationThicknessValue) *Element {
	return e.setValue("text-decoration-thickness", v)
}
func (expression) isTextDecorationThickness()  {}
func (wideKeyword) isTextDecorationThickness() {}

// TextEmphasisValue is a value for the CSS text-emphasis property.
type TextEmphasisValue interface {
	cssValue
	isTextEmphasis()
}

// TextEmphasis sets the CSS text-emphasis property. Use RawValue for complex CSS syntax.
func (e *Element) TextEmphasis(v TextEmphasisValue) *Element { return e.setValue("text-emphasis", v) }
func (expression) isTextEmphasis()                           {}
func (wideKeyword) isTextEmphasis()                          {}

// TextEmphasisColorValue is a value for the CSS text-emphasis-color property.
type TextEmphasisColorValue interface {
	cssValue
	isTextEmphasisColor()
}

// TextEmphasisColor sets the CSS text-emphasis-color property. Use RawValue for complex CSS syntax.
func (e *Element) TextEmphasisColor(v TextEmphasisColorValue) *Element {
	return e.setValue("text-emphasis-color", v)
}
func (expression) isTextEmphasisColor()  {}
func (wideKeyword) isTextEmphasisColor() {}
func (Color) isTextEmphasisColor()       {}

// TextEmphasisPositionValue is a value for the CSS text-emphasis-position property.
type TextEmphasisPositionValue interface {
	cssValue
	isTextEmphasisPosition()
}

// TextEmphasisPosition sets the CSS text-emphasis-position property. Use RawValue for complex CSS syntax.
func (e *Element) TextEmphasisPosition(v TextEmphasisPositionValue) *Element {
	return e.setValue("text-emphasis-position", v)
}
func (expression) isTextEmphasisPosition()  {}
func (wideKeyword) isTextEmphasisPosition() {}

// TextEmphasisSkipValue is a value for the CSS text-emphasis-skip property.
type TextEmphasisSkipValue interface {
	cssValue
	isTextEmphasisSkip()
}

// TextEmphasisSkip sets the CSS text-emphasis-skip property. Use RawValue for complex CSS syntax.
func (e *Element) TextEmphasisSkip(v TextEmphasisSkipValue) *Element {
	return e.setValue("text-emphasis-skip", v)
}
func (expression) isTextEmphasisSkip()  {}
func (wideKeyword) isTextEmphasisSkip() {}

// TextEmphasisStyleValue is a value for the CSS text-emphasis-style property.
type TextEmphasisStyleValue interface {
	cssValue
	isTextEmphasisStyle()
}

// TextEmphasisStyle sets the CSS text-emphasis-style property. Use RawValue for complex CSS syntax.
func (e *Element) TextEmphasisStyle(v TextEmphasisStyleValue) *Element {
	return e.setValue("text-emphasis-style", v)
}
func (expression) isTextEmphasisStyle()  {}
func (wideKeyword) isTextEmphasisStyle() {}

// TextGroupAlignValue is a value for the CSS text-group-align property.
type TextGroupAlignValue interface {
	cssValue
	isTextGroupAlign()
}

// TextGroupAlign sets the CSS text-group-align property. Use RawValue for complex CSS syntax.
func (e *Element) TextGroupAlign(v TextGroupAlignValue) *Element {
	return e.setValue("text-group-align", v)
}
func (expression) isTextGroupAlign()  {}
func (wideKeyword) isTextGroupAlign() {}

// TextIndentValue is a value for the CSS text-indent property.
type TextIndentValue interface {
	cssValue
	isTextIndent()
}

// TextIndent sets the CSS text-indent property. Use RawValue for complex CSS syntax.
func (e *Element) TextIndent(v TextIndentValue) *Element { return e.setValue("text-indent", v) }
func (expression) isTextIndent()                         {}
func (wideKeyword) isTextIndent()                        {}
func (Length) isTextIndent()                             {}
func (Percentage) isTextIndent()                         {}

// TextJustifyValue is a value for the CSS text-justify property.
type TextJustifyValue interface {
	cssValue
	isTextJustify()
}

// TextJustify sets the CSS text-justify property. Use RawValue for complex CSS syntax.
func (e *Element) TextJustify(v TextJustifyValue) *Element { return e.setValue("text-justify", v) }
func (expression) isTextJustify()                          {}
func (wideKeyword) isTextJustify()                         {}

// TextOrientationValue is a value for the CSS text-orientation property.
type TextOrientationValue interface {
	cssValue
	isTextOrientation()
}

// TextOrientation sets the CSS text-orientation property. Use RawValue for complex CSS syntax.
func (e *Element) TextOrientation(v TextOrientationValue) *Element {
	return e.setValue("text-orientation", v)
}
func (expression) isTextOrientation()  {}
func (wideKeyword) isTextOrientation() {}

// TextOverflowValue is a value for the CSS text-overflow property.
type TextOverflowValue interface {
	cssValue
	isTextOverflow()
}

// TextOverflow sets the CSS text-overflow property. Use RawValue for complex CSS syntax.
func (e *Element) TextOverflow(v TextOverflowValue) *Element { return e.setValue("text-overflow", v) }
func (expression) isTextOverflow()                           {}
func (wideKeyword) isTextOverflow()                          {}

type textOverflowKeyword uint16

const (
	// TextOverflowClip sets text-overflow to clip.
	TextOverflowClip textOverflowKeyword = 1
	// TextOverflowEllipsis sets text-overflow to ellipsis.
	TextOverflowEllipsis textOverflowKeyword = 2
)

func (textOverflowKeyword) isTextOverflow() {}
func (v textOverflowKeyword) cssText() (string, error) {
	switch v {
	case TextOverflowClip:
		return "clip", nil
	case TextOverflowEllipsis:
		return "ellipsis", nil
	default:
		return "", fmt.Errorf("ui: invalid text-overflow keyword %d", v)
	}
}

// TextShadowValue is a value for the CSS text-shadow property.
type TextShadowValue interface {
	cssValue
	isTextShadow()
}

// TextShadow sets the CSS text-shadow property. Use RawValue for complex CSS syntax.
func (e *Element) TextShadow(v TextShadowValue) *Element { return e.setValue("text-shadow", v) }
func (expression) isTextShadow()                         {}
func (wideKeyword) isTextShadow()                        {}

// TextSpacingValue is a value for the CSS text-spacing property.
type TextSpacingValue interface {
	cssValue
	isTextSpacing()
}

// TextSpacing sets the CSS text-spacing property. Use RawValue for complex CSS syntax.
func (e *Element) TextSpacing(v TextSpacingValue) *Element { return e.setValue("text-spacing", v) }
func (expression) isTextSpacing()                          {}
func (wideKeyword) isTextSpacing()                         {}

// TextSpacingTrimValue is a value for the CSS text-spacing-trim property.
type TextSpacingTrimValue interface {
	cssValue
	isTextSpacingTrim()
}

// TextSpacingTrim sets the CSS text-spacing-trim property. Use RawValue for complex CSS syntax.
func (e *Element) TextSpacingTrim(v TextSpacingTrimValue) *Element {
	return e.setValue("text-spacing-trim", v)
}
func (expression) isTextSpacingTrim()  {}
func (wideKeyword) isTextSpacingTrim() {}

// TextTransformValue is a value for the CSS text-transform property.
type TextTransformValue interface {
	cssValue
	isTextTransform()
}

// TextTransform sets the CSS text-transform property. Use RawValue for complex CSS syntax.
func (e *Element) TextTransform(v TextTransformValue) *Element {
	return e.setValue("text-transform", v)
}
func (expression) isTextTransform()  {}
func (wideKeyword) isTextTransform() {}

// TextUnderlineOffsetValue is a value for the CSS text-underline-offset property.
type TextUnderlineOffsetValue interface {
	cssValue
	isTextUnderlineOffset()
}

// TextUnderlineOffset sets the CSS text-underline-offset property. Use RawValue for complex CSS syntax.
func (e *Element) TextUnderlineOffset(v TextUnderlineOffsetValue) *Element {
	return e.setValue("text-underline-offset", v)
}
func (expression) isTextUnderlineOffset()  {}
func (wideKeyword) isTextUnderlineOffset() {}

// TextUnderlinePositionValue is a value for the CSS text-underline-position property.
type TextUnderlinePositionValue interface {
	cssValue
	isTextUnderlinePosition()
}

// TextUnderlinePosition sets the CSS text-underline-position property. Use RawValue for complex CSS syntax.
func (e *Element) TextUnderlinePosition(v TextUnderlinePositionValue) *Element {
	return e.setValue("text-underline-position", v)
}
func (expression) isTextUnderlinePosition()  {}
func (wideKeyword) isTextUnderlinePosition() {}

// TextWrapValue is a value for the CSS text-wrap property.
type TextWrapValue interface {
	cssValue
	isTextWrap()
}

// TextWrap sets the CSS text-wrap property. Use RawValue for complex CSS syntax.
func (e *Element) TextWrap(v TextWrapValue) *Element { return e.setValue("text-wrap", v) }
func (expression) isTextWrap()                       {}
func (wideKeyword) isTextWrap()                      {}

// TextWrapModeValue is a value for the CSS text-wrap-mode property.
type TextWrapModeValue interface {
	cssValue
	isTextWrapMode()
}

// TextWrapMode sets the CSS text-wrap-mode property. Use RawValue for complex CSS syntax.
func (e *Element) TextWrapMode(v TextWrapModeValue) *Element { return e.setValue("text-wrap-mode", v) }
func (expression) isTextWrapMode()                           {}
func (wideKeyword) isTextWrapMode()                          {}

// TextWrapStyleValue is a value for the CSS text-wrap-style property.
type TextWrapStyleValue interface {
	cssValue
	isTextWrapStyle()
}

// TextWrapStyle sets the CSS text-wrap-style property. Use RawValue for complex CSS syntax.
func (e *Element) TextWrapStyle(v TextWrapStyleValue) *Element {
	return e.setValue("text-wrap-style", v)
}
func (expression) isTextWrapStyle()  {}
func (wideKeyword) isTextWrapStyle() {}

// TopValue is a value for the CSS top property.
type TopValue interface {
	cssValue
	isTop()
}

// Top sets the CSS top property. Use RawValue for complex CSS syntax.
func (e *Element) Top(v TopValue) *Element { return e.setValue("top", v) }
func (expression) isTop()                  {}
func (wideKeyword) isTop()                 {}
func (Length) isTop()                      {}
func (Percentage) isTop()                  {}

// TransformValue is a value for the CSS transform property.
type TransformValue interface {
	cssValue
	isTransform()
}

// Transform sets the CSS transform property. Use RawValue for complex CSS syntax.
func (e *Element) Transform(v TransformValue) *Element { return e.setValue("transform", v) }
func (expression) isTransform()                        {}
func (wideKeyword) isTransform()                       {}
func (Transform) isTransform()                         {}

// TransformBoxValue is a value for the CSS transform-box property.
type TransformBoxValue interface {
	cssValue
	isTransformBox()
}

// TransformBox sets the CSS transform-box property. Use RawValue for complex CSS syntax.
func (e *Element) TransformBox(v TransformBoxValue) *Element { return e.setValue("transform-box", v) }
func (expression) isTransformBox()                           {}
func (wideKeyword) isTransformBox()                          {}

// TransformOriginValue is a value for the CSS transform-origin property.
type TransformOriginValue interface {
	cssValue
	isTransformOrigin()
}

// TransformOrigin sets the CSS transform-origin property. Use RawValue for complex CSS syntax.
func (e *Element) TransformOrigin(v TransformOriginValue) *Element {
	return e.setValue("transform-origin", v)
}
func (expression) isTransformOrigin()  {}
func (wideKeyword) isTransformOrigin() {}

// TransformStyleValue is a value for the CSS transform-style property.
type TransformStyleValue interface {
	cssValue
	isTransformStyle()
}

// TransformStyle sets the CSS transform-style property. Use RawValue for complex CSS syntax.
func (e *Element) TransformStyle(v TransformStyleValue) *Element {
	return e.setValue("transform-style", v)
}
func (expression) isTransformStyle()  {}
func (wideKeyword) isTransformStyle() {}

// TransitionValue is a value for the CSS transition property.
type TransitionValue interface {
	cssValue
	isTransition()
}

// Transition sets the CSS transition property. Use RawValue for complex CSS syntax.
func (e *Element) Transition(v TransitionValue) *Element { return e.setValue("transition", v) }
func (expression) isTransition()                         {}
func (wideKeyword) isTransition()                        {}

// TransitionBehaviorValue is a value for the CSS transition-behavior property.
type TransitionBehaviorValue interface {
	cssValue
	isTransitionBehavior()
}

// TransitionBehavior sets the CSS transition-behavior property. Use RawValue for complex CSS syntax.
func (e *Element) TransitionBehavior(v TransitionBehaviorValue) *Element {
	return e.setValue("transition-behavior", v)
}
func (expression) isTransitionBehavior()  {}
func (wideKeyword) isTransitionBehavior() {}

// TransitionDelayValue is a value for the CSS transition-delay property.
type TransitionDelayValue interface {
	cssValue
	isTransitionDelay()
}

// TransitionDelay sets the CSS transition-delay property. Use RawValue for complex CSS syntax.
func (e *Element) TransitionDelay(v TransitionDelayValue) *Element {
	return e.setValue("transition-delay", v)
}
func (expression) isTransitionDelay()  {}
func (wideKeyword) isTransitionDelay() {}

// TransitionDurationValue is a value for the CSS transition-duration property.
type TransitionDurationValue interface {
	cssValue
	isTransitionDuration()
}

// TransitionDuration sets the CSS transition-duration property. Use RawValue for complex CSS syntax.
func (e *Element) TransitionDuration(v TransitionDurationValue) *Element {
	return e.setValue("transition-duration", v)
}
func (expression) isTransitionDuration()  {}
func (wideKeyword) isTransitionDuration() {}

// TransitionPropertyValue is a value for the CSS transition-property property.
type TransitionPropertyValue interface {
	cssValue
	isTransitionProperty()
}

// TransitionProperty sets the CSS transition-property property. Use RawValue for complex CSS syntax.
func (e *Element) TransitionProperty(v TransitionPropertyValue) *Element {
	return e.setValue("transition-property", v)
}
func (expression) isTransitionProperty()  {}
func (wideKeyword) isTransitionProperty() {}

// TransitionTimingFunctionValue is a value for the CSS transition-timing-function property.
type TransitionTimingFunctionValue interface {
	cssValue
	isTransitionTimingFunction()
}

// TransitionTimingFunction sets the CSS transition-timing-function property. Use RawValue for complex CSS syntax.
func (e *Element) TransitionTimingFunction(v TransitionTimingFunctionValue) *Element {
	return e.setValue("transition-timing-function", v)
}
func (expression) isTransitionTimingFunction()  {}
func (wideKeyword) isTransitionTimingFunction() {}

// TranslateValue is a value for the CSS translate property.
type TranslateValue interface {
	cssValue
	isTranslate()
}

// Translate sets the CSS translate property. Use RawValue for complex CSS syntax.
func (e *Element) Translate(v TranslateValue) *Element { return e.setValue("translate", v) }
func (expression) isTranslate()                        {}
func (wideKeyword) isTranslate()                       {}

// TriggerScopeValue is a value for the CSS trigger-scope property.
type TriggerScopeValue interface {
	cssValue
	isTriggerScope()
}

// TriggerScope sets the CSS trigger-scope property. Use RawValue for complex CSS syntax.
func (e *Element) TriggerScope(v TriggerScopeValue) *Element { return e.setValue("trigger-scope", v) }
func (expression) isTriggerScope()                           {}
func (wideKeyword) isTriggerScope()                          {}

// UnicodeBidiValue is a value for the CSS unicode-bidi property.
type UnicodeBidiValue interface {
	cssValue
	isUnicodeBidi()
}

// UnicodeBidi sets the CSS unicode-bidi property. Use RawValue for complex CSS syntax.
func (e *Element) UnicodeBidi(v UnicodeBidiValue) *Element { return e.setValue("unicode-bidi", v) }
func (expression) isUnicodeBidi()                          {}
func (wideKeyword) isUnicodeBidi()                         {}

// VerticalAlignValue is a value for the CSS vertical-align property.
type VerticalAlignValue interface {
	cssValue
	isVerticalAlign()
}

// VerticalAlign sets the CSS vertical-align property. Use RawValue for complex CSS syntax.
func (e *Element) VerticalAlign(v VerticalAlignValue) *Element {
	return e.setValue("vertical-align", v)
}
func (expression) isVerticalAlign()  {}
func (wideKeyword) isVerticalAlign() {}

// ViewTransitionClassValue is a value for the CSS view-transition-class property.
type ViewTransitionClassValue interface {
	cssValue
	isViewTransitionClass()
}

// ViewTransitionClass sets the CSS view-transition-class property. Use RawValue for complex CSS syntax.
func (e *Element) ViewTransitionClass(v ViewTransitionClassValue) *Element {
	return e.setValue("view-transition-class", v)
}
func (expression) isViewTransitionClass()  {}
func (wideKeyword) isViewTransitionClass() {}

// ViewTransitionGroupValue is a value for the CSS view-transition-group property.
type ViewTransitionGroupValue interface {
	cssValue
	isViewTransitionGroup()
}

// ViewTransitionGroup sets the CSS view-transition-group property. Use RawValue for complex CSS syntax.
func (e *Element) ViewTransitionGroup(v ViewTransitionGroupValue) *Element {
	return e.setValue("view-transition-group", v)
}
func (expression) isViewTransitionGroup()  {}
func (wideKeyword) isViewTransitionGroup() {}

// ViewTransitionNameValue is a value for the CSS view-transition-name property.
type ViewTransitionNameValue interface {
	cssValue
	isViewTransitionName()
}

// ViewTransitionName sets the CSS view-transition-name property. Use RawValue for complex CSS syntax.
func (e *Element) ViewTransitionName(v ViewTransitionNameValue) *Element {
	return e.setValue("view-transition-name", v)
}
func (expression) isViewTransitionName()  {}
func (wideKeyword) isViewTransitionName() {}

// ViewTransitionScopeValue is a value for the CSS view-transition-scope property.
type ViewTransitionScopeValue interface {
	cssValue
	isViewTransitionScope()
}

// ViewTransitionScope sets the CSS view-transition-scope property. Use RawValue for complex CSS syntax.
func (e *Element) ViewTransitionScope(v ViewTransitionScopeValue) *Element {
	return e.setValue("view-transition-scope", v)
}
func (expression) isViewTransitionScope()  {}
func (wideKeyword) isViewTransitionScope() {}

// VisibilityValue is a value for the CSS visibility property.
type VisibilityValue interface {
	cssValue
	isVisibility()
}

// Visibility sets the CSS visibility property. Use RawValue for complex CSS syntax.
func (e *Element) Visibility(v VisibilityValue) *Element { return e.setValue("visibility", v) }
func (expression) isVisibility()                         {}
func (wideKeyword) isVisibility()                        {}

type visibilityKeyword uint16

const (
	// VisibilityVisible sets visibility to visible.
	VisibilityVisible visibilityKeyword = 1
	// VisibilityHidden sets visibility to hidden.
	VisibilityHidden visibilityKeyword = 2
	// VisibilityCollapse sets visibility to collapse.
	VisibilityCollapse visibilityKeyword = 3
)

func (visibilityKeyword) isVisibility() {}
func (v visibilityKeyword) cssText() (string, error) {
	switch v {
	case VisibilityVisible:
		return "visible", nil
	case VisibilityHidden:
		return "hidden", nil
	case VisibilityCollapse:
		return "collapse", nil
	default:
		return "", fmt.Errorf("ui: invalid visibility keyword %d", v)
	}
}

// WhiteSpaceValue is a value for the CSS white-space property.
type WhiteSpaceValue interface {
	cssValue
	isWhiteSpace()
}

// WhiteSpace sets the CSS white-space property. Use RawValue for complex CSS syntax.
func (e *Element) WhiteSpace(v WhiteSpaceValue) *Element { return e.setValue("white-space", v) }
func (expression) isWhiteSpace()                         {}
func (wideKeyword) isWhiteSpace()                        {}

type whiteSpaceKeyword uint16

const (
	// WhiteSpaceNormal sets white-space to normal.
	WhiteSpaceNormal whiteSpaceKeyword = 1
	// WhiteSpaceNowrap sets white-space to nowrap.
	WhiteSpaceNowrap whiteSpaceKeyword = 2
	// WhiteSpacePre sets white-space to pre.
	WhiteSpacePre whiteSpaceKeyword = 3
	// WhiteSpacePreWrap sets white-space to pre-wrap.
	WhiteSpacePreWrap whiteSpaceKeyword = 4
	// WhiteSpacePreLine sets white-space to pre-line.
	WhiteSpacePreLine whiteSpaceKeyword = 5
	// WhiteSpaceBreakSpaces sets white-space to break-spaces.
	WhiteSpaceBreakSpaces whiteSpaceKeyword = 6
)

func (whiteSpaceKeyword) isWhiteSpace() {}
func (v whiteSpaceKeyword) cssText() (string, error) {
	switch v {
	case WhiteSpaceNormal:
		return "normal", nil
	case WhiteSpaceNowrap:
		return "nowrap", nil
	case WhiteSpacePre:
		return "pre", nil
	case WhiteSpacePreWrap:
		return "pre-wrap", nil
	case WhiteSpacePreLine:
		return "pre-line", nil
	case WhiteSpaceBreakSpaces:
		return "break-spaces", nil
	default:
		return "", fmt.Errorf("ui: invalid white-space keyword %d", v)
	}
}

// WhiteSpaceCollapseValue is a value for the CSS white-space-collapse property.
type WhiteSpaceCollapseValue interface {
	cssValue
	isWhiteSpaceCollapse()
}

// WhiteSpaceCollapse sets the CSS white-space-collapse property. Use RawValue for complex CSS syntax.
func (e *Element) WhiteSpaceCollapse(v WhiteSpaceCollapseValue) *Element {
	return e.setValue("white-space-collapse", v)
}
func (expression) isWhiteSpaceCollapse()  {}
func (wideKeyword) isWhiteSpaceCollapse() {}

// WhiteSpaceTrimValue is a value for the CSS white-space-trim property.
type WhiteSpaceTrimValue interface {
	cssValue
	isWhiteSpaceTrim()
}

// WhiteSpaceTrim sets the CSS white-space-trim property. Use RawValue for complex CSS syntax.
func (e *Element) WhiteSpaceTrim(v WhiteSpaceTrimValue) *Element {
	return e.setValue("white-space-trim", v)
}
func (expression) isWhiteSpaceTrim()  {}
func (wideKeyword) isWhiteSpaceTrim() {}

// WidowsValue is a value for the CSS widows property.
type WidowsValue interface {
	cssValue
	isWidows()
}

// Widows sets the CSS widows property. Use RawValue for complex CSS syntax.
func (e *Element) Widows(v WidowsValue) *Element { return e.setValue("widows", v) }
func (expression) isWidows()                     {}
func (wideKeyword) isWidows()                    {}
func (Numeric) isWidows()                        {}

// WidthValue is a value for the CSS width property.
type WidthValue interface {
	cssValue
	isWidth()
}

// Width sets the CSS width property. Use RawValue for complex CSS syntax.
func (e *Element) Width(v WidthValue) *Element { return e.setValue("width", v) }
func (expression) isWidth()                    {}
func (wideKeyword) isWidth()                   {}
func (Length) isWidth()                        {}
func (Percentage) isWidth()                    {}

type widthKeyword uint16

const (
	// WidthAuto sets width to auto.
	WidthAuto widthKeyword = 1
	// WidthMinContent sets width to min-content.
	WidthMinContent widthKeyword = 2
	// WidthMaxContent sets width to max-content.
	WidthMaxContent widthKeyword = 3
	// WidthFitContent sets width to fit-content.
	WidthFitContent widthKeyword = 4
)

func (widthKeyword) isWidth() {}
func (v widthKeyword) cssText() (string, error) {
	switch v {
	case WidthAuto:
		return "auto", nil
	case WidthMinContent:
		return "min-content", nil
	case WidthMaxContent:
		return "max-content", nil
	case WidthFitContent:
		return "fit-content", nil
	default:
		return "", fmt.Errorf("ui: invalid width keyword %d", v)
	}
}

// WordBreakValue is a value for the CSS word-break property.
type WordBreakValue interface {
	cssValue
	isWordBreak()
}

// WordBreak sets the CSS word-break property. Use RawValue for complex CSS syntax.
func (e *Element) WordBreak(v WordBreakValue) *Element { return e.setValue("word-break", v) }
func (expression) isWordBreak()                        {}
func (wideKeyword) isWordBreak()                       {}

// WordSpacingValue is a value for the CSS word-spacing property.
type WordSpacingValue interface {
	cssValue
	isWordSpacing()
}

// WordSpacing sets the CSS word-spacing property. Use RawValue for complex CSS syntax.
func (e *Element) WordSpacing(v WordSpacingValue) *Element { return e.setValue("word-spacing", v) }
func (expression) isWordSpacing()                          {}
func (wideKeyword) isWordSpacing()                         {}
func (Length) isWordSpacing()                              {}

// WrapAfterValue is a value for the CSS wrap-after property.
type WrapAfterValue interface {
	cssValue
	isWrapAfter()
}

// WrapAfter sets the CSS wrap-after property. Use RawValue for complex CSS syntax.
func (e *Element) WrapAfter(v WrapAfterValue) *Element { return e.setValue("wrap-after", v) }
func (expression) isWrapAfter()                        {}
func (wideKeyword) isWrapAfter()                       {}

// WrapBeforeValue is a value for the CSS wrap-before property.
type WrapBeforeValue interface {
	cssValue
	isWrapBefore()
}

// WrapBefore sets the CSS wrap-before property. Use RawValue for complex CSS syntax.
func (e *Element) WrapBefore(v WrapBeforeValue) *Element { return e.setValue("wrap-before", v) }
func (expression) isWrapBefore()                         {}
func (wideKeyword) isWrapBefore()                        {}

// WrapFlowValue is a value for the CSS wrap-flow property.
type WrapFlowValue interface {
	cssValue
	isWrapFlow()
}

// WrapFlow sets the CSS wrap-flow property. Use RawValue for complex CSS syntax.
func (e *Element) WrapFlow(v WrapFlowValue) *Element { return e.setValue("wrap-flow", v) }
func (expression) isWrapFlow()                       {}
func (wideKeyword) isWrapFlow()                      {}

// WrapInsideValue is a value for the CSS wrap-inside property.
type WrapInsideValue interface {
	cssValue
	isWrapInside()
}

// WrapInside sets the CSS wrap-inside property. Use RawValue for complex CSS syntax.
func (e *Element) WrapInside(v WrapInsideValue) *Element { return e.setValue("wrap-inside", v) }
func (expression) isWrapInside()                         {}
func (wideKeyword) isWrapInside()                        {}

// WrapThroughValue is a value for the CSS wrap-through property.
type WrapThroughValue interface {
	cssValue
	isWrapThrough()
}

// WrapThrough sets the CSS wrap-through property. Use RawValue for complex CSS syntax.
func (e *Element) WrapThrough(v WrapThroughValue) *Element { return e.setValue("wrap-through", v) }
func (expression) isWrapThrough()                          {}
func (wideKeyword) isWrapThrough()                         {}

// WritingModeValue is a value for the CSS writing-mode property.
type WritingModeValue interface {
	cssValue
	isWritingMode()
}

// WritingMode sets the CSS writing-mode property. Use RawValue for complex CSS syntax.
func (e *Element) WritingMode(v WritingModeValue) *Element { return e.setValue("writing-mode", v) }
func (expression) isWritingMode()                          {}
func (wideKeyword) isWritingMode()                         {}

// ZIndexValue is a value for the CSS z-index property.
type ZIndexValue interface {
	cssValue
	isZIndex()
}

// ZIndex sets the CSS z-index property. Use RawValue for complex CSS syntax.
func (e *Element) ZIndex(v ZIndexValue) *Element { return e.setValue("z-index", v) }
func (expression) isZIndex()                     {}
func (wideKeyword) isZIndex()                    {}
func (Numeric) isZIndex()                        {}
