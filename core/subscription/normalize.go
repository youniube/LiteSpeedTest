package subscription

import (
	"fmt"
	"strings"
)

func NormalizeNodes(nodes []ProxyNode) []ProxyNode {
	res := make([]ProxyNode, 0, len(nodes))
	seen := map[string]struct{}{}
	for _, node := range nodes {
		n, ok := NormalizeNode(node)
		if !ok {
			continue
		}
		key := fmt.Sprintf("%s|%s|%d|%s", n.Type, n.Server, n.Port, n.Name)
		if _, exist := seen[key]; exist {
			continue
		}
		seen[key] = struct{}{}
		res = append(res, n)
	}
	return res
}

func NormalizeNode(node ProxyNode) (ProxyNode, bool) {
	n := node.clone()
	n.Type = normalizeProxyType(n.Type)
	n.Server = strings.TrimSpace(n.Server)
	n.Name = strings.TrimSpace(n.Name)
	n.Network = strings.ToLower(strings.TrimSpace(n.Network))
	n.Path = strings.TrimSpace(n.Path)
	n.Host = strings.TrimSpace(n.Host)
	n.SNI = strings.TrimSpace(n.SNI)
	if n.Type == "https" {
		n.Type = "http"
		n.TLS = true
	}
	if n.Port <= 0 || n.Server == "" || n.Type == "" {
		return ProxyNode{}, false
	}
	if n.Name == "" {
		n.Name = fmt.Sprintf("%s:%d", n.Server, n.Port)
	}
	switch n.Type {
	case "socks":
		n.Type = "socks5"
	case "vmess", "vless", "trojan", "ss", "ssr", "http", "socks5":
	default:
		return ProxyNode{}, false
	}
	if n.Network == "" {
		n.Network = "tcp"
	}
	if n.Type == "trojan" && !n.TLS {
		n.TLS = true
	}
	return n, true
}
