package subscription

import (
	"regexp"
	"strings"

	"github.com/xxf098/lite-proxy/config"
)

var regProfile = regexp.MustCompile(`(?i)(?:vmess|vless|trojan|ssr|ss|http)://[^\s]+`)

func ParseURIText(content string) ([]ProxyNode, error) {
	links := ExtractURILinks(content)
	if len(links) == 0 {
		return nil, ErrNoProxyFound
	}
	nodes := make([]ProxyNode, 0, len(links))
	for _, link := range links {
		node, err := LinkToNode(link)
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

func ExtractURILinks(content string) []string {
	matches := regProfile.FindAllString(content, -1)
	res := make([]string, 0, len(matches))
	for _, link := range matches {
		link = strings.TrimSpace(link)
		if config.RegShadowrocketVmess.MatchString(link) {
			if fixed, err := config.ShadowrocketLinkToVmessLink(link); err == nil {
				link = fixed
			}
		}
		res = append(res, link)
	}
	return res
}

func LinkToNode(link string) (ProxyNode, error) {
	lower := strings.ToLower(link)
	switch {
	case strings.HasPrefix(lower, "vmess://"):
		opt, err := config.VmessLinkToVmessOption(link)
		if err != nil {
			return ProxyNode{}, err
		}
		node := ProxyNode{
			Name:           firstNotEmpty(opt.Name, opt.Server),
			Type:           "vmess",
			Server:         opt.Server,
			Port:           int(opt.Port),
			UUID:           opt.UUID,
			Password:       opt.Password,
			Cipher:         opt.Cipher,
			AlterID:        opt.AlterID,
			TLS:            opt.TLS,
			SkipCertVerify: opt.SkipCertVerify,
			Network:        opt.Network,
			SourceFormat:   FormatURI,
			Raw:            map[string]any{"direct_link": link},
		}
		if opt.ServerName != "" {
			node.SNI = opt.ServerName
		}
		if opt.WSPath != "" {
			node.Path = opt.WSPath
		}
		if h, ok := opt.WSHeaders["Host"]; ok {
			node.Host = h
		}
		if h, ok := opt.WSOpts.Headers["Host"]; ok && node.Host == "" {
			node.Host = h
		}
		return node, nil
	case strings.HasPrefix(lower, "vless://"):
		opt, err := config.VlessLinkToVlessOption(link)
		if err != nil {
			return ProxyNode{}, err
		}
		return ProxyNode{
			Name:           opt.Name,
			Type:           "vless",
			Server:         opt.Server,
			Port:           opt.Port,
			UUID:           opt.UUID,
			TLS:            opt.TLS,
			SNI:            opt.SNI,
			SkipCertVerify: opt.SkipCertVerify,
			Network:        opt.Network,
			Host:           opt.Host,
			Path:           opt.Path,
			Flow:           opt.Flow,
			Fingerprint:    opt.Fingerprint,
			ServiceName:    opt.ServiceName,
			PublicKey:      opt.PublicKey,
			ShortID:        opt.ShortID,
			SourceFormat:   FormatURI,
			Raw:            map[string]any{"direct_link": link},
		}, nil
	case strings.HasPrefix(lower, "trojan://"):
		opt, err := config.TrojanLinkToTrojanOption(link)
		if err != nil {
			return ProxyNode{}, err
		}
		node := ProxyNode{
			Name:           firstNotEmpty(opt.Remarks, opt.Name),
			Type:           "trojan",
			Server:         opt.Server,
			Port:           opt.Port,
			Password:       opt.Password,
			TLS:            true,
			SNI:            opt.SNI,
			ALPN:           append([]string{}, opt.ALPN...),
			SkipCertVerify: opt.SkipCertVerify,
			Network:        opt.Network,
			SourceFormat:   FormatURI,
			Raw:            map[string]any{"direct_link": link},
		}
		if opt.WSOpts.Path != "" {
			node.Path = opt.WSOpts.Path
		}
		if h, ok := opt.WSOpts.Headers["host"]; ok {
			node.Host = h
		}
		if h, ok := opt.WSOpts.Headers["Host"]; ok && node.Host == "" {
			node.Host = h
		}
		node.ServiceName = opt.GrpcOpts.GrpcServiceName
		return node, nil
	case strings.HasPrefix(lower, "ssr://"):
		opt, err := config.SSRLinkToSSROption(link)
		if err != nil {
			return ProxyNode{}, err
		}
		return ProxyNode{
			Name:         firstNotEmpty(opt.Remarks, opt.Name),
			Type:         "ssr",
			Server:       opt.Server,
			Port:         opt.Port,
			Password:     opt.Password,
			Method:       opt.Cipher,
			Cipher:       opt.Cipher,
			Protocol:     opt.Protocol,
			Obfs:         opt.Obfs,
			ObfsParam:    opt.ObfsParam,
			ProtoParam:   opt.ProtocolParam,
			SourceFormat: FormatURI,
			Raw:          map[string]any{"direct_link": link},
		}, nil
	case strings.HasPrefix(lower, "ss://"):
		opt, err := config.SSLinkToSSOption(link)
		if err != nil {
			return ProxyNode{}, err
		}
		return ProxyNode{
			Name:         firstNotEmpty(opt.Remarks, opt.Name),
			Type:         "ss",
			Server:       opt.Server,
			Port:         opt.Port,
			Password:     opt.Password,
			Method:       opt.Cipher,
			Plugin:       opt.Plugin,
			PluginOpts:   opt.PluginOpts,
			SourceFormat: FormatURI,
			Raw:          map[string]any{"direct_link": link},
		}, nil
	case strings.HasPrefix(lower, "http://"):
		opt, err := config.HttpLinkToHttpOption(link)
		if err != nil {
			return ProxyNode{}, err
		}
		return ProxyNode{
			Name:           firstNotEmpty(opt.Remarks, opt.Name),
			Type:           "http",
			Server:         opt.Server,
			Port:           opt.Port,
			Username:       opt.UserName,
			Password:       opt.Password,
			TLS:            opt.TLS,
			SNI:            opt.SNI,
			SkipCertVerify: opt.SkipCertVerify,
			SourceFormat:   FormatURI,
			Raw:            map[string]any{"direct_link": link},
		}, nil
	default:
		return ProxyNode{}, ErrUnsupportedNodeType
	}
}
