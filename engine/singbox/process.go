package singbox

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	binPath := strings.TrimSpace(pickFirstNonEmpty(r.binPath, opt.SingboxBin))
	if binPath == "" {
		binPath = "sing-box"
	}
	workDir := strings.TrimSpace(pickFirstNonEmpty(r.workDir, opt.WorkDir))
	if workDir == "" {
		workDir = ".lite-singbox"
	}
	if opt.StartupWait <= 0 {
		opt.StartupWait = 5 * time.Second
	}

	listenPort, err := pickFreeTCPPort()
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

	cmd := exec.CommandContext(ctx, binPath, "run", "-c", configPath)
	cmd.Dir = sessionDir

	logFilePath := filepath.Join(sessionDir, "sing-box.log")
	logFile, openErr := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if openErr == nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	if err = cmd.Start(); err != nil {
		if logFile != nil {
			_ = logFile.Close()
		}
		if !opt.KeepTempFile {
			_ = os.RemoveAll(sessionDir)
		}
		return nil, err
	}

	cleanup := func(force bool) error {
		var result error
		if cmd.Process != nil {
			if force {
				_ = cmd.Process.Kill()
			}
			_, _ = cmd.Process.Wait()
		}
		if logFile != nil {
			_ = logFile.Close()
		}
		if !opt.KeepTempFile {
			if err := os.RemoveAll(sessionDir); err != nil && result == nil {
				result = err
			}
		}
		return result
	}

	readyCtx, cancel := context.WithTimeout(ctx, opt.StartupWait)
	defer cancel()
	if err = waitTCPReady(readyCtx, net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", listenPort))); err != nil {
		_ = cleanup(true)
		return nil, fmt.Errorf("start sing-box failed: %w", err)
	}

	var once sync.Once
	closeFunc := func(closeCtx context.Context) error {
		var closeErr error
		once.Do(func() {
			done := make(chan struct{})
			go func() {
				closeErr = cleanup(true)
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
		HTTPAddr:  net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", listenPort)),
		SOCKSAddr: net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", listenPort)),
		CloseFunc: closeFunc,
	}, nil
}

func pickFreeTCPPort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.New("invalid tcp addr")
	}
	return addr.Port, nil
}

func waitTCPReady(ctx context.Context, addr string) error {
	d := net.Dialer{Timeout: 300 * time.Millisecond}
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()
	for {
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func pickFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
