package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// InputClient handles the secondary WebSocket for pointer/button input.
type InputClient struct {
	conn *websocket.Conn
}

// ConnectInput gets the pointer input socket path from the TV and connects.
func (c *TVClient) ConnectInput() (*InputClient, error) {
	raw, err := c.Request("ssap://com.webos.service.networkinput/getPointerInputSocket", nil)
	if err != nil {
		return nil, fmt.Errorf("getting input socket: %w", err)
	}

	var resp struct {
		SocketPath  string `json:"socketPath"`
		ReturnValue bool   `json:"returnValue"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parsing socket path: %w", err)
	}

	if resp.SocketPath == "" {
		return nil, fmt.Errorf("empty socket path returned")
	}

	// The socket path is ws:// (unencrypted), connect to it
	dialer := websocket.Dialer{
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
		HandshakeTimeout: connectTimeout,
		NetDialContext:    (&net.Dialer{Timeout: connectTimeout}).DialContext,
	}

	conn, _, err := dialer.Dial(resp.SocketPath, nil)
	if err != nil {
		return nil, fmt.Errorf("connecting to input socket: %w", err)
	}

	return &InputClient{conn: conn}, nil
}

// Button sends a button press command.
func (ic *InputClient) Button(name string) error {
	msg := fmt.Sprintf("type:button\nname:%s\n\n", name)
	return ic.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

// Click sends a click at current pointer position.
func (ic *InputClient) Click() error {
	msg := "type:click\n\n"
	return ic.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

// Move sends a pointer move command (relative delta).
func (ic *InputClient) Move(dx, dy int) error {
	msg := fmt.Sprintf("type:move\ndx:%d\ndy:%d\ndown:0\n\n", dx, dy)
	return ic.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

// Scroll sends a scroll command.
func (ic *InputClient) Scroll(dx, dy int) error {
	msg := fmt.Sprintf("type:scroll\ndx:%d\ndy:%d\n\n", dx, dy)
	return ic.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (ic *InputClient) Close() {
	if ic.conn != nil {
		ic.conn.Close()
	}
}

// knownButtons lists recognized button names for the help text.
var knownButtons = []string{
	"UP", "DOWN", "LEFT", "RIGHT", "ENTER",
	"HOME", "BACK", "EXIT",
	"RED", "GREEN", "YELLOW", "BLUE",
	"MENU", "MUTE",
	"VOLUMEUP", "VOLUMEDOWN",
	"CHANNELUP", "CHANNELDOWN",
	"CC", "DASH", "INFO",
	"1", "2", "3", "4", "5", "6", "7", "8", "9", "0",
}

// handleButtonCmd connects the input socket and sends button(s).
func handleButtonCmd(tv *TVClient, cmdArgs []string) error {
	if len(cmdArgs) == 0 {
		fmt.Println("Available buttons:")
		for i := 0; i < len(knownButtons); i += 8 {
			end := i + 8
			if end > len(knownButtons) {
				end = len(knownButtons)
			}
			fmt.Printf("  %s\n", strings.Join(knownButtons[i:end], ", "))
		}
		return nil
	}

	ic, err := tv.ConnectInput()
	if err != nil {
		return err
	}
	defer ic.Close()

	for _, btn := range cmdArgs {
		if err := ic.Button(strings.ToUpper(btn)); err != nil {
			return fmt.Errorf("sending button %s: %w", btn, err)
		}
		// Small delay between multiple buttons
		if len(cmdArgs) > 1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}
