package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/xxf098/lite-proxy/utils"
)

func ParseSubscription(ctx context.Context, input string) ([]ProxyNode, InputFormat, error) {
	if strings.TrimSpace(input) == "" {
		return nil, FormatUnknown, ErrInvalidSubscription
	}
	content, hint, err := readInput(ctx, input)
	if err != nil {
		return nil, FormatUnknown, err
	}
	nodes, format, err := parseTextContent(string(content), hint)
	if err != nil {
		return nil, format, err
	}
	nodes = NormalizeNodes(nodes)
	if len(nodes) == 0 {
		return nil, format, ErrNoProxyFound
	}
	return nodes, format, nil
}

func ParseToLinks(ctx context.Context, input string) ([]string, InputFormat, error) {
	nodes, format, err := ParseSubscription(ctx, input)
	if err != nil {
		return nil, format, err
	}
	links, err := ConvertNodesToLinks(nodes)
	if err != nil {
		return nil, format, err
	}
	if len(links) == 0 {
		return nil, format, ErrNoProxyFound
	}
	return links, format, nil
}

func ParseText(content string, format InputFormat) ([]ProxyNode, error) {
	nodes, _, err := parseTextContent(content, format)
	if err != nil {
		return nil, err
	}
	nodes = NormalizeNodes(nodes)
	if len(nodes) == 0 {
		return nil, ErrNoProxyFound
	}
	return nodes, nil
}

func ParseTextToLinks(content string, format InputFormat) ([]string, error) {
	nodes, err := ParseText(content, format)
	if err != nil {
		return nil, err
	}
	return ConvertNodesToLinks(nodes)
}

func readInput(ctx context.Context, input string) ([]byte, InputFormat, error) {
	trimmed := strings.TrimSpace(input)
	if utils.IsUrl(trimmed) {
		data, _, err := FetchSubscription(ctx, trimmed)
		return data, DetectURLHint(trimmed), err
	}
	if utils.IsFilePath(trimmed) {
		data, err := os.ReadFile(trimmed)
		return data, DetectURLHint(trimmed), err
	}
	return []byte(input), FormatUnknown, nil
}

func parseTextContent(content string, hint InputFormat) ([]ProxyNode, InputFormat, error) {
	format := hint
	if format == FormatUnknown {
		format = DetectTextFormat(content)
	}

	if format != FormatUnknown {
		nodes, err := parseTextByFormat(content, format)
		if err == nil && len(nodes) > 0 {
			return nodes, format, nil
		}
	}

	for _, f := range []InputFormat{FormatURI, FormatBase64, FormatClash, FormatLoon, FormatSurge, FormatQX, FormatSingBox} {
		if f == format {
			continue
		}
		nodes, err := parseTextByFormat(content, f)
		if err == nil && len(nodes) > 0 {
			return nodes, f, nil
		}
	}
	return nil, format, ErrUnsupportedFormat
}

func parseTextByFormat(content string, format InputFormat) ([]ProxyNode, error) {
	switch format {
	case FormatURI:
		return ParseURIText(content)
	case FormatBase64:
		return ParseBase64Text(content)
	case FormatClash:
		return ParseClashText(content)
	case FormatSurge:
		return ParseSurgeText(content)
	case FormatLoon:
		return ParseLoonText(content)
	case FormatQX:
		return ParseQuantumultXText(content)
	case FormatSingBox:
		return ParseSingBoxJSON(content)
	default:
		return nil, ErrUnsupportedFormat
	}
}

func getString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch t := v.(type) {
			case string:
				if s := strings.TrimSpace(t); s != "" {
					return s
				}
			case fmt.Stringer:
				if s := strings.TrimSpace(t.String()); s != "" {
					return s
				}
			case json.Number:
				return t.String()
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
				return fmt.Sprintf("%v", t)
			}
		}
	}
	return ""
}

func getBool(m map[string]any, keys ...string) bool {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch t := v.(type) {
			case bool:
				return t
			case string:
				switch strings.ToLower(strings.TrimSpace(t)) {
				case "1", "true", "yes", "on":
					return true
				}
			case int:
				return t != 0
			case float64:
				return t != 0
			}
		}
	}
	return false
}

func getInt(m map[string]any, keys ...string) int {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch t := v.(type) {
			case int:
				return t
			case int64:
				return int(t)
			case uint16:
				return int(t)
			case uint64:
				return int(t)
			case float64:
				return int(t)
			case string:
				i, err := strconv.Atoi(strings.TrimSpace(t))
				if err == nil {
					return i
				}
			}
		}
	}
	return 0
}

func cloneMap(m map[string]any) map[string]any {
	if len(m) == 0 {
		return nil
	}
	cp := make(map[string]any, len(m))
	for k, v := range m {
		cp[k] = v
	}
	return cp
}

func splitHostPort(raw string) (string, int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", 0, fmt.Errorf("empty address")
	}
	if host, port, err := net.SplitHostPort(raw); err == nil {
		p, err := strconv.Atoi(port)
		return host, p, err
	}
	idx := strings.LastIndex(raw, ":")
	if idx <= 0 || idx == len(raw)-1 {
		return "", 0, fmt.Errorf("invalid address: %s", raw)
	}
	host := strings.TrimSpace(raw[:idx])
	port, err := strconv.Atoi(strings.TrimSpace(raw[idx+1:]))
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

func splitCSV(line string) []string {
	parts := strings.Split(line, ",")
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}

func parseKeyValueTokens(parts []string) map[string]string {
	m := map[string]string{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			m[strings.ToLower(strings.TrimSpace(kv[0]))] = trimWrapping(kv[1])
		}
	}
	return m
}

func trimWrapping(v string) string {
	v = strings.TrimSpace(v)
	v = strings.Trim(v, "\"'")
	return strings.TrimSpace(v)
}

func parseSimpleHeaders(raw string) map[string]string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	res := map[string]string{}
	for _, item := range strings.FieldsFunc(raw, func(r rune) bool { return r == '|' || r == '&' }) {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		kv := strings.SplitN(item, ":", 2)
		if len(kv) != 2 {
			kv = strings.SplitN(item, "=", 2)
		}
		if len(kv) == 2 {
			res[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

func getMap(m map[string]any, keys ...string) map[string]any {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			if mv, ok := v.(map[string]any); ok {
				return mv
			}
		}
	}
	return nil
}
