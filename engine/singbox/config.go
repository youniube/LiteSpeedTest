package singbox

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xxf098/lite-proxy/engine"
)

func writeConfig(link, workDir string, listenPort int, opt engine.StartOptions) (string, string, error) {
	if strings.TrimSpace(workDir) == "" {
		workDir = ".lite-singbox"
	}
	sessionDir := filepath.Join(workDir, fmt.Sprintf("session-%d", time.Now().UnixNano()))
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return "", "", err
	}

	proxyOutbound, err := BuildOutbound(link)
	if err != nil {
		return "", sessionDir, err
	}

	configObject := map[string]any{
		"log": map[string]any{
			"level": normalizeLogLevel(opt.LogLevel),
		},
		"inbounds": []any{
			map[string]any{
				"type":        "mixed",
				"tag":         "mixed-in",
				"listen":      "127.0.0.1",
				"listen_port": listenPort,
			},
		},
		"outbounds": []any{
			proxyOutbound,
			map[string]any{"type": "direct", "tag": "direct"},
			map[string]any{"type": "block", "tag": "block"},
		},
		"route": map[string]any{
			"final":                 "proxy",
			"auto_detect_interface": true,
		},
	}

	configPath := filepath.Join(sessionDir, "config.json")
	data, err := json.MarshalIndent(configObject, "", "  ")
	if err != nil {
		return "", sessionDir, err
	}
	if err = os.WriteFile(configPath, data, 0o644); err != nil {
		return "", sessionDir, err
	}
	return configPath, sessionDir, nil
}

func normalizeLogLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "trace", "debug", "info", "warn", "error", "fatal", "panic":
		return strings.ToLower(strings.TrimSpace(level))
	default:
		return "warn"
	}
}
