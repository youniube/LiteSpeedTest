package subscription

import "strings"

type InputFormat string

const (
	FormatUnknown InputFormat = "unknown"
	FormatURI     InputFormat = "uri"
	FormatBase64  InputFormat = "base64"
	FormatClash   InputFormat = "clash"
	FormatSurge   InputFormat = "surge"
	FormatLoon    InputFormat = "loon"
	FormatQX      InputFormat = "quantumultx"
	FormatSingBox InputFormat = "singbox"
)

type ProxyNode struct {
	Name   string
	Type   string
	Server string
	Port   int

	UUID     string
	Password string
	Method   string

	TLS            bool
	SNI            string
	ALPN           []string
	SkipCertVerify bool

	Network string
	Host    string
	Path    string
	Flow    string

	UDP bool

	Plugin     string
	PluginOpts map[string]any

	Username    string
	Fingerprint string
	ServiceName string
	PublicKey   string
	ShortID     string
	Cipher      string
	Protocol    string
	Obfs        string
	ObfsParam   string
	ProtoParam  string
	AlterID     int

	SourceFormat InputFormat
	Raw          map[string]any
}

func (n ProxyNode) clone() ProxyNode {
	cp := n
	if len(n.ALPN) > 0 {
		cp.ALPN = append([]string{}, n.ALPN...)
	}
	if len(n.PluginOpts) > 0 {
		cp.PluginOpts = make(map[string]any, len(n.PluginOpts))
		for k, v := range n.PluginOpts {
			cp.PluginOpts[k] = v
		}
	}
	if len(n.Raw) > 0 {
		cp.Raw = make(map[string]any, len(n.Raw))
		for k, v := range n.Raw {
			cp.Raw[k] = v
		}
	}
	return cp
}

func normalizeProxyType(t string) string {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "shadowsocks":
		return "ss"
	case "shadowsocksr":
		return "ssr"
	case "socks":
		return "socks5"
	case "https":
		return "https"
	default:
		return strings.ToLower(strings.TrimSpace(t))
	}
}
