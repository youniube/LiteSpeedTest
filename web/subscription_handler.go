package web

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/xxf098/lite-proxy/config"
	"github.com/xxf098/lite-proxy/utils"
)

type GetSubscriptionLink struct {
	FilePath string `json:"filePath"`
	Group    string `json:"group"`
}

type getSubscriptionLinkResponse struct {
	Link string `json:"link"`
}

var subscriptionLinkMap = map[string]string{}

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

	body := GetSubscriptionLink{}
	if r.Body == nil {
		writeAPIError(w, http.StatusBadRequest, "Invalid Parameter")
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "Invalid Parameter")
		return
	}
	if err = json.Unmarshal(data, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(body.FilePath) == 0 || len(body.Group) == 0 {
		writeAPIError(w, http.StatusBadRequest, "Invalid Parameter")
		return
	}

	md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(body.FilePath)))
	subscriptionLinkMap[md5Hash] = body.FilePath
	subscriptionLink := buildSubscriptionLink(r, md5Hash, body.Group)
	writeJSON(w, http.StatusOK, getSubscriptionLinkResponse{Link: subscriptionLink})
}

func getSubscription(w http.ResponseWriter, r *http.Request) {
	queries := r.URL.Query()
	key := queries.Get("key")
	if len(key) < 1 {
		http.Error(w, "Key not found", http.StatusBadRequest)
		return
	}
	sub := queries.Get("sub")
	filePath, ok := subscriptionLinkMap[key]
	if !ok {
		http.Error(w, "Wrong key", http.StatusBadRequest)
		return
	}

	if isYamlFile(filePath) && utils.IsUrl(filePath) {
		links, err := getSubscriptionLinks(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		b64Data := base64.StdEncoding.EncodeToString([]byte(strings.Join(links, "\n")))
		writePlainText(w, http.StatusOK, b64Data)
		return
	}

	if isYamlFile(filePath) {
		data, err := writeClash(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writePlainText(w, http.StatusOK, string(data))
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(data) > 128 && strings.Contains(string(data[:128]), "proxies:") {
		if dataClash, err := writeClash(filePath); err == nil && len(dataClash) > 0 {
			data = dataClash
		}
	}
	if sub == "v2ray" {
		if dataShadowrocket, err := writeShadowrocket(data); err == nil && len(dataShadowrocket) > 0 {
			data = dataShadowrocket
		}
	}

	writePlainText(w, http.StatusOK, string(data))
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
