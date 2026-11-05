// Voice status strip for the discord demo. The strip sits above the composer
// while the user is in a voice channel; the always-on user bar in the sidebar
// footer lives in servers.go.
//
// Actions emitted (handled by the host click dispatcher in another file):
// voice:mute, voice:deafen, voice:leave.
package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// voiceBar is the joined-channel strip with mute, deafen, and leave controls
// plus speaking dots. It returns nil while not in a voice channel.
func (a *app) voiceBar() *ui.Element {
	if a.voiceID == "" {
		return nil
	}
	name := "Voice channel"
	if ch := a.channelByID(a.serverID, a.voiceID); ch.Name != "" {
		name = ch.Name
	}
	muteLabel, muteIcon := "Mute", "◉"
	if a.voiceMuted {
		muteLabel, muteIcon = "Unmute", "⊘"
	}
	deafenLabel, deafenIcon := "Deafen", "♪"
	if a.voiceDeafened {
		deafenLabel, deafenIcon = "Undeafen", "⊘"
	}
	muted := "false"
	if a.voiceMuted {
		muted = "true"
	}
	deafened := "false"
	if a.voiceDeafened {
		deafened = "true"
	}
	mute := chadcn.Button(muteIcon, chadcn.ButtonOptions{
		Variant: chadcn.ButtonText,
		Size:    chadcn.ButtonSmall,
	}).ID("dc-voicebar-mute").Attr("data-action", "voice:mute").
		Attr("aria-label", muteLabel).
		Attr("aria-pressed", muted)
	deafen := chadcn.Button(deafenIcon, chadcn.ButtonOptions{
		Variant: chadcn.ButtonText,
		Size:    chadcn.ButtonSmall,
	}).ID("dc-voicebar-deafen").Attr("data-action", "voice:deafen").
		Attr("aria-label", deafenLabel).
		Attr("aria-pressed", deafened)
	leave := chadcn.Button("⏻", chadcn.ButtonOptions{
		Variant: chadcn.ButtonText,
		Size:    chadcn.ButtonSmall,
	}).ID("dc-voicebar-leave").Attr("data-action", "voice:leave").
		Attr("aria-label", "Leave voice channel")
	return ui.Div().Class("dc-voice").
		Attr("aria-label", "Voice status").
		Children(
			dcSpeakingDots(),
			ui.Div().Class("dc-voice-info").Children(
				ui.Text("Voice connected").Class("dc-voice-state"),
				ui.Text(name).Class("dc-voice-name"),
			),
			ui.Div().Class("dc-voice-btns").Children(
				chadcn.Tooltip(chadcn.TooltipOptions{
					ID:       "voice-mute-tip",
					CSSHover: true,
					Content:  muteLabel,
					Side:     chadcn.PopupTop,
				}, mute),
				chadcn.Tooltip(chadcn.TooltipOptions{
					ID:       "voice-deafen-tip",
					CSSHover: true,
					Content:  deafenLabel,
					Side:     chadcn.PopupTop,
				}, deafen),
				chadcn.Tooltip(chadcn.TooltipOptions{
					ID:       "voice-leave-tip",
					CSSHover: true,
					Content:  "Leave voice channel",
					Side:     chadcn.PopupTop,
				}, leave),
			),
		)
}

// dcSpeakingDots renders three green dots marking live voice activity.
func dcSpeakingDots() *ui.Element {
	return ui.Span().Class("dc-speaking").
		Attr("role", "img").
		Attr("aria-label", "Someone is speaking").
		Display(ui.DisplayFlex).
		Gap(ui.Px(2)).
		Children(
			dcSpeakingDot(),
			dcSpeakingDot(),
			dcSpeakingDot(),
		)
}

// dcSpeakingDot is one dot of the speaking indicator.
func dcSpeakingDot() *ui.Element {
	return ui.Span().Class("dc-speaking-dot").
		Attr("aria-hidden", "true").
		Display(ui.DisplayInlineBlock).
		Width(ui.Px(6)).
		Height(ui.Px(6)).
		BorderRadius(ui.Px(3)).
		BackgroundColor(ui.Hex("#23a559"))
}
