package mobile

// UsageAccessRequested returns and clears the request to open Android settings.
func UsageAccessRequested() bool {
	return app != nil && app.TakeAccessRequest()
}

// SetUsageAccess queues the system permission state.
func SetUsageAccess(allowed bool) {
	if app != nil {
		app.QueueAccess(allowed)
	}
}

// SetAppsJSON sends the app labels used by the one-app picker.
func SetAppsJSON(data string) {
	if app != nil {
		app.QueueApps(data)
	}
}

// SetUsageJSON sends the selected app's daily totals.
func SetUsageJSON(data string) {
	if app != nil {
		app.QueueUsage(data)
	}
}

// RestoreSelection restores the last app picked on this device.
func RestoreSelection(pkg string) {
	if app != nil {
		app.QueueRestore(pkg)
	}
}

// SelectedPackage returns the app chosen in the page.
func SelectedPackage() string {
	if app == nil {
		return ""
	}
	return app.SelectedPackage()
}

// Back returns to the app picker before Android closes the activity.
func Back() bool { return app != nil && app.RequestBack() }

// CancelTouches clears a pointer gesture when Android pauses the activity.
func CancelTouches() {
	if inputPage != nil {
		inputPage.CancelTouches()
	}
}
