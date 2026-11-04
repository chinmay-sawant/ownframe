package main

import "strings"

// navID normalizes a sidebar label (or raw tab fragment) to a canonical tab
// ID. Canonical IDs: overview, audience, campaigns, content, engagement,
// revenue, subscribers, automations, segments, integrations, analytics,
// reports, learn, help, settings.
func navID(label string) string {
	s := strings.ToLower(strings.TrimSpace(label))
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.Join(strings.Fields(s), " ")
	switch s {
	case "help center", "helpcenter", "help":
		return "help"
	case "system settings", "systemsettings", "settings", "setting":
		return "settings"
	}
	return strings.ReplaceAll(s, " ", "")
}

// validTabs is the set of canonical tab IDs contentBody understands.
func validTabs() map[string]bool {
	return map[string]bool{
		"overview":     true,
		"audience":     true,
		"campaigns":    true,
		"content":      true,
		"engagement":   true,
		"revenue":      true,
		"subscribers":  true,
		"automations":  true,
		"segments":     true,
		"integrations": true,
		"analytics":    true,
		"reports":      true,
		"learn":        true,
		"help":         true,
		"settings":     true,
	}
}

// tabFromAction maps a click action to a canonical tab ID. It accepts the
// live "nav:<label>" form the sidebar emits (labels are lowercase display
// names such as "nav:help center" or "nav:system settings"), the legacy
// "nav-<id>" form stored in navSections data (e.g. "nav-help"), and bare
// labels or IDs. Unknown tabs yield "" so the caller can ignore the click.
func tabFromAction(action string) string {
	s := strings.TrimSpace(action)
	switch {
	case strings.HasPrefix(s, "nav:"):
		s = strings.TrimPrefix(s, "nav:")
	case strings.HasPrefix(s, "nav-"):
		s = strings.TrimPrefix(s, "nav-")
	case strings.HasPrefix(s, "nav "):
		s = strings.TrimPrefix(s, "nav ")
	case strings.HasPrefix(s, "tab:"):
		s = strings.TrimPrefix(s, "tab:")
	}
	tab := navID(s)
	if !validTabs()[tab] {
		return ""
	}
	return tab
}
