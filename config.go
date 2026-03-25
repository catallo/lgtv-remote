package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds all lgtv configuration.
type Config struct {
	IP           string            `json:"ip,omitempty"`
	MAC          string            `json:"mac,omitempty"`
	KeyFile      string            `json:"key_file,omitempty"`
	SoundAliases map[string]string `json:"sound_aliases,omitempty"`
}

// configDir returns ~/.config/lgtv
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "lgtv-remote"), nil
}

// configPath returns the full path to config.json
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// clientKeyPath returns the full path to client_key.txt
func clientKeyPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "client_key.txt"), nil
}

// LoadConfig loads config from ~/.config/lgtv-remote/config.json.
// Returns a zero Config (not an error) if the file doesn't exist.
func LoadConfig() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}
	return LoadConfigFrom(path)
}

// LoadConfigFrom loads config from a specific path.
func LoadConfigFrom(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config %s: %w", path, err)
	}
	return cfg, nil
}

// SaveConfig writes config to ~/.config/lgtv-remote/config.json.
func SaveConfig(cfg Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	return SaveConfigTo(cfg, path)
}

// SaveConfigTo writes config to a specific path.
func SaveConfigTo(cfg Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

// SaveClientKey saves the client key to ~/.config/lgtv-remote/client_key.txt.
func SaveClientKey(key string) error {
	path, err := clientKeyPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	return os.WriteFile(path, []byte(key+"\n"), 0600)
}

// LoadClientKeyFromConfig loads the client key using config priority:
// CLI flag > config file key_file > default path.
func LoadClientKeyFromConfig(flagKeyFile string, cfg Config) (string, error) {
	path := ""
	switch {
	case flagKeyFile != "":
		path = flagKeyFile
	case cfg.KeyFile != "":
		path = cfg.KeyFile
	default:
		p, err := clientKeyPath()
		if err != nil {
			return "", err
		}
		path = p
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("client key not found at %s\nRun 'lgtv setup' to pair with your TV", path)
		}
		return "", fmt.Errorf("reading client key from %s: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// ResolveIP determines the TV IP from CLI flag > config > env var.
func ResolveIP(flagIP string, cfg Config) string {
	if flagIP != "" {
		return flagIP
	}
	if cfg.IP != "" {
		return cfg.IP
	}
	if env := os.Getenv("LGTV_IP"); env != "" {
		return env
	}
	return ""
}

// ResolveMAC determines the TV MAC from config > env var.
func ResolveMAC(cfg Config) string {
	if cfg.MAC != "" {
		return cfg.MAC
	}
	if env := os.Getenv("LGTV_MAC"); env != "" {
		return env
	}
	return ""
}
