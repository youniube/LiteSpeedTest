package singbox

import (
	"fmt"
	"strings"

	"github.com/xxf098/lite-proxy/config"
)

func BuildOutbound(link string) (map[string]any, error) {
	switch {
	case strings.HasPrefix(strings.ToLower(link), "vless://"):
		return buildVlessOutbound(link)
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

	if tlsObj, err := buildTLS(opt); err != nil {
		return nil, err
	} else if tlsObj != nil {
		out["tls"] = tlsObj
	}
	if transport, err := buildTransport(opt); err != nil {
		return nil, err
	} else if transport != nil {
		out["transport"] = transport
	}
	return out, nil
}

func buildTLS(opt *config.VlessOption) (map[string]any, error) {
	security := strings.ToLower(opt.Security)
	if security == "" || security == "none" {
		return nil, nil
	}

	tlsObj := map[string]any{"enabled": true}
	serverName := firstNonEmpty(opt.SNI, opt.ServerName)
	if serverName != "" {
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
		if opt.ShortID == "" {
			return nil, fmt.Errorf("vless reality requires short id")
		}
		tlsObj["reality"] = map[string]any{
			"enabled":    true,
			"public_key": opt.PublicKey,
			"short_id":   opt.ShortID,
		}
	}
	return tlsObj, nil
}

func buildTransport(opt *config.VlessOption) (map[string]any, error) {
	switch strings.ToLower(opt.Network) {
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
	case "http":
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

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
