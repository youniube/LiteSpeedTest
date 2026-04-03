package request

import (
	"context"
	"fmt"
	"net"
)

func PingWithDial(ctx context.Context, dial func(network, addr string) (net.Conn, error), opt PingOption) (int64, error) {
	if opt.Attempts < 1 {
		opt.Attempts = 1
	}
	var lastErr error
	for i := 0; i < opt.Attempts; i++ {
		select {
		case <-ctx.Done():
			if lastErr != nil {
				return 0, lastErr
			}
			return 0, ctx.Err()
		default:
		}
		remoteConn, err := dial("tcp", net.JoinHostPort(remoteHost, "80"))
		if err != nil {
			lastErr = err
			continue
		}
		elapse, err := pingInternal(remoteConn)
		_ = remoteConn.Close()
		if err == nil && elapse > 0 {
			return elapse, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("ping failed")
	}
	return 0, lastErr
}
