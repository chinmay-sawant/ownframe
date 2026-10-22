package remote

import "strconv"

// btCommand is the HID report the phone should send.
// key is a keyboard usage. con is a consumer-control usage.
// Controls that have no HID usage return false and stay on Wi-Fi.
func btCommand(id string) (string, bool) {
	if code, ok := btKey[id]; ok {
		return "key:" + strconv.Itoa(code), true
	}

	if code, ok := btCon[id]; ok {
		return "con:" + strconv.Itoa(code), true
	}

	return "", false
}

var btKey = map[string]int{
	"up":    hidUp,
	"down":  hidDown,
	"left":  hidLeft,
	"right": hidRight,
	"ok":    hidOk,
	"back":  hidBack,
	"exit":  hidBack,
	"n1":    hidN1,
	"n2":    hidN2,
	"n3":    hidN3,
	"n4":    hidN4,
	"n5":    hidN5,
	"n6":    hidN6,
	"n7":    hidN7,
	"n8":    hidN8,
	"n9":    hidN9,
	"n0":    hidN0,
}

var btCon = map[string]int{
	"volup":  hidVolUp,
	"voldn":  hidVolDn,
	"mute":   hidMute,
	"chup":   hidChUp,
	"chdn":   hidChDn,
	"play":   hidPlay,
	"pause":  hidPause,
	"stop":   hidStop,
	"rew":    hidRew,
	"ff":     hidFf,
	"power":  hidPower,
	"home":   hidHome,
	"menu":   hidMenu,
	"record": hidRec,
}
