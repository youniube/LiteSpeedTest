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

func createSink(ctx context.Context, o Options) (tunnel.Client, error) {
	if engine.NeedExternalEngine(o.Engine, o.Link) {
		return createExternalEngineSink(ctx, c)
	}
	return createInternalSink(ctx, c)
}

func createExternalEngineSink(ctx context.Context, o Options) (tunnel.Client, error) {
	runner := singbox.New(o.SingboxBin, o.SingboxWorkDir)
	lp, err := runner.Start(ctx, o.Link, engine.StartOptions{
		SingboxBin:   o.SingboxBin,
		WorkDir:      o.SingboxWorkDir,
		LogLevel:     "warn",
		StartupWait:  5 * time.Second,
		KeepTempFile: o.KeepTempFile,
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

func createInternalSink(ctx context.Context, o Options) (tunnel.Client, error) {
	matches, err := utils.CheckLink(o.Link)
	if err != nil {
		return nil, err
	}

	creator, err := outbound.GetDialerCreator(matches[1])
	if err != nil {
		return nil, err
	}

	d, err := creator(o.Link)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, errors.New("not supported link")
	}

	return proxy.NewClient(ctx, d), nil
}
