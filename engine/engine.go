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
	if strings.EqualFold(engineName, "singbox") {
		return true
	}
	return strings.HasPrefix(strings.ToLower(link), "vless://")
}
