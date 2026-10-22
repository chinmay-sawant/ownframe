//go:build linux

package collector

import (
	"errors"
	"strconv"
	"strings"
)

// procStat is the part of /proc/<pid>/stat this example reads. The comm field
// can hold spaces and parentheses, so parsing starts after the last ')'.
type procStat struct {
	comm       string
	state      string
	ppid       int32
	utime      uint64
	stime      uint64
	nice       int64
	threads    int
	startTicks uint64
	vsize      uint64
	rss        uint64
}

// errBadStat marks a stat line that does not match the expected shape.
var errBadStat = errors.New("procfs: unreadable stat line")

// parseProcStat parses one /proc/<pid>/stat line.
func parseProcStat(data []byte) (procStat, error) {
	s := string(data)

	open := strings.IndexByte(s, '(')
	closing := strings.LastIndexByte(s, ')')
	if open < 0 || closing <= open || closing+2 > len(s) {
		return procStat{}, errBadStat
	}

	st := procStat{comm: s[open+1 : closing]}

	fields := strings.Fields(s[closing+1:])
	if len(fields) < minStatFields {
		return procStat{}, errBadStat
	}

	st.state = fields[0]
	st.ppid = parseInt32(fields[1])
	st.utime = parseUint(fields[11])
	st.stime = parseUint(fields[12])
	st.nice = parseSigned(fields[16])
	st.threads = parseCount(fields[17])
	st.startTicks = parseUint(fields[19])
	st.vsize = parseUint(fields[20])
	st.rss = parseUint(fields[21])

	return st, nil
}

// parseUint returns zero for a missing or malformed field.
func parseUint(s string) uint64 {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0
	}

	return v
}
