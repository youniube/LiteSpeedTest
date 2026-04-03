package singbox

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Log       LogConfig   `json:"log,omitempty"`
	Inbounds  []Inbound   `json:"inbounds"`
	Outbounds []any       `json:"outbounds"`
	Route     RouteConfig `json:"route,omitempty"`
}

type LogConfig struct {
	Disabled  bool   `json:"disabled,omitempty"`
	Level     string `json:"level,omitempty"`
	Output    string `json:"output,omitempty"`
	Timestamp bool   `json:"timestamp,omitempty"`
}

type Inbound struct {
	Type           string        `json:"type"`
	Tag            string        `json:"tag,omitempty"`
	Listen         string        `json:"listen,omitempty"`
	ListenPort     int           `json:"listen_port,omitempty"`
	Users          []InboundUser `json:"users,omitempty"`
	SetSystemProxy bool          `json:"set_system_proxy,omitempty"`
}

type InboundUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RouteConfig struct {
	Final string `json:"final,omitempty"`
}

func NewSingleNodeConfig(port int, outbound any, logLevel string) Config {
	if logLevel == "" {
		logLevel = "warn"
	}

	return Config{
		Log: LogConfig{
			Level:     logLevel,
			Output:    "box.log",
			Timestamp: true,
		},
		Inbounds: []Inbound{
			{
				Type:       "mixed",
				Tag:        "mixed-in",
				Listen:     "127.0.0.1",
				ListenPort: port,
			},
		},
		Outbounds: []any{
			outbound,
			map[string]any{
				"type": "direct",
				"tag":  "direct",
			},
		},
		Route: RouteConfig{
			Final: "proxy",
		},
	}
}

func WriteConfig(workDir string, cfg Config) (string, func() error, error) {
	if workDir == "" {
		return "", nil, fmt.Errorf("empty workDir")
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return "", nil, err
	}

	configPath := filepath.Join(workDir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", nil, err
	}
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return "", nil, err
	}

	cleanup := func() error {
		return os.RemoveAll(workDir)
	}
	return configPath, cleanup, nil
}
