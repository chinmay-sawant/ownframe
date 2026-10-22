package remote

import "strings"

const faceSpec = `
power|Power|wide
wake|Wake|wide
voldn|Vol-
volup|Vol+
mute|Mute
chdn|Ch-
chup|Ch+
up|Up
left|Left
ok|OK
right|Right
down|Down
back|Back
home|Home
exit|Exit
input|Input
n1|1
n2|2
n3|3
n4|4
n5|5
n6|6
n7|7
n8|8
n9|9
n0|0
play|Play
pause|Pause
stop|Stop
rew|Rew
ff|FF
red|Red|c-red
green|Green|c-green
yellow|Yellow|c-yellow
blue|Blue|c-blue
hotstar|Hotstar
prime|Prime
youtube|YouTube
guide|Guide
info|Info
menu|Menu
`

const moreSpec = `
cc|CC
list|List
qmenu|Q.Menu
dash|Dash
star|*
record|Rec
sap|SAP
ad|AD
aspect|Aspect
threed|3D
apps|Apps
recent|Recent
zoom|Zoom
text|Text
screenoff|Screen off
screenon|Screen on
livetv|Live TV
hdmi1|HDMI 1
hdmi2|HDMI 2
hdmi3|HDMI 3
magnify|Magnify
program|Program
`

func faceKeys() []Key { return keysFrom(faceSpec) }

func moreKeys() []Key { return keysFrom(moreSpec) }

func keysFrom(spec string) []Key {
	out := []Key{}

	for _, line := range strings.Split(spec, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		key := Key{ID: parts[0], Label: parts[1]}
		if len(parts) > maxKeyParts {
			key.Class = parts[2]
		}

		out = append(out, key)
	}

	return out
}
