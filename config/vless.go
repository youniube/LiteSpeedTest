package config

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type VlessWSOptions struct {
	Path    string            `proxy:"path,omitempty"`
	Headers map[string]string `proxy:"headers,omitempty"`
}

type VlessGrpcOptions struct {
	GrpcServiceName string `proxy:"grpc-service-name,omitempty"`
}

type VlessRealityOptions struct {
	PublicKey string `proxy:"public-key,omitempty"`
	ShortID   string `proxy:"short-id,omitempty"`
}

type VlessOption struct {
	Name string `proxy:"name,omitempty"`

	Server string `proxy:"server"`
	Port   int    `proxy:"port"`
	UUID   string `proxy:"uuid"`

	TLS            bool   `proxy:"tls,omitempty"`
	ServerName     string `proxy:"servername,omitempty"`
	SkipCertVerify bool   `proxy:"skip-cert-verify,omitempty"`
	UDP            bool   `proxy:"udp,omitempty"`
	Network        string `proxy:"network,omitempty"`
	Flow           string `proxy:"flow,omitempty"`
	Fingerprint    string `proxy:"client-fingerprint,omitempty"`

	WSOpts      VlessWSOptions      `proxy:"ws-opts,omitempty"`
	GrpcOpts    VlessGrpcOptions    `proxy:"grpc-opts,omitempty"`
	RealityOpts VlessRealityOptions `proxy:"reality-opts,omitempty"`

	Security    string `proxy:"security,omitempty"`
	SNI         string `proxy:"sni,omitempty"`
	Host        string `proxy:"host,omitempty"`
	Path        string `proxy:"path,omitempty"`
	ServiceName string `proxy:"serviceName,omitempty"`
	PublicKey   string `proxy:"pbk,omitempty"`
	ShortID     string `proxy:"sid,omitempty"`
	Insecure    bool   `proxy:"allowInsecure,omitempty"`
}

func (o *VlessOption) Normalize() {
	if o.Network == "" {
		o.Network = "tcp"
	}
	o.Network = strings.ToLower(o.Network)

	if o.SNI == "" && o.ServerName != "" {
		o.SNI = o.ServerName
	}
	if o.ServerName == "" && o.SNI != "" {
		o.ServerName = o.SNI
	}
	if o.Insecure || o.SkipCertVerify {
		o.Insecure = true
		o.SkipCertVerify = true
	}
	if o.Path == "" && o.WSOpts.Path != "" {
		o.Path = o.WSOpts.Path
	}
	if o.Host == "" && len(o.WSOpts.Headers) > 0 {
		if h, ok := o.WSOpts.Headers["Host"]; ok {
			o.Host = h
		} else if h, ok := o.WSOpts.Headers["host"]; ok {
			o.Host = h
		}
	}
	if o.ServiceName == "" && o.GrpcOpts.GrpcServiceName != "" {
		o.ServiceName = o.GrpcOpts.GrpcServiceName
	}
	if o.PublicKey == "" && o.RealityOpts.PublicKey != "" {
		o.PublicKey = o.RealityOpts.PublicKey
	}
	if o.ShortID == "" && o.RealityOpts.ShortID != "" {
		o.ShortID = o.RealityOpts.ShortID
	}

	o.Security = strings.ToLower(o.Security)
	switch {
	case o.Security != "":
	case o.PublicKey != "" || o.ShortID != "":
		o.Security = "reality"
	case o.TLS:
		o.Security = "tls"
	default:
		o.Security = "none"
	}
}

func VlessLinkToVlessOption(link string) (*VlessOption, error) {
	u, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(u.Scheme, "vless") {
		return nil, fmt.Errorf("not a vless link")
	}
	if u.Hostname() == "" {
		return nil, fmt.Errorf("missing server")
	}
	if u.Port() == "" {
		return nil, fmt.Errorf("missing port")
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, err
	}

	name, _ := url.QueryUnescape(u.Fragment)
	q := u.Query()
	path := q.Get("path")
	if path != "" {
		if unescaped, err := url.QueryUnescape(path); err == nil {
			path = unescaped
		}
	}

	opt := &VlessOption{
		Name:        name,
		Server:      u.Hostname(),
		Port:        port,
		UUID:        u.User.Username(),
		Network:     strings.ToLower(q.Get("type")),
		Security:    strings.ToLower(q.Get("security")),
		Flow:        q.Get("flow"),
		SNI:         q.Get("sni"),
		Host:        q.Get("host"),
		Path:        path,
		ServiceName: firstNonEmpty(q.Get("serviceName"), q.Get("service_name")),
		Fingerprint: firstNonEmpty(q.Get("fp"), q.Get("fingerprint")),
		PublicKey:   firstNonEmpty(q.Get("pbk"), q.Get("publicKey")),
		ShortID:     firstNonEmpty(q.Get("sid"), q.Get("shortId")),
	}
	if isTrue(q.Get("allowInsecure")) || isTrue(q.Get("insecure")) {
		opt.Insecure = true
		opt.SkipCertVerify = true
	}
	if opt.SNI != "" {
		opt.ServerName = opt.SNI
	}
	if opt.Security == "" && isTrue(q.Get("tls")) {
		opt.TLS = true
	}
	if opt.Security == "tls" || opt.Security == "reality" {
		opt.TLS = true
	}
	opt.Normalize()
	if opt.UUID == "" {
		return nil, fmt.Errorf("missing uuid")
	}
	return opt, nil
}

func VlessOptionToLink(o *VlessOption, namePrefix string) (string, error) {
	if o == nil {
		return "", fmt.Errorf("nil vless option")
	}
	if o.Server == "" || o.Port <= 0 || o.UUID == "" {
		return "", fmt.Errorf("invalid vless option")
	}
	opt := *o
	opt.Normalize()

	u := &url.URL{Scheme: "vless", Host: net.JoinHostPort(opt.Server, strconv.Itoa(opt.Port))}
	u.User = url.User(opt.UUID)
	q := url.Values{}
	if opt.Network != "" && opt.Network != "tcp" {
		q.Set("type", opt.Network)
	}
	if opt.Security != "" && opt.Security != "none" {
		q.Set("security", opt.Security)
	}
	if opt.Flow != "" {
		q.Set("flow", opt.Flow)
	}
	if opt.SNI != "" {
		q.Set("sni", opt.SNI)
	}
	if opt.Host != "" {
		q.Set("host", opt.Host)
	}
	if opt.Path != "" {
		q.Set("path", opt.Path)
	}
	if opt.ServiceName != "" {
		q.Set("serviceName", opt.ServiceName)
	}
	if opt.Fingerprint != "" {
		q.Set("fp", opt.Fingerprint)
	}
	if opt.PublicKey != "" {
		q.Set("pbk", opt.PublicKey)
	}
	if opt.ShortID != "" {
		q.Set("sid", opt.ShortID)
	}
	if opt.Insecure || opt.SkipCertVerify {
		q.Set("allowInsecure", "1")
	}
	u.RawQuery = q.Encode()
	if opt.Name != "" || namePrefix != "" {
		u.Fragment = url.QueryEscape(namePrefix + opt.Name)
	}
	return u.String(), nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func isTrue(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
