package tv

import (
	"context"
	"errors"
	"net"
)

// Magic builds a Wake-on-LAN packet for mac.
func Magic(mac string) ([]byte, error) {
	hw, err := net.ParseMAC(mac)
	if err != nil || len(hw) != 6 {
		return nil, errors.New("the saved TV address is unusable")
	}

	pkt := make([]byte, wakePktSize)
	for i := 0; i < 6; i++ {
		pkt[i] = 0xff
	}

	for i := 1; i <= 16; i++ {
		copy(pkt[i*6:], hw)
	}

	return pkt, nil
}

// Wake sends the magic packet on the local broadcast.
func Wake(mac string) error {
	return sendWake(mac, "")
}

func sendWake(mac, host string) error {
	pkt, err := Magic(mac)
	if err != nil {
		return err
	}

	conn, err := (&net.ListenConfig{}).ListenPacket(context.Background(), "udp4", ":0")
	if err != nil {
		return errors.New("the wake packet could not be sent")
	}

	defer func() { _ = conn.Close() }()
	sent := false
	for n := 0; n < 3; n++ {
		for _, raw := range wakeTargets(host) {
			addr, err := net.ResolveUDPAddr("udp4", raw)
			if err != nil {
				continue
			}
			if _, err = conn.WriteTo(pkt, addr); err != nil {
				continue
			}
			sent = true
		}
	}

	if !sent {
		return errors.New("the wake packet could not be sent")
	}

	return nil
}

func wakeTargets(host string) []string {
	out := []string{"255.255.255.255:9", "255.255.255.255:7"}
	if b := subnet(host); b != "" {
		out = append(out, b+":9", b+":7")
	}
	if host != "" {
		out = append(out, host+":9", host+":7")
	}

	return out
}

func subnet(host string) string {
	ip := net.ParseIP(host).To4()
	if ip == nil {
		return ""
	}

	ip[3] = 255

	return ip.String()
}
