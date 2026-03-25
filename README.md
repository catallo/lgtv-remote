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
lgtv-remote on                  # Wake-on-LAN
lgtv-remote off                 # Turn off
lgtv-remote volume              # Show volume
lgtv-remote volume 25           # Set volume
lgtv-remote volume up|down      # Adjust
lgtv-remote mute                # Toggle mute
lgtv-remote sound tv            # TV speakers
lgtv-remote sound headphone     # Wired output
lgtv-remote sound arc           # HDMI ARC
lgtv-remote input HDMI_1        # Switch input
lgtv-remote inputs              # List inputs
lgtv-remote launch YouTube      # Open app
lgtv-remote apps                # List apps
lgtv-remote btn UP DOWN ENTER   # Remote buttons
lgtv-remote screen-off          # Screen off, audio stays
lgtv-remote toast "Hello"       # Show notification
lgtv-remote info                # System info
lgtv-remote discover            # Find TVs on network
```

Media: `play`, `pause`, `stop`, `rewind`, `ff`
Pointer: `click`, `move <dx> <dy>`, `scroll <dx> <dy>`
Channels: `channel`, `channels`, `channel <number>`

## Config

Stored in `~/.config/lgtv-remote/`:
- `config.json` — IP, MAC address
- `client_key.txt` — pairing key (auto-saved)

Override with flags (`-i`, `--key-file`) or env vars (`LGTV_IP`, `LGTV_MAC`).

## JSON output

Use `-j` for machine-readable output.

## Compatibility

Tested with webOS 6.0 (2021). Should work with any LG webOS TV from 2016+ (webOS 3.0 and newer). Older models (2014-2015, webOS 1-2) used a different connection method and are not supported.

## License

MIT
