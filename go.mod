module github.com/chinmay-sawant/ownframe

go 1.26.10

require (
	github.com/chinmay-sawant/blinkless v0.0.4-0.20261024120823-8d4693ea24ec
	github.com/hajimehoshi/ebiten/v2 v2.10.4
	golang.org/x/image v0.45.0
	golang.org/x/text v0.43.0
	golang.org/x/time v0.14.0
)

require (
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/telemetry v0.0.0-20261010193011-f71dc92da45e // indirect
	golang.org/x/tools v0.52.0 // indirect
	golang.org/x/vuln v1.9.0 // indirect
)

require (
	// Upstream tracking: ancient indirects below (freetype-go and
	// graphics-go 2016, xgb 2021, poly2tri-go 2017, freetype 2017,
	// rasterx 2022, scanx 2019) have no in-repo upgrade path. Track
	// upstream Ebiten and blinkless bumps; pinned hashes limit tampering.
	codeberg.org/go-pdf/fpdf v0.11.1 // indirect
	github.com/BurntSushi/freetype-go v0.0.0-20160129220410-b763ddbfe298 // indirect
	github.com/BurntSushi/graphics-go v0.0.0-20160129215708-b43f31a4a966 // indirect
	github.com/BurntSushi/xgb v0.0.0-20210121224620-deaf085860bc // indirect
	github.com/BurntSushi/xgbutil v0.0.0-20190907113008-ad855c713046 // indirect
	github.com/ByteArena/poly2tri-go v0.0.0-20170716161910-d102ad91854f // indirect
	github.com/andybalholm/brotli v1.2.1 // indirect
	github.com/benoitkugler/textlayout v0.3.2 // indirect
	github.com/benoitkugler/textprocessing v0.0.6 // indirect
	github.com/ebitengine/gomobile v0.0.0-20260820040257-d11f821a26a6 // indirect
	github.com/ebitengine/hideconsole v1.0.0 // indirect
	github.com/ebitengine/oto/v3 v3.5.0 // indirect
	github.com/ebitengine/purego v0.11.0 // indirect
	github.com/go-fonts/latin-modern v0.3.3 // indirect
	github.com/go-text/typesetting v0.3.5 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/jfreymuth/pulse v0.1.3 // indirect
	github.com/srwiley/rasterx v0.0.0-20220730225603-2ab79fcdd4ef // indirect
	github.com/srwiley/scanx v0.0.0-20190309010443-e94503791388 // indirect
	github.com/tdewolff/canvas v0.0.0-20260803134256-8e86b9abb917 // indirect
	github.com/tdewolff/font v0.0.0-20260424075104-b5eeb1e23189 // indirect
	github.com/tdewolff/minify/v2 v2.24.13 // indirect
	github.com/tdewolff/parse/v2 v2.8.12 // indirect
	github.com/yuin/goldmark v1.8.2 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0
	modernc.org/knuth v0.5.5 // indirect
	modernc.org/token v1.1.0 // indirect
	star-tex.org/x/tex v0.7.1 // indirect
)

tool golang.org/x/vuln/cmd/govulncheck
