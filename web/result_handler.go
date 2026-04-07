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
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	result := TestResult{}
	if r.Body == nil {
		writeAPIError(w, http.StatusBadRequest, "Please send a request body")
		return
	}
	data, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "Please send a request body")
		return
	}
	if err = json.Unmarshal(data, &result); err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}

	fontPath := "WenQuanYiMicroHei-01.ttf"
	options := render.NewTableOptions(40, 30, 0.5, 0.5, result.FontSize, 0.5, fontPath, result.Language, result.Theme, "Asia/Shanghai", FontBytes)
	table, err := render.NewTableWithOption(result.Nodes, &options)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
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
	picdata, err := table.EncodeB64(msg)
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, err.Error())
		return
	}
	writePlainText(w, http.StatusOK, picdata)
}
