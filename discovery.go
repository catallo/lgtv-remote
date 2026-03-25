package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	ssdpAddr    = "239.255.255.250:1900"
	ssdpTimeout = 3 * time.Second
)

// DiscoveredTV represents a TV found via SSDP.
type DiscoveredTV struct {
	IP       string
	Name     string
	Location string
}

// ssdpMSearchMessage builds an M-SEARCH request for SSDP discovery.
func ssdpMSearchMessage() []byte {
	return []byte("M-SEARCH * HTTP/1.1\r\n" +
		"HOST: 239.255.255.250:1900\r\n" +
		"MAN: \"ssdp:discover\"\r\n" +
		"MX: 3\r\n" +
		"ST: urn:lge-com:service:webos-second-screen:1\r\n" +
		"\r\n")
}

// DiscoverTVs sends SSDP M-SEARCH and returns LG webOS TVs found on the network.
func DiscoverTVs() ([]DiscoveredTV, error) {
	return DiscoverTVsWithTimeout(ssdpTimeout)
}

// DiscoverTVsWithTimeout sends SSDP M-SEARCH with a custom timeout.
func DiscoverTVsWithTimeout(timeout time.Duration) ([]DiscoveredTV, error) {
	addr, err := net.ResolveUDPAddr("udp4", ssdpAddr)
	if err != nil {
		return nil, fmt.Errorf("resolving SSDP address: %w", err)
	}

	conn, err := net.ListenUDP("udp4", nil)
	if err != nil {
		return nil, fmt.Errorf("opening UDP socket: %w", err)
	}
	defer conn.Close()

	msg := ssdpMSearchMessage()
	if _, err := conn.WriteToUDP(msg, addr); err != nil {
		return nil, fmt.Errorf("sending M-SEARCH: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(timeout))

	seen := make(map[string]bool)
	var tvs []DiscoveredTV
	buf := make([]byte, 4096)

	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			// Timeout is expected — we're done listening
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			return tvs, nil
		}

		headers := parseSSDPResponse(string(buf[:n]))
		st := headers["st"]
		if !strings.Contains(st, "lge") && !strings.Contains(st, "webos") {
			continue
		}

		ip := src.IP.String()
		if seen[ip] {
			continue
		}
		seen[ip] = true

		name := ""
		if server := headers["server"]; server != "" {
			name = server
		}
		if dlnaName := headers["dlna.devicename"]; dlnaName != "" {
			name = dlnaName
		}

		tvs = append(tvs, DiscoveredTV{
			IP:       ip,
			Name:     name,
			Location: headers["location"],
		})
	}

	return tvs, nil
}

// parseSSDPResponse parses HTTP-like SSDP headers into a map.
func parseSSDPResponse(data string) map[string]string {
	headers := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.IndexByte(line, ':'); idx > 0 {
			key := strings.ToLower(strings.TrimSpace(line[:idx]))
			val := strings.TrimSpace(line[idx+1:])
			headers[key] = val
		}
	}
	return headers
}
