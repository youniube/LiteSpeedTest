package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/xxf098/lite-proxy/web/render"
)

type TestResult struct {
	TotalTraffic string       `json:"totalTraffic"`
	TotalTime    string       `json:"totalTime"`
	Language     string       `json:"language"`
	FontSize     int          `json:"fontSize"`
	Theme        string       `json:"theme"`
	Nodes        render.Nodes `json:"nodes"`
}

func generateResult(w http.ResponseWriter, r *http.Request) {
	result := TestResult{}
	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(data, &result); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fontPath := "WenQuanYiMicroHei-01.ttf"
	options := render.NewTableOptions(40, 30, 0.5, 0.5, result.FontSize, 0.5, fontPath, result.Language, result.Theme, "Asia/Shanghai", FontBytes)
	table, err := render.NewTableWithOption(result.Nodes, &options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	linksCount := 0
	successCount := 0
	for _, v := range result.Nodes {
		linksCount++
		if v.IsOk {
			successCount++
		}
	}
	msg := table.FormatTraffic(result.TotalTraffic, result.TotalTime, fmt.Sprintf("%d/%d", successCount, linksCount))
	if picdata, err := table.EncodeB64(msg); err == nil {
		fmt.Fprint(w, picdata)
	}
}
