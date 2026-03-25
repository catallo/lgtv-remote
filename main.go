package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var jsonOutput bool

func usage() {
	fmt.Fprintf(os.Stderr, `lgtv-remote — LG webOS TV remote control

Usage: lgtv-remote [flags] <command> [args...]

Commands:
  setup                         Interactive first-run setup wizard
  discover                      Discover LG TVs on the network (SSDP)
  on                            Wake-on-LAN (turn on)
  off                           Turn off
  volume                        Get current volume
  volume <0-100>                Set volume
  volume up|down                Volume up/down
  mute                          Toggle mute
  mute on|off                   Mute/unmute
  app                           Show current app
  apps                          List all installed apps
  launch <app-id>               Launch app by ID
  input                         Show current input
  inputs                        List all inputs
  input <HDMI_1|HDMI_2|...>     Switch input
  channel                       Current channel
  channels                      List channels
  channel <number>              Switch to channel
  toast <message>               Show toast notification
  play                          Media play
  pause                         Media pause
  stop                          Media stop
  rewind                        Media rewind
  ff                            Media fast forward
  info                          System info
  screen-off                    Turn screen off (audio stays)
  screen-on                     Turn screen on
  text <text>                   Insert text (for search fields)
  sound                         Get current sound output
  sound <output>                Set sound output (tv/headphone/arc/bt/optical)
  btn [button...]               Send button press (UP/DOWN/LEFT/RIGHT/ENTER/HOME/BACK...)
  click                         Click at current pointer position
  move <dx> <dy>                Move pointer (relative)
  scroll <dx> <dy>              Scroll

Flags:
  -i, --ip <addr>               TV IP address
  --key-file <path>             Client key file
  -j, --json                    JSON output
  -h, --help                    Show this help

Configuration:
  Config file: ~/.config/lgtv-remote/config.json
  Client key:  ~/.config/lgtv-remote/client_key.txt
  Env vars:    LGTV_IP, LGTV_MAC

  Priority: CLI flags > config file > environment variables
  Run 'lgtv setup' for interactive first-run configuration.
`)
}

func main() {
	var flagIP, flagKeyFile string
	var args []string

	// Parse flags
	rawArgs := os.Args[1:]
	for i := 0; i < len(rawArgs); i++ {
		switch rawArgs[i] {
		case "-h", "--help":
			usage()
			os.Exit(0)
		case "-j", "--json":
			jsonOutput = true
		case "-i", "--ip":
			i++
			if i >= len(rawArgs) {
				fatal("--ip requires an argument")
			}
			flagIP = rawArgs[i]
		case "--key-file":
			i++
			if i >= len(rawArgs) {
				fatal("--key-file requires an argument")
			}
			flagKeyFile = rawArgs[i]
		default:
			args = append(args, rawArgs[i])
		}
	}

	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	cmd := args[0]
	cmdArgs := args[1:]

	// Commands that don't need a TV connection
	switch cmd {
	case "setup":
		runSetup()
		return
	case "discover":
		runDiscover()
		return
	}

	// Load config
	cfg, err := LoadConfig()
	if err != nil {
		fatal("loading config: %v", err)
	}

	// WOL doesn't need a WebSocket connection
	if cmd == "on" {
		mac := ResolveMAC(cfg)
		if mac == "" {
			fatal("No MAC address configured. Run 'lgtv setup' or set LGTV_MAC.")
		}
		if err := SendWOL(mac); err != nil {
			fatal("WOL failed: %v", err)
		}
		output("Wake-on-LAN packet sent to %s", mac)
		return
	}

	// All other commands need a TV connection
	ip := ResolveIP(flagIP, cfg)
	if ip == "" {
		fatal("No TV IP address configured. Run 'lgtv setup' or set LGTV_IP.")
	}

	clientKey, err := LoadClientKeyFromConfig(flagKeyFile, cfg)
	if err != nil {
		fatal("%v", err)
	}

	tv := NewTVClient(ip, clientKey)
	if err := tv.Connect(); err != nil {
		fatal("%v", err)
	}
	defer tv.Close()

	if err := tv.Register(); err != nil {
		fatal("%v", err)
	}

	switch cmd {
	case "off":
		_, err = tv.Request("ssap://system/turnOff", nil)
		if err == nil {
			output("TV turned off")
		}

	case "volume":
		if len(cmdArgs) == 0 {
			raw, e := tv.Request("ssap://audio/getVolume", nil)
			err = e
			if err == nil {
				printJSON(raw, func() {
					var v map[string]interface{}
					json.Unmarshal(raw, &v)
					if vs, ok := v["volumeStatus"].(map[string]interface{}); ok {
						fmt.Printf("Volume: %.0f (muted: %v, output: %s)\n", vs["volume"], vs["muteStatus"], vs["soundOutput"])
					} else {
						fmt.Printf("Volume: %.0f (muted: %v)\n", v["volume"], v["muted"])
					}
				})
			}
		} else {
			switch cmdArgs[0] {
			case "up":
				_, err = tv.Request("ssap://audio/volumeUp", nil)
				if err == nil {
					output("Volume up")
				}
			case "down":
				_, err = tv.Request("ssap://audio/volumeDown", nil)
				if err == nil {
					output("Volume down")
				}
			default:
				vol, e := strconv.Atoi(cmdArgs[0])
				if e != nil || vol < 0 || vol > 100 {
					fatal("Volume must be 0-100 or up/down")
				}
				_, err = tv.Request("ssap://audio/setVolume", map[string]int{"volume": vol})
				if err == nil {
					output("Volume set to %d", vol)
				}
			}
		}

	case "mute":
		if len(cmdArgs) == 0 {
			// Toggle: get current then flip
			raw, e := tv.Request("ssap://audio/getVolume", nil)
			if e != nil {
				err = e
			} else {
				var v map[string]interface{}
				json.Unmarshal(raw, &v)
				muted := false
				if vs, ok := v["volumeStatus"].(map[string]interface{}); ok {
					muted, _ = vs["muteStatus"].(bool)
				} else {
					muted, _ = v["muted"].(bool)
				}
				_, err = tv.Request("ssap://audio/setMute", map[string]bool{"mute": !muted})
				if err == nil {
					output("Mute: %v", !muted)
				}
			}
		} else {
			mute := cmdArgs[0] == "on"
			_, err = tv.Request("ssap://audio/setMute", map[string]bool{"mute": mute})
			if err == nil {
				output("Mute: %v", mute)
			}
		}

	case "app":
		raw, e := tv.Request("ssap://com.webos.applicationManager/getForegroundAppInfo", nil)
		err = e
		if err == nil {
			printJSON(raw, func() {
				var v map[string]interface{}
				json.Unmarshal(raw, &v)
				fmt.Printf("Current app: %s\n", v["appId"])
			})
		}

	case "apps":
		raw, e := tv.Request("ssap://com.webos.applicationManager/listApps", nil)
		err = e
		if err == nil {
			printJSON(raw, func() {
				var v map[string]interface{}
				json.Unmarshal(raw, &v)
				if apps, ok := v["apps"].([]interface{}); ok {
					for _, a := range apps {
						app := a.(map[string]interface{})
						fmt.Printf("%-40s %s\n", app["id"], app["title"])
					}
				}
			})
		}

	case "launch":
		if len(cmdArgs) == 0 {
			fatal("Usage: lgtv-remote launch <app-id-or-name>")
		}
		appArg := strings.Join(cmdArgs, " ")
		appID := appArg
		// Try to resolve by name if it doesn't look like an app ID (no dots)
		if !strings.Contains(appArg, ".") {
			raw, e := tv.Request("ssap://com.webos.applicationManager/listApps", nil)
			if e == nil {
				var v map[string]interface{}
				json.Unmarshal(raw, &v)
				if apps, ok := v["apps"].([]interface{}); ok {
					search := strings.ToLower(appArg)
					var matches []map[string]interface{}
					for _, a := range apps {
						app := a.(map[string]interface{})
						title := strings.ToLower(fmt.Sprintf("%v", app["title"]))
						id := strings.ToLower(fmt.Sprintf("%v", app["id"]))
						if title == search || id == search {
							matches = []map[string]interface{}{app}
							break
						}
						if strings.Contains(title, search) || strings.Contains(id, search) {
							matches = append(matches, app)
						}
					}
					if len(matches) == 0 {
						fatal("No app found matching %q (use 'lgtv apps' to list)", appArg)
					} else if len(matches) > 1 {
						fmt.Fprintf(os.Stderr, "Multiple matches for %q:\n", appArg)
						for _, m := range matches {
							fmt.Fprintf(os.Stderr, "  %-40s %s\n", m["id"], m["title"])
						}
						fatal("Be more specific or use the app ID directly")
					}
					appID = fmt.Sprintf("%v", matches[0]["id"])
				}
			}
		}
		_, err = tv.Request("ssap://com.webos.applicationManager/launch", map[string]string{"id": appID})
		if err == nil {
			output("Launched %s (%s)", appArg, appID)
		}

	case "input":
		if len(cmdArgs) == 0 {
			raw, e := tv.Request("ssap://com.webos.applicationManager/getForegroundAppInfo", nil)
			err = e
			if err == nil {
				printJSON(raw, func() {
					var v map[string]interface{}
					json.Unmarshal(raw, &v)
					fmt.Printf("Current app: %s\n", v["appId"])
				})
			}
		} else {
			_, err = tv.Request("ssap://tv/switchInput", map[string]string{"inputId": cmdArgs[0]})
			if err == nil {
				output("Switched to %s", cmdArgs[0])
			}
		}

	case "inputs":
		raw, e := tv.Request("ssap://tv/getExternalInputList", nil)
		err = e
		if err == nil {
			printJSON(raw, func() {
				var v map[string]interface{}
				json.Unmarshal(raw, &v)
				if devs, ok := v["devices"].([]interface{}); ok {
					for _, d := range devs {
						dev := d.(map[string]interface{})
						fmt.Printf("%-12s %s\n", dev["id"], dev["label"])
					}
				}
			})
		}

	case "channel":
		if len(cmdArgs) == 0 {
			raw, e := tv.Request("ssap://tv/getCurrentChannel", nil)
			err = e
			if err == nil {
				printJSON(raw, func() {
					var v map[string]interface{}
					json.Unmarshal(raw, &v)
					fmt.Printf("Channel: %s — %s\n", v["channelNumber"], v["channelName"])
				})
			}
		} else {
			_, err = tv.Request("ssap://tv/openChannel", map[string]string{"channelNumber": cmdArgs[0]})
			if err == nil {
				output("Switched to channel %s", cmdArgs[0])
			}
		}

	case "channels":
		raw, e := tv.Request("ssap://tv/getChannelList", nil)
		err = e
		if err == nil {
			printJSON(raw, func() {
				var v map[string]interface{}
				json.Unmarshal(raw, &v)
				if chs, ok := v["channelList"].([]interface{}); ok {
					for _, c := range chs {
						ch := c.(map[string]interface{})
						fmt.Printf("%-6s %s\n", ch["channelNumber"], ch["channelName"])
					}
				}
			})
		}

	case "toast":
		if len(cmdArgs) == 0 {
			fatal("Usage: lgtv-remote toast <message>")
		}
		msg := strings.Join(cmdArgs, " ")
		_, err = tv.Request("ssap://system.notifications/createToast", map[string]string{"message": msg})
		if err == nil {
			output("Toast sent")
		}

	case "play":
		_, err = tv.Request("ssap://media.controls/play", nil)
		if err == nil { output("Play") }
	case "pause":
		_, err = tv.Request("ssap://media.controls/pause", nil)
		if err == nil { output("Pause") }
	case "stop":
		_, err = tv.Request("ssap://media.controls/stop", nil)
		if err == nil { output("Stop") }
	case "rewind":
		_, err = tv.Request("ssap://media.controls/rewind", nil)
		if err == nil { output("Rewind") }
	case "ff":
		_, err = tv.Request("ssap://media.controls/fastForward", nil)
		if err == nil { output("Fast Forward") }

	case "info":
		// Try software info first, fall back gracefully
		raw, e := tv.Request("ssap://com.webos.service.update/getCurrentSWInformation", nil)
		if e != nil {
			// Fallback: just show current app + volume
			raw2, e2 := tv.Request("ssap://com.webos.applicationManager/getForegroundAppInfo", nil)
			raw3, e3 := tv.Request("ssap://audio/getVolume", nil)
			if e2 == nil && e3 == nil {
				if jsonOutput {
					combined := map[string]json.RawMessage{"app": raw2, "audio": raw3}
					b, _ := json.MarshalIndent(combined, "", "  ")
					fmt.Println(string(b))
				} else {
					var app, vol map[string]interface{}
					json.Unmarshal(raw2, &app)
					json.Unmarshal(raw3, &vol)
					fmt.Printf("Current app: %s\n", app["appId"])
					if vs, ok := vol["volumeStatus"].(map[string]interface{}); ok {
						fmt.Printf("Volume: %.0f (muted: %v)\n", vs["volume"], vs["muteStatus"])
					}
					fmt.Println("(detailed system info requires re-pairing)")
				}
			} else {
				err = fmt.Errorf("could not get TV info")
			}
		} else {
			printJSON(raw, func() {
				var v map[string]interface{}
				json.Unmarshal(raw, &v)
				fmt.Printf("Product: %s\n", v["product_name"])
				fmt.Printf("Model: %s\n", v["model_name"])
				fmt.Printf("Firmware: %s.%s\n", v["major_ver"], v["minor_ver"])
				fmt.Printf("Country: %s\n", v["country"])
			})
		}

	case "screen-off":
		_, err = tv.Request("ssap://com.webos.service.tv.power/turnOffScreen", nil)
		if err == nil { output("Screen off") }
	case "screen-on":
		_, err = tv.Request("ssap://com.webos.service.tv.power/turnOnScreen", nil)
		if err == nil { output("Screen on") }

	case "text":
		if len(cmdArgs) == 0 {
			fatal("Usage: lgtv-remote text <text>")
		}
		txt := strings.Join(cmdArgs, " ")
		_, err = tv.Request("ssap://com.webos.service.ime/insertText", map[string]string{"text": txt})
		if err == nil { output("Text inserted") }

	case "sound":
		if len(cmdArgs) == 0 {
			// Get current sound output via audio status
			raw, e := tv.Request("ssap://audio/getStatus", nil)
			if e != nil {
				// Fallback: check volume info which has soundOutput
				raw2, e2 := tv.Request("ssap://audio/getVolume", nil)
				if e2 != nil {
					err = e2
				} else {
					printJSON(raw2, func() {
						var v map[string]interface{}
						json.Unmarshal(raw2, &v)
						if vs, ok := v["volumeStatus"].(map[string]interface{}); ok {
							fmt.Printf("Sound output: %s\n", vs["soundOutput"])
						}
					})
				}
			} else {
				printJSON(raw, func() {
					var v map[string]interface{}
					json.Unmarshal(raw, &v)
					if vs, ok := v["volumeStatus"].(map[string]interface{}); ok {
						fmt.Printf("Sound output: %s\n", vs["soundOutput"])
					} else if so, ok := v["soundOutput"]; ok {
						fmt.Printf("Sound output: %s\n", so)
					} else {
						fmt.Printf("Sound output: %v\n", v)
					}
				})
			}
		} else {
			target := resolveSoundAlias(cmdArgs[0])
			_, err = tv.Request("ssap://com.webos.service.apiadapter/audio/changeSoundOutput", map[string]string{"output": target})
			if err == nil {
				output("Sound output: %s", target)
			}
		}

	case "btn", "button", "key":
		if err := handleButtonCmd(tv, cmdArgs); err != nil {
			fatal("%v", err)
		} else if len(cmdArgs) > 0 {
			output("Button: %s", strings.Join(cmdArgs, " "))
		}

	case "click":
		ic, e := tv.ConnectInput()
		if e != nil {
			fatal("%v", e)
		}
		defer ic.Close()
		err = ic.Click()
		if err == nil {
			output("Click")
		}

	case "move":
		if len(cmdArgs) < 2 {
			fatal("Usage: lgtv-remote move <dx> <dy>")
		}
		dx, e1 := strconv.Atoi(cmdArgs[0])
		dy, e2 := strconv.Atoi(cmdArgs[1])
		if e1 != nil || e2 != nil {
			fatal("dx and dy must be integers")
		}
		ic, e := tv.ConnectInput()
		if e != nil {
			fatal("%v", e)
		}
		defer ic.Close()
		err = ic.Move(dx, dy)
		if err == nil {
			output("Move dx=%d dy=%d", dx, dy)
		}

	case "scroll":
		if len(cmdArgs) < 2 {
			fatal("Usage: lgtv-remote scroll <dx> <dy>")
		}
		dx, e1 := strconv.Atoi(cmdArgs[0])
		dy, e2 := strconv.Atoi(cmdArgs[1])
		if e1 != nil || e2 != nil {
			fatal("dx and dy must be integers")
		}
		ic, e := tv.ConnectInput()
		if e != nil {
			fatal("%v", e)
		}
		defer ic.Close()
		err = ic.Scroll(dx, dy)
		if err == nil {
			output("Scroll dx=%d dy=%d", dx, dy)
		}

	default:
		fatal("Unknown command: %s (use -h for help)", cmd)
	}

	if err != nil {
		fatal("%v", err)
	}
}

// resolveSoundAlias maps friendly sound output names to TV API identifiers.
func resolveSoundAlias(name string) string {
	switch strings.ToLower(name) {
	case "tv", "speaker", "tv_speaker":
		return "tv_speaker"
	case "headphone", "headphones", "wired", "lautsprecher":
		return "headphone"
	case "arc", "hdmi_arc", "earc":
		return "external_arc"
	case "bt", "bluetooth":
		return "bt_soundbar"
	case "optical":
		return "external_optical"
	default:
		return name
	}
}

func output(format string, a ...interface{}) {
	if jsonOutput {
		msg := fmt.Sprintf(format, a...)
		b, _ := json.Marshal(map[string]string{"status": "ok", "message": msg})
		fmt.Println(string(b))
	} else {
		fmt.Printf(format+"\n", a...)
	}
}

func printJSON(raw json.RawMessage, humanFn func()) {
	if jsonOutput {
		var v interface{}
		json.Unmarshal(raw, &v)
		b, _ := json.MarshalIndent(v, "", "  ")
		fmt.Println(string(b))
	} else {
		humanFn()
	}
}

func fatal(format string, a ...interface{}) {
	msg := fmt.Sprintf(format, a...)
	if jsonOutput {
		b, _ := json.Marshal(map[string]string{"status": "error", "error": msg})
		fmt.Fprintln(os.Stderr, string(b))
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	}
	os.Exit(1)
}
