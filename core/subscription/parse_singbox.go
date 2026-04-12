package subscription

import (
	"encoding/json"
	"fmt"
)

func ParseSingBoxJSON(content string) ([]ProxyNode, error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return nil, err
	}
	list, ok := raw["outbounds"].([]any)
	if !ok {
		return nil, ErrNoProxyFound
	}
	nodes := make([]ProxyNode, 0, len(list))
	for _, item := range list {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		node, err := ParseSingBoxOutbound(m)
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

func ParseSingBoxOutbound(m map[string]any) (ProxyNode, error) {
	proxyType := normalizeProxyType(getString(m, "type"))
	switch proxyType {
	case "selector", "urltest", "direct", "block", "dns":
		return ProxyNode{}, ErrUnsupportedNodeType
	}
	node := ProxyNode{
		Name:           firstNotEmpty(getString(m, "tag"), getString(m, "name")),
		Type:           proxyType,
		Server:         getString(m, "server"),
		Port:           getInt(m, "server_port", "port"),
		UUID:           getString(m, "uuid"),
		Password:       getString(m, "password"),
		Method:         getString(m, "method"),
		TLS:            getBool(m, "tls", "tls_enabled"),
		SNI:            getString(m, "server_name", "sni"),
		SkipCertVerify: getBool(m, "insecure", "skip-cert-verify"),
		SourceFormat:   FormatSingBox,
		Raw:            cloneMap(m),
	}
	if node.Type == "shadowsocks" {
		node.Type = "ss"
	}
	if node.Type == "socks" {
		node.Type = "socks5"
		node.Username = getString(m, "username", "user")
	}
	if node.Type == "http" {
		node.Username = getString(m, "username", "user")
	}
	if node.Type == "vmess" {
		node.AlterID = getInt(m, "alter_id")
		node.Cipher = firstNotEmpty(getString(m, "security"), "none")
	}
	if node.Type == "vless" {
		node.Flow = getString(m, "flow")
		node.Fingerprint = getString(m, "tls_fingerprint")
		node.PublicKey = getString(m, "public_key")
		node.ShortID = getString(m, "short_id")
	}
	if transport := getMap(m, "transport"); len(transport) > 0 {
		node.Network = normalizeProxyType(getString(transport, "type"))
		switch node.Network {
		case "ws":
			node.Path = getString(transport, "path")
			if headers := getMap(transport, "headers"); len(headers) > 0 {
				node.Host = getString(headers, "Host", "host")
			}
		case "grpc":
			node.ServiceName = getString(transport, "service_name")
		}
	}
	switch node.Type {
	case "ss", "vmess", "vless", "trojan", "http", "socks5":
		return node, nil
	default:
		return ProxyNode{}, fmt.Errorf("unsupported sing-box type: %s", node.Type)
	}
}
