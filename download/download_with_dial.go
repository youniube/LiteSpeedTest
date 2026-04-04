package download

import (
	"context"
	"net"
	"time"
)

func DownloadWithDial(ctx context.Context, option DownloadOption, resultChan chan<- int64, startChan chan<- time.Time, dial func(network, addr string) (net.Conn, error)) (int64, error) {
	if option.URL == "" {
		option.URL = downloadLink
	}
	if option.DownloadTimeout <= 0 {
		option.DownloadTimeout = 15 * time.Second
	}
	if option.HandshakeTimeout <= 0 {
		option.HandshakeTimeout = option.DownloadTimeout
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, option.DownloadTimeout)
	defer cancel()
	return downloadInternal(ctx, option, resultChan, startChan, dial)
}
