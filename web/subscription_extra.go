package web

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/xxf098/lite-proxy/config"
)

func parseExtraContent(data string) ([]string, error) {
	data = strings.TrimSpace(strings.ReplaceAll(data, "\r\n", "\n"))
	if data == "" {
		return nil, fmt.Errorf("empty input")
	}

	if links, err := parseSingBoxJSONContent(data); err == nil && len(links) > 0 {
		return links, nil
	}
	if links, err := parseLoonContent(data); err == nil && len(links) > 0 {
		return links, nil
	}
	if links, err := parseBareProxyLines(data); err == nil && len(links) > 0 {
		return links, nil
	}
	return nil, fmt.Errorf("unsupported extra subscription format")
}

func parseLoonContent(data string) ([]string, error) {
	lines := strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n")
	inProxy := false
	links := make([]string, 0)
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
			inProxy = section == "proxy"
			continue
		}
		if !inProxy {
			continue
		}
		if link, ok := parseProxyLineToLink(line); ok {
			links = append(links, link)
		}
	}
	if len(links) == 0 {
		return nil, fmt.Errorf("no proxy found in loon content")
	}
	return links, nil
}

func parseBareProxyLines(data string) ([]string, error) {
	lines := strings.Split(strings.ReplaceAll(data, "\r\n", "\n"), "\n")
	links := make([]string, 0)
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "//") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			continue
		}
		if link, ok := parseProxyLineToLink(line); ok {
			links = append(links, link)
		}
	}
	if len(links) == 0 {
		return nil, fmt.Errorf("no proxy lines found")
	}
	return links, nil
}

func parseProxyLineToLink(line string) (string, bool) {
	if !strings.Contains(line, "=") {
		return "", false
	}
	parts := strings.SplitN(line, "=", 2)
	name := strings.TrimSpace(parts[0])
	rhs := strings.TrimSpace(parts[1])
	if name == "" || rhs == "" {
		return "", false
	}
	fields := splitCSV(rhs)
	if len(fields) < 3 {
		return "", false
	}
	proto := normalizeProto(fields[0])
	switch proto {
	case "trojan":
		link, err := buildTrojanLinkFromFields(name, fields)
		return link, err == nil
	case "vless":
		link, err := buildVlessLinkFromFields(name, fields)
		return link, err == nil
	case "ss", "shadowsocks":
		link, err := buildSSLinkFromFields(name, fields)
		return link, err == nil
	default:
		return "", false
	}
}

func splitCSV(s string) []string {
	out := make([]string, 0)
	var b strings.Builder
	var quote rune
	for _, r := range s {
		switch r {
		case '"', '\'':
			if quote == 0 {
				quote = r
			} else if quote == r {
				quote = 0
			}
			b.WriteRune(r)
		case ',':
			if quote == 0 {
				out = append(out, strings.TrimSpace(b.String()))
				b.Reset()
			} else {
				b.WriteRune(r)
			}
		default:
			b.WriteRune(r)
		}
	}
	out = append(out, strings.TrimSpace(b.String()))
	return out
}

func unquoteValue(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func normalizeProto(s string) string {
	s = strings.ToLower(strings.TrimSpace(unquoteValue(s)))
	switch s {
	case "shadowsocks":
		return "ss"
	default:
		return s
	}
}

func parseKVParts(fields []string, start int) map[string]string {
	attrs := map[string]string{}
	for i := start; i < len(fields); i++ {
		item := strings.TrimSpace(fields[i])
		if item == "" {
			continue
		}
		if strings.Contains(item, "=") {
			kv := strings.SplitN(item, "=", 2)
			attrs[strings.ToLower(strings.TrimSpace(kv[0]))] = unquoteValue(kv[1])
			continue
		}
		if strings.Contains(item, ":") {
			kv := strings.SplitN(item, ":", 2)
			attrs[strings.ToLower(strings.TrimSpace(kv[0]))] = unquoteValue(kv[1])
		}
	}
	return attrs
}

func isTruthy(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func buildTrojanLinkFromFields(name string, fields []string) (string, error) {
	if len(fields) < 4 {
		return "", fmt.Errorf("invalid trojan line")
	}
	server := strings.TrimSpace(unquoteValue(fields[1]))
	port, err := strconv.Atoi(strings.TrimSpace(unquoteValue(fields[2])))
	if err != nil {
		return "", err
	}
	password := strings.TrimSpace(unquoteValue(fields[3]))
	attrs := parseKVParts(fields, 4)

	u := &url.URL{Scheme: "trojan", Host: net.JoinHostPort(server, strconv.Itoa(port))}
	u.User = url.User(password)
	q := url.Values{}
	if isTruthy(attrs["skip-cert-verify"]) || isTruthy(attrs["allowinsecure"]) {
		q.Set("allowInsecure", "1")
	} else {
		q.Set("security", "tls")
	}
	if sni := firstNonEmpty(attrs["sni"], attrs["tls-name"], attrs["servername"]); sni != "" {
		q.Set("sni", sni)
	}
	switch strings.ToLower(strings.TrimSpace(firstNonEmpty(attrs["transport"], attrs["network"]))) {
	case "ws":
		q.Set("type", "ws")
		if p := firstNonEmpty(attrs["ws-path"], attrs["path"], attrs["obfs-uri"]); p != "" {
			q.Set("path", p)
		}
		if host := firstNonEmpty(attrs["host"], attrs["obfs-host"]); host != "" {
			q.Set("host", host)
		}
	case "grpc":
		q.Set("type", "grpc")
		if svc := firstNonEmpty(attrs["service-name"], attrs["grpc-service-name"], attrs["servicename"]); svc != "" {
			q.Set("serviceName", svc)
		}
	}
	u.RawQuery = q.Encode()
	u.Fragment = url.QueryEscape(name)
	return u.String(), nil
}

func buildVlessLinkFromFields(name string, fields []string) (string, error) {
	if len(fields) < 4 {
		return "", fmt.Errorf("invalid vless line")
	}
	server := strings.TrimSpace(unquoteValue(fields[1]))
	port, err := strconv.Atoi(strings.TrimSpace(unquoteValue(fields[2])))
	if err != nil {
		return "", err
	}
	uuid := strings.TrimSpace(unquoteValue(fields[3]))
	attrs := parseKVParts(fields, 4)

	opt := &config.VlessOption{
		Name:        name,
		Server:      server,
		Port:        port,
		UUID:        uuid,
		Flow:        attrs["flow"],
		SNI:         firstNonEmpty(attrs["sni"], attrs["tls-name"], attrs["servername"]),
		Host:        firstNonEmpty(attrs["host"], attrs["obfs-host"]),
		Path:        firstNonEmpty(attrs["ws-path"], attrs["path"], attrs["obfs-uri"]),
		Network:     strings.ToLower(strings.TrimSpace(firstNonEmpty(attrs["transport"], attrs["network"]))),
		PublicKey:   firstNonEmpty(attrs["public-key"], attrs["pbk"]),
		ShortID:     firstNonEmpty(attrs["short-id"], attrs["sid"]),
		Fingerprint: firstNonEmpty(attrs["fp"], attrs["fingerprint"], attrs["client-fingerprint"]),
	}
	if opt.Network == "" {
		opt.Network = "tcp"
	}
	if isTruthy(attrs["skip-cert-verify"]) || isTruthy(attrs["allowinsecure"]) {
		opt.Insecure = true
		opt.SkipCertVerify = true
	}
	if isTruthy(attrs["over-tls"]) || isTruthy(attrs["tls"]) {
		opt.TLS = true
	}
	switch {
	case opt.PublicKey != "":
		opt.Security = "reality"
		opt.TLS = true
	case opt.TLS:
		opt.Security = "tls"
	}
	return config.VlessOptionToLink(opt, "")
}

func buildSSLinkFromFields(name string, fields []string) (string, error) {
	if len(fields) < 5 {
		return "", fmt.Errorf("invalid ss line")
	}
	server := strings.TrimSpace(unquoteValue(fields[1]))
	port, err := strconv.Atoi(strings.TrimSpace(unquoteValue(fields[2])))
	if err != nil {
		return "", err
	}
	attrs := parseKVParts(fields, 3)
	method := firstNonEmpty(attrs["encrypt-method"], strings.TrimSpace(unquoteValue(fields[3])))
	password := firstNonEmpty(attrs["password"], strings.TrimSpace(unquoteValue(fields[4])))
	if method == "" || password == "" {
		return "", fmt.Errorf("invalid ss line")
	}
	userinfo := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", method, password)))
	link := fmt.Sprintf("ss://%s@%s", userinfo, net.JoinHostPort(server, strconv.Itoa(port)))
	if name != "" {
		link += "#" + url.QueryEscape(name)
	}
	return link, nil
}

func parseSingBoxJSONContent(data string) ([]string, error) {
	type sbConfig struct {
		Outbounds []map[string]any `json:"outbounds"`
	}
	cfg := sbConfig{}
	if err := json.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, err
	}
	if len(cfg.Outbounds) == 0 {
		return nil, fmt.Errorf("no outbounds")
	}
	links := make([]string, 0)
	for _, ob := range cfg.Outbounds {
		link, err := singBoxOutboundToLink(ob)
		if err == nil && link != "" {
			links = append(links, link)
		}
	}
	if len(links) == 0 {
		return nil, fmt.Errorf("no supported outbounds")
	}
	return links, nil
}

func singBoxOutboundToLink(ob map[string]any) (string, error) {
	t := strings.ToLower(strings.TrimSpace(asString(ob["type"])))
	switch t {
	case "trojan":
		return buildTrojanLinkFromSingBox(ob)
	case "vless":
		return buildVlessLinkFromSingBox(ob)
	case "shadowsocks":
		return buildSSLinkFromSingBox(ob)
	default:
		return "", fmt.Errorf("unsupported sing-box outbound type: %s", t)
	}
}

func buildTrojanLinkFromSingBox(ob map[string]any) (string, error) {
	server := asString(ob["server"])
	port := asInt(ob["server_port"])
	password := asString(ob["password"])
	if server == "" || port <= 0 || password == "" {
		return "", fmt.Errorf("invalid trojan outbound")
	}
	u := &url.URL{Scheme: "trojan", Host: net.JoinHostPort(server, strconv.Itoa(port))}
	u.User = url.User(password)
	q := url.Values{}
	tls := asMap(ob["tls"])
	if len(tls) > 0 {
		if isTruthy(asString(tls["insecure"])) {
			q.Set("allowInsecure", "1")
		} else if isTruthy(asString(tls["enabled"])) {
			q.Set("security", "tls")
		}
		if sni := asString(tls["server_name"]); sni != "" {
			q.Set("sni", sni)
		}
	}
	transport := asMap(ob["transport"])
	switch strings.ToLower(asString(transport["type"])) {
	case "ws":
		q.Set("type", "ws")
		if p := asString(transport["path"]); p != "" {
			q.Set("path", p)
		}
		if headers := asMap(transport["headers"]); len(headers) > 0 {
			if host := asString(headers["Host"]); host != "" {
				q.Set("host", host)
			}
			if host := asString(headers["host"]); host != "" {
				q.Set("host", host)
			}
		}
	case "grpc":
		q.Set("type", "grpc")
		if svc := asString(transport["service_name"]); svc != "" {
			q.Set("serviceName", svc)
		}
	}
	u.RawQuery = q.Encode()
	u.Fragment = url.QueryEscape(asString(ob["tag"]))
	return u.String(), nil
}

func buildVlessLinkFromSingBox(ob map[string]any) (string, error) {
	opt := &config.VlessOption{
		Name:    asString(ob["tag"]),
		Server:  asString(ob["server"]),
		Port:    asInt(ob["server_port"]),
		UUID:    asString(ob["uuid"]),
		Flow:    asString(ob["flow"]),
		Network: "tcp",
	}
	if opt.Server == "" || opt.Port <= 0 || opt.UUID == "" {
		return "", fmt.Errorf("invalid vless outbound")
	}
	tls := asMap(ob["tls"])
	if len(tls) > 0 {
		if isTruthy(asString(tls["enabled"])) {
			opt.TLS = true
		}
		opt.SNI = asString(tls["server_name"])
		if isTruthy(asString(tls["insecure"])) {
			opt.Insecure = true
			opt.SkipCertVerify = true
		}
		if utls := asMap(tls["utls"]); len(utls) > 0 {
			if isTruthy(asString(utls["enabled"])) {
				opt.Fingerprint = asString(utls["fingerprint"])
			}
		}
		if reality := asMap(tls["reality"]); len(reality) > 0 && isTruthy(asString(reality["enabled"])) {
			opt.Security = "reality"
			opt.PublicKey = asString(reality["public_key"])
			opt.ShortID = asString(reality["short_id"])
			opt.TLS = true
		}
	}
	if opt.Security == "" && opt.TLS {
		opt.Security = "tls"
	}
	transport := asMap(ob["transport"])
	if len(transport) > 0 {
		opt.Network = strings.ToLower(asString(transport["type"]))
		switch opt.Network {
		case "", "tcp":
			opt.Network = "tcp"
		case "ws":
			opt.Path = asString(transport["path"])
			if headers := asMap(transport["headers"]); len(headers) > 0 {
				opt.Host = firstNonEmpty(asString(headers["Host"]), asString(headers["host"]))
			}
		case "grpc":
			opt.ServiceName = asString(transport["service_name"])
		case "httpupgrade", "http-upgrade":
			opt.Network = "httpupgrade"
			opt.Host = asString(transport["host"])
			opt.Path = asString(transport["path"])
		case "http":
			opt.Host = asString(transport["host"])
			if opt.Host == "" {
				if hosts, ok := transport["host"].([]any); ok && len(hosts) > 0 {
					opt.Host = asString(hosts[0])
				}
			}
			opt.Path = asString(transport["path"])
		}
	}
	return config.VlessOptionToLink(opt, "")
}

func buildSSLinkFromSingBox(ob map[string]any) (string, error) {
	server := asString(ob["server"])
	port := asInt(ob["server_port"])
	method := asString(ob["method"])
	password := asString(ob["password"])
	if server == "" || port <= 0 || method == "" || password == "" {
		return "", fmt.Errorf("invalid shadowsocks outbound")
	}
	userinfo := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", method, password)))
	link := fmt.Sprintf("ss://%s@%s", userinfo, net.JoinHostPort(server, strconv.Itoa(port)))
	if tag := asString(ob["tag"]); tag != "" {
		link += "#" + url.QueryEscape(tag)
	}
	return link, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func asString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	case json.Number:
		return x.String()
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatInt(int64(x), 10)
	case int:
		return strconv.Itoa(x)
	case int64:
		return strconv.FormatInt(x, 10)
	default:
		return ""
	}
}

func asInt(v any) int {
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case json.Number:
		i, _ := x.Int64()
		return int(i)
	case string:
		i, _ := strconv.Atoi(strings.TrimSpace(x))
		return i
	default:
		return 0
	}
}

func asMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}
