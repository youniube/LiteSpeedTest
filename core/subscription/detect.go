package subscription

import (
	"regexp"
	"strings"

	"github.com/xxf098/lite-proxy/utils"
)

var (
	regHasURI = regexp.MustCompile(`(?i)(?:^|\s)(?:vmess|vless|trojan|ssr|ss|http)://`)
	regQXLine = regexp.MustCompile(`(?im)^\s*(shadowsocks|vmess|trojan|http|socks5)\s*=`)
)

func DetectTextFormat(content string) InputFormat {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return FormatUnknown
	}

	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, "[server_local]") || strings.Contains(lower, "[server_remote]") || regQXLine.MatchString(trimmed) {
		return FormatQX
	}
	if strings.HasPrefix(trimmed, "{") && strings.Contains(lower, "\"outbounds\"") {
		return FormatSingBox
	}
	if strings.Contains(lower, "[remote proxy]") || strings.Contains(lower, "[remote filter]") {
		return FormatLoon
	}
	if strings.Contains(lower, "[proxy]") {
		if strings.Contains(lower, "transport:") || strings.Contains(lower, "over-tls:") || strings.Contains(lower, "tls-name:") || strings.Contains(lower, "shadowsocks,") || strings.Contains(lower, "shadowsocksr,") {
			return FormatLoon
		}
	}
	if strings.Contains(lower, "proxies:") || strings.Contains(lower, "proxy-groups:") || strings.Contains(lower, "mixed-port:") || strings.Contains(lower, "allow-lan:") {
		return FormatClash
	}
	if strings.Contains(lower, "[proxy]") || strings.Contains(lower, "[proxy group]") || strings.Contains(lower, "[general]") || strings.Contains(lower, "[rule]") {
		return FormatSurge
	}
	if HasURIScheme(trimmed) {
		return FormatURI
	}
	if IsLikelyBase64Text(trimmed) {
		return FormatBase64
	}
	return FormatUnknown
}

func DetectURLHint(input string) InputFormat {
	lower := strings.ToLower(strings.TrimSpace(input))
	switch {
	case strings.Contains(lower, ".yaml") || strings.Contains(lower, ".yml"):
		return FormatClash
	case strings.Contains(lower, ".json"):
		return FormatSingBox
	case strings.Contains(lower, ".conf"):
		return FormatSurge
	default:
		return FormatUnknown
	}
}

func HasURIScheme(content string) bool {
	return regHasURI.MatchString(content)
}

func IsLikelyBase64Text(content string) bool {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" || strings.Contains(trimmed, " ") || strings.Contains(trimmed, "\t") {
		return false
	}
	decoded, err := utils.DecodeB64(trimmed)
	if err != nil {
		return false
	}
	decoded = strings.TrimSpace(decoded)
	if decoded == "" {
		return false
	}
	format := DetectTextFormat(decoded)
	if format != FormatUnknown && format != FormatBase64 {
		return true
	}
	return strings.Contains(decoded, "\n")
}
