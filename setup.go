package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func runSetup() {
	fmt.Println("lgtv setup — configure your LG webOS TV")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Step 1: Discover TVs
	fmt.Println("Searching for LG webOS TVs on your network...")
	tvs, err := DiscoverTVs()
	if err != nil {
		fmt.Printf("Discovery error: %v\n", err)
	}

	var ip string
	if len(tvs) == 1 {
		fmt.Printf("Found TV: %s (%s)\n", tvs[0].Name, tvs[0].IP)
		fmt.Print("Use this TV? [Y/n] ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer == "" || answer == "y" || answer == "yes" {
			ip = tvs[0].IP
		}
	} else if len(tvs) > 1 {
		fmt.Println("Found multiple TVs:")
		for i, tv := range tvs {
			fmt.Printf("  %d) %s (%s)\n", i+1, tv.Name, tv.IP)
		}
		fmt.Print("Choose TV number: ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		idx, err := strconv.Atoi(answer)
		if err == nil && idx >= 1 && idx <= len(tvs) {
			ip = tvs[idx-1].IP
		}
	} else {
		fmt.Println("No TVs found automatically.")
	}

	if ip == "" {
		fmt.Print("Enter TV IP address: ")
		answer, _ := reader.ReadString('\n')
		ip = strings.TrimSpace(answer)
	}

	if ip == "" {
		fmt.Fprintln(os.Stderr, "Error: no IP address provided")
		os.Exit(1)
	}

	// Step 2: Pair with TV
	fmt.Println()
	fmt.Println("Pairing with TV at", ip, "...")
	fmt.Println(">>> Please accept the pairing prompt on your TV <<<")

	tv := NewTVClient(ip, "")
	if err := tv.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer tv.Close()

	clientKey, err := registerAndGetKey(tv)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Paired successfully!")

	// Save client key
	if err := SaveClientKey(clientKey); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving client key: %v\n", err)
		os.Exit(1)
	}
	keyPath, _ := clientKeyPath()
	fmt.Printf("Client key saved to %s\n", keyPath)

	// Step 3: Ask for MAC address (needed for WOL)
	fmt.Println()
	fmt.Println("MAC address is needed for Wake-on-LAN (turning on the TV).")
	fmt.Println("You can find it in TV Settings > Network > Wi-Fi > Advanced.")
	fmt.Print("Enter TV MAC address (or press Enter to skip): ")
	macAnswer, _ := reader.ReadString('\n')
	mac := strings.TrimSpace(macAnswer)

	// Validate MAC if provided
	if mac != "" {
		if _, err := parseMAC(mac); err != nil {
			fmt.Printf("Warning: invalid MAC address %q, skipping\n", mac)
			mac = ""
		}
	}

	// Step 4: Save config
	cfg := Config{
		IP:  ip,
		MAC: mac,
	}
	if err := SaveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		os.Exit(1)
	}
	cfgPath, _ := configPath()
	fmt.Printf("Config saved to %s\n", cfgPath)

	fmt.Println()
	fmt.Println("Setup complete! Try running:")
	fmt.Println("  lgtv info")
	fmt.Println("  lgtv volume")
	if mac != "" {
		fmt.Println("  lgtv on       (Wake-on-LAN)")
	} else {
		fmt.Println("  (Run setup again to add MAC for Wake-on-LAN)")
	}
}

// registerAndGetKey performs registration and extracts the client key from the response.
func registerAndGetKey(tv *TVClient) (string, error) {
	reg := buildRegistrationMessage("")
	if err := tv.conn.WriteJSON(reg); err != nil {
		return "", fmt.Errorf("sending registration: %w", err)
	}

	for {
		var resp struct {
			Type    string          `json:"type"`
			ID      string          `json:"id"`
			Payload json.RawMessage `json:"payload,omitempty"`
			Error   string          `json:"error,omitempty"`
		}
		if err := tv.conn.ReadJSON(&resp); err != nil {
			return "", fmt.Errorf("reading registration response: %w", err)
		}
		if resp.Type == "registered" {
			var payload struct {
				ClientKey string `json:"client-key"`
			}
			if err := json.Unmarshal(resp.Payload, &payload); err != nil {
				return "", fmt.Errorf("parsing client key: %w", err)
			}
			if payload.ClientKey == "" {
				return "", fmt.Errorf("TV returned empty client key")
			}
			return payload.ClientKey, nil
		}
		if resp.Type == "error" {
			return "", fmt.Errorf("registration failed: %s", resp.Error)
		}
	}
}

func runDiscover() {
	fmt.Println("Searching for LG webOS TVs...")
	tvs, err := DiscoverTVs()
	if err != nil {
		fatal("Discovery failed: %v", err)
	}
	if len(tvs) == 0 {
		fmt.Println("No TVs found.")
		return
	}
	for _, tv := range tvs {
		if jsonOutput {
			b, _ := json.MarshalIndent(tv, "", "  ")
			fmt.Println(string(b))
		} else {
			if tv.Name != "" {
				fmt.Printf("%s (%s)\n", tv.Name, tv.IP)
			} else {
				fmt.Println(tv.IP)
			}
		}
	}
}
