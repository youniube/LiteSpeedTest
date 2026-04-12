package subscription

import (
	"fmt"
	"strings"
)

func ParseSurgeText(content string) ([]ProxyNode, error) {
	sections := SplitSections(content)
	lines := sections["proxy"]
	if len(lines) == 0 {
		return nil, ErrNoProxyFound
	}
	nodes := make([]ProxyNode, 0, len(lines))
	for _, line := range lines {
		node, err := ParseSurgeProxyLine(line)
		if err != nil {
			continue
		}
		nodes = append(nodes, node)
	}
	if len(nodes) == 0 {
		return nil, ErrNoProxyFound
	}
	return nodes, nil
}

func ParseSurgeProxyLine(line string) (ProxyNode, error) {
	kv := strings.SplitN(line, "=", 2)
	if len(kv) != 2 {
		return ProxyNode{}, fmt.Errorf("invalid proxy line")
	}
	name := trimWrapping(kv[0])
	parts := splitCSV(kv[1])
	if len(parts) < 3 {
		return ProxyNode{}, fmt.Errorf("invalid proxy fields")
	}
	proxyType := normalizeProxyType(parts[0])
	host := trimWrapping(parts[1])
	port := 0
	fmt.Sscanf(parts[2], "%d", &port)
	positional, attrs := parseProxyTokens(parts[3:], false)
	node := ProxyNode{
		Name:         name,
		Type:         proxyType,
		Server:       host,
		Port:         port,
		SourceFormat: FormatSurge,
		Raw:          map[string]any{"line": line},
	}
	switch proxyType {
	case "ss":
		node.Method = firstNotEmpty(attrs["encrypt-method"], attrs["method"], attrs["cipher"])
		if node.Method == "" && len(positional) >= 1 {
			node.Method = positional[0]
		}
		node.Password = attrs["password"]
		if node.Password == "" && len(positional) >= 2 {
			node.Password = positional[1]
		}
	case "ssr":
		if len(positional) >= 1 {
			node.Cipher = positional[0]
		}
		if len(positional) >= 2 {
			node.Password = positional[1]
		}
		if len(positional) >= 3 {
			node.Protocol = positional[2]
		}
		if len(positional) >= 4 {
			node.ProtoParam = normalizeEmptyPlaceholder(positional[3])
		}
		if len(positional) >= 5 {
			node.Obfs = positional[4]
		}
		if len(positional) >= 6 {
			node.ObfsParam = normalizeEmptyPlaceholder(positional[5])
		}
	case "vmess":
		node.Cipher = firstNotEmpty(attrs["cipher"], attrs["method"], "none")
		if len(positional) >= 1 && node.Cipher == "none" {
			node.Cipher = positional[0]
		}
		node.UUID = firstNotEmpty(attrs["username"], attrs["uuid"], attrs["password"])
		if node.UUID == "" && len(positional) >= 2 {
			node.UUID = positional[1]
		}
		node.AlterID = 0
		node.TLS = getBoolFromString(firstNotEmpty(attrs["tls"], attrs["over-tls"]))
		node.SNI = firstNotEmpty(attrs["sni"], attrs["servername"], attrs["tls-name"], attrs["tls-host"])
		node.SkipCertVerify = isSkipVerify(attrs)
		applyTransportAttrs(&node, attrs)
	case "trojan":
		node.Password = attrs["password"]
		if node.Password == "" && len(positional) >= 1 {
			node.Password = positional[0]
		}
		node.TLS = true
		node.SNI = firstNotEmpty(attrs["sni"], attrs["peer"], attrs["tls-name"], attrs["tls-host"])
		node.SkipCertVerify = isSkipVerify(attrs)
		applyTransportAttrs(&node, attrs)
	case "http", "https":
		node.Type = "http"
		node.Username = firstNotEmpty(attrs["username"], attrs["user"])
		node.Password = attrs["password"]
		if node.Username == "" && len(positional) >= 1 {
			node.Username = positional[0]
		}
		if node.Password == "" && len(positional) >= 2 {
			node.Password = positional[1]
		}
		node.TLS = proxyType == "https" || getBoolFromString(attrs["tls"])
		node.SNI = firstNotEmpty(attrs["sni"], attrs["servername"], attrs["tls-name"], attrs["tls-host"])
		node.SkipCertVerify = isSkipVerify(attrs)
	case "socks5", "socks", "socks5-tls":
		node.Type = "socks5"
		node.Username = firstNotEmpty(attrs["username"], attrs["user"])
		node.Password = attrs["password"]
		if node.Username == "" && len(positional) >= 1 {
			node.Username = positional[0]
		}
		if node.Password == "" && len(positional) >= 2 {
			node.Password = positional[1]
		}
		node.TLS = proxyType == "socks5-tls" || getBoolFromString(attrs["tls"])
		node.SNI = firstNotEmpty(attrs["sni"], attrs["servername"], attrs["tls-name"], attrs["tls-host"])
		node.SkipCertVerify = isSkipVerify(attrs)
	default:
		return ProxyNode{}, ErrUnsupportedNodeType
	}
	return node, nil
}

func SplitSections(content string) map[string][]string {
	sections := map[string][]string{}
	current := ""
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(strings.TrimRight(rawLine, "\r"))
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			current = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		if current != "" {
			sections[current] = append(sections[current], line)
		}
	}
	return sections
}

func getBoolFromString(v string) bool {
	switch strings.ToLower(strings.TrimSpace(strings.ReplaceAll(v, " ", ""))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func parseHeaderHost(v string) string {
	headers := parseSimpleHeaders(v)
	return firstNotEmpty(headers["Host"], headers["host"])
}

func parseProxyTokens(parts []string, allowColon bool) ([]string, map[string]string) {
	attrs := map[string]string{}
	positional := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if k, v, ok := splitKeyValueToken(part, allowColon); ok {
			attrs[k] = trimWrapping(v)
			continue
		}
		positional = append(positional, trimWrapping(part))
	}
	return positional, attrs
}

func splitKeyValueToken(part string, allowColon bool) (string, string, bool) {
	if kv := strings.SplitN(part, "=", 2); len(kv) == 2 {
		return strings.ToLower(strings.TrimSpace(kv[0])), strings.TrimSpace(kv[1]), true
	}
	if allowColon {
		idx := strings.Index(part, ":")
		if idx > 0 {
			key := strings.ToLower(strings.TrimSpace(part[:idx]))
			val := strings.TrimSpace(part[idx+1:])
			if key != "http-header" && key != "ws-headers" && key != "headers" {
				return key, val, true
			}
		}
	}
	return "", "", false
}

func applyTransportAttrs(node *ProxyNode, attrs map[string]string) {
	network := strings.ToLower(firstNotEmpty(attrs["network"], attrs["transport"], attrs["obfs"]))
	switch network {
	case "", "tcp":
		if getBoolFromString(attrs["ws"]) {
			network = "ws"
		}
	case "websocket":
		network = "ws"
	}
	if network != "" {
		node.Network = network
	}
	if node.Network == "ws" {
		node.Path = firstNotEmpty(attrs["ws-path"], attrs["path"], node.Path, "/")
		node.Host = firstNotEmpty(attrs["host"], parseHeaderHost(attrs["ws-headers"]), parseHeaderHost(attrs["headers"]), node.Host)
	}
	if node.Network == "grpc" {
		node.ServiceName = firstNotEmpty(attrs["service-name"], attrs["grpc-service-name"], attrs["servicename"], attrs["grpc-service"])
	}
}

func isSkipVerify(attrs map[string]string) bool {
	if v, ok := attrs["skip-cert-verify"]; ok {
		return getBoolFromString(v)
	}
	if v, ok := attrs["skip-common-name-verify"]; ok {
		return getBoolFromString(v)
	}
	if v, ok := attrs["tls-verification"]; ok && strings.TrimSpace(v) != "" {
		return !getBoolFromString(v)
	}
	return false
}

func normalizeEmptyPlaceholder(v string) string {
	v = trimWrapping(v)
	if v == "{}" || v == "" {
		return ""
	}
	return v
}
