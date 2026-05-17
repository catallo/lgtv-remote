# lgtv-remote

A simple CLI to control LG webOS TVs. Written in Go, single binary, no dependencies. Use it in your scripts, cron jobs, or hand it to your AI agent.

Works with LG webOS TVs from 2016 onwards (webOS 3.0+).

### Pre-built binaries

Binaries are available for every platform Go supports:

**Linux:** amd64, arm64, arm, 386, mips, mips64, mipsle, mips64le, ppc64, ppc64le, riscv64, s390x, loong64

**macOS:** amd64 (Intel), arm64 (Apple Silicon)

**Windows:** amd64, arm64, 386

**FreeBSD:** amd64, arm64, arm, 386

**OpenBSD:** amd64, arm64, arm, 386, ppc64, riscv64

**NetBSD:** amd64, arm64, arm, 386

Download from the [Releases](https://github.com/catallo/lgtv-remote/releases) page.

## Install

Download the binary for your platform from [Releases](https://github.com/catallo/lgtv-remote/releases) and put it in your PATH:

```
# Example for Linux amd64
curl -L https://github.com/catallo/lgtv-remote/releases/latest/download/lgtv-remote-linux-amd64 -o lgtv-remote
chmod +x lgtv-remote
sudo mv lgtv-remote /usr/local/bin/
```

Or build from source:

```
go install github.com/catallo/lgtv-remote@latest
```

## Setup

```
lgtv-remote setup
```

This will:
1. Scan your network for LG TVs (SSDP)
2. Pair with your TV (accept the prompt on screen)
3. Save config to `~/.config/lgtv-remote/`

## Usage

```
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
...

Configuration:
  Config file: ~/.config/lgtv-remote/config.json
  Client key:  ~/.config/lgtv-remote/client_key.txt
  Env vars:    LGTV_IP, LGTV_MAC

  Priority: CLI flags > config file > environment variables
  Run 'lgtv setup' for interactive first-run configuration.

Override with flags (`-i`, `--key-file`) or env vars (`LGTV_IP`, `LGTV_MAC`).

## Compatibility

Tested with webOS 6.0 (2021). Should work with any LG webOS TV from 2016+ (webOS 3.0 and newer). Older models (2014-2015, webOS 1-2) used a different connection method and are not supported.

## License

MIT
