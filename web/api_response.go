package web

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type apiErrorResponse struct {
	Error string `json:"error"`
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}
	writeAPIError(w, http.StatusMethodNotAllowed, fmt.Sprintf("method %s not allowed", r.Method))
	return false
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writePlainText(w http.ResponseWriter, status int, text string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(text))
}

func writeAPIError(w http.ResponseWriter, status int, errMsg string) {
	writeJSON(w, status, apiErrorResponse{Error: errMsg})
}
