package subscription

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/xxf098/lite-proxy/utils"
)

func FetchSubscription(ctx context.Context, input string) ([]byte, string, error) {
	if !utils.IsUrl(input) {
		return nil, "", ErrInvalidSubscription
	}
	client := &http.Client{Timeout: 20 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSpace(input), nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return data, resp.Header.Get("Content-Type"), nil
}
