package subscription

import (
	"fmt"
	"strings"

	"github.com/xxf098/lite-proxy/config"
)

func ParseClashText(content string) ([]ProxyNode, error) {
	rawCfg, err := config.UnmarshalRawConfig([]byte(content))
	if err != nil {
		return nil, err
	}
	nodes := make([]ProxyNode, 0, len(rawCfg.Proxy))
	for _, mapping := range rawCfg.Proxy {
		node, err := ParseClashMapping(mapping)
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

func ParseClashMapping(mapping map[string]any) (ProxyNode, error) {
	proxyType := normalizeProxyType(getString(mapping, "type"))
	if proxyType == "" {
		return ProxyNode{}, fmt.Errorf("missing type")
	}
	if proxyType == "socks5" {
		return ProxyNode{
			Name:           getString(mapping, "name"),
			Type:           "socks5",
			Server:         getString(mapping, "server"),
			Port:           getInt(mapping, "port"),
			Username:       getString(mapping, "username", "user"),
			Password:       getString(mapping, "password", "passwd"),
			TLS:            getBool(mapping, "tls"),
			SNI:            getString(mapping, "sni", "servername"),
			SkipCertVerify: getBool(mapping, "skip-cert-verify", "skip-cert-verify"),
			SourceFormat:   FormatClash,
			Raw:            cloneMap(mapping),
		}, nil
	}
	if proxyType == "https" {
		proxyType = "http"
	}
	link, err := config.ParseProxy(mapping, "")
	if err != nil {
		return ProxyNode{}, err
	}
	node, err := LinkToNode(link)
	if err != nil {
		return ProxyNode{}, err
	}
	node.Type = proxyType
	node.SourceFormat = FormatClash
	node.Raw = cloneMap(mapping)
	if proxyType == "vmess" {
		if node.Host == "" {
			if wsHeaders := getMap(mapping, "ws-headers"); len(wsHeaders) > 0 {
				node.Host = getString(wsHeaders, "Host", "host")
			}
			if wsOpts := getMap(mapping, "ws-opts"); len(wsOpts) > 0 {
				node.Path = firstNotEmpty(node.Path, getString(wsOpts, "path"))
				if headers := getMap(wsOpts, "headers"); len(headers) > 0 {
					node.Host = firstNotEmpty(node.Host, getString(headers, "Host", "host"))
				}
			}
		}
	}
	if proxyType == "vless" {
		node.Flow = firstNotEmpty(node.Flow, getString(mapping, "flow"))
		node.Fingerprint = firstNotEmpty(node.Fingerprint, getString(mapping, "client-fingerprint"))
		if wsOpts := getMap(mapping, "ws-opts"); len(wsOpts) > 0 {
			node.Path = firstNotEmpty(node.Path, getString(wsOpts, "path"))
			if headers := getMap(wsOpts, "headers"); len(headers) > 0 {
				node.Host = firstNotEmpty(node.Host, getString(headers, "Host", "host"))
			}
		}
		if reality := getMap(mapping, "reality-opts"); len(reality) > 0 {
			node.PublicKey = firstNotEmpty(node.PublicKey, getString(reality, "public-key"))
			node.ShortID = firstNotEmpty(node.ShortID, getString(reality, "short-id"))
		}
	}
	if proxyType == "trojan" {
		if strings.TrimSpace(node.SNI) == "" {
			node.SNI = getString(mapping, "sni")
		}
	}
	return node, nil
}
