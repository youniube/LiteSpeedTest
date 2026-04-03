package singbox

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/xxf098/lite-proxy/engine"
)

type Engine struct {
	BinPath  string
	WorkRoot string
	LogLevel string
	KeepTemp bool
}

func New(binPath, workRoot string) *Engine {
	return &Engine{
		BinPath:  binPath,
		WorkRoot: workRoot,
		LogLevel: "warn",
	}
}

func (e *Engine) Name() string { return "singbox" }

func (e *Engine) Start(ctx context.Context, link string, opt engine.StartOptions) (*engine.LocalProxy, error) {
	port, err := ReservePort()
	if err != nil {
		return nil, err
	}

	outbound, err := BuildOutbound(link)
	if err != nil {
		return nil, err
	}

	cfg := NewSingleNodeConfig(port, outbound, opt.LogLevel)
	workDir := filepath.Join(opt.WorkDir, fmt.Sprintf("node-%d", port))
	configPath, cleanupFiles, err := WriteConfig(workDir, cfg)
	if err != nil {
		return nil, err
	}

	proc, err := StartProcess(ctx, opt.SingboxBin, configPath)
	if err != nil {
		_ = cleanupFiles()
		return nil, err
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	if err := WaitReady(addr, opt.StartupWait); err != nil {
		_ = proc.Close(context.Background())
		_ = cleanupFiles()
		return nil, err
	}

	return &engine.LocalProxy{
		HTTPAddr:  addr,
		SOCKSAddr: addr,
		CloseFunc: func(closeCtx context.Context) error {
			err := proc.Close(closeCtx)
			if !opt.KeepTempFile {
				_ = cleanupFiles()
			}
			return err
		},
	}, nil
}
