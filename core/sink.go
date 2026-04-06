package core

import (
	"context"
	"errors"
	"time"

	"github.com/xxf098/lite-proxy/engine"
	"github.com/xxf098/lite-proxy/engine/singbox"
	"github.com/xxf098/lite-proxy/outbound"
	"github.com/xxf098/lite-proxy/proxy"
	"github.com/xxf098/lite-proxy/tunnel"
	"github.com/xxf098/lite-proxy/utils"
)

func createSink(ctx context.Context, c Config) (tunnel.Client, error) {
	if engine.NeedExternalEngine(c.Engine, c.Link) {
		return createExternalEngineSink(ctx, c)
	}
	return createInternalSink(ctx, c)
}

func createExternalEngineSink(ctx context.Context, c Config) (tunnel.Client, error) {
	runner := singbox.New(c.SingboxBin, c.SingboxWorkDir)
	lp, err := runner.Start(ctx, c.Link, engine.StartOptions{
		SingboxBin:   c.SingboxBin,
		WorkDir:      c.SingboxWorkDir,
		LogLevel:     "warn",
		StartupWait:  5 * time.Second,
		KeepTempFile: c.KeepTempFile,
	})
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		_ = lp.Close(context.Background())
	}()

	return proxy.NewLocalSocksClient(ctx, lp.SOCKSAddr), nil
}

func createInternalSink(ctx context.Context, c Config) (tunnel.Client, error) {
	matches, err := utils.CheckLink(c.Link)
	if err != nil {
		return nil, err
	}

	creator, err := outbound.GetDialerCreator(matches[1])
	if err != nil {
		return nil, err
	}

	d, err := creator(c.Link)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, errors.New("not supported link")
	}

	return proxy.NewClient(ctx, d), nil
}
