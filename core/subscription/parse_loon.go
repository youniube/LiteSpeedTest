package subscription

import (
	"fmt"
	"strings"
)

func ParseLoonText(content string) ([]ProxyNode, error) {
	sections := SplitSections(content)
	lines := sections["proxy"]
	if len(lines) == 0 {
		return nil, ErrNoProxyFound
	}
	nodes := make([]ProxyNode, 0, len(lines))
	for _, line := range lines {
		node, err := ParseLoonProxyLine(line)
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

func ParseLoonProxyLine(line string) (ProxyNode, error) {
	kv := strings.SplitN(line, "=", 2)
	if len(kv) != 2 {
		return ProxyNode{}, fmt.Errorf("invalid loon proxy line")
	}
	name := trimWrapping(kv[0])
	parts := splitCSV(kv[1])
	if len(parts) < 3 {
		return ProxyNode{}, fmt.Errorf("invalid loon proxy fields")
	}
	proxyType := normalizeProxyType(parts[0])
	host := trimWrapping(parts[1])
	port := 0
	fmt.Sscanf(parts[2], "%d", &port)
	positional, attrs := parseProxyTokens(parts[3:], true)
	node := ProxyNode{
		Name:         name,
		Type:         proxyType,
		Server:       host,
		Port:         port,
		SourceFormat: FormatLoon,
		Raw:          map[string]any{"line": line},
	}
	switch proxyType {
	case "http", "https":
		node.Type = "http"
		if len(positional) >= 1 {
			node.Username = positional[0]
		}
		if len(positional) >= 2 {
			node.Password = positional[1]
		}
		node.Username = firstNotEmpty(node.Username, attrs["username"], attrs["user"])
		node.Password = firstNotEmpty(node.Password, attrs["password"])
		node.TLS = proxyType == "https" || getBoolFromString(firstNotEmpty(attrs["tls"], attrs["over-tls"]))
		node.SNI = firstNotEmpty(attrs["tls-name"], attrs["tls-host"], attrs["sni"], attrs["servername"])
		node.SkipCertVerify = isSkipVerify(attrs)
	case "ss":
		if len(positional) >= 1 {
			node.Method = positional[0]
		}
		if len(positional) >= 2 {
			node.Password = positional[1]
		}
		node.Method = firstNotEmpty(node.Method, attrs["method"], attrs["cipher"], attrs["encrypt-method"])
		node.Password = firstNotEmpty(node.Password, attrs["password"])
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
		if len(positional) >= 1 {
			node.Cipher = positional[0]
		}
		if len(positional) >= 2 {
			node.UUID = positional[1]
		}
		node.Cipher = firstNotEmpty(node.Cipher, attrs["method"], attrs["cipher"], "none")
		node.UUID = firstNotEmpty(node.UUID, attrs["username"], attrs["uuid"], attrs["password"])
		node.TLS = getBoolFromString(firstNotEmpty(attrs["over-tls"], attrs["tls"]))
		node.SNI = firstNotEmpty(attrs["tls-name"], attrs["tls-host"], attrs["sni"], attrs["servername"])
		node.SkipCertVerify = isSkipVerify(attrs)
		applyTransportAttrs(&node, attrs)
	case "trojan":
		if len(positional) >= 1 {
			node.Password = positional[0]
		}
		node.Password = firstNotEmpty(node.Password, attrs["password"])
		node.TLS = true
		node.SNI = firstNotEmpty(attrs["tls-name"], attrs["tls-host"], attrs["sni"], attrs["peer"])
		node.SkipCertVerify = isSkipVerify(attrs)
		applyTransportAttrs(&node, attrs)
	case "socks5", "socks":
		node.Type = "socks5"
		if len(positional) >= 1 {
			node.Username = positional[0]
		}
		if len(positional) >= 2 {
			node.Password = positional[1]
		}
		node.Username = firstNotEmpty(node.Username, attrs["username"], attrs["user"])
		node.Password = firstNotEmpty(node.Password, attrs["password"])
		node.TLS = getBoolFromString(firstNotEmpty(attrs["tls"], attrs["over-tls"]))
		node.SNI = firstNotEmpty(attrs["tls-name"], attrs["tls-host"], attrs["sni"], attrs["servername"])
		node.SkipCertVerify = isSkipVerify(attrs)
	default:
		return ProxyNode{}, ErrUnsupportedNodeType
	}
	return node, nil
}
