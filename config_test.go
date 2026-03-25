package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := Config{
		IP:  "192.168.1.100",
		MAC: "AA:BB:CC:DD:EE:FF",
	}

	if err := SaveConfigTo(cfg, path); err != nil {
		t.Fatalf("SaveConfigTo: %v", err)
	}

	loaded, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	if loaded.IP != cfg.IP {
		t.Errorf("IP = %q, want %q", loaded.IP, cfg.IP)
	}
	if loaded.MAC != cfg.MAC {
		t.Errorf("MAC = %q, want %q", loaded.MAC, cfg.MAC)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	cfg, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if cfg.IP != "" || cfg.MAC != "" {
		t.Errorf("expected zero config, got: %+v", cfg)
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	os.WriteFile(path, []byte("not json"), 0644)

	_, err := LoadConfigFrom(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestSaveConfigCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "config.json")

	cfg := Config{IP: "10.0.0.1"}
	if err := SaveConfigTo(cfg, path); err != nil {
		t.Fatalf("SaveConfigTo: %v", err)
	}

	loaded, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}
	if loaded.IP != "10.0.0.1" {
		t.Errorf("IP = %q, want %q", loaded.IP, "10.0.0.1")
	}
}

func TestConfigWithSoundAliases(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg := Config{
		IP: "10.0.0.5",
		SoundAliases: map[string]string{
			"living_room": "external_arc",
			"bedroom":     "bt_soundbar",
		},
	}

	if err := SaveConfigTo(cfg, path); err != nil {
		t.Fatalf("SaveConfigTo: %v", err)
	}

	loaded, err := LoadConfigFrom(path)
	if err != nil {
		t.Fatalf("LoadConfigFrom: %v", err)
	}

	if loaded.SoundAliases["living_room"] != "external_arc" {
		t.Errorf("sound alias living_room = %q, want external_arc", loaded.SoundAliases["living_room"])
	}
}

func TestResolveIP(t *testing.T) {
	cfg := Config{IP: "192.168.1.50"}

	// CLI flag takes priority
	if got := ResolveIP("10.0.0.1", cfg); got != "10.0.0.1" {
		t.Errorf("ResolveIP with flag = %q, want 10.0.0.1", got)
	}

	// Config used when no flag
	if got := ResolveIP("", cfg); got != "192.168.1.50" {
		t.Errorf("ResolveIP with config = %q, want 192.168.1.50", got)
	}

	// Env var fallback
	os.Setenv("LGTV_IP", "172.16.0.1")
	defer os.Unsetenv("LGTV_IP")
	if got := ResolveIP("", Config{}); got != "172.16.0.1" {
		t.Errorf("ResolveIP with env = %q, want 172.16.0.1", got)
	}
}

func TestResolveMAC(t *testing.T) {
	cfg := Config{MAC: "AA:BB:CC:DD:EE:FF"}

	if got := ResolveMAC(cfg); got != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("ResolveMAC with config = %q, want AA:BB:CC:DD:EE:FF", got)
	}

	os.Setenv("LGTV_MAC", "11:22:33:44:55:66")
	defer os.Unsetenv("LGTV_MAC")
	if got := ResolveMAC(Config{}); got != "11:22:33:44:55:66" {
		t.Errorf("ResolveMAC with env = %q, want 11:22:33:44:55:66", got)
	}
}

func TestLoadClientKeyFromConfig(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key.txt")
	os.WriteFile(keyPath, []byte("mykey123\n"), 0600)

	// CLI flag path
	key, err := LoadClientKeyFromConfig(keyPath, Config{})
	if err != nil {
		t.Fatalf("LoadClientKeyFromConfig: %v", err)
	}
	if key != "mykey123" {
		t.Errorf("key = %q, want mykey123", key)
	}

	// Config key_file path
	key, err = LoadClientKeyFromConfig("", Config{KeyFile: keyPath})
	if err != nil {
		t.Fatalf("LoadClientKeyFromConfig: %v", err)
	}
	if key != "mykey123" {
		t.Errorf("key = %q, want mykey123", key)
	}
}
