package mobile

import "github.com/chinmay-sawant/ownframe/examples/pet-care/app"

var application *app.App

func Back() bool { return application != nil && application.RequestBack() }

func CancelTouches() {
	if application != nil {
		application.Page().Edit().CancelTouches()
	}
}

func DarkTheme() bool { return application != nil && application.DarkTheme() }

func PickerRequest() string {
	if application == nil {
		return ""
	}
	return application.TakePickerRequest()
}

func BackupData() []byte {
	if application == nil {
		return nil
	}
	return application.TakeBackupData()
}

func SetBackup(data []byte) {
	if application != nil {
		application.QueueBackupImport(data)
	}
}
