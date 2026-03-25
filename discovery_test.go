package main

import (
	"strings"
	"testing"
)

func TestParseSSDPResponse(t *testing.T) {
	response := "HTTP/1.1 200 OK\r\n" +
		"CACHE-CONTROL: max-age=1800\r\n" +
		"ST: urn:lge-com:service:webos-second-screen:1\r\n" +
		"USN: uuid:abcd-1234::urn:lge-com:service:webos-second-screen:1\r\n" +
		"LOCATION: http://192.168.1.100:3000/\r\n" +
		"SERVER: WebOS/4.0\r\n" +
		"DLNADeviceName.lge.com: [LG] webOS TV OLED55C1\r\n" +
		"\r\n"

	headers := parseSSDPResponse(response)

	if headers["st"] != "urn:lge-com:service:webos-second-screen:1" {
		t.Errorf("ST = %q", headers["st"])
	}

	if headers["location"] != "http://192.168.1.100:3000/" {
		t.Errorf("Location = %q", headers["location"])
	}

	if headers["server"] != "WebOS/4.0" {
		t.Errorf("Server = %q", headers["server"])
	}
}

func TestParseSSDPResponseCaseInsensitive(t *testing.T) {
	response := "HTTP/1.1 200 OK\r\n" +
		"St: urn:lge-com:service:webos-second-screen:1\r\n" +
		"Location: http://10.0.0.5:3000/\r\n" +
		"\r\n"

	headers := parseSSDPResponse(response)

	if headers["st"] != "urn:lge-com:service:webos-second-screen:1" {
		t.Errorf("ST = %q", headers["st"])
	}
	if headers["location"] != "http://10.0.0.5:3000/" {
		t.Errorf("Location = %q", headers["location"])
	}
}

func TestSSDPMSearchMessage(t *testing.T) {
	msg := string(ssdpMSearchMessage())

	if !strings.Contains(msg, "M-SEARCH") {
		t.Error("missing M-SEARCH method")
	}
	if !strings.Contains(msg, "239.255.255.250:1900") {
		t.Error("missing multicast address")
	}
	if !strings.Contains(msg, "ssdp:discover") {
		t.Error("missing ssdp:discover")
	}
	if !strings.Contains(msg, "lge-com") {
		t.Error("missing LG service type")
	}
	if !strings.Contains(msg, "webos-second-screen") {
		t.Error("missing webos-second-screen")
	}
}

func TestParseSSDPResponseEmpty(t *testing.T) {
	headers := parseSSDPResponse("")
	if len(headers) != 0 {
		t.Errorf("expected empty map, got %d entries", len(headers))
	}
}

func TestParseSSDPResponseNoColon(t *testing.T) {
	response := "HTTP/1.1 200 OK\r\nNO-COLON-LINE\r\n"
	headers := parseSSDPResponse(response)
	// The HTTP status line has a colon in "HTTP/1.1 200 OK" — no, it doesn't have one at all
	// "NO-COLON-LINE" has no colon, should be skipped
	if _, ok := headers["no-colon-line"]; ok {
		t.Error("should not parse line without colon")
	}
}
