package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// Alert-rule bounds. The dialog asks for a six-digit code and keeps the
// threshold inside a band the pager tolerates.
const (
	dialogOTPSlots       = 6
	dialogThresholdMin   = 20
	dialogThresholdMax   = 200
	dialogThresholdStep  = 5
	dialogThresholdShift = 10
)

// globalDialogs mounts the page-level modal surfaces and the toast. The shell
// calls it once at the screen root, so the dialog backdrop covers the window
// and the toast lands in the bottom-right corner.
func (a *app) globalDialogs() *ui.Element {
	return ui.Div().Display(ui.DisplayBlock).Children(
		a.dialogCreateAlert(),
		a.dialogDeleteView(),
		a.dialogOnCall(),
		chadcn.Toast(chadcn.ToastOptions{
			Open:        a.toastOpen,
			Title:       a.toastTitle,
			Description: a.toastBody,
			CloseAction: "toast:close",
		}),
	)
}

// dialogCreateAlert builds the rule editor from the page and dialog state. The
// page header opens it through a.dialogOpen.
func (a *app) dialogCreateAlert() *ui.Element {
	if !a.dialogOpen {
		return nil
	}
	return chadcn.Dialog(chadcn.DialogOptions{
		Interactions: a.interactions,
		ID:           "create-alert-dialog",
		Open:         a.dialogOpen,
		Title:        "Create alert rule",
		Description:  "Notify the channel that owns this service when the condition holds for five minutes.",
		Width:        ui.Px(620),
		OnOpenChange: a.dialogSetOpen,
		Footer: []*ui.Element{
			chadcn.Button("Cancel", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined}).
				OnClick(func(ctx context.Context) error {
					return a.dialogSetOpen(ctx, false)
				}),
			chadcn.Button("Save", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).
				OnClick(a.dialogSaveAlert),
		},
	},
		ui.Div().Class("app-stack").Gap(ui.Px(8)).Children(
			a.dialogIdentityRow(),
		),
		chadcn.Separator(chadcn.SeparatorOptions{}),
		a.dialogDeliveryRow(),
		a.dialogTogglesRow(),
	)
}

// dialogSetOpen writes the create-dialog flag for the backdrop, the close
// control, and Cancel.
func (a *app) dialogSetOpen(ctx context.Context, open bool) error {
	if a.dialogOpen == open {
		return nil
	}
	a.dialogOpen = open
	return a.refresh(ctx)
}

// dialogIdentityRow pairs the delivery address with the escalation severity,
// which keeps the body short enough for one dialog view.
func (a *app) dialogIdentityRow() *ui.Element {
	return ui.Div().Class("app-row").AlignItems(ui.AlignItemsFlexStart).Children(
		chadcn.Field(chadcn.FieldOptions{
			Label:       "Notify email",
			For:         "alert-email",
			Control:     chadcn.InputWithOptions("alert-email", a.email, chadcn.InputOptions{Type: "email", Placeholder: "oncall@acme.dev"}),
			Description: "Pages and digests. Blank uses the team alias.",
		}).Width(ui.Percent(48)),
		a.dialogSeverityField().Width(ui.Percent(48)),
	)
}

// dialogDeliveryRow pairs the threshold with the verification code.
func (a *app) dialogDeliveryRow() *ui.Element {
	return ui.Div().Class("app-row").AlignItems(ui.AlignItemsFlexStart).Children(
		a.dialogThresholdField().Width(ui.Percent(52)),
		a.dialogOTPField().Width(ui.Percent(44)),
	)
}

// dialogThresholdField pairs direct slider input with threshold preset buttons.
func (a *app) dialogThresholdField() *ui.Element {
	return chadcn.Field(chadcn.FieldOptions{
		Label:       "Error-rate threshold",
		Description: fmt.Sprintf("Five minutes above %d per 1,000 requests.", a.threshold),
		Control: ui.Div().Class("app-col").Children(
			chadcn.Slider(chadcn.SliderOptions{
				Interactions: a.interactions,
				OnValueChange: func(ctx context.Context, value float64) error {
					a.threshold = int(value)
					return a.patchSlider(ctx, "alert-threshold-slider", value, dialogThresholdMin, dialogThresholdMax, fmt.Sprintf("%d per 1k", a.threshold))
				},
				ID:    "alert-threshold-slider",
				Label: "Error-rate threshold",
				Value: float64(a.threshold),
				Min:   dialogThresholdMin,
				Max:   dialogThresholdMax,
				Step:  dialogThresholdStep,
			}),
			ui.Div().Class("app-row app-row--between").Children(
				ui.Div().Class("app-row").Children(
					a.dialogNudge(-dialogThresholdShift),
					a.dialogNudge(dialogThresholdShift),
				),
				chadcn.Badge(fmt.Sprintf("%d per 1k", a.threshold), chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}).ID("alert-threshold-badge"),
			),
		),
	})
}

// dialogNudge moves the threshold by delta and clamps it to the slider range.
func (a *app) dialogNudge(delta int) *ui.Element {
	return chadcn.Button(fmt.Sprintf("%+d", delta), chadcn.ButtonOptions{Variant: chadcn.ButtonText, Size: chadcn.ButtonSmall}).
		OnClick(func(ctx context.Context) error {
			a.threshold = min(dialogThresholdMax, max(dialogThresholdMin, a.threshold+delta))
			return a.patchSlider(ctx, "alert-threshold-slider", float64(a.threshold), dialogThresholdMin, dialogThresholdMax, fmt.Sprintf("%d per 1k", a.threshold))
		})
}

// dialogTogglesRow holds the on/off switch and the on-call escalation box in
// two fixed halves, so the checkbox cannot wrap below the dialog body.
func (a *app) dialogTogglesRow() *ui.Element {
	return ui.Div().Class("app-row").Gap(ui.Px(12)).Children(
		ui.Div().Width(ui.Percent(48)).Children(
			ui.Div().Class("app-col").Children(
				chadcn.SwitchWithOptions("Alerts enabled", a.alertsEnabled, chadcn.SwitchOptions{ID: "alert-enabled"}).
					OnClick(func(ctx context.Context) error {
						a.alertsEnabled = !a.alertsEnabled
						return a.patchSwitch(ctx, "alert-enabled", a.alertsEnabled)
					}),
				ui.Text("Off keeps the rule saved without sending anything.").Class("app-muted"),
			),
		),
		ui.Div().Width(ui.Percent(48)).Children(
			chadcn.CheckboxWithOptions("Page the on-call engineer", a.checkValue, chadcn.CheckboxOptions{
				ID:          "alert-page",
				Description: "Escalate to PagerDuty when the rule repeats twice.",
			}).OnClick(func(ctx context.Context) error {
				a.checkValue = !a.checkValue
				return a.patchCheckbox(ctx, "alert-page", a.checkValue)
			}),
		),
	)
}

// dialogSeverityField selects the escalation tone for the rule.
func (a *app) dialogSeverityField() *ui.Element {
	return ui.Div().Class("app-col").Children(
		ui.Text("Severity").Class("app-section-label"),
		chadcn.RadioGroup(chadcn.RadioGroupOptions{
			ID:    "alert-severity",
			Label: "Severity",
			Value: a.severityValue,
			Items: []chadcn.RadioItem{
				{Value: "info", Label: "Info"},
				{Value: "warning", Label: "Warning"},
				{Value: "critical", Label: "Critical"},
			},
			OnSelect: func(ctx context.Context, value string) error {
				a.severityValue = value
				return a.refresh(ctx)
			},
		}),
	)
}

// dialogOTPField confirms the person saving the rule.
func (a *app) dialogOTPField() *ui.Element {
	return ui.Div().Class("app-col").Children(
		ui.Div().Class("app-row app-row--between").Children(
			chadcn.Label("Verification code"),
			chadcn.Badge("2FA", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
		),
		chadcn.InputOTP(chadcn.InputOTPOptions{
			ID:     "alert-otp",
			Label:  "Verification code",
			Length: dialogOTPSlots,
			Value:  a.otp,
		}),
		ui.Text("Six digits from the enrolled authenticator.").Class("app-muted"),
	)
}

// dialogSaveAlert reads the form-owned values, stores them on the app, closes
// the dialog, and reports the result through the shared toast.
func (a *app) dialogSaveAlert(ctx context.Context) error {
	email := strings.TrimSpace(a.formValue("alert-email"))
	if email == "" {
		email = "oncall@acme.dev"
	}
	a.email = email
	a.otp = a.dialogOTPDigits()
	a.dialogOpen = false
	route := "Slack"
	if a.checkValue {
		route = "PagerDuty"
	}
	return a.showToast(ctx, "Alert rule saved",
		fmt.Sprintf("%s · %d per 1k · %s via %s", a.severityValue, a.threshold, a.email, route))
}

// dialogOTPDigits joins the typed slot values, falling back to the stored code
// when the slots were never touched.
func (a *app) dialogOTPDigits() string {
	var digits strings.Builder
	for slot := range dialogOTPSlots {
		digits.WriteString(strings.TrimSpace(a.formValue(fmt.Sprintf("alert-otp-%d", slot))))
	}
	if digits.Len() == 0 {
		return a.otp
	}
	return digits.String()
}

// dialogDeleteView confirms removal of the saved EU latency view.
func (a *app) dialogDeleteView() *ui.Element {
	if !a.alertOpen {
		return nil
	}
	return chadcn.AlertDialog(chadcn.AlertDialogOptions{
		Interactions: a.interactions,
		ID:           "delete-view-dialog",
		Open:         a.alertOpen,
		Title:        "Delete saved view?",
		Description:  "The EU latency view and its pinned filters are removed for everyone in acme / production. This cannot be undone.",
		ConfirmLabel: "Delete",
		CancelLabel:  "Cancel",
		Destructive:  true,
		OnConfirm: func(ctx context.Context) error {
			a.alertOpen = false
			return a.showToast(ctx, "View deleted", "EU latency is gone from saved views.")
		},
		OnCancel: func(ctx context.Context) error {
			a.alertOpen = false
			return a.showToast(ctx, "View kept", "EU latency is still in saved views.")
		},
	})
}

// dialogOnCall lets the user reassign the on-call responder for the chosen date.
func (a *app) dialogOnCall() *ui.Element {
	if !a.onCallDialogOpen {
		return nil
	}
	currentDay := a.onCallDialogDate
	if currentDay.IsZero() {
		currentDay = a.onCallDate
	}
	if currentDay.IsZero() {
		currentDay = time.Now()
	}
	currentResponder := a.onCallFor(currentDay)

	var rows []*ui.Element
	for _, r := range defaultResponders() {
		resp := r
		isCurrent := resp.Name == currentResponder.Name
		badgeText := "Available"
		badgeVariant := chadcn.BadgeOutline
		btnVariant := chadcn.ButtonOutlined
		btnText := "Assign"
		if isCurrent {
			badgeText = "Current"
			badgeVariant = chadcn.BadgeSecondary
			btnVariant = chadcn.ButtonFilled
			btnText = "Active"
		}

		item := ui.Div().Class("app-row app-row--between").
			Padding(ui.Edges(ui.Px(8), ui.Px(10))).
			Border(ui.SolidBorder(ui.Px(1), ui.ColorVar("--of-border", ui.Hex("#e5e5e5")))).
			BorderRadius(ui.Px(6)).
			Children(
				ui.Div().Class("app-row").Gap(ui.Px(10)).Children(
					chadcn.Avatar(chadcn.AvatarOptions{Initials: resp.Initials, Size: chadcn.AvatarSmall, Status: chadcn.AvatarStatusOnline}),
					ui.Div().Class("app-col").Gap(ui.Px(2)).Children(
						ui.Text(resp.Name).FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
						ui.Text(resp.Role+" · "+resp.Region).Class("app-muted"),
					),
				),
				ui.Div().Class("app-row").Gap(ui.Px(8)).Children(
					chadcn.Badge(badgeText, chadcn.BadgeOptions{Variant: badgeVariant}),
					chadcn.Button(btnText, chadcn.ButtonOptions{
						Variant: btnVariant,
						Size:    chadcn.ButtonSmall,
					}).OnClick(func(ctx context.Context) error {
						a.onCallDialogOpen = false
						return a.setOnCall(ctx, currentDay, resp)
					}),
				),
			)
		rows = append(rows, item)
	}

	return chadcn.Dialog(chadcn.DialogOptions{
		Interactions: a.interactions,
		ID:           "oncall-swap-dialog",
		Open:         a.onCallDialogOpen,
		Title:        "Swap On-Call Coverage",
		Description:  "Assign primary responder for " + currentDay.Format("Monday, Jan 2, 2006") + ".",
		Width:        ui.Px(540),
		OnOpenChange: func(ctx context.Context, open bool) error {
			a.onCallDialogOpen = open
			return a.refresh(ctx)
		},
		Footer: []*ui.Element{
			chadcn.Button("Close", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined}).
				OnClick(func(ctx context.Context) error {
					a.onCallDialogOpen = false
					return a.refresh(ctx)
				}),
		},
	},
		ui.Div().Class("app-stack").Gap(ui.Px(10)).Children(rows...),
	)
}
