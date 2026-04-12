package subscription

import (
	"strings"

	"github.com/xxf098/lite-proxy/utils"
)

func ParseBase64Text(content string) ([]ProxyNode, error) {
	decoded, err := DecodeBase64Text(content)
	if err != nil {
		return nil, err
	}
	format := DetectTextFormat(decoded)
	if format == FormatUnknown || format == FormatBase64 {
		format = FormatURI
	}
	return parseTextByFormat(decoded, format)
}

func DecodeBase64Text(content string) (string, error) {
	return utils.DecodeB64(strings.TrimSpace(content))
}
