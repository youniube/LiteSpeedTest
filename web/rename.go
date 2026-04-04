package web

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xxf098/lite-proxy/config"
)

type renameNode struct {
	ID       int    `json:"id"`
	Remark   string `json:"remark"`
	Server   string `json:"server"`
	Protocol string `json:"protocol"`
	Link     string `json:"link"`
}

type renameRequest struct {
	Nodes       []renameNode `json:"nodes"`
	UseExternal bool         `json:"useExternal"`
	IntervalMs  int          `json:"intervalMs"`
}

type renameResponse struct {
	Nodes []renameNode `json:"nodes"`
}

type renameGeoCacheEntry struct {
	Label     string    `json:"label"`
	ExpiresAt time.Time `json:"expiresAt"`
}

var (
	renameCacheMu       sync.Mutex
	renameLastExternal  time.Time
	renameGeoCache      map[string]renameGeoCacheEntry
	renameGeoCacheOnce  sync.Once
	renameSeparatorExpr = regexp.MustCompile(`[|_/]+`)
	renameSpaceExpr     = regexp.MustCompile(`\s+`)
	renameProfileExpr   = regexp.MustCompile(`(?i)^profile\s*\d+$`)
	renameNoiseExpr     = regexp.MustCompile(`(?i)^(default|node|server|unnamed|untitled|unknown|test)$`)
)

func renameNodesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()

	req := renameRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Nodes) == 0 {
		_ = json.NewEncoder(w).Encode(renameResponse{Nodes: []renameNode{}})
		return
	}

	interval := time.Duration(req.IntervalMs) * time.Millisecond
	if interval < 300*time.Millisecond {
		interval = 1200 * time.Millisecond
	}

	nodes, err := smartRenameNodes(r.Context(), req.Nodes, req.UseExternal, interval)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(renameResponse{Nodes: nodes})
}

func smartRenameNodes(ctx context.Context, nodes []renameNode, useExternal bool, interval time.Duration) ([]renameNode, error) {
	outputs := make([]renameNode, len(nodes))
	counts := map[string]int{}
	totals := map[string]int{}
	bases := make([]string, len(nodes))

	for i, node := range nodes {
		base := buildRenameBase(ctx, node, useExternal, interval)
		if base == "" {
			base = fmt.Sprintf("Node %d", node.ID+1)
		}
		bases[i] = base
		totals[base]++
	}

	for i, node := range nodes {
		base := bases[i]
		counts[base]++
		node.Remark = base
		if totals[base] > 1 {
			node.Remark = fmt.Sprintf("%s %02d", base, counts[base])
		}
		outputs[i] = node
	}
	return outputs, nil
}

func buildRenameBase(ctx context.Context, node renameNode, useExternal bool, interval time.Duration) string {
	remark := strings.TrimSpace(node.Remark)
	server := strings.TrimSpace(node.Server)
	protocol := normalizeProtocolLabel(node.Protocol)

	if cfg, err := config.Link2Config(node.Link); err == nil {
		if strings.TrimSpace(cfg.Remarks) != "" {
			remark = cfg.Remarks
		}
		if strings.TrimSpace(cfg.Server) != "" {
			server = cfg.Server
		}
		if strings.TrimSpace(cfg.Protocol) != "" {
			protocol = normalizeProtocolLabel(cfg.Protocol)
			if cfg.Net != "" && !strings.Contains(protocol, "/") {
				protocol = fmt.Sprintf("%s/%s", protocol, strings.ToLower(strings.TrimSpace(cfg.Net)))
			}
		}
	}

	cleaned := normalizeRemark(remark)
	if isUsefulRemark(cleaned) {
		return cleaned
	}

	var geoLabel string
	if useExternal {
		geoLabel = lookupGeoLabel(ctx, server, interval)
	}
	if geoLabel == "" {
		geoLabel = normalizeHostLabel(server)
	}
	if geoLabel == "" {
		geoLabel = firstNonEmpty(protocol, "Node")
	}
	return geoLabel
}

func normalizeProtocolLabel(protocol string) string {
	protocol = strings.TrimSpace(strings.ToLower(protocol))
	if protocol == "" {
		return ""
	}
	return protocol
}

func normalizeRemark(remark string) string {
	remark = strings.TrimSpace(strings.ToValidUTF8(remark, ""))
	if remark == "" {
		return ""
	}
	remark = html.UnescapeString(remark)
	for i := 0; i < 2; i++ {
		if decoded, err := url.QueryUnescape(remark); err == nil && decoded != "" {
			remark = decoded
		}
	}
	remark = strings.ReplaceAll(remark, "+", " ")
	remark = strings.ReplaceAll(remark, "\t", " ")
	remark = strings.ReplaceAll(remark, "\n", " ")
	remark = strings.ReplaceAll(remark, "\r", " ")
	remark = renameSeparatorExpr.ReplaceAllString(remark, " · ")
	remark = strings.ReplaceAll(remark, "- ", "-")
	remark = strings.ReplaceAll(remark, " -", "-")
	remark = renameSpaceExpr.ReplaceAllString(remark, " ")
	remark = strings.Trim(remark, " .·|-_[]{}()<>")
	return remark
}

func isUsefulRemark(remark string) bool {
	remark = strings.TrimSpace(remark)
	if remark == "" {
		return false
	}
	if renameProfileExpr.MatchString(remark) || renameNoiseExpr.MatchString(remark) {
		return false
	}

	useful := 0
	noise := 0
	for _, r := range remark {
		switch {
		case r >= '0' && r <= '9':
			useful++
		case r >= 'a' && r <= 'z':
			useful++
		case r >= 'A' && r <= 'Z':
			useful++
		case r >= 0x4e00 && r <= 0x9fff:
			useful++
		case strings.ContainsRune(" ·-_.:%", r):
		default:
			noise++
		}
	}
	if useful < 2 {
		return false
	}
	if noise > useful*2 {
		return false
	}
	return true
}

func normalizeHostLabel(server string) string {
	host := extractHost(server)
	if host == "" {
		return ""
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	host = strings.TrimPrefix(host, "www.")
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	parts := strings.Split(host, ".")
	if len(parts) >= 3 {
		parts = parts[len(parts)-3:]
	}
	return strings.Join(parts, ".")
}

func extractHost(server string) string {
	server = strings.TrimSpace(server)
	if server == "" {
		return ""
	}
	if strings.HasPrefix(server, "[") {
		if host, _, err := net.SplitHostPort(server); err == nil {
			return strings.Trim(host, "[]")
		}
	}
	if host, _, err := net.SplitHostPort(server); err == nil {
		return host
	}
	if ip := net.ParseIP(server); ip != nil {
		return ip.String()
	}
	if idx := strings.LastIndex(server, ":"); idx > 0 && !strings.Contains(server[idx+1:], ":") {
		if _, err := strconv.Atoi(server[idx+1:]); err == nil {
			return server[:idx]
		}
	}
	return server
}

func lookupGeoLabel(ctx context.Context, server string, interval time.Duration) string {
	host := extractHost(server)
	if host == "" {
		return ""
	}

	ip := host
	if parsed := net.ParseIP(host); parsed == nil {
		resolved := resolvePublicIP(host)
		if resolved == "" {
			return ""
		}
		ip = resolved
	}

	renameGeoCacheOnce.Do(loadRenameGeoCache)

	renameCacheMu.Lock()
	if entry, ok := renameGeoCache[ip]; ok && time.Now().Before(entry.ExpiresAt) {
		renameCacheMu.Unlock()
		return entry.Label
	}
	wait := interval - time.Since(renameLastExternal)
	if wait > 0 {
		renameCacheMu.Unlock()
		select {
		case <-ctx.Done():
			return ""
		case <-time.After(wait):
		}
		renameCacheMu.Lock()
	}
	renameLastExternal = time.Now()
	renameCacheMu.Unlock()

	lookupCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(lookupCtx, http.MethodGet, fmt.Sprintf("https://ipwho.is/%s", url.PathEscape(ip)), nil)
	if err != nil {
		return ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}

	var data struct {
		Success     bool   `json:"success"`
		CountryCode string `json:"country_code"`
		Country     string `json:"country"`
		Region      string `json:"region"`
		City        string `json:"city"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return ""
	}
	if !data.Success {
		return ""
	}

	parts := []string{}
	if data.CountryCode != "" {
		parts = append(parts, strings.ToUpper(strings.TrimSpace(data.CountryCode)))
	} else if data.Country != "" {
		parts = append(parts, strings.TrimSpace(data.Country))
	}
	if data.City != "" {
		parts = append(parts, strings.TrimSpace(data.City))
	} else if data.Region != "" {
		parts = append(parts, strings.TrimSpace(data.Region))
	}
	label := strings.TrimSpace(strings.Join(parts, "-"))
	if label == "" {
		return ""
	}

	renameCacheMu.Lock()
	renameGeoCache[ip] = renameGeoCacheEntry{Label: label, ExpiresAt: time.Now().Add(12 * time.Hour)}
	renameCacheMu.Unlock()
	saveRenameGeoCache()
	return label
}

func resolvePublicIP(host string) string {
	ips, err := net.LookupIP(host)
	if err != nil {
		return ""
	}
	sort.SliceStable(ips, func(i, j int) bool {
		return len(ips[i]) < len(ips[j])
	})
	for _, ip := range ips {
		if ip == nil || ip.IsLoopback() || ip.IsUnspecified() || isPrivateIP(ip) {
			continue
		}
		return ip.String()
	}
	for _, ip := range ips {
		if ip != nil {
			return ip.String()
		}
	}
	return ""
}

func renameCacheFilePath() string {
	dir := filepath.Join(".lite-singbox", "cache")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "rename_geo_cache.json")
}

func loadRenameGeoCache() {
	renameGeoCache = map[string]renameGeoCacheEntry{}
	data, err := os.ReadFile(renameCacheFilePath())
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &renameGeoCache)
}

func saveRenameGeoCache() {
	renameCacheMu.Lock()
	snapshot := make(map[string]renameGeoCacheEntry, len(renameGeoCache))
	for k, v := range renameGeoCache {
		snapshot[k] = v
	}
	renameCacheMu.Unlock()

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(renameCacheFilePath(), data, 0o644)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
