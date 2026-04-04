package engine

import (
	"context"
	"strings"
	"time"
)

type StartOptions struct {
	SingboxBin   string
	WorkDir      string
	LogLevel     string
	StartupWait  time.Duration
	KeepTempFile bool
}

type LocalProxy struct {
	HTTPAddr  string
	SOCKSAddr string
	CloseFunc func(context.Context) error
}

func (p *LocalProxy) Close(ctx context.Context) error {
	if p == nil || p.CloseFunc == nil {
		return nil
	}
	return p.CloseFunc(ctx)
}

type Runner interface {
	Name() string
	Start(ctx context.Context, link string, opt StartOptions) (*LocalProxy, error)
}

func NeedExternalEngine(engineName, link string) bool {
	if strings.EqualFold(strings.TrimSpace(engineName), "native") {
		return false
	}
	return SupportsSingbox(link)
}

func SupportsSingbox(link string) bool {
	link = strings.ToLower(strings.TrimSpace(link))
	switch {
	case strings.HasPrefix(link, "vless://"):
		return true
	case strings.HasPrefix(link, "vmess://"):
		return true
	case strings.HasPrefix(link, "trojan://"):
		return true
	case strings.HasPrefix(link, "ss://"):
		return true
	default:
		return false
	}
}
