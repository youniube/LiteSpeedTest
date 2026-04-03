package download

import (
	"context"
	"net"
	"time"
)

func DownloadWithDial(ctx context.Context, option DownloadOption, resultChan chan<- int64, startChan chan<- time.Time, dial func(network, addr string) (net.Conn, error)) (int64, error) {
	return downloadInternal(ctx, option, resultChan, startChan, dial)
}
