package web

import (
	"context"
	"encoding/json"
	"fmt"
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
	Info      renameLocation `json:"info"`
	ExpiresAt time.Time      `json:"expiresAt"`
}

type renameDNSCacheEntry struct {
	IP        string    `json:"ip"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type renameLookupState struct {
	dnsBatch map[string]string
}

type renameLocation struct {
	Code string `json:"code"`
	Zh   string `json:"zh"`
}

type renameAlias struct {
	Code     string
	Zh       string
	Keywords []string
}

var (
	renameCacheMu      sync.Mutex
	renameLastExternal time.Time
	renameGeoCache     map[string]renameGeoCacheEntry
	renameDNSCache     map[string]renameDNSCacheEntry
	renameGeoCacheOnce sync.Once

	renameSplitExpr  = regexp.MustCompile(`[\s|/_\\\-]+`)
	renameCleanExpr  = regexp.MustCompile(`[^a-z0-9\p{Han}]+`)
	renameSpaceExpr  = regexp.MustCompile(`\s+`)
	renameIPv4Expr   = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	renamePortExpr   = regexp.MustCompile(`:\d+$`)
	renameNumberExpr = regexp.MustCompile(`\d+`)

	renameAliases = []renameAlias{
		{Code: "HK", Zh: "香港", Keywords: []string{"香港", "hongkong", "hong-kong", "hong_kong", "hk"}},
		{Code: "TW", Zh: "台湾", Keywords: []string{"台湾", "taiwan", "taipei", "tw"}},
		{Code: "MO", Zh: "澳门", Keywords: []string{"澳门", "macao", "macau", "mo"}},
		{Code: "SG", Zh: "新加坡", Keywords: []string{"新加坡", "singapore", "sg"}},
		{Code: "JP", Zh: "日本", Keywords: []string{"日本", "东京", "大阪", "japan", "tokyo", "osaka", "jp"}},
		{Code: "KR", Zh: "韩国", Keywords: []string{"韩国", "首尔", "korea", "seoul", "kr"}},
		{Code: "US", Zh: "美国", Keywords: []string{"美国", "洛杉矶", "圣何塞", "纽约", "usa", "unitedstates", "united-states", "america", "losangeles", "sanjose", "newyork", "us"}},
		{Code: "CA", Zh: "加拿大", Keywords: []string{"加拿大", "多伦多", "温哥华", "canada", "toronto", "vancouver", "ca"}},
		{Code: "GB", Zh: "英国", Keywords: []string{"英国", "伦敦", "uk", "unitedkingdom", "united-kingdom", "england", "london", "gb"}},
		{Code: "DE", Zh: "德国", Keywords: []string{"德国", "法兰克福", "germany", "frankfurt", "de"}},
		{Code: "FR", Zh: "法国", Keywords: []string{"法国", "巴黎", "france", "paris", "fr"}},
		{Code: "NL", Zh: "荷兰", Keywords: []string{"荷兰", "阿姆斯特丹", "netherlands", "amsterdam", "nl"}},
		{Code: "AU", Zh: "澳大利亚", Keywords: []string{"澳大利亚", "悉尼", "墨尔本", "australia", "sydney", "melbourne", "au"}},
		{Code: "MY", Zh: "马来西亚", Keywords: []string{"马来西亚", "吉隆坡", "malaysia", "kualalumpur", "my"}},
		{Code: "TH", Zh: "泰国", Keywords: []string{"泰国", "曼谷", "thailand", "bangkok", "th"}},
		{Code: "VN", Zh: "越南", Keywords: []string{"越南", "河内", "胡志明", "vietnam", "hanoi", "hochiminh", "vn"}},
		{Code: "PH", Zh: "菲律宾", Keywords: []string{"菲律宾", "马尼拉", "philippines", "manila", "ph"}},
		{Code: "IN", Zh: "印度", Keywords: []string{"印度", "孟买", "德里", "india", "mumbai", "delhi", "in"}},
		{Code: "RU", Zh: "俄罗斯", Keywords: []string{"俄罗斯", "莫斯科", "russia", "moscow", "ru"}},
		{Code: "TR", Zh: "土耳其", Keywords: []string{"土耳其", "turkey", "istanbul", "tr"}},
		{Code: "BR", Zh: "巴西", Keywords: []string{"巴西", "brazil", "saopaulo", "br"}},
		{Code: "AR", Zh: "阿根廷", Keywords: []string{"阿根廷", "argentina", "ar"}},
		{Code: "AE", Zh: "阿联酋", Keywords: []string{"阿联酋", "迪拜", "uae", "dubai", "emirates", "ae"}},
		{Code: "ID", Zh: "印度尼西亚", Keywords: []string{"印度尼西亚", "雅加达", "indonesia", "jakarta", "id"}},
		{Code: "CN", Zh: "中国", Keywords: []string{"中国", "北京", "上海", "广州", "深圳", "china", "beijing", "shanghai", "guangzhou", "shenzhen", "cn"}},
	}

	renameCodeMap = map[string]string{}
)

func init() {
	for _, alias := range renameAliases {
		renameCodeMap[strings.ToUpper(alias.Code)] = alias.Zh
	}
}

func renameNodesHandler(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	defer r.Body.Close()

	req := renameRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(req.Nodes) == 0 {
		writeJSON(w, http.StatusOK, renameResponse{Nodes: []renameNode{}})
		return
	}

	interval := time.Duration(req.IntervalMs) * time.Millisecond
	switch {
	case req.IntervalMs <= 0:
		interval = 2000 * time.Millisecond
	case interval < 1200*time.Millisecond:
		interval = 1200 * time.Millisecond
	}

	nodes, err := smartRenameNodes(r.Context(), req.Nodes, req.UseExternal, interval)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, renameResponse{Nodes: nodes})
}

func smartRenameNodes(ctx context.Context, nodes []renameNode, useExternal bool, interval time.Duration) ([]renameNode, error) {
	outputs := make([]renameNode, len(nodes))
	totals := map[string]int{}
	counts := map[string]int{}
	bases := make([]string, len(nodes))
	lookupState := &renameLookupState{dnsBatch: map[string]string{}}

	for i, node := range nodes {
		base := buildRenameBase(ctx, node, useExternal, interval, lookupState)
		if base == "" {
			base = "🌐UN-未知"
		}
		bases[i] = base
		totals[base]++
	}

	for i, node := range nodes {
		base := bases[i]
		counts[base]++
		node.Remark = fmt.Sprintf("%s %02d", base, counts[base])
		outputs[i] = node
	}
	return outputs, nil
}

func buildRenameBase(ctx context.Context, node renameNode, useExternal bool, interval time.Duration, lookupState *renameLookupState) string {
	remark := strings.TrimSpace(node.Remark)
	server := strings.TrimSpace(node.Server)
	link := strings.TrimSpace(node.Link)

	if cfg, err := config.Link2Config(link); err == nil {
		if strings.TrimSpace(cfg.Remarks) != "" {
			remark = cfg.Remarks
		}
		if strings.TrimSpace(cfg.Server) != "" {
			server = cfg.Server
		}
	}

	if loc := inferLocationFromRemark(remark); loc.Code != "" {
		return formatRenameBase(loc)
	}
	if loc := inferLocationFromServer(server); loc.Code != "" {
		return formatRenameBase(loc)
	}
	if useExternal {
		if loc := lookupGeoLocation(ctx, server, interval, lookupState); loc.Code != "" {
			return formatRenameBase(loc)
		}
	}
	if fallback := fallbackRemarkBase(remark); fallback != "" {
		return fallback
	}
	return formatRenameBase(renameLocation{Code: "UN", Zh: "未知"})
}

func fallbackRemarkBase(remark string) string {
	remark = strings.TrimSpace(strings.ToValidUTF8(remark, ""))
	if remark == "" {
		return ""
	}
	compact := strings.ToLower(strings.ReplaceAll(remark, " ", ""))
	blocked := []string{
		"剩余流量", "流量", "到期", "订阅", "subscription", "traffic", "expire",
		"used", "total", "balance", "套餐", "官网", "客服", "最新网址",
	}
	for _, item := range blocked {
		if strings.Contains(compact, item) {
			return ""
		}
	}
	return remark
}

func inferLocationFromRemark(remark string) renameLocation {
	if strings.TrimSpace(remark) == "" {
		return renameLocation{}
	}
	normalized := normalizeMatchText(remark)
	if normalized == "" {
		return renameLocation{}
	}

	tokens := renameSplitExpr.Split(normalized, -1)
	for i, token := range tokens {
		if i >= 4 {
			break
		}
		if loc := lookupAliasToken(token, true); loc.Code != "" {
			return loc
		}
	}
	for _, token := range tokens {
		if loc := lookupAliasToken(token, false); loc.Code != "" {
			return loc
		}
	}
	compact := strings.ReplaceAll(normalized, " ", "")
	for _, alias := range renameAliases {
		for _, kw := range alias.Keywords {
			compactKW := strings.ReplaceAll(strings.ToLower(kw), " ", "")
			if compactKW == "" {
				continue
			}
			if strings.Contains(compact, compactKW) {
				return renameLocation{Code: alias.Code, Zh: alias.Zh}
			}
		}
	}
	return renameLocation{}
}

func lookupAliasToken(token string, prefixOnly bool) renameLocation {
	token = normalizeMatchText(token)
	token = strings.ReplaceAll(token, " ", "")
	token = renameNumberExpr.ReplaceAllString(token, "")
	if token == "" {
		return renameLocation{}
	}
	for _, alias := range renameAliases {
		for _, kw := range alias.Keywords {
			candidate := strings.ReplaceAll(strings.ToLower(kw), " ", "")
			if candidate == "" {
				continue
			}
			if prefixOnly {
				if token == candidate || strings.HasPrefix(token, candidate) {
					return renameLocation{Code: alias.Code, Zh: alias.Zh}
				}
				continue
			}
			if token == candidate {
				return renameLocation{Code: alias.Code, Zh: alias.Zh}
			}
		}
	}
	return renameLocation{}
}

func inferLocationFromServer(server string) renameLocation {
	host := strings.ToLower(strings.TrimSpace(extractHost(server)))
	if host == "" {
		return renameLocation{}
	}
	if strings.HasSuffix(host, ".hk") {
		return renameLocation{Code: "HK", Zh: "香港"}
	}
	if strings.HasSuffix(host, ".tw") {
		return renameLocation{Code: "TW", Zh: "台湾"}
	}
	if strings.HasSuffix(host, ".jp") {
		return renameLocation{Code: "JP", Zh: "日本"}
	}
	if strings.HasSuffix(host, ".sg") {
		return renameLocation{Code: "SG", Zh: "新加坡"}
	}
	if strings.HasSuffix(host, ".kr") {
		return renameLocation{Code: "KR", Zh: "韩国"}
	}
	if strings.HasSuffix(host, ".us") {
		return renameLocation{Code: "US", Zh: "美国"}
	}
	if strings.HasSuffix(host, ".uk") || strings.HasSuffix(host, ".gb") {
		return renameLocation{Code: "GB", Zh: "英国"}
	}

	tokens := renameSplitExpr.Split(normalizeMatchText(host), -1)
	for _, token := range tokens {
		if loc := lookupAliasToken(token, false); loc.Code != "" {
			return loc
		}
	}
	return renameLocation{}
}

func lookupGeoLocation(ctx context.Context, server string, interval time.Duration, lookupState *renameLookupState) renameLocation {
	host := extractHost(server)
	if host == "" {
		return renameLocation{}
	}

	ip := host
	if parsed := net.ParseIP(host); parsed == nil {
		resolved := resolvePublicIP(host, lookupState)
		if resolved == "" {
			return renameLocation{}
		}
		ip = resolved
	}

	renameGeoCacheOnce.Do(loadRenameGeoCache)

	renameCacheMu.Lock()
	if entry, ok := renameGeoCache[ip]; ok && time.Now().Before(entry.ExpiresAt) {
		renameCacheMu.Unlock()
		return entry.Info
	}
	wait := interval - time.Since(renameLastExternal)
	if wait > 0 {
		renameCacheMu.Unlock()
		select {
		case <-ctx.Done():
			return renameLocation{}
		case <-time.After(wait):
		}
		renameCacheMu.Lock()
	}
	renameLastExternal = time.Now()
	renameCacheMu.Unlock()

	info, ok := queryGeoLocationWithRetry(ctx, ip)
	if !ok {
		return renameLocation{}
	}

	renameCacheMu.Lock()
	renameGeoCache[ip] = renameGeoCacheEntry{Info: info, ExpiresAt: time.Now().Add(12 * time.Hour)}
	renameCacheMu.Unlock()
	saveRenameGeoCache()
	return info
}

func queryGeoLocationWithRetry(ctx context.Context, ip string) (renameLocation, bool) {
	info, ok := queryGeoLocationOnce(ctx, ip)
	if ok {
		return info, true
	}

	retryDelay := 2 * time.Second
	select {
	case <-ctx.Done():
		return renameLocation{}, false
	case <-time.After(retryDelay):
	}

	return queryGeoLocationOnce(ctx, ip)
}

func queryGeoLocationOnce(ctx context.Context, ip string) (renameLocation, bool) {
	lookupCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(lookupCtx, http.MethodGet, fmt.Sprintf("https://ipwho.is/%s", url.PathEscape(ip)), nil)
	if err != nil {
		return renameLocation{}, false
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return renameLocation{}, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return renameLocation{}, false
	}

	var data struct {
		Success     bool   `json:"success"`
		CountryCode string `json:"country_code"`
		Country     string `json:"country"`
		Region      string `json:"region"`
		City        string `json:"city"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return renameLocation{}, false
	}
	if !data.Success {
		return renameLocation{}, false
	}

	code := strings.ToUpper(strings.TrimSpace(data.CountryCode))
	zh := renameCodeMap[code]
	if zh == "" {
		zh = strings.TrimSpace(data.Country)
	}
	info := renameLocation{Code: code, Zh: zh}
	if info.Code == "" || info.Zh == "" {
		return renameLocation{}, false
	}
	return info, true
}

func formatRenameBase(loc renameLocation) string {
	code := strings.ToUpper(strings.TrimSpace(loc.Code))
	zh := strings.TrimSpace(loc.Zh)
	if code == "" {
		code = "UN"
	}
	if zh == "" {
		zh = renameCodeMap[code]
	}
	if zh == "" {
		zh = "未知"
	}
	return fmt.Sprintf("%s%s-%s", flagEmoji(code), code, zh)
}

func flagEmoji(code string) string {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 2 || code == "UN" {
		return "🌐"
	}
	r1 := rune(code[0])
	r2 := rune(code[1])
	if r1 < 'A' || r1 > 'Z' || r2 < 'A' || r2 > 'Z' {
		return "🌐"
	}
	return string([]rune{0x1F1E6 + (r1 - 'A'), 0x1F1E6 + (r2 - 'A')})
}

func normalizeMatchText(value string) string {
	value = strings.TrimSpace(strings.ToValidUTF8(value, ""))
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "+", " ")
	value = strings.ReplaceAll(value, "_", " ")
	value = strings.ReplaceAll(value, "-", " ")
	value = renamePortExpr.ReplaceAllString(value, "")
	value = renameCleanExpr.ReplaceAllString(value, " ")
	value = renameSpaceExpr.ReplaceAllString(value, " ")
	return strings.TrimSpace(value)
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
	if renameIPv4Expr.MatchString(server) {
		return server
	}
	if idx := strings.LastIndex(server, ":"); idx > 0 && !strings.Contains(server[idx+1:], ":") {
		if _, err := strconv.Atoi(server[idx+1:]); err == nil {
			return server[:idx]
		}
	}
	return server
}

func resolvePublicIP(host string, lookupState *renameLookupState) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}

	if lookupState != nil {
		if ip, ok := lookupState.dnsBatch[host]; ok {
			return ip
		}
	}

	renameGeoCacheOnce.Do(loadRenameGeoCache)

	now := time.Now()
	renameCacheMu.Lock()
	if entry, ok := renameDNSCache[host]; ok {
		if now.Before(entry.ExpiresAt) {
			ip := entry.IP
			renameCacheMu.Unlock()
			if lookupState != nil && ip != "" {
				lookupState.dnsBatch[host] = ip
			}
			return ip
		}
		delete(renameDNSCache, host)
	}
	renameCacheMu.Unlock()

	ips, err := net.LookupIP(host)
	if err != nil {
		return ""
	}
	sort.SliceStable(ips, func(i, j int) bool {
		return len(ips[i]) < len(ips[j])
	})

	resolved := ""
	for _, ip := range ips {
		if ip == nil || ip.IsLoopback() || ip.IsUnspecified() || isPrivateIP(ip) {
			continue
		}
		resolved = ip.String()
		break
	}
	if resolved == "" {
		for _, ip := range ips {
			if ip != nil {
				resolved = ip.String()
				break
			}
		}
	}

	if lookupState != nil && resolved != "" {
		lookupState.dnsBatch[host] = resolved
	}
	if resolved == "" {
		return ""
	}

	renameCacheMu.Lock()
	renameDNSCache[host] = renameDNSCacheEntry{IP: resolved, ExpiresAt: time.Now().Add(24 * time.Hour)}
	renameCacheMu.Unlock()
	return resolved
}

func renameCacheFilePath() string {
	dir := filepath.Join(".lite-singbox", "cache")
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, "rename_geo_cache.json")
}

func loadRenameGeoCache() {
	renameGeoCache = map[string]renameGeoCacheEntry{}
	renameDNSCache = map[string]renameDNSCacheEntry{}
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
