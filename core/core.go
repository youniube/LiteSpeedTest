package core

import "github.com/xxf098/lite-proxy/proxy"

func StartInstance(c Config) (*proxy.Proxy, error) {
	ctx, cancel := newInstanceContext(c)

	sources, err := buildSources(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	sink, err := createSink(ctx, c)
	if err != nil {
		cancel()
		return nil, err
	}

	startStartupPing(c)
	setDefaultResolver()

	return proxy.NewProxy(ctx, cancel, sources, sink), nil
}
