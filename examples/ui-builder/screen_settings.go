package main

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// settingsState holds the settings tab's form values, inline validation
// errors, and whether the last save succeeded.
// It is stored on builderApp as a.settings by the app owner.
type settingsState struct {
	Name   string
	Email  string
	Nick   string
	Role   string
	Notify bool
	Errors map[string]string
	Saved  bool
}

func defaultSettingsState() settingsState {
	return settingsState{
		Name:   "Ada Lovelace",
		Email:  "ada@example.com",
		Nick:   "",
		Role:   "editor",
		Notify: true,
	}
}

// settingsMaxNickRunes bounds the optional nickname field.
const settingsMaxNickRunes = 12

// settingsRoles lists the roles offered by the role select row.
func settingsRoles() []string {
	return []string{"viewer", "editor", "admin"}
}

func (a *builderApp) settingsPanel() *ui.Element {
	return ui.Section().Flex(ui.RawValue("1")).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).
		Children(
			ui.Div().Children(
				a.panelHeading("Settings"),
				ui.P().Text("A form built purely with the ui builder.").
					Margin(ui.RawValue("7px 0 0")).Color(ui.RawValue("var(--of-muted,#a9bbd1)")).FontSize(ui.Px(12)),
				ui.P().Text("Click a field and type — Save reads the live control values.").
					Margin(ui.RawValue("4px 0 0")).Color(ui.RawValue("var(--of-muted,#a9bbd1)")).FontSize(ui.Px(12)),
			),
			ui.Div().Padding(ui.Px(16)).BorderRadius(ui.Px(11)).BackgroundColor(ui.RawValue("var(--of-surface,#202d42)")).
				Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(12)).
				Children(
					a.settingsField("Name", "settings-name", "name", "text", a.settings.Name, "Ada Lovelace"),
					a.settingsField("Email", "settings-email", "email", "text", a.settings.Email, "ada@example.com"),
					a.settingsField("Nickname (optional, max 12 chars)", "settings-nick", "nick", "text", a.settings.Nick, "ada"),
					a.settingsRoleRow(),
					a.settingsNotifyRow(),
					ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).Children(
						a.settingsFormButton("Save", "settings:save", true),
						a.settingsFormButton("Reset", "settings:reset", false),
					),
					a.settingsNotice(),
				),
		)
}

func (a *builderApp) settingsField(label, id, name, inputType, value, placeholder string) *ui.Element {
	row := []*ui.Element{
		ui.Text(label).FontSize(ui.Px(13)).FontWeight(ui.FontWeightBold).Color(ui.RawValue("var(--of-muted,#b5c7df)")),
		ui.Tag("input").ID(id).Attr("type", inputType).Attr("name", name).
			Attr("value", value).Attr("placeholder", placeholder).
			Width(ui.Percent(100)).BoxSizing(ui.BoxSizingBorderBox).
			Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.Px(8)).
			BackgroundColor(ui.RawValue("var(--of-control,#26364e)")).Color(ui.RawValue("var(--of-text,#ecf2fb)")).
			Border(ui.RawValue("1px solid var(--of-border,#3e526e)")).FontSize(ui.Px(13)),
	}
	if msg := a.settings.Errors[id]; msg != "" {
		row = append(row, ui.Text(msg).FontSize(ui.Px(12)).Color(ui.Hex("#ff8a8a")))
	}
	return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(6)).Children(row...)
}

func (a *builderApp) settingsRoleRow() *ui.Element {
	roles := settingsRoles()
	buttons := make([]*ui.Element, 0, len(roles))
	for _, role := range roles {
		selected := a.settings.Role == role
		buttons = append(buttons, a.settingsRoleButton(role, selected))
	}
	row := []*ui.Element{
		ui.Text("Role").FontSize(ui.Px(13)).FontWeight(ui.FontWeightBold).Color(ui.RawValue("var(--of-muted,#b5c7df)")),
		ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).Children(buttons...),
	}
	if msg := a.settings.Errors["role"]; msg != "" {
		row = append(row, ui.Text(msg).FontSize(ui.Px(12)).Color(ui.Hex("#ff8a8a")))
	}
	return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(6)).Children(row...)
}

func (a *builderApp) settingsRoleButton(role string, selected bool) *ui.Element {
	bg, fg, border := "#2b3950", "#dce6f6", "1px solid #42536b"
	label := "○ " + role
	if selected {
		bg, fg, border = "#355596", "#ffffff", "1px solid #5b7cc4"
		label = "● " + role
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", "settings:role:"+role).
		BackgroundColor(ui.RawValue(bg)).Color(ui.RawValue(fg)).Border(ui.RawValue(border)).
		Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.Px(8)).
		FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold)
}

func (a *builderApp) settingsNotifyRow() *ui.Element {
	label := "☐ Notifications off"
	if a.settings.Notify {
		label = "☑ Notifications on"
	}
	bg := "#2b3950"
	if a.settings.Notify {
		bg = "#355596"
	}
	return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(6)).Children(
		ui.Text("Notifications").FontSize(ui.Px(13)).FontWeight(ui.FontWeightBold).Color(ui.RawValue("var(--of-muted,#b5c7df)")),
		ui.Button(label).Attr("type", "button").Attr("data-action", "settings:notify").
			AlignSelf(ui.RawValue("flex-start")).
			BackgroundColor(ui.RawValue(bg)).Color(ui.RawValue("var(--of-text,#f1f5fd)")).
			Border(ui.RawValue("1px solid var(--of-border,#40546f)")).
			Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.Px(8)).FontWeight(ui.FontWeightBold),
	)
}

func (a *builderApp) settingsFormButton(label, action string, primary bool) *ui.Element {
	bg, fg, border := "#2b3950", "#dce6f6", "1px solid #42536b"
	if primary {
		bg, fg, border = "#4677ed", "#ffffff", "1px solid #4677ed"
	}
	return ui.Button(label).Attr("type", "button").Attr("data-action", action).
		BackgroundColor(ui.RawValue(bg)).Color(ui.RawValue(fg)).Border(ui.RawValue(border)).
		Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.Px(8)).
		FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold)
}

func (a *builderApp) settingsNotice() *ui.Element {
	if a.settings.Saved {
		return ui.Text("Settings saved (demo — nothing was written).").FontSize(ui.Px(12)).Color(ui.Hex("#7ee2a8"))
	}
	return ui.Text("Type in the fields above, then Save — validation runs on the live control values.").
		FontSize(ui.Px(11)).Color(ui.RawValue("var(--of-muted,#a0b6d0)"))
}

// handleSettings handles settings:* actions, updating a.settings and a.notice.
// It reports whether the action was consumed.
func (a *builderApp) handleSettings(action string) bool {
	switch action {
	case "settings:save":
		a.settings.Name = strings.TrimSpace(a.page.Form().FormValue("settings-name"))
		a.settings.Email = strings.TrimSpace(a.page.Form().FormValue("settings-email"))
		a.settings.Nick = strings.TrimSpace(a.page.Form().FormValue("settings-nick"))
		a.settings.Errors = map[string]string{}
		if a.settings.Name == "" {
			a.settings.Errors["settings-name"] = "Name must not be empty."
		}
		if !strings.Contains(a.settings.Email, "@") {
			a.settings.Errors["settings-email"] = "Email must contain @."
		}
		if len([]rune(a.settings.Nick)) > settingsMaxNickRunes {
			a.settings.Errors["settings-nick"] = "Nickname must be 12 characters or fewer."
		}
		if len(a.settings.Errors) > 0 {
			a.settings.Saved = false
			a.notice = "Settings have validation errors — fix the highlighted fields."
			return true
		}
		a.settings.Saved = true
		a.notice = "Settings saved (demo — nothing was written)."
		return true
	case "settings:reset":
		a.settings = defaultSettingsState()
		a.page.Form().SetFormValue("settings-name", a.settings.Name)
		a.page.Form().SetFormValue("settings-email", a.settings.Email)
		a.page.Form().SetFormValue("settings-nick", a.settings.Nick)
		a.notice = "Settings reset to defaults."
		return true
	case "settings:notify":
		a.settings.Notify = !a.settings.Notify
		a.settings.Saved = false
		if a.settings.Notify {
			a.notice = "Notifications turned on."
		} else {
			a.notice = "Notifications turned off."
		}
		return true
	}
	if role, ok := strings.CutPrefix(action, "settings:role:"); ok {
		for _, valid := range settingsRoles() {
			if role == valid {
				a.settings.Role = role
				a.settings.Saved = false
				a.notice = "Role set to " + role + "."
				return true
			}
		}
		return true
	}
	return false
}
