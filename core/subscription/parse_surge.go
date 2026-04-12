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
	attrs := parseKeyValueTokens(parts[3:])
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
		node.Password = attrs["password"]
	case "vmess":
		node.UUID = firstNotEmpty(attrs["username"], attrs["uuid"], attrs["password"])
		node.AlterID = 0
		node.Cipher = firstNotEmpty(attrs["cipher"], attrs["method"], "none")
		node.TLS = getBoolFromString(attrs["tls"])
		node.SNI = firstNotEmpty(attrs["sni"], attrs["servername"])
		if getBoolFromString(attrs["ws"]) || strings.EqualFold(attrs["transport"], "ws") {
			node.Network = "ws"
			node.Path = firstNotEmpty(attrs["ws-path"], attrs["path"], "/")
			node.Host = firstNotEmpty(attrs["host"], parseHeaderHost(attrs["ws-headers"]))
		} else if strings.EqualFold(attrs["transport"], "grpc") || strings.EqualFold(attrs["type"], "grpc") {
			node.Network = "grpc"
			node.ServiceName = firstNotEmpty(attrs["service-name"], attrs["servicename"])
		}
	case "trojan":
		node.Password = attrs["password"]
		node.TLS = true
		node.SNI = firstNotEmpty(attrs["sni"], attrs["peer"])
		node.SkipCertVerify = getBoolFromString(attrs["skip-cert-verify"])
		if strings.EqualFold(attrs["ws"], "true") || strings.EqualFold(attrs["network"], "ws") {
			node.Network = "ws"
			node.Path = firstNotEmpty(attrs["ws-path"], attrs["path"], "/")
			node.Host = firstNotEmpty(attrs["host"], parseHeaderHost(attrs["ws-headers"]))
		} else if strings.EqualFold(attrs["network"], "grpc") || strings.EqualFold(attrs["type"], "grpc") {
			node.Network = "grpc"
			node.ServiceName = firstNotEmpty(attrs["service-name"], attrs["servicename"])
		}
	case "http", "https":
		node.Type = "http"
		node.Username = firstNotEmpty(attrs["username"], attrs["user"])
		node.Password = attrs["password"]
		node.TLS = proxyType == "https" || getBoolFromString(attrs["tls"])
		node.SNI = firstNotEmpty(attrs["sni"], attrs["servername"])
		node.SkipCertVerify = getBoolFromString(attrs["skip-cert-verify"])
	case "socks5", "socks":
		node.Type = "socks5"
		node.Username = firstNotEmpty(attrs["username"], attrs["user"])
		node.Password = attrs["password"]
		node.TLS = getBoolFromString(attrs["tls"])
		node.SNI = firstNotEmpty(attrs["sni"], attrs["servername"])
		node.SkipCertVerify = getBoolFromString(attrs["skip-cert-verify"])
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
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
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
	switch strings.ToLower(strings.TrimSpace(v)) {
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
