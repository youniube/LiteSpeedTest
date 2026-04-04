package singbox

import (
	"fmt"
	"strings"

	"github.com/xxf098/lite-proxy/config"
	"github.com/xxf098/lite-proxy/outbound"
)

func BuildOutbound(link string) (map[string]any, error) {
	switch {
	case hasLinkScheme(link, "vless://"):
		return buildVlessOutbound(link)
	case hasLinkScheme(link, "vmess://"):
		return buildVmessOutbound(link)
	case hasLinkScheme(link, "trojan://"):
		return buildTrojanOutbound(link)
	case hasLinkScheme(link, "ss://"):
		return buildShadowsocksOutbound(link)
	default:
		return nil, fmt.Errorf("unsupported link for sing-box engine")
	}
}

func buildVlessOutbound(link string) (map[string]any, error) {
	opt, err := config.VlessLinkToVlessOption(link)
	if err != nil {
		return nil, err
	}
	opt.Normalize()

	out := map[string]any{
		"type":        "vless",
		"tag":         "proxy",
		"server":      opt.Server,
		"server_port": opt.Port,
		"uuid":        opt.UUID,
	}
	if opt.Flow != "" {
		out["flow"] = opt.Flow
	}
	if tlsObj, err := buildTLSFromVless(opt); err != nil {
		return nil, err
	} else if tlsObj != nil {
		out["tls"] = tlsObj
	}
	if transport, err := buildTransportFromVless(opt); err != nil {
		return nil, err
	} else if transport != nil {
		out["transport"] = transport
	}
	return out, nil
}

func buildVmessOutbound(link string) (map[string]any, error) {
	opt, err := config.VmessLinkToVmessOption(link)
	if err != nil {
		return nil, err
	}

	out := map[string]any{
		"type":        "vmess",
		"tag":         "proxy",
		"server":      opt.Server,
		"server_port": int(opt.Port),
		"uuid":        firstNonEmpty(opt.UUID, opt.Password),
		"security":    normalizeVmessSecurity(opt.Cipher),
	}
	if opt.AlterID > 0 {
		out["alter_id"] = opt.AlterID
	}
	if tlsObj := buildTLSForVMess(opt); tlsObj != nil {
		out["tls"] = tlsObj
	}
	if transport, err := buildTransportFromVMess(opt); err != nil {
		return nil, err
	} else if transport != nil {
		out["transport"] = transport
	}
	return out, nil
}

func buildTrojanOutbound(link string) (map[string]any, error) {
	opt, err := config.TrojanLinkToTrojanOption(link)
	if err != nil {
		return nil, err
	}

	out := map[string]any{
		"type":        "trojan",
		"tag":         "proxy",
		"server":      opt.Server,
		"server_port": opt.Port,
		"password":    opt.Password,
		"tls":         buildTLSForTrojan(opt),
	}
	if transport, err := buildTransportFromTrojan(opt); err != nil {
		return nil, err
	} else if transport != nil {
		out["transport"] = transport
	}
	return out, nil
}

func buildShadowsocksOutbound(link string) (map[string]any, error) {
	opt, err := config.SSLinkToSSOption(link)
	if err != nil {
		return nil, err
	}

	out := map[string]any{
		"type":        "shadowsocks",
		"tag":         "proxy",
		"server":      opt.Server,
		"server_port": opt.Port,
		"method":      normalizeShadowsocksMethod(opt.Cipher),
		"password":    opt.Password,
	}
	if pluginObj := buildShadowsocksPlugin(opt); pluginObj != nil {
		for k, v := range pluginObj {
			out[k] = v
		}
	}
	return out, nil
}

func buildTLSFromVless(opt *config.VlessOption) (map[string]any, error) {
	security := strings.ToLower(strings.TrimSpace(opt.Security))
	if security == "" || security == "none" {
		return nil, nil
	}

	tlsObj := map[string]any{"enabled": true}
	if serverName := firstNonEmpty(opt.SNI, opt.ServerName); serverName != "" {
		tlsObj["server_name"] = serverName
	}
	if opt.Insecure || opt.SkipCertVerify {
		tlsObj["insecure"] = true
	}
	if opt.Fingerprint != "" {
		tlsObj["utls"] = map[string]any{
			"enabled":     true,
			"fingerprint": opt.Fingerprint,
		}
	}
	if security == "reality" {
		if opt.PublicKey == "" {
			return nil, fmt.Errorf("vless reality requires public key")
		}
		reality := map[string]any{
			"enabled":    true,
			"public_key": opt.PublicKey,
		}
		if opt.ShortID != "" {
			reality["short_id"] = opt.ShortID
		}
		tlsObj["reality"] = reality
	}
	return tlsObj, nil
}

func buildTLSForVMess(opt *outbound.VmessOption) map[string]any {
	if !opt.TLS {
		return nil
	}
	tlsObj := map[string]any{"enabled": true}
	if serverName := strings.TrimSpace(opt.ServerName); serverName != "" {
		tlsObj["server_name"] = serverName
	}
	if opt.SkipCertVerify {
		tlsObj["insecure"] = true
	}
	return tlsObj
}

func buildTLSForTrojan(opt *outbound.TrojanOption) map[string]any {
	tlsObj := map[string]any{"enabled": true}
	if serverName := strings.TrimSpace(opt.SNI); serverName != "" {
		tlsObj["server_name"] = serverName
	}
	if opt.SkipCertVerify {
		tlsObj["insecure"] = true
	}
	if len(opt.ALPN) > 0 {
		tlsObj["alpn"] = compactStrings(opt.ALPN)
	}
	return tlsObj
}

func buildTransportFromVless(opt *config.VlessOption) (map[string]any, error) {
	switch strings.ToLower(strings.TrimSpace(opt.Network)) {
	case "", "tcp":
		return nil, nil
	case "ws":
		t := map[string]any{"type": "ws"}
		if opt.Path != "" {
			t["path"] = opt.Path
		}
		if opt.Host != "" {
			t["headers"] = map[string]any{"Host": opt.Host}
		}
		return t, nil
	case "grpc":
		t := map[string]any{"type": "grpc"}
		if opt.ServiceName != "" {
			t["service_name"] = opt.ServiceName
		}
		return t, nil
	case "httpupgrade", "http-upgrade":
		t := map[string]any{"type": "httpupgrade"}
		if opt.Host != "" {
			t["host"] = opt.Host
		}
		if opt.Path != "" {
			t["path"] = opt.Path
		}
		return t, nil
	case "http", "h2":
		t := map[string]any{"type": "http"}
		if opt.Host != "" {
			t["host"] = []string{opt.Host}
		}
		if opt.Path != "" {
			t["path"] = opt.Path
		}
		return t, nil
	case "quic":
		return map[string]any{"type": "quic"}, nil
	default:
		return nil, fmt.Errorf("unsupported vless transport for sing-box: %s", opt.Network)
	}
}

func buildTransportFromVMess(opt *outbound.VmessOption) (map[string]any, error) {
	switch network := strings.ToLower(strings.TrimSpace(opt.Network)); network {
	case "", "tcp":
		return nil, nil
	case "ws":
		t := map[string]any{"type": "ws"}
		if path := strings.TrimSpace(opt.WSPath); path != "" {
			t["path"] = path
		}
		if headers := normalizeStringHeaders(opt.WSHeaders); len(headers) > 0 {
			t["headers"] = headers
		}
		return t, nil
	case "grpc":
		return map[string]any{"type": "grpc"}, nil
	case "httpupgrade", "http-upgrade":
		t := map[string]any{"type": "httpupgrade"}
		if host := firstHeaderValue(opt.WSHeaders, "Host", "host"); host != "" {
			t["host"] = host
		}
		if path := strings.TrimSpace(opt.WSPath); path != "" {
			t["path"] = path
		}
		return t, nil
	case "http":
		t := map[string]any{"type": "http"}
		if hosts := headerValues(opt.HTTPOpts.Headers, "Host", "host"); len(hosts) > 0 {
			t["host"] = hosts
		}
		if paths := compactStrings(opt.HTTPOpts.Path); len(paths) > 0 {
			t["path"] = paths[0]
		}
		if method := strings.TrimSpace(opt.HTTPOpts.Method); method != "" {
			t["method"] = method
		}
		return t, nil
	case "h2":
		t := map[string]any{"type": "http"}
		if hosts := compactStrings(opt.HTTP2Opts.Host); len(hosts) > 0 {
			t["host"] = hosts
		}
		if path := strings.TrimSpace(opt.HTTP2Opts.Path); path != "" {
			t["path"] = path
		}
		return t, nil
	case "quic":
		return map[string]any{"type": "quic"}, nil
	default:
		return nil, fmt.Errorf("unsupported vmess transport for sing-box: %s", network)
	}
}

func buildTransportFromTrojan(opt *outbound.TrojanOption) (map[string]any, error) {
	switch network := strings.ToLower(strings.TrimSpace(opt.Network)); network {
	case "", "tcp":
		return nil, nil
	case "ws":
		t := map[string]any{"type": "ws"}
		if path := strings.TrimSpace(opt.WSOpts.Path); path != "" {
			t["path"] = path
		}
		if headers := normalizeStringHeaders(opt.WSOpts.Headers); len(headers) > 0 {
			t["headers"] = headers
		}
		return t, nil
	case "grpc":
		t := map[string]any{"type": "grpc"}
		if serviceName := strings.TrimSpace(opt.GrpcOpts.GrpcServiceName); serviceName != "" {
			t["service_name"] = serviceName
		}
		return t, nil
	default:
		return nil, fmt.Errorf("unsupported trojan transport for sing-box: %s", network)
	}
}

func buildShadowsocksPlugin(opt *outbound.ShadowSocksOption) map[string]any {
	plugin := strings.TrimSpace(opt.Plugin)
	if plugin == "" {
		return nil
	}
	out := map[string]any{}
	switch plugin {
	case "obfs":
		out["plugin"] = "obfs-local"
	case "obfs-local", "v2ray-plugin":
		out["plugin"] = plugin
	default:
		out["plugin"] = plugin
	}
	if len(opt.PluginOpts) > 0 {
		out["plugin_opts"] = opt.PluginOpts
	}
	return out
}

func hasLinkScheme(link, scheme string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(link)), scheme)
}

func normalizeVmessSecurity(security string) string {
	security = strings.ToLower(strings.TrimSpace(security))
	switch security {
	case "", "auto":
		return "auto"
	case "aes-128-gcm", "chacha20-poly1305", "zero", "none":
		return security
	default:
		return security
	}
}

func normalizeShadowsocksMethod(method string) string {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "chacha20-poly1305":
		return "chacha20-ietf-poly1305"
	case "xchacha20-poly1305":
		return "xchacha20-ietf-poly1305"
	default:
		return method
	}
}

func normalizeStringHeaders(headers map[string]string) map[string]any {
	if len(headers) == 0 {
		return nil
	}
	out := map[string]any{}
	for k, v := range headers {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func headerValues(headers map[string][]string, keys ...string) []string {
	for _, key := range keys {
		if values, ok := headers[key]; ok {
			return compactStrings(values)
		}
	}
	return nil
}

func firstHeaderValue(headers map[string]string, keys ...string) string {
	for _, key := range keys {
		if value, ok := headers[key]; ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func compactStrings(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
