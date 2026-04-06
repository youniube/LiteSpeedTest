package core

import (
	"context"

	"github.com/xxf098/lite-proxy/tunnel"
	"github.com/xxf098/lite-proxy/tunnel/adapter"
	tunnelhttp "github.com/xxf098/lite-proxy/tunnel/http"
	"github.com/xxf098/lite-proxy/tunnel/socks"
)

func newInstanceContext(c Config) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "LocalHost", c.LocalHost)
	ctx = context.WithValue(ctx, "LocalPort", c.LocalPort)
	return ctx, cancel
}

func buildSources(ctx context.Context) ([]tunnel.Server, error) {
	adapterServer, err := adapter.NewServer(ctx, nil)
	if err != nil {
		return nil, err
	}

	httpServer, err := tunnelhttp.NewServer(ctx, adapterServer)
	if err != nil {
		return nil, err
	}

	socksServer, err := socks.NewServer(ctx, adapterServer)
	if err != nil {
		return nil, err
	}

	return []tunnel.Server{httpServer, socksServer}, nil
}
