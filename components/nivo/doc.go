// Package nivo provides div and CSS charts for OwnFrame.
//
// Charts are HTML built through the ui builder. They return fresh trees, take
// their data through options structs, and leave state in the host. There is no
// canvas, SVG, or script. Paint lives in Stylesheet, which an application
// appends to its theme:
//
//	Theme: theme.Dark + nivo.Stylesheet
//
// Lengths are percentages of a measured maximum, so a drag or a new value is a
// new tree, not a canvas redraw. Choropleth and GeoMap are labeled region
// cells, not geographic maps. CirclePacking is a depth-ordered list of rows,
// not packed circles.
package nivo
