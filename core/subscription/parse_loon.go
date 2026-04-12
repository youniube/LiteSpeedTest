package subscription

import (
	"fmt"
	"strings"
)

func ParseLoonText(content string) ([]ProxyNode, error) {
	sections := SplitSections(content)
	lines := extractProxyLikeLines(content, sections["proxy"])
	if len(lines) == 0 {
		return nil, ErrNoProxyFound
	}
	nodes := make([]ProxyNode, 0, len(lines))
	for _, line := range lines {
		node, err := ParseLoonProxyLine(line)
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

func ParseLoonProxyLine(line string) (ProxyNode, error) {
	node, err := ParseSurgeProxyLine(line)
	if err != nil {
		return ProxyNode{}, err
	}
	node.SourceFormat = FormatLoon
	node.Raw = map[string]any{"line": line}
	kv := strings.SplitN(line, "=", 2)
	if len(kv) != 2 {
		return node, nil
	}
	parts := splitCSV(kv[1])
	if len(parts) < 3 {
		return ProxyNode{}, fmt.Errorf("invalid loon proxy line")
	}
	attrs := parseKeyValueTokens(parts[3:])
	if node.Type == "vmess" {
		if strings.EqualFold(attrs["transport"], "ws") {
			node.Network = "ws"
			node.Path = firstNotEmpty(node.Path, attrs["ws-path"], attrs["path"], "/")
			node.Host = firstNotEmpty(node.Host, attrs["host"])
		}
	}
	return node, nil
}
