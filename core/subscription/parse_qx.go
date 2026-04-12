package subscription

import (
	"fmt"
	"strings"
)

func ParseQuantumultXText(content string) ([]ProxyNode, error) {
	sections := SplitSections(content)
	lines := append([]string{}, sections["server_local"]...)
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(strings.TrimRight(rawLine, "\r"))
		if regQXLine.MatchString(line) {
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return nil, ErrNoProxyFound
	}
	nodes := make([]ProxyNode, 0, len(lines))
	seen := map[string]struct{}{}
	for _, line := range lines {
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		node, err := ParseQXNodeLine(line)
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

func ParseQXNodeLine(line string) (ProxyNode, error) {
	kv := strings.SplitN(line, "=", 2)
	if len(kv) != 2 {
		return ProxyNode{}, fmt.Errorf("invalid qx node line")
	}
	proxyType := normalizeProxyType(kv[0])
	parts := splitCSV(kv[1])
	if len(parts) == 0 {
		return ProxyNode{}, fmt.Errorf("invalid qx node line")
	}
	host, port, err := splitHostPort(parts[0])
	if err != nil {
		return ProxyNode{}, err
	}
	attrs := parseKeyValueTokens(parts[1:])
	node := ProxyNode{
		Name:         firstNotEmpty(attrs["tag"], attrs["remarks"], host),
		Type:         proxyType,
		Server:       host,
		Port:         port,
		SourceFormat: FormatQX,
		Raw:          map[string]any{"line": line},
	}
	switch proxyType {
	case "ss":
		node.Method = firstNotEmpty(attrs["method"], attrs["cipher"])
		node.Password = attrs["password"]
	case "vmess":
		node.UUID = firstNotEmpty(attrs["password"], attrs["uuid"])
		node.Cipher = firstNotEmpty(attrs["method"], "none")
		node.Network = strings.ToLower(firstNotEmpty(attrs["obfs"], attrs["transport"], "tcp"))
		node.Path = firstNotEmpty(attrs["obfs-uri"], attrs["path"])
		node.Host = firstNotEmpty(attrs["obfs-host"], attrs["host"])
		node.TLS = getBoolFromString(attrs["over-tls"]) || getBoolFromString(attrs["tls"])
		node.SNI = firstNotEmpty(attrs["tls-host"], attrs["sni"])
		node.SkipCertVerify = getBoolFromString(attrs["tls-verification"]) == false && attrs["tls-verification"] != ""
	case "trojan":
		node.Password = attrs["password"]
		node.TLS = true
		node.SNI = firstNotEmpty(attrs["tls-host"], attrs["sni"])
		node.SkipCertVerify = getBoolFromString(attrs["tls-verification"]) == false && attrs["tls-verification"] != ""
		node.Network = strings.ToLower(firstNotEmpty(attrs["obfs"], attrs["transport"], "tcp"))
		node.Path = firstNotEmpty(attrs["obfs-uri"], attrs["path"])
		node.Host = firstNotEmpty(attrs["obfs-host"], attrs["host"])
	case "http":
		node.Username = attrs["username"]
		node.Password = attrs["password"]
		node.TLS = getBoolFromString(attrs["over-tls"]) || getBoolFromString(attrs["tls"])
		node.SNI = firstNotEmpty(attrs["tls-host"], attrs["sni"])
		node.SkipCertVerify = getBoolFromString(attrs["tls-verification"]) == false && attrs["tls-verification"] != ""
	case "socks5":
		node.Username = attrs["username"]
		node.Password = attrs["password"]
	default:
		return ProxyNode{}, ErrUnsupportedNodeType
	}
	return node, nil
}
