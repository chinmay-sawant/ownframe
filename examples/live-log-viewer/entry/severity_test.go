package entry

import "testing"

func TestParseSeverity(t *testing.T) {
	cases := []struct {
		in   string
		want Severity
	}{
		{"info", Info},
		{"[WARN]", Warn},
		{"warning", Warn},
		{"ERR", Error},
		{"fatal", Fatal},
		{"panic", Fatal},
		{"verbose", Trace},
		{"", Unknown},
		{"banana", Unknown},
	}

	for _, c := range cases {
		if got := ParseSeverity(c.in); got != c.want {
			t.Errorf("ParseSeverity(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestSeverityOrderAndLabels(t *testing.T) {
	if Unknown >= Trace || Trace >= Debug || Debug >= Info ||
		Info >= Warn || Warn >= Error || Error >= Fatal {
		t.Fatal("severity order is not ascending")
	}

	labels := map[Severity]string{
		Unknown: "UNKNOWN", Trace: "TRACE", Debug: "DEBUG", Info: "INFO",
		Warn: "WARN", Error: "ERROR", Fatal: "FATAL",
	}

	for sev, want := range labels {
		if got := sev.String(); got != want {
			t.Errorf("%d.String() = %q, want %q", sev, got, want)
		}
	}

	if got := Severity(99).String(); got != "UNKNOWN" {
		t.Errorf("out-of-range severity = %q", got)
	}
}
