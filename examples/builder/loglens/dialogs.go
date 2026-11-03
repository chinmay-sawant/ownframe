package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const alertNotesRows = 2

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
	return chadcn.Dialog(chadcn.DialogOptions{
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
		ui.Div().Class("app-stack").Children(
			a.dialogIdentityRow(),
			a.dialogNotesField(),
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
			Label:   "Notify email",
			For:     "alert-email",
			Control: chadcn.InputWithOptions("alert-email", a.email, chadcn.InputOptions{Type: "email", Placeholder: "oncall@acme.dev"}),
			Description: "Pages, escalations, and the daily digest for this rule go here. " +
				"Leave it blank to use the team alias.",
		}).Width(ui.Percent(48)),
		a.dialogSeverityField().Width(ui.Percent(48)),
	)
}

// dialogNotesField keeps the free-text intent query next to the rule.
func (a *app) dialogNotesField() *ui.Element {
	return chadcn.Field(chadcn.FieldOptions{
		Label: "Condition notes",
		For:   "alert-notes",
		Control: chadcn.TextareaWithOptions("alert-notes", a.notes, chadcn.TextareaOptions{
			Rows:        alertNotesRows,
			Placeholder: "service:api-gateway level:error status:504",
		}),
	})
}

// dialogDeliveryRow pairs the threshold with the verification code.
func (a *app) dialogDeliveryRow() *ui.Element {
	return ui.Div().Class("app-row").AlignItems(ui.AlignItemsFlexStart).Children(
		a.dialogThresholdField().Width(ui.Percent(52)),
		a.dialogOTPField().Width(ui.Percent(44)),
	)
}

// dialogThresholdField pairs the slider with two nudge buttons, because the
// slider itself only reports its value.
func (a *app) dialogThresholdField() *ui.Element {
	return chadcn.Field(chadcn.FieldOptions{
		Label:       "Error-rate threshold",
		Description: fmt.Sprintf("Page when errors stay above %d per 1,000 requests for five minutes.", a.threshold),
		Control: ui.Div().Class("app-col").Children(
			chadcn.Slider(chadcn.SliderOptions{
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
				chadcn.Badge(fmt.Sprintf("%d per 1k", a.threshold), chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
			),
		),
	})
}

// dialogNudge moves the threshold by delta and clamps it to the slider range.
func (a *app) dialogNudge(delta int) *ui.Element {
	return chadcn.Button(fmt.Sprintf("%+d", delta), chadcn.ButtonOptions{Variant: chadcn.ButtonText, Size: chadcn.ButtonSmall}).
		OnClick(func(ctx context.Context) error {
			a.threshold = min(dialogThresholdMax, max(dialogThresholdMin, a.threshold+delta))
			return a.refresh(ctx)
		})
}

// dialogTogglesRow holds the on/off switch and the on-call escalation box.
func (a *app) dialogTogglesRow() *ui.Element {
	return ui.Div().Class("app-row app-row--between").Children(
		ui.Div().Class("app-col").Children(
			chadcn.Switch("Alerts enabled", a.alertsEnabled).ID("alert-enabled").
				OnClick(func(ctx context.Context) error {
					a.alertsEnabled = !a.alertsEnabled
					return a.refresh(ctx)
				}),
			ui.Text("Off keeps the rule saved without sending anything.").Class("app-muted"),
		),
		chadcn.CheckboxWithOptions("Page the on-call engineer", a.checkValue, chadcn.CheckboxOptions{
			Description: "Escalate to PagerDuty when the rule repeats twice.",
		}).ID("alert-page").
			OnClick(func(ctx context.Context) error {
				a.checkValue = !a.checkValue
				return a.refresh(ctx)
			}),
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
				{Value: "info", Label: "Info", Description: "Daily digest only"},
				{Value: "warning", Label: "Warning", Description: "Slack #alerts"},
				{Value: "critical", Label: "Critical", Description: "Page on-call"},
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
		ui.Text("Six digits from the authenticator enrolled for this workspace.").Class("app-muted"),
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
	a.notes = strings.TrimSpace(a.formValue("alert-notes"))
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
	return chadcn.AlertDialog(chadcn.AlertDialogOptions{
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
