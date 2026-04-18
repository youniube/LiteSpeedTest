package web

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/xxf098/lite-proxy/config"
	"github.com/xxf098/lite-proxy/utils"
)

type GetSubscriptionLink struct {
	FilePath string   `json:"filePath"`
	Group    string   `json:"group"`
	Links    []string `json:"links,omitempty"`
}

type getSubscriptionLinkResponse struct {
	Link string `json:"link"`
}

type subscriptionEntry struct {
	Group string
	Links []string
}

var (
	subscriptionLinkMu  sync.RWMutex
	subscriptionLinkMap = map[string]subscriptionEntry{}
)

func saveSubscriptionEntry(key string, entry subscriptionEntry) {
	subscriptionLinkMu.Lock()
	defer subscriptionLinkMu.Unlock()
	subscriptionLinkMap[key] = entry
}

func getSubscriptionEntry(key string) (subscriptionEntry, bool) {
	subscriptionLinkMu.RLock()
	defer subscriptionLinkMu.RUnlock()
	entry, ok := subscriptionLinkMap[key]
	return entry, ok
}

func buildSubscriptionLink(r *http.Request, key string, group string) string {
	scheme := forwardedScheme(r)
	host := strings.TrimSpace(r.Host)
	if host == "" {
		if ipAddr, err := localIP(); err == nil {
			host = netJoinHostPortIfNeeded(ipAddr.String(), "10888")
		} else {
			host = "127.0.0.1:10888"
		}
	}
	return fmt.Sprintf("%s://%s/getSubscription?key=%s&group=%s", scheme, host, url.QueryEscape(key), url.QueryEscape(group))
}

func forwardedScheme(r *http.Request) string {
	if proto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); proto != "" {
		return proto
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func netJoinHostPortIfNeeded(host string, port string) string {
	if strings.Contains(host, ":") {
		return host
	}
	return host + ":" + port
}

func getSubscriptionLink(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}
	defer r.Body.Close()

	body := GetSubscriptionLink{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	group := strings.TrimSpace(body.Group)
	if group == "" {
		writeAPIError(w, http.StatusBadRequest, "Invalid Parameter")
		return
	}

	links := sanitizeLinks(body.Links)
	filePath := strings.TrimSpace(body.FilePath)
	if len(links) == 0 {
		if filePath == "" {
			writeAPIError(w, http.StatusBadRequest, "Invalid Parameter")
			return
		}
		if !utils.IsUrl(filePath) {
			writeAPIError(w, http.StatusBadRequest, "local file path is not allowed")
			return
		}
		var err error
		links, err = getSubscriptionLinks(filePath)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, err.Error())
			return
		}
		links = sanitizeLinks(links)
	}
	if len(links) == 0 {
		writeAPIError(w, http.StatusBadRequest, "No links found")
		return
	}

	payloadKey := filePath + "\n" + group + "\n" + strings.Join(links, "\n")
	md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(payloadKey)))
	saveSubscriptionEntry(md5Hash, subscriptionEntry{
		Group: group,
		Links: links,
	})
	subscriptionLink := buildSubscriptionLink(r, md5Hash, group)
	writeJSON(w, http.StatusOK, getSubscriptionLinkResponse{Link: subscriptionLink})
}

func sanitizeLinks(links []string) []string {
	output := make([]string, 0, len(links))
	for _, link := range links {
		trimmed := strings.TrimSpace(link)
		if trimmed == "" {
			continue
		}
		output = append(output, trimmed)
	}
	return output
}

func encodeLinksSubscription(links []string) string {
	joined := strings.Join(sanitizeLinks(links), "\n")
	if joined == "" {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(joined))
}

func getSubscription(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	key := queries.Get("key")
	if len(key) < 1 {
		http.Error(w, "Key not found", http.StatusBadRequest)
		return
	}
	entry, ok := getSubscriptionEntry(key)
	if !ok {
		http.Error(w, "Wrong key", http.StatusBadRequest)
		return
	}
	payload := encodeLinksSubscription(entry.Links)
	if payload == "" {
		http.Error(w, "No links found", http.StatusBadRequest)
		return
	}
	writePlainText(w, http.StatusOK, payload)
}

func writeClash(filePath string) ([]byte, error) {
	links, err := parseClashFileByLine(filePath)
	if err != nil {
		return nil, err
	}
	subscription := []byte(strings.Join(links, "\n"))
	data := make([]byte, base64.StdEncoding.EncodedLen(len(subscription)))
	base64.StdEncoding.Encode(data, subscription)
	return data, nil
}

func writeShadowrocket(data []byte) ([]byte, error) {
	links, err := ParseLinks(string(data))
	if err != nil {
		return nil, err
	}
	newLinks := make([]string, 0, len(links))
	for _, link := range links {
		if strings.HasPrefix(link, "vmess://") && strings.Contains(link, "&") {
			if newLink, err := config.ShadowrocketLinkToVmessLink(link); err == nil {
				newLinks = append(newLinks, newLink)
			}
		} else {
			newLinks = append(newLinks, link)
		}
	}
	subscription := []byte(strings.Join(newLinks, "\n"))
	data = make([]byte, base64.StdEncoding.EncodedLen(len(subscription)))
	base64.StdEncoding.Encode(data, subscription)
	return data, nil
}
