package singbox

import (
	"context"
	"fmt"
	"net"
	"path/filepath"
	"sync"
	"time"

	"github.com/xxf098/lite-proxy/engine"
)

type Runner struct {
	binPath string
	workDir string
}

func New(binPath, workDir string) *Runner {
	return &Runner{binPath: binPath, workDir: workDir}
}

func (r *Runner) Name() string {
	return "sing-box"
}

func (r *Runner) Start(ctx context.Context, link string, opt engine.StartOptions) (*engine.LocalProxy, error) {
	binPath, err := resolveSingboxBinary(firstNonEmpty(opt.SingboxBin, r.binPath))
	if err != nil {
		return nil, err
	}
	workDir := firstNonEmpty(opt.WorkDir, r.workDir)
	if workDir == "" {
		workDir = defaultWorkDir()
	}
	if opt.StartupWait <= 0 {
		opt.StartupWait = 5 * time.Second
	}

	listenPort, err := ReservePort()
	if err != nil {
		return nil, err
	}
	configPath, sessionDir, err := writeConfig(link, workDir, listenPort, opt)
	if err != nil {
		return nil, err
	}
	if absSessionDir, absErr := filepath.Abs(sessionDir); absErr == nil {
		sessionDir = absSessionDir
	}
	if absConfigPath, absErr := filepath.Abs(configPath); absErr == nil {
		configPath = absConfigPath
	}

	proc, err := startSingboxProcess(ctx, binPath, sessionDir, configPath)
	if err != nil {
		if !opt.KeepTempFile {
			_ = cleanupSessionDir(sessionDir)
		}
		return nil, err
	}

	addr := net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", listenPort))
	readyCtx, cancel := context.WithTimeout(ctx, opt.StartupWait)
	defer cancel()
	if err := waitTCPReady(readyCtx, addr); err != nil {
		_ = proc.Close(true)
		if !opt.KeepTempFile {
			_ = cleanupSessionDir(sessionDir)
		}
		return nil, fmt.Errorf("start sing-box failed: %w", err)
	}

	var once sync.Once
	closeFunc := func(closeCtx context.Context) error {
		var closeErr error
		once.Do(func() {
			done := make(chan struct{})
			go func() {
				closeErr = proc.Close(true)
				if !opt.KeepTempFile {
					if err := cleanupSessionDir(sessionDir); err != nil && closeErr == nil {
						closeErr = err
					}
				}
				close(done)
			}()
			select {
			case <-done:
			case <-closeCtx.Done():
				closeErr = closeCtx.Err()
			}
		})
		return closeErr
	}

	return &engine.LocalProxy{
		HTTPAddr:  addr,
		SOCKSAddr: addr,
		CloseFunc: closeFunc,
	}, nil
}
