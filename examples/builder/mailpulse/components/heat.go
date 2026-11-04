package components

import "github.com/chinmay-sawant/ownframe/ui"

// HeatCell renders one heatmap swatch for a level from 0 to 4.
func HeatCell(level int) *ui.Element {
	return HeatCellID("", level)
}

// HeatCellID renders a heatmap swatch with a stable element id so hover and
// click handlers can identify the weekday and hour it represents.
func HeatCellID(id string, level int) *ui.Element {
	if level < 0 {
		level = 0
	}
	if level > 4 {
		level = 4
	}
	digit := string([]byte{byte('0' + level)})
	cell := ui.Div().Class("mp-heat-cell mp-heat-cell--l"+digit+" mp-heat--"+digit).Attr("title", digit)
	if id != "" {
		cell.ID(id)
	}
	return cell
}
