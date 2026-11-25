package main

import (
	"math"
	"strconv"
	"strings"
)

// String renders a value the way string functions see it (General format).
func (v Value) String() string {
	switch v.Kind {
	case KindNumber:
		return xlGeneral(v.Num)
	case KindString:
		return v.Str
	case KindBool:
		if v.Bool {
			return "TRUE"
		}
		return "FALSE"
	case KindError:
		return v.Err
	}
	return ""
}

const (
	xlDecimalBase     = 10  // decimal rounding radix
	xlPercentScale    = 100 // Percent format shows the fraction times this
	xlRunTwoDigit     = 2   // dd/mm/MM style run: zero-padded
	xlRunAbbrev       = 3   // ddd/mmm style run: abbreviated name
	xlDefaultDecimals = 2   // Number, Currency, Accounting and Scientific default
	xlExpDigits       = 2   // minimum digits in a scientific exponent
	xlQuoteSkip       = 2   // a quoted literal spans its two quote marks
)

// xlClean15 drops binary noise beyond 15 significant digits.
func xlClean15(f float64) float64 {
	if c, err := strconv.ParseFloat(strconv.FormatFloat(f, 'g', 15, 64), 64); err == nil {
		return c
	}
	return f
}

// xlRound rounds half away from zero to dec decimals (dec may be negative).
func xlRound(f float64, dec int) float64 {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return f
	}
	// strip binary noise so 1.005 behaves like the typed decimal
	f = xlClean15(f)
	p := math.Pow(xlDecimalBase, float64(dec))
	return math.Round(xlClean15(f*p)) / p
}

// xlGeneral formats like Excel's General: up to 10 significant digits.
func xlGeneral(f float64) string {
	if f == 0 {
		return "0"
	}
	if math.IsNaN(f) {
		return ErrNum
	}
	if math.IsInf(f, 0) {
		return ErrNum
	}
	abs := math.Abs(f)
	if abs >= 1e11 || abs < 1e-5 {
		s := strconv.FormatFloat(f, 'E', 5, 64)
		mant, exp, _ := strings.Cut(s, "E")
		if strings.Contains(mant, ".") {
			mant = strings.TrimRight(strings.TrimRight(mant, "0"), ".")
		}
		sign := exp[0]
		digits := strings.TrimLeft(exp[1:], "0")
		if len(digits) < xlExpDigits {
			digits = strings.Repeat("0", xlExpDigits-len(digits)) + digits
		}
		return mant + "E" + string(sign) + digits
	}
	r, _ := strconv.ParseFloat(strconv.FormatFloat(f, 'g', 10, 64), 64)
	return strconv.FormatFloat(r, 'f', -1, 64)
}

func xlFixed(f float64, dec int) string {
	f = xlRound(f, dec)
	s := strconv.FormatFloat(f, 'f', max(dec, 0), 64)
	if strings.HasPrefix(s, "-") && strings.Trim(s, "-0.") == "" {
		s = s[1:]
	}
	return s
}

func xlThousands(s string) string {
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	ip, fp, hasFrac := strings.Cut(s, ".")
	var sb strings.Builder
	for i, c := range ip {
		if i > 0 && (len(ip)-i)%3 == 0 {
			sb.WriteByte(',')
		}
		sb.WriteRune(c)
	}
	out := sb.String()
	if hasFrac {
		out += "." + fp
	}
	if neg {
		out = "-" + out
	}
	return out
}

func xlDecimals(st Style, def int) int {
	switch {
	case st.Decimals < 0:
		return 0
	case st.Decimals == 0:
		return def
	}
	return st.Decimals
}

// FormatValue renders a value using the style's number format.
func FormatValue(v Value, st Style) string {
	if v.Kind != KindNumber {
		return v.String()
	}
	f := v.Num
	switch st.NumFmt {
	case "", "General", "Text":
		if st.Decimals > 0 && st.NumFmt != "Text" {
			return xlFixed(f, st.Decimals)
		}
		return xlGeneral(f)
	case "Number":
		return xlFixed(f, xlDecimals(st, xlDefaultDecimals))
	case "Currency":
		s := xlThousands(xlFixed(math.Abs(f), xlDecimals(st, xlDefaultDecimals)))
		if f < 0 && strings.Trim(s, "0.,") != "" {
			return "-$" + s
		}
		return "$" + s
	case "Accounting":
		s := xlThousands(xlFixed(math.Abs(f), xlDecimals(st, xlDefaultDecimals)))
		switch {
		case f == 0 || strings.Trim(s, "0.,") == "":
			return "$ -"
		case f < 0:
			return "($ " + s + ")"
		}
		return "$ " + s
	case "Percent":
		return xlFixed(f*xlPercentScale, xlDecimals(st, 0)) + "%"
	case "Scientific":
		return xlSci(f, xlDecimals(st, xlDefaultDecimals))
	case "Date":
		return xlSerialToTime(f).Format("2006-01-02")
	case "LongDate":
		return xlSerialToTime(f).Format("Mon, Jan 2, 2006")
	case "Time":
		return xlSerialToTime(f).Format("15:04:05")
	}
	return xlFormatCode(f, st.NumFmt)
}

func xlSci(f float64, dec int) string {
	s := strconv.FormatFloat(f, 'E', dec, 64)
	mant, exp, _ := strings.Cut(s, "E")
	sign := exp[0]
	digits := strings.TrimLeft(exp[1:], "0")
	if len(digits) < xlExpDigits {
		digits = strings.Repeat("0", xlExpDigits-len(digits)) + digits
	}
	return mant + "E" + string(sign) + digits
}

// xlFormatCode renders f with an Excel-style format code: numeric codes such as
// "#,##0.00", "0%", "$#,##0" and date codes such as "yyyy-mm-dd" or "hh:mm".
func xlFormatCode(f float64, code string) string {
	secs := strings.Split(code, ";")
	sec := secs[0]
	neg := f < 0
	if neg && len(secs) > 1 {
		sec = secs[1]
		f = -f
		neg = false
	}
	if xlIsDateCode(sec) {
		return xlFormatDate(f, sec)
	}
	// literal prefix, placeholders, literal suffix
	var prefix, body, suffix strings.Builder
	state := 0
	percent := false
	for i := 0; i < len(sec); i++ {
		c := sec[i]
		isPh := c == '0' || c == '#' || c == '?' || ((c == '.' || c == ',') && state == 1) || (c == '.' && state == 0)
		switch {
		case c == '"':
			j := strings.IndexByte(sec[i+1:], '"')
			if j < 0 {
				j = len(sec) - i - 1
			}
			lit := sec[i+1 : i+1+j]
			if state == 0 {
				prefix.WriteString(lit)
			} else {
				state = 2
				suffix.WriteString(lit)
			}
			i += j + 1
		case c == '\\' && i+1 < len(sec):
			i++
			if state == 0 {
				prefix.WriteByte(sec[i])
			} else {
				state = 2
				suffix.WriteByte(sec[i])
			}
		case c == '%':
			percent = true
			if state == 0 {
				prefix.WriteByte(c)
			} else {
				state = 2
				suffix.WriteByte(c)
			}
		case isPh && state < 2:
			state = 1
			body.WriteByte(c)
		case state == 0:
			prefix.WriteByte(c)
		default:
			state = 2
			suffix.WriteByte(c)
		}
	}
	b := body.String()
	if percent {
		f *= xlPercentScale
	}
	dec, thousands := 0, strings.Contains(b, ",")
	if _, frac, ok := strings.Cut(b, "."); ok {
		dec = len(strings.NewReplacer(",", "").Replace(frac))
	}
	intPart, _, _ := strings.Cut(b, ".")
	minInt := strings.Count(intPart, "0")
	s := xlFixed(math.Abs(f), dec)
	ip, fp, hasFrac := strings.Cut(s, ".")
	if ip == "0" && minInt == 0 {
		ip = ""
	}
	for len(ip) < minInt {
		ip = "0" + ip
	}
	if thousands {
		ip = xlThousands(ip)
	}
	out := ip
	if hasFrac {
		out += "." + fp
	}
	sign := ""
	if f < 0 || (neg && f != 0) {
		sign = "-"
	}
	if strings.Trim(s, "0.") == "" {
		sign = ""
	}
	return sign + prefix.String() + out + suffix.String()
}

func xlIsDateCode(code string) bool {
	l := strings.ToLower(code)
	for _, t := range []string{"yy", "dd", "mmm", "hh", "ss", "am/pm", "d/", "/d", "m/d", "d-m", "h:", ":m"} {
		if strings.Contains(l, t) {
			return true
		}
	}
	return false
}

// xlFormatDate renders a serial using y/m/d/h/s tokens.
func xlFormatDate(f float64, code string) string {
	t := xlSerialToTime(f)
	l := strings.ToLower(code)
	ampm := strings.Contains(l, "am/pm")
	var sb strings.Builder
	for i := 0; i < len(l); {
		run := 1
		for i+run < len(l) && l[i+run] == l[i] {
			run++
		}
		rest := l[i:]
		switch {
		case strings.HasPrefix(rest, "am/pm"):
			sb.WriteString(t.Format("PM"))
			i += 5
			continue
		case l[i] == 'y':
			if run >= xlRunAbbrev {
				sb.WriteString(t.Format("2006"))
			} else {
				sb.WriteString(t.Format("06"))
			}
		case l[i] == 'm':
			// minutes when adjacent to hours or seconds
			prev := strings.TrimRight(l[:i], " ")
			after := strings.TrimLeft(l[i+run:], " :")
			isMin := strings.HasSuffix(prev, "h:") || strings.HasSuffix(prev, "hh:") || strings.HasPrefix(after, "s")
			switch {
			case isMin && run == 1:
				sb.WriteString(strconv.Itoa(t.Minute()))
			case isMin:
				sb.WriteString(t.Format("04"))
			case run == 1:
				sb.WriteString(t.Format("1"))
			case run == xlRunTwoDigit:
				sb.WriteString(t.Format("01"))
			case run == xlRunAbbrev:
				sb.WriteString(t.Format("Jan"))
			default:
				sb.WriteString(t.Format("January"))
			}
		case l[i] == 'd':
			switch run {
			case 1:
				sb.WriteString(t.Format("2"))
			case xlRunTwoDigit:
				sb.WriteString(t.Format("02"))
			case xlRunAbbrev:
				sb.WriteString(t.Format("Mon"))
			default:
				sb.WriteString(t.Format("Monday"))
			}
		case l[i] == 'h':
			layout := "15"
			if ampm {
				layout = "3"
				if run >= xlRunTwoDigit {
					layout = "03"
				}
			} else if run == 1 {
				sb.WriteString(strconv.Itoa(t.Hour()))
				i += run
				continue
			}
			sb.WriteString(t.Format(layout))
		case l[i] == 's':
			if run == 1 {
				sb.WriteString(strconv.Itoa(t.Second()))
			} else {
				sb.WriteString(t.Format("05"))
			}
		case l[i] == '"':
			j := strings.IndexByte(l[i+1:], '"')
			if j < 0 {
				j = len(l) - i - 1
			}
			sb.WriteString(code[i+1 : i+1+j])
			i += j + xlQuoteSkip
			continue
		default:
			sb.WriteString(code[i : i+run])
		}
		i += run
	}
	return sb.String()
}
