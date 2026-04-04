package singbox

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type singboxProcess struct {
	cmd     *exec.Cmd
	logFile *os.File
}

func startSingboxProcess(ctx context.Context, binPath, sessionDir, configPath string) (*singboxProcess, error) {
	cmd := exec.CommandContext(ctx, binPath, "run", "-c", configPath)
	cmd.Dir = sessionDir

	logFilePath := filepath.Join(sessionDir, "sing-box.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err == nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}

	if err := cmd.Start(); err != nil {
		if logFile != nil {
			_ = logFile.Close()
		}
		return nil, err
	}
	return &singboxProcess{cmd: cmd, logFile: logFile}, nil
}

func (p *singboxProcess) Close(force bool) error {
	if p == nil {
		return nil
	}
	if p.cmd != nil && p.cmd.Process != nil {
		if force {
			_ = p.cmd.Process.Kill()
		}
		_, _ = p.cmd.Process.Wait()
	}
	if p.logFile != nil {
		_ = p.logFile.Close()
	}
	return nil
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

func resolveSingboxBinary(raw string) (string, error) {
	candidate := strings.TrimSpace(raw)
	candidates := make([]string, 0, 4)
	if candidate != "" {
		candidates = append(candidates, candidate)
	}
	defaultName := "sing-box"
	if runtime.GOOS == "windows" {
		defaultName = "sing-box.exe"
	}
	candidates = append(candidates, defaultName)
	if runtime.GOOS == "windows" {
		candidates = append(candidates, "sing-box")
	}
	if exeDir := executableDir(); exeDir != "" {
		candidates = append(candidates, filepath.Join(exeDir, defaultName))
		if runtime.GOOS == "windows" {
			candidates = append(candidates, filepath.Join(exeDir, "sing-box.exe"))
		}
	}

	seen := map[string]struct{}{}
	for _, item := range candidates {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}

		if strings.ContainsRune(item, os.PathSeparator) || filepath.IsAbs(item) {
			if _, err := os.Stat(item); err == nil {
				if abs, absErr := filepath.Abs(item); absErr == nil {
					return abs, nil
				}
				return item, nil
			}
			continue
		}
		if path, err := exec.LookPath(item); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("sing-box binary not found")
}

func executableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exePath)
}

func defaultWorkDir() string {
	if exeDir := executableDir(); exeDir != "" {
		return filepath.Join(exeDir, ".lite-singbox")
	}
	return ".lite-singbox"
}

func cleanupSessionDir(sessionDir string) error {
	if strings.TrimSpace(sessionDir) == "" {
		return nil
	}
	return os.RemoveAll(sessionDir)
}
