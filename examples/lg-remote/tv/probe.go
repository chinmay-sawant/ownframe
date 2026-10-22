package tv

import (
	"net"
	"sync"
)

func probeAll() []Hit {
	out := []Hit{}

	for _, ip := range localIPv4() {
		out = append(out, probe(ip)...)
	}

	return out
}

func localIPv4() []net.IP {
	out := []net.IP{}
	ifaces, err := net.Interfaces()
	if err != nil {
		return out
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipnet.IP.To4()
			if ip == nil || ip.IsLoopback() || ip[0] == 169 {
				continue
			}

			out = append(out, ip)
		}
	}

	return out
}

func probe(local net.IP) []Hit {
	var mu sync.Mutex
	var wg sync.WaitGroup
	out := []Hit{}
	sem := make(chan struct{}, probeWorkers)

	for host := 1; host < 255; host++ {
		if byte(host) == local[3] {
			continue
		}

		ip := net.IPv4(local[0], local[1], local[2], byte(host)).String()
		wg.Add(1)
		sem <- struct{}{}

		go func(ip string) {
			defer wg.Done()
			defer func() { <-sem }()

			if !portOpen(ip, "3000") {
				return
			}

			name := nameAt("http://" + ip + ":1962/")
			if name == "" || !looksLikeTV(name) {
				return
			}

			mu.Lock()
			out = append(out, Hit{IP: ip, Name: name})
			mu.Unlock()
		}(ip)
	}

	wg.Wait()

	return out
}

func portOpen(ip, port string) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), probeTimeout)
	if err != nil {
		return false
	}

	_ = conn.Close()
	return true
}
