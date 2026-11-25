package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// CondRule is one conditional formatting rule.
type CondRule struct {
	// Kind: "gt","lt","eq","between","text","dupes","top","bottom","above",
	// "below","databar","scale3".
	Kind string
	// A and B are rule arguments: a threshold (gt/lt/eq), bounds (between),
	// the substring (text) or the count N (top and bottom, default 10).
	A, B string
	// Preset: "red","yellow","green" (default red; databar default blue).
	Preset string
}

type xlCondEntry struct {
	rng  Range
	rule CondRule
	off  bool

	ver    int // sheet version the cache was built for
	lo, hi float64
	mid    float64
	counts map[string]int
	thresh float64
	// hasNums records that the range held numbers (above/below average).
	hasNums bool
}

func (s *Sheet) AddCondFormat(r Range, rule CondRule) {
	s.conds = append(s.conds, &xlCondEntry{rng: r, rule: rule, ver: -1})
	s.ver++
}

// CondInfo describes one stored rule for the Manage Rules dialog.
type CondInfo struct {
	Range Range
	Rule  CondRule
	Off   bool
}

// CondRules lists the rules in priority order, including disabled ones.
func (s *Sheet) CondRules() []CondInfo {
	out := make([]CondInfo, len(s.conds))
	for i, e := range s.conds {
		out[i] = CondInfo{e.rng, e.rule, e.off}
	}
	return out
}

// SetCondOff enables or disables rule i.
func (s *Sheet) SetCondOff(i int, off bool) {
	if i >= 0 && i < len(s.conds) {
		s.conds[i].off = off
		s.ver++
	}
}

// DeleteCond removes rule i.
func (s *Sheet) DeleteCond(i int) {
	if i >= 0 && i < len(s.conds) {
		s.conds = append(s.conds[:i], s.conds[i+1:]...)
		s.ver++
	}
}

// MoveCondUp swaps rule i with the one before it.
func (s *Sheet) MoveCondUp(i int) {
	if i > 0 && i < len(s.conds) {
		s.conds[i-1], s.conds[i] = s.conds[i], s.conds[i-1]
		s.ver++
	}
}

// ClearCondFormat removes every rule whose range intersects r.
func (s *Sheet) ClearCondFormat(r Range) {
	keep := s.conds[:0]
	for _, e := range s.conds {
		if !e.rng.Intersects(r) {
			keep = append(keep, e)
		}
	}
	s.conds = keep
	s.ver++
}

// CondFormats lists the rules and their ranges.
func (s *Sheet) CondFormats() (rs []Range, rules []CondRule) {
	for _, e := range s.conds {
		rs = append(rs, e.rng)
		rules = append(rules, e.rule)
	}
	return rs, rules
}

type xlPreset struct{ fg, bg, bar string }

func xlPresets() map[string]xlPreset {
	return map[string]xlPreset{
		"red":    {"#9C0006", "#FFC7CE", "#F8696B"},
		"yellow": {"#9C5700", "#FFEB9C", "#FFB628"},
		"green":  {"#006100", "#C6EFCE", "#63BE7B"},
	}
}

func xlPresetOf(name, def string) xlPreset {
	if p, ok := xlPresets()[name]; ok {
		return p
	}
	return xlPresets()[def]
}

// refresh rebuilds the per-rule statistics when the sheet changed.
func (s *Sheet) refresh(e *xlCondEntry) {
	if e.ver == s.ver {
		return
	}
	e.ver = s.ver
	e.lo, e.hi = math.Inf(1), math.Inf(-1)
	e.counts = nil
	var nums []float64
	r := e.rng.clip()
	u := s.UsedRange()
	r.Max.Row, r.Max.Col = min(r.Max.Row, u.Max.Row), min(r.Max.Col, u.Max.Col)
	for row := r.Min.Row; row <= r.Max.Row; row++ {
		for col := r.Min.Col; col <= r.Max.Col; col++ {
			v := s.cellVal(CellRef{row, col})
			switch e.rule.Kind {
			case "dupes":
				if v.Kind != KindEmpty {
					if e.counts == nil {
						e.counts = map[string]int{}
					}
					e.counts[strings.ToLower(v.String())]++
				}
			default:
				if v.Kind == KindNumber {
					nums = append(nums, v.Num)
					e.lo, e.hi = math.Min(e.lo, v.Num), math.Max(e.hi, v.Num)
				}
			}
		}
	}
	sort.Float64s(nums)
	if len(nums) > 0 {
		e.mid = nums[len(nums)/2]
		if len(nums)%2 == 0 {
			e.mid = (nums[len(nums)/2-1] + nums[len(nums)/2]) / xlMidpointDiv
		}
	}
	switch e.rule.Kind {
	case "top", "bottom":
		n, err := strconv.Atoi(strings.TrimSpace(e.rule.A))
		if err != nil || n < 1 {
			n = 10
		}
		n = min(n, len(nums))
		switch {
		case n == 0 && e.rule.Kind == "top":
			e.thresh = math.Inf(1)
		case n == 0:
			e.thresh = math.Inf(-1)
		case e.rule.Kind == "top":
			e.thresh = nums[len(nums)-n]
		default:
			e.thresh = nums[n-1]
		}
	case "above", "below":
		e.thresh = 0
		for _, f := range nums {
			e.thresh += f
		}
		if len(nums) > 0 {
			e.thresh /= float64(len(nums))
		}
		e.counts = nil
		e.hasNums = len(nums) > 0
	}
}

func xlCondNum(s string) (float64, bool) {
	return xlParseNumber(strings.TrimSpace(s))
}

// condHit evaluates a highlighting rule for a value.
func (s *Sheet) condHit(e *xlCondEntry, v Value) bool {
	switch e.rule.Kind {
	case "gt", "lt", "eq":
		if v.Kind != KindNumber {
			if e.rule.Kind == "eq" && v.Kind == KindString {
				return strings.EqualFold(v.Str, strings.TrimSpace(e.rule.A))
			}
			return false
		}
		a, ok := xlCondNum(e.rule.A)
		if !ok {
			return false
		}
		switch e.rule.Kind {
		case "gt":
			return v.Num > a
		case "lt":
			return v.Num < a
		}
		return v.Num == a
	case "between":
		a, ok1 := xlCondNum(e.rule.A)
		b, ok2 := xlCondNum(e.rule.B)
		if v.Kind != KindNumber || !ok1 || !ok2 {
			return false
		}
		if a > b {
			a, b = b, a
		}
		return v.Num >= a && v.Num <= b
	case "text":
		return e.rule.A != "" && v.Kind != KindEmpty && strings.Contains(strings.ToLower(v.String()), strings.ToLower(e.rule.A))
	case "dupes":
		return v.Kind != KindEmpty && e.counts[strings.ToLower(v.String())] > 1
	case "top":
		return v.Kind == KindNumber && v.Num >= e.thresh
	case "bottom":
		return v.Kind == KindNumber && v.Num <= e.thresh
	case "above":
		return v.Kind == KindNumber && e.hasNums && v.Num > e.thresh
	case "below":
		return v.Kind == KindNumber && e.hasNums && v.Num < e.thresh
	}
	return false
}

// StyleAt returns the base style merged with conditional formatting.
func (s *Sheet) StyleAt(row, col int) Style {
	ref := CellRef{row, col}
	st := s.styles[ref]
	if len(s.conds) == 0 {
		return st
	}
	v := s.cellVal(ref)
	for _, e := range s.conds {
		if e.off || !e.rng.Contains(ref) {
			continue
		}
		s.refresh(e)
		switch e.rule.Kind {
		case "databar":
		case "scale3":
			if v.Kind == KindNumber && e.hi > e.lo {
				st.BG = xlScale3(v.Num, e.lo, e.mid, e.hi)
			}
		default:
			if s.condHit(e, v) {
				p := xlPresetOf(e.rule.Preset, "red")
				st.FG, st.BG = p.fg, p.bg
			}
		}
	}
	return st
}

// xlScale3 maps a value onto a red-yellow-green three-color scale.
func xlScale3(v, lo, mid, hi float64) string {
	lerp := func(a, b [3]float64, t float64) string {
		return fmt.Sprintf("#%02X%02X%02X", int(a[0]+(b[0]-a[0])*t), int(a[1]+(b[1]-a[1])*t), int(a[2]+(b[2]-a[2])*t))
	}
	red, yellow, green := [3]float64{248, 105, 107}, [3]float64{255, 235, 132}, [3]float64{99, 190, 123}
	if v <= mid {
		if mid == lo {
			return lerp(yellow, yellow, 0)
		}
		return lerp(red, yellow, (v-lo)/(mid-lo))
	}
	if hi == mid {
		return lerp(green, green, 0)
	}
	return lerp(yellow, green, (v-mid)/(hi-mid))
}

// DataBar returns the bar fill fraction (0..1) and color for a cell covered by a
// "databar" rule.
func (s *Sheet) DataBar(row, col int) (frac float64, color string, ok bool) {
	ref := CellRef{row, col}
	v := s.cellVal(ref)
	if v.Kind != KindNumber {
		return 0, "", false
	}
	for i := len(s.conds) - 1; i >= 0; i-- {
		e := s.conds[i]
		if e.off || e.rule.Kind != "databar" || !e.rng.Contains(ref) {
			continue
		}
		s.refresh(e)
		lo := math.Min(0, e.lo)
		if e.hi <= lo {
			return 0, xlPresetOfBar(e.rule.Preset), true
		}
		return math.Max(0, math.Min(1, (v.Num-lo)/(e.hi-lo))), xlPresetOfBar(e.rule.Preset), true
	}
	return 0, "", false
}

func xlPresetOfBar(name string) string {
	if p, ok := xlPresets()[name]; ok {
		return p.bar
	}
	return "#638EC6"
}
