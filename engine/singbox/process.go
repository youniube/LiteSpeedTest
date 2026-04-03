package singbox

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Process struct {
	cmd        *exec.Cmd
	stdoutFile *os.File
	stderrFile *os.File
}

func StartProcess(ctx context.Context, binPath, configPath string) (*Process, error) {
	if configPath == "" {
		return nil, fmt.Errorf("empty config path")
	}
	if binPath == "" {
		binPath = "sing-box"
	}
	resolvedBin, err := resolveBin(binPath)
	if err != nil {
		return nil, err
	}
	workDir := filepath.Dir(configPath)
	if err := runCheck(ctx, resolvedBin, configPath, workDir); err != nil {
		return nil, err
	}

	stdoutPath := filepath.Join(workDir, "singbox.stdout.log")
	stderrPath := filepath.Join(workDir, "singbox.stderr.log")
	stdoutFile, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}
	stderrFile, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		_ = stdoutFile.Close()
		return nil, err
	}

	cmd := exec.CommandContext(ctx, resolvedBin, "run", "-c", configPath)
	cmd.Dir = workDir
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile
	if err := cmd.Start(); err != nil {
		_ = stdoutFile.Close()
		_ = stderrFile.Close()
		return nil, fmt.Errorf("start sing-box failed: %w", err)
	}
	return &Process{cmd: cmd, stdoutFile: stdoutFile, stderrFile: stderrFile}, nil
}

func runCheck(ctx context.Context, binPath, configPath, workDir string) error {
	cmd := exec.CommandContext(ctx, binPath, "check", "-c", configPath)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("sing-box check failed: %w\n%s", err, string(output))
	}
	return nil
}

func WaitReady(addr string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		lastErr = err
		time.Sleep(150 * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("timeout")
	}
	return fmt.Errorf("sing-box not ready on %s: %w", addr, lastErr)
}

func (p *Process) Close(ctx context.Context) error {
	if p == nil {
		return nil
	}
	var firstErr error
	addErr := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if p.cmd != nil && p.cmd.Process != nil {
		_ = p.cmd.Process.Kill()
		waitCh := make(chan error, 1)
		go func() { waitCh <- p.cmd.Wait() }()
		select {
		case <-ctx.Done():
			addErr(ctx.Err())
		case err := <-waitCh:
			if err != nil {
				addErr(err)
			}
		}
	}
	if p.stdoutFile != nil {
		addErr(p.stdoutFile.Close())
	}
	if p.stderrFile != nil {
		addErr(p.stderrFile.Close())
	}
	return firstErr
}

func resolveBin(binPath string) (string, error) {
	if filepath.IsAbs(binPath) || filepath.Dir(binPath) != "." {
		if _, err := os.Stat(binPath); err != nil {
			return "", fmt.Errorf("sing-box binary not found: %s", binPath)
		}
		return binPath, nil
	}
	resolved, err := exec.LookPath(binPath)
	if err != nil {
		return "", fmt.Errorf("cannot find sing-box binary %q in PATH", binPath)
	}
	return resolved, nil
}
