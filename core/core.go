package core

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/xxf098/lite-proxy/config"
	_ "github.com/xxf098/lite-proxy/config"
	"github.com/xxf098/lite-proxy/dns"
	"github.com/xxf098/lite-proxy/engine"
	"github.com/xxf098/lite-proxy/engine/singbox"
	"github.com/xxf098/lite-proxy/outbound"
	"github.com/xxf098/lite-proxy/proxy"
	"github.com/xxf098/lite-proxy/request"
	"github.com/xxf098/lite-proxy/transport/resolver"
	"github.com/xxf098/lite-proxy/tunnel"
	"github.com/xxf098/lite-proxy/tunnel/adapter"
	"github.com/xxf098/lite-proxy/tunnel/http"
	"github.com/xxf098/lite-proxy/tunnel/socks"
	"github.com/xxf098/lite-proxy/utils"
)

func StartInstance(c Config) (*proxy.Proxy, error) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "LocalHost", c.LocalHost)
	ctx = context.WithValue(ctx, "LocalPort", c.LocalPort)

	adapterServer, err := adapter.NewServer(ctx, nil)
	if err != nil {
		cancel()
		return nil, err
	}
	httpServer, err := http.NewServer(ctx, adapterServer)
	if err != nil {
		cancel()
		return nil, err
	}
	socksServer, err := socks.NewServer(ctx, adapterServer)
	if err != nil {
		cancel()
		return nil, err
	}
	sources := []tunnel.Server{httpServer, socksServer}

	sink, err := createSink(ctx, c)
	if err != nil {
		cancel()
		return nil, err
	}

	go func(link string) {
		if c.Ping < 1 {
			return
		}
		if cfg, err := config.Link2Config(c.Link); err == nil {
			info := fmt.Sprintf("%s %s", cfg.Remarks, net.JoinHostPort(cfg.Server, strconv.Itoa(cfg.Port)))
			if engine.NeedExternalEngine(c.Engine, link) {
				log.Print(info)
				return
			}
			opt := request.PingOption{Attempts: c.Ping, TimeOut: 1200 * time.Millisecond}
			if elapse, err := request.PingLinkInternal(link, opt); err == nil {
				info = fmt.Sprintf("%s \033[32m%dms\033[0m", info, elapse)
			} else {
				info = fmt.Sprintf("\033[31m%s\033[0m", err.Error())
			}
			log.Print(info)
		}
	}(c.Link)

	setDefaultResolver()
	p := proxy.NewProxy(ctx, cancel, sources, sink)
	return p, nil
}

func createSink(ctx context.Context, c Config) (tunnel.Client, error) {
	if engine.NeedExternalEngine(c.Engine, c.Link) {
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

	var d outbound.Dialer
	matches, err := utils.CheckLink(c.Link)
	if err != nil {
		return nil, err
	}
	creator, err := outbound.GetDialerCreator(matches[1])
	if err != nil {
		return nil, err
	}
	d, err = creator(c.Link)
	if err != nil {
		return nil, err
	}
	if d != nil {
		return proxy.NewClient(ctx, d), nil
	}
	return nil, errors.New("not supported link")
}

func setDefaultResolver() {
	resolver.DefaultResolver = dns.DefaultResolver()
}
