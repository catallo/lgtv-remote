package main

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"
)

// SendWOL sends a Wake-on-LAN magic packet to the given MAC address.
func SendWOL(macAddr string) error {
	mac, err := parseMAC(macAddr)
	if err != nil {
		return fmt.Errorf("invalid MAC address %q: %w", macAddr, err)
	}

	// Magic packet: 6 bytes of 0xFF followed by MAC repeated 16 times
	packet := make([]byte, 102)
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:], mac)
	}

	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 9,
	})
	if err != nil {
		return fmt.Errorf("creating UDP connection: %w", err)
	}
	defer conn.Close()

	_, err = conn.Write(packet)
	if err != nil {
		return fmt.Errorf("sending magic packet: %w", err)
	}
	return nil
}

func parseMAC(s string) ([]byte, error) {
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 12 {
		return nil, fmt.Errorf("expected 12 hex digits, got %d", len(s))
	}
	return hex.DecodeString(s)
}
