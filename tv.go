package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	connectTimeout = 5 * time.Second
	commandTimeout = 10 * time.Second
)

// TVClient communicates with an LG webOS TV over SSAP WebSocket protocol.
type TVClient struct {
	ip        string
	clientKey string
	conn      *websocket.Conn
	msgID     atomic.Int64
	mu        sync.Mutex
}

// SSAPMessage is the JSON-RPC style message for SSAP commands.
type SSAPMessage struct {
	Type    string      `json:"type"`
	ID      string      `json:"id,omitempty"`
	URI     string      `json:"uri,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

// SSAPResponse is the response from the TV.
type SSAPResponse struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Error   string          `json:"error,omitempty"`
}

func NewTVClient(ip, clientKey string) *TVClient {
	return &TVClient{
		ip:        ip,
		clientKey: clientKey,
	}
}

func (c *TVClient) Connect() error {
	dialer := websocket.Dialer{
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
		HandshakeTimeout: connectTimeout,
		NetDialContext:    (&net.Dialer{Timeout: connectTimeout}).DialContext,
	}

	url := fmt.Sprintf("wss://%s:3001", c.ip)
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return fmt.Errorf("connecting to TV at %s: %w", c.ip, err)
	}
	c.conn = conn
	return nil
}

func (c *TVClient) Register() error {
	reg := buildRegistrationMessage(c.clientKey)
	if err := c.conn.WriteJSON(reg); err != nil {
		return fmt.Errorf("sending registration: %w", err)
	}

	// Read responses until we get registered confirmation
	deadline := time.Now().Add(commandTimeout)
	c.conn.SetReadDeadline(deadline)

	for {
		var resp SSAPResponse
		if err := c.conn.ReadJSON(&resp); err != nil {
			return fmt.Errorf("reading registration response: %w", err)
		}
		if resp.Type == "registered" {
			return nil
		}
		if resp.Type == "error" {
			return fmt.Errorf("registration failed: %s", resp.Error)
		}
		// response type could be other things during handshake; keep reading
	}
}

func (c *TVClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *TVClient) nextID() string {
	return fmt.Sprintf("msg_%d", c.msgID.Add(1))
}

// Request sends a command and returns the response payload.
func (c *TVClient) Request(uri string, payload interface{}) (json.RawMessage, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.nextID()
	msg := SSAPMessage{
		Type:    "request",
		ID:      id,
		URI:     uri,
		Payload: payload,
	}

	if err := c.conn.WriteJSON(msg); err != nil {
		return nil, fmt.Errorf("sending command: %w", err)
	}

	c.conn.SetReadDeadline(time.Now().Add(commandTimeout))

	for {
		var resp SSAPResponse
		if err := c.conn.ReadJSON(&resp); err != nil {
			return nil, fmt.Errorf("reading response: %w", err)
		}
		if resp.ID == id {
			if resp.Type == "error" {
				return nil, fmt.Errorf("TV error: %s", resp.Error)
			}
			return resp.Payload, nil
		}
	}
}

func buildRegistrationMessage(clientKey string) map[string]interface{} {
	return map[string]interface{}{
		"type": "register",
		"id":   "register_0",
		"payload": map[string]interface{}{
			"forcePairing": false,
			"pairingType": "PROMPT",
			"client-key":  clientKey,
			"manifest": map[string]interface{}{
				"manifestVersion": 1,
				"appVersion":      "1.1",
				"signed": map[string]interface{}{
					"created":  "20140509",
					"appId":    "com.lge.test",
					"vendorId": "com.lge",
					"localizedAppNames": map[string]interface{}{
						"":       "LG Remote App",
						"ko-KR":  "리모컨 앱",
						"zxx-XX": "ЛГ Rэмotэ AПП",
					},
					"localizedVendorNames": map[string]interface{}{
						"": "LG Electronics",
					},
					"permissions": []string{
						"TEST_SECURE",
						"CONTROL_INPUT_TEXT",
						"CONTROL_MOUSE_AND_KEYBOARD",
						"READ_INSTALLED_APPS",
						"READ_LGE_SDX",
						"READ_NOTIFICATIONS",
						"SEARCH",
						"WRITE_SETTINGS",
						"WRITE_NOTIFICATION_ALERT",
						"CONTROL_POWER",
						"READ_CURRENT_CHANNEL",
						"READ_RUNNING_APPS",
						"READ_UPDATE_INFO",
						"UPDATE_FROM_REMOTE_APP",
						"READ_LGE_TV_INPUT_EVENTS",
						"READ_TV_CURRENT_TIME",
					},
					"serial": "2f930e2d2cfe083771f68e4fe7bb07",
				},
				"permissions": []string{
					"LAUNCH",
					"LAUNCH_WEBAPP",
					"APP_TO_APP",
					"CLOSE",
					"TEST_OPEN",
					"TEST_PROTECTED",
					"CONTROL_AUDIO",
					"CONTROL_DISPLAY",
					"CONTROL_INPUT_JOYSTICK",
					"CONTROL_INPUT_MEDIA_RECORDING",
					"CONTROL_INPUT_MEDIA_PLAYBACK",
					"CONTROL_INPUT_TV",
					"CONTROL_POWER",
					"READ_APP_STATUS",
					"READ_CURRENT_CHANNEL",
					"READ_INPUT_DEVICE_LIST",
					"READ_NETWORK_STATE",
					"READ_RUNNING_APPS",
					"READ_TV_CHANNEL_LIST",
					"WRITE_NOTIFICATION_TOAST",
					"READ_POWER_STATE",
					"READ_COUNTRY_INFO",
					"READ_SETTINGS",
					"CONTROL_TV_SCREEN",
					"CONTROL_TV_STANBY",
					"CONTROL_FAVORITE_GROUP",
					"CONTROL_USER_INFO",
					"CHECK_BLUETOOTH_DEVICE",
					"CONTROL_BLUETOOTH",
					"CONTROL_TIMER_INFO",
					"STB_INTERNAL_CONNECTION",
					"CONTROL_RECORDING",
					"READ_RECORDING_STATE",
					"WRITE_RECORDING_LIST",
					"READ_RECORDING_LIST",
					"READ_RECORDING_SCHEDULE",
					"WRITE_RECORDING_SCHEDULE",
				},
				"signatures": []map[string]interface{}{
					{
						"signatureVersion": 1,
						"signature":        "eyJhbGdvcml0aG0iOiJSU0EtU0hBMjU2Iiwia2V5SWQiOiJ0ZXN0LXNpZ25pbmctY2VydCIsInNpZ25hdHVyZVZlcnNpb24iOjF9.hrVRgjCwXVvE2OOSpDZ58hR+59aFNwYDyjQgKk3auukd7pcegmE2CzPCa0bJ0ZsRAcKkCTJrWo5iDzNhMBWRyaMOv5zWSrthlf7G128qvIlpMT0YNY+n/FaOHE73uLrS/g7swl3/qH/BGFG2Hu4RlL48eb3lLKqTt2xKHdCs6Cd4RMfJPYnzgvI4BNrFUKsjkcu+WD4OO2A27Pq1n50cMchmcaXadJhGrOqH5YmHdOCj5NSHzJYrsW0HPlpuAx/ECMeIZYDh6RMqaFM2DXzdKX9NmmyqzJ3o/0lkk/N97gfVRLW5hA29yeAwaCViZNCP8iC9aO0q9fQojoa7NQnAtw==",
					},
				},
			},
		},
	}
}
