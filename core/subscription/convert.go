package subscription

import (
	"fmt"

	"github.com/xxf098/lite-proxy/config"
)

func ConvertNodesToLinks(nodes []ProxyNode) ([]string, error) {
	links := make([]string, 0, len(nodes))
	for _, node := range nodes {
		link, err := ConvertNodeToLink(node)
		if err != nil {
			continue
		}
		if link != "" {
			links = append(links, link)
		}
	}
	if len(links) == 0 {
		return nil, ErrNoProxyFound
	}
	return links, nil
}

func ConvertNodeToLink(node ProxyNode) (string, error) {
	normalized := NormalizeNodes([]ProxyNode{node})
	if len(normalized) == 0 {
		return "", ErrNoProxyFound
	}
	node = normalized[0]
	mapping := NodeToProxyMapping(node)
	if mapping == nil {
		return "", ErrUnsupportedNodeType
	}
	return config.ParseProxy(mapping, "")
}

func NodeToProxyMapping(node ProxyNode) map[string]any {
	t := normalizeProxyType(node.Type)
	switch t {
	case "ss", "ssr", "vmess", "trojan", "http", "vless":
	default:
		return nil
	}
	m := map[string]any{
		"type":   t,
		"name":   node.Name,
		"server": node.Server,
		"port":   node.Port,
	}
	switch t {
	case "ss":
		m["cipher"] = node.Method
		m["password"] = node.Password
		if node.Plugin != "" {
			m["plugin"] = node.Plugin
		}
		if len(node.PluginOpts) > 0 {
			m["plugin-opts"] = node.PluginOpts
		}
	case "ssr":
		m["cipher"] = firstNotEmpty(node.Cipher, node.Method)
		m["password"] = node.Password
		m["protocol"] = node.Protocol
		m["obfs"] = node.Obfs
		if node.ObfsParam != "" {
			m["obfs-param"] = node.ObfsParam
		}
		if node.ProtoParam != "" {
			m["protocol-param"] = node.ProtoParam
		}
	case "vmess":
		m["uuid"] = firstNotEmpty(node.UUID, node.Password)
		m["alterId"] = node.AlterID
		m["cipher"] = firstNotEmpty(node.Cipher, "none")
		if node.TLS {
			m["tls"] = true
		}
		if node.SkipCertVerify {
			m["skip-cert-verify"] = true
		}
		if node.SNI != "" {
			m["servername"] = node.SNI
		}
		if node.Network != "" && node.Network != "tcp" {
			m["network"] = node.Network
		}
		if node.Network == "ws" {
			m["ws-path"] = firstNotEmpty(node.Path, "/")
			if node.Host != "" {
				m["ws-headers"] = map[string]string{"Host": node.Host}
			}
		}
	case "trojan":
		m["password"] = node.Password
		if node.TLS {
			m["sni"] = node.SNI
		}
		if node.SkipCertVerify {
			m["skip-cert-verify"] = true
		}
		if node.Network != "" && node.Network != "tcp" {
			m["network"] = node.Network
		}
		if node.Network == "ws" {
			m["ws-opts"] = map[string]any{"path": firstNotEmpty(node.Path, "/")}
			if node.Host != "" {
				m["ws-opts"] = map[string]any{"path": firstNotEmpty(node.Path, "/"), "headers": map[string]string{"Host": node.Host}}
			}
		}
		if node.Network == "grpc" && node.ServiceName != "" {
			m["grpc-opts"] = map[string]any{"grpc-service-name": node.ServiceName}
		}
	case "http":
		m["username"] = node.Username
		m["password"] = node.Password
		if node.TLS {
			m["tls"] = true
		}
		if node.SNI != "" {
			m["sni"] = node.SNI
		}
		if node.SkipCertVerify {
			m["skip-cert-verify"] = true
		}
	case "vless":
		m["uuid"] = node.UUID
		if node.TLS {
			m["tls"] = true
		}
		if node.SNI != "" {
			m["sni"] = node.SNI
		}
		if node.SkipCertVerify {
			m["skip-cert-verify"] = true
			m["allowInsecure"] = true
		}
		if node.Network != "" && node.Network != "tcp" {
			m["network"] = node.Network
		}
		if node.Flow != "" {
			m["flow"] = node.Flow
		}
		if node.Fingerprint != "" {
			m["client-fingerprint"] = node.Fingerprint
		}
		if node.Network == "ws" {
			ws := map[string]any{"path": firstNotEmpty(node.Path, "/")}
			if node.Host != "" {
				ws["headers"] = map[string]string{"Host": node.Host}
			}
			m["ws-opts"] = ws
		}
		if node.Network == "grpc" && node.ServiceName != "" {
			m["grpc-opts"] = map[string]any{"grpc-service-name": node.ServiceName}
		}
		if node.PublicKey != "" || node.ShortID != "" {
			m["reality-opts"] = map[string]any{"public-key": node.PublicKey, "short-id": node.ShortID}
		}
	}
	return m
}

func firstNotEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func debugNode(node ProxyNode) string {
	return fmt.Sprintf("%s|%s|%s|%d", node.Type, node.Name, node.Server, node.Port)
}
