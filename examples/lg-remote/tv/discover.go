package tv

import (
	"net"
	"strings"
	"time"
)

// Hit is a TV that answered on the LAN.
type Hit struct {
	IP   string
	Name string
}

const ssdpAll = "M-SEARCH * HTTP/1.1\r\n" +
	"HOST: 239.255.255.250:1900\r\n" +
	"MAN: \"ssdp:discover\"\r\n" +
	"MX: 1\r\n" +
	"ST: ssdp:all\r\n\r\n"

const ssdpLG = "M-SEARCH * HTTP/1.1\r\n" +
	"HOST: 239.255.255.250:1900\r\n" +
	"MAN: \"ssdp:discover\"\r\n" +
	"MX: 1\r\n" +
	"ST: urn:lge-com:service:webos-second-screen:1\r\n\r\n"

// Discover asks the Wi-Fi for a webOS TV.
// A multicast search runs first. If nothing answers, it probes the local
// subnet for the TV's control port.
func Discover() []Hit {
	hits := ssdp()
	if len(hits) > 0 {
		return hits
	}

	return probeAll()
}

func ssdp() []Hit {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return []Hit{}
	}

	defer func() { _ = conn.Close() }()
	dst := &net.UDPAddr{IP: net.ParseIP("239.255.255.250"), Port: ssdpPort}
	_, _ = conn.WriteTo([]byte(ssdpLG), dst)
	_, _ = conn.WriteTo([]byte(ssdpAll), dst)
	_ = conn.SetReadDeadline(time.Now().Add(ssdpTimeout))

	found := map[string]Hit{}
	buf := make([]byte, ssdpBufSize)

	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
		}

		if addr == nil || addr.IP == nil {
			continue
		}

		text := string(buf[:n])
		if !looksLikeTV(text) {
			continue
		}

		ip := addr.IP.String()
		found[ip] = Hit{IP: ip, Name: nameFrom(text)}
	}

	out := make([]Hit, 0, len(found))
	for _, hit := range found {
		out = append(out, hit)
	}

	return out
}

func looksLikeTV(text string) bool {
	low := strings.ToLower(text)

	return strings.Contains(low, "webos") || strings.Contains(low, "lge")
}

func choose(hits []Hit) Hit {
	for _, hit := range hits {
		up := strings.ToUpper(hit.Name)
		if strings.Contains(up, "UP7750") || strings.Contains(up, "WEBOS TV") {
			return hit
		}
	}

	return hits[0]
}
