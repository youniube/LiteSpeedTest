package core

import "github.com/xxf098/lite-proxy/proxy"

func StartInstance(o Options) (*proxy.Proxy, error) {
	ctx, cancel := newInstanceContext(o)

	sources, err := buildSources(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	sink, err := createSink(ctx, o)
	if err != nil {
		cancel()
		return nil, err
	}

	startStartupPing(o)
	setDefaultResolver()

	return proxy.NewProxy(ctx, cancel, sources, sink), nil
}
