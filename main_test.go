package main

import (
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

func TestParseMAC(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantErr bool
	}{
		{"b0:37:95:81:74:c0", "b037958174c0", false},
		{"B0-37-95-81-74-C0", "b037958174c0", false},
		{"b037958174c0", "b037958174c0", false},
		{"AA:BB:CC:DD:EE:FF", "aabbccddeeff", false},
		{"invalid", "", true},
		{"b0:37:95:81:74", "", true},
	}
	for _, tt := range tests {
		got, err := parseMAC(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseMAC(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if err == nil && hex.EncodeToString(got) != tt.want {
			t.Errorf("parseMAC(%q) = %x, want %s", tt.input, got, tt.want)
		}
	}
}

func TestBuildRegistrationMessage(t *testing.T) {
	msg := buildRegistrationMessage("testkey123")

	if msg["type"] != "register" {
		t.Errorf("type = %v, want register", msg["type"])
	}

	payload, ok := msg["payload"].(map[string]interface{})
	if !ok {
		t.Fatal("payload is not a map")
	}

	if payload["client-key"] != "testkey123" {
		t.Errorf("client-key = %v, want testkey123", payload["client-key"])
	}

	if payload["pairingType"] != "PROMPT" {
		t.Errorf("pairingType = %v, want PROMPT", payload["pairingType"])
	}

	manifest, ok := payload["manifest"].(map[string]interface{})
	if !ok {
		t.Fatal("manifest is not a map")
	}

	perms, ok := manifest["permissions"].([]string)
	if !ok {
		t.Fatal("permissions is not []string")
	}
	if len(perms) == 0 {
		t.Error("permissions is empty")
	}
}

func TestBuildRegistrationMessageEmptyKey(t *testing.T) {
	msg := buildRegistrationMessage("")
	payload := msg["payload"].(map[string]interface{})
	if payload["client-key"] != "" {
		t.Errorf("client-key = %v, want empty string", payload["client-key"])
	}
}

func TestSSAPMessageJSON(t *testing.T) {
	msg := SSAPMessage{
		Type:    "request",
		ID:      "msg_1",
		URI:     "ssap://audio/setVolume",
		Payload: map[string]int{"volume": 50},
	}

	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	s := string(b)
	if !strings.Contains(s, `"type":"request"`) {
		t.Error("missing type field")
	}
	if !strings.Contains(s, `"uri":"ssap://audio/setVolume"`) {
		t.Error("missing uri field")
	}
	if !strings.Contains(s, `"volume":50`) {
		t.Error("missing volume in payload")
	}
}

func TestResolveSoundAlias(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"tv", "tv_speaker"},
		{"speaker", "tv_speaker"},
		{"TV_SPEAKER", "tv_speaker"},
		{"headphone", "headphone"},
		{"headphones", "headphone"},
		{"wired", "headphone"},
		{"lautsprecher", "headphone"},
		{"arc", "external_arc"},
		{"hdmi_arc", "external_arc"},
		{"earc", "external_arc"},
		{"bt", "bt_soundbar"},
		{"bluetooth", "bt_soundbar"},
		{"optical", "external_optical"},
		{"some_custom_output", "some_custom_output"},
	}
	for _, tt := range tests {
		got := resolveSoundAlias(tt.input)
		if got != tt.want {
			t.Errorf("resolveSoundAlias(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
