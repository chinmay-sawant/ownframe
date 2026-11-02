package theme

import "github.com/chinmay-sawant/ownframe/ui"

// Background is the application canvas color.
func Background() ui.Color { return ui.ColorVar("--of-background", ui.Hex("#f5f7fb")) }

// Surface is the color of cards, panels, and reading surfaces.
func Surface() ui.Color { return ui.ColorVar("--of-surface", ui.Hex("#ffffff")) }

// Control is the color of secondary controls and inset surfaces.
func Control() ui.Color { return ui.ColorVar("--of-control", ui.Hex("#eef2f8")) }

// Text is the primary foreground color.
func Text() ui.Color { return ui.ColorVar("--of-text", ui.Hex("#172b4d")) }

// Muted is the secondary foreground color.
func Muted() ui.Color { return ui.ColorVar("--of-muted", ui.Hex("#5c6d85")) }

// Primary is the accent color for primary actions.
func Primary() ui.Color { return ui.ColorVar("--of-primary", ui.Hex("#315fca")) }

// OnPrimary is the foreground color on the primary accent.
func OnPrimary() ui.Color { return ui.ColorVar("--of-on-primary", ui.Hex("#ffffff")) }

// BorderColor is the color of separators and control borders.
func BorderColor() ui.Color { return ui.ColorVar("--of-border", ui.Hex("#dce4f0")) }

// Selected is the background of a selected item.
func Selected() ui.Color { return ui.ColorVar("--of-selected", ui.Hex("#e5edff")) }

// Scrollbar is the thumb color, separate from every pane surface.
func Scrollbar() ui.Color { return ui.ColorVar("--of-scrollbar", ui.Hex("#899ab1")) }

// Border is the standard one-pixel separator or control border.
func Border() ui.BorderStroke { return ui.SolidBorder(ui.Px(1), BorderColor()) }

// Radius is the standard control and panel corner radius.
func Radius() ui.Length { return ui.LengthVar("--of-radius", ui.Px(8)) }

// Spacing is the standard card padding.
func Spacing() ui.Length { return ui.LengthVar("--of-spacing", ui.Px(16)) }

// ButtonPaddingX is horizontal button padding.
func ButtonPaddingX() ui.Length { return ui.LengthVar("--of-button-padding-x", ui.Px(12)) }

// ButtonPaddingY is vertical button padding.
func ButtonPaddingY() ui.Length { return ui.LengthVar("--of-button-padding-y", ui.Px(9)) }

// Backdrop is the translucent overlay behind a dialog.
func Backdrop() ui.Color { return ui.ColorVar("--of-backdrop", ui.Hex("#0000008c")) }

// Danger is the fill for destructive actions.
func Danger() ui.Color { return ui.ColorVar("--of-danger", ui.Hex("#b3363f")) }

// OnDanger is the foreground on a destructive action.
func OnDanger() ui.Color { return ui.ColorVar("--of-on-danger", ui.Hex("#ffffff")) }
