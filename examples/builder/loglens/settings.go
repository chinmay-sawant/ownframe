package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	settingsSliderStep = 5
	settingsSliderMax  = 100
	settingsNotesRows  = 3
)

const settingsOTPLength = 6

// settingsSheet is the project settings side panel. The sheet has no trigger:
// the sidebar gear sets settingsOpen. Every control writes its value straight
// to the app state, and Save reads the text fields back from Page.Form.
func (a *app) settingsSheet() *ui.Element {
	plan := "Scale"
	if value, ok := a.extra["settings-plan"].(string); ok && strings.TrimSpace(value) != "" {
		plan = strings.TrimSpace(value)
	}

	channels := make([]chadcn.SelectValue, 0, len(channelOptions))
	for _, option := range channelOptions {
		channels = append(channels, chadcn.SelectValue{Value: option, Label: option})
	}

	readOTP := func() string {
		var code strings.Builder
		for index := range settingsOTPLength {
			code.WriteString(strings.TrimSpace(a.formValue("settings-otp-" + strconv.Itoa(index))))
		}
		return code.String()
	}

	return chadcn.Sheet(chadcn.SheetOptions{
		ID:          "settings-sheet",
		Open:        a.settingsOpen,
		Side:        chadcn.SheetRight,
		Title:       "Project settings",
		Description: "Notifications, routing, and verification.",
		OnOpenChange: func(ctx context.Context, open bool) error {
			a.settingsOpen = open
			return a.refresh(ctx)
		},
		Footer: []*ui.Element{
			chadcn.Button("Cancel", chadcn.ButtonOptions{Variant: chadcn.ButtonText}).
				OnClick(func(ctx context.Context) error {
					a.settingsOpen = false
					return a.refresh(ctx)
				}),
			chadcn.Button("Save", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).
				OnClick(func(ctx context.Context) error {
					a.email = strings.TrimSpace(a.formValue("settings-email"))
					a.channel = a.formValue("settings-channel")
					if a.channel == "" {
						a.channel = "Slack"
					}
					a.webhook = strings.TrimPrefix(strings.TrimSpace(a.formValue("settings-webhook")), "https://")
					a.otp = readOTP()
					a.notes = strings.TrimSpace(a.formValue("settings-notes"))
					a.settingsOpen = false
					return a.showToast(ctx, "Settings saved",
						a.channel+" receives "+a.severityValue+" alerts for the "+plan+" plan.")
				}),
		},
	},
		ui.Div().Class("app-stack").Children(
			a.settingsRoutingSection(plan, channels),
			chadcn.Separator(chadcn.SeparatorOptions{Orientation: chadcn.SeparatorHorizontal}),
			a.settingsNoiseSection(),
			chadcn.Separator(chadcn.SeparatorOptions{Orientation: chadcn.SeparatorHorizontal}),
			a.settingsVerificationSection(),
		),
	)
}

func (a *app) settingsRoutingSection(plan string, channels []chadcn.SelectValue) *ui.Element {
	return ui.Div().Class("app-col").Children(
		ui.Div().Class("app-row app-row--between").Children(
			ui.Text("Alert routing").Class("app-section-label"),
			chadcn.Badge(plan+" plan", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
		),
		chadcn.Field(chadcn.FieldOptions{
			Label:       "Destination email",
			For:         "settings-email",
			Description: "Digests and paging receipts land here.",
			Control: chadcn.InputWithOptions("settings-email", a.email, chadcn.InputOptions{
				Type:        "email",
				Placeholder: "oncall@acme.dev",
			}),
		}),
		chadcn.Field(chadcn.FieldOptions{
			Label:   "Primary channel",
			For:     "settings-channel",
			Control: chadcn.NativeSelect("settings-channel", channels, a.channel),
		}),
		chadcn.Field(chadcn.FieldOptions{
			Label:       "Webhook",
			For:         "settings-webhook",
			Description: "Slack and PagerDuty events are mirrored to this endpoint.",
			Control: chadcn.InputGroup(chadcn.InputGroupOptions{
				ID:          "settings-webhook",
				Label:       "Webhook",
				Value:       a.webhook,
				Placeholder: "hooks.acme.dev/loglens",
				Prefix:      ui.Text("https://"),
			}),
		}),
	)
}

func (a *app) settingsNoiseSection() *ui.Element {
	presets := []int{50, 85, 95}
	buttons := make([]*ui.Element, 0, len(presets))
	for _, preset := range presets {
		variant := chadcn.ButtonOutlined
		if a.threshold == preset {
			variant = chadcn.ButtonFilled
		}
		buttons = append(buttons, chadcn.Button(strconv.Itoa(preset)+"%", chadcn.ButtonOptions{
			Variant: variant,
			Size:    chadcn.ButtonSmall,
		}).OnClick(func(ctx context.Context) error {
			a.threshold = preset
			return a.refresh(ctx)
		}))
	}

	return ui.Div().Class("app-col").Children(
		ui.Text("Noise control").Class("app-section-label"),
		chadcn.CheckboxWithOptions("Mute resolved incidents", a.mute, chadcn.CheckboxOptions{
			Description: "Skip notifications when an incident resolves before anyone acks it.",
		}).OnClick(func(ctx context.Context) error {
			a.mute = !a.mute
			return a.refresh(ctx)
		}),
		ui.Div().Class("app-row app-row--between").Children(
			chadcn.Switch("Alerting enabled", a.alertsEnabled).
				OnClick(func(ctx context.Context) error {
					a.alertsEnabled = !a.alertsEnabled
					return a.refresh(ctx)
				}),
			settingsAlertBadge(a.alertsEnabled),
		),
		ui.Div().Class("app-col").Children(
			ui.Div().Class("app-row app-row--between").Children(
				chadcn.Label("Sample threshold"),
				ui.Text(strconv.Itoa(a.threshold)+"%").Class("app-muted"),
			),
			chadcn.Slider(chadcn.SliderOptions{
				Label: "Sample threshold",
				Value: float64(a.threshold),
				Min:   0,
				Max:   settingsSliderMax,
				Step:  settingsSliderStep,
			}).ID("settings-threshold"),
			chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "Threshold presets"}, buttons...),
		),
		ui.Div().Class("app-col").Children(
			ui.Text("Minimum severity"),
			chadcn.RadioGroup(chadcn.RadioGroupOptions{
				ID:    "settings-severity",
				Label: "Minimum severity",
				Value: a.severityValue,
				Items: []chadcn.RadioItem{
					{Value: "info", Label: "Info", Description: "Everything above debug."},
					{Value: "warning", Label: "Warning", Description: "Warnings and errors."},
					{Value: "critical", Label: "Critical", Description: "Paging incidents only."},
				},
				OnSelect: func(ctx context.Context, value string) error {
					a.severityValue = value
					return a.refresh(ctx)
				},
			}),
		),
	)
}

func settingsAlertBadge(enabled bool) *ui.Element {
	if enabled {
		return chadcn.Badge("Live", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary})
	}
	return chadcn.Badge("Paused", chadcn.BadgeOptions{Variant: chadcn.BadgeDanger})
}

func (a *app) settingsVerificationSection() *ui.Element {
	return ui.Div().Class("app-col").Children(
		ui.Div().Class("app-row app-row--between").Children(
			ui.Text("Verification").Class("app-section-label"),
			chadcn.Badge("2FA required", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
		),
		chadcn.Label("Authenticator code"),
		chadcn.InputOTP(chadcn.InputOTPOptions{
			ID:     "settings-otp",
			Label:  "Authenticator code",
			Length: settingsOTPLength,
			Value:  a.otp,
		}),
		ui.Text("Six digits from the on-call authenticator. Save to verify.").Class("app-muted"),
		chadcn.Label("Handover notes", chadcn.LabelOptions{For: "settings-notes"}),
		chadcn.TextareaWithOptions("settings-notes", a.notes, chadcn.TextareaOptions{
			Placeholder: "Escalate to #incident-room if the pager is not acked in five minutes.",
			Rows:        settingsNotesRows,
		}),
	)
}
