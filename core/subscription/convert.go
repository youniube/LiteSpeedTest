package subscription

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

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
	if link, ok := directLinkFromRaw(node); ok && strings.TrimSpace(link) != "" {
		return link, nil
	}
	if link, err := buildDirectLink(node); err == nil && link != "" {
		return link, nil
	}
	mapping := NodeToProxyMapping(node)
	if mapping == nil {
		return "", ErrUnsupportedNodeType
	}
	return config.ParseProxy(mapping, "")
}

func directLinkFromRaw(node ProxyNode) (string, bool) {
	if len(node.Raw) == 0 {
		return "", false
	}
	if link, ok := node.Raw["direct_link"].(string); ok && strings.TrimSpace(link) != "" {
		return link, true
	}
	if link, ok := node.Raw["link"].(string); ok && strings.TrimSpace(link) != "" {
		return link, true
	}
	return "", false
}

func buildDirectLink(node ProxyNode) (string, error) {
	switch normalizeProxyType(node.Type) {
	case "trojan":
		return buildTrojanLink(node)
	case "ss":
		return buildShadowsocksLink(node)
	case "http":
		return buildHttpLink(node)
	case "vless":
		return buildVlessLink(node)
	default:
		return "", ErrUnsupportedNodeType
	}
}

func buildTrojanLink(node ProxyNode) (string, error) {
	if node.Server == "" || node.Port <= 0 || node.Password == "" {
		return "", ErrInvalidSubscription
	}
	u := &url.URL{
		Scheme:   "trojan",
		User:     url.User(node.Password),
		Host:     net.JoinHostPort(node.Server, strconv.Itoa(node.Port)),
		Fragment: node.Name,
	}
	q := url.Values{}
	q.Set("security", "tls")
	if node.SkipCertVerify {
		q.Set("allowInsecure", "1")
	}
	if sni := strings.TrimSpace(node.SNI); sni != "" {
		q.Set("sni", sni)
	}
	switch node.Network {
	case "ws":
		q.Set("type", "ws")
		if path := firstNotEmpty(node.Path, "/"); path != "" {
			q.Set("path", path)
		}
		if host := strings.TrimSpace(node.Host); host != "" {
			q.Set("host", host)
		}
	case "grpc":
		q.Set("type", "grpc")
		if svc := strings.TrimSpace(node.ServiceName); svc != "" {
			q.Set("serviceName", svc)
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func buildShadowsocksLink(node ProxyNode) (string, error) {
	method := firstNotEmpty(node.Method, node.Cipher)
	if node.Server == "" || node.Port <= 0 || method == "" || node.Password == "" {
		return "", ErrInvalidSubscription
	}
	if strings.HasPrefix(strings.ToLower(method), "2022-") {
		return "", ErrUnsupportedNodeType
	}
	u := &url.URL{
		Scheme:   "ss",
		User:     url.UserPassword(method, node.Password),
		Host:     net.JoinHostPort(node.Server, strconv.Itoa(node.Port)),
		Fragment: node.Name,
	}
	if node.Plugin != "" {
		plugin := node.Plugin
		if len(node.PluginOpts) > 0 {
			parts := make([]string, 0, len(node.PluginOpts)+1)
			parts = append(parts, plugin)
			for k, v := range node.PluginOpts {
				if fmt.Sprintf("%v", v) == "" {
					continue
				}
				parts = append(parts, fmt.Sprintf("%s=%v", k, v))
			}
			plugin = strings.Join(parts, ";")
		}
		q := url.Values{}
		q.Set("plugin", plugin)
		u.RawQuery = q.Encode()
	}
	return u.String(), nil
}

func buildHttpLink(node ProxyNode) (string, error) {
	if node.Server == "" || node.Port <= 0 {
		return "", ErrInvalidSubscription
	}
	u := &url.URL{
		Scheme:   "http",
		User:     url.User(node.Password),
		Host:     net.JoinHostPort(node.Server, strconv.Itoa(node.Port)),
		Fragment: node.Name,
	}
	q := url.Values{}
	q.Set("tls", strconv.FormatBool(node.TLS))
	if node.Username != "" {
		q.Set("username", node.Username)
	}
	if node.SkipCertVerify {
		q.Set("allowInsecure", "1")
	}
	if node.SNI != "" {
		q.Set("sni", node.SNI)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func buildVlessLink(node ProxyNode) (string, error) {
	opt := &config.VlessOption{
		Name:           node.Name,
		Server:         node.Server,
		Port:           node.Port,
		UUID:           node.UUID,
		TLS:            node.TLS,
		SNI:            node.SNI,
		ServerName:     node.SNI,
		SkipCertVerify: node.SkipCertVerify,
		Network:        node.Network,
		Flow:           node.Flow,
		Fingerprint:    node.Fingerprint,
		Host:           node.Host,
		Path:           node.Path,
		ServiceName:    node.ServiceName,
		PublicKey:      node.PublicKey,
		ShortID:        node.ShortID,
	}
	if node.SkipCertVerify {
		opt.Insecure = true
	}
	return config.VlessOptionToLink(opt, "")
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
