package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	grpcServer "github.com/xxf098/lite-proxy/api/rpc/liteserver"
	C "github.com/xxf098/lite-proxy/constant"
	"github.com/xxf098/lite-proxy/core"
	"github.com/xxf098/lite-proxy/utils"
	webServer "github.com/xxf098/lite-proxy/web"
)

type runMode int

const (
	runModeTest runMode = iota
	runModeGRPC
	runModeWeb
	runModeInstance
)

type closableRunner interface {
	Run() error
	Close() error
}

func run() error {
	if *version {
		printVersion()
		return nil
	}

	link := detectInputLink(os.Args[1:])

	switch resolveRunMode(link) {
	case runModeTest:
		return runTestMode()
	case runModeGRPC:
		return runGRPCMode()
	case runModeWeb:
		return runWebMode()
	default:
		return runInstanceMode(link)
	}
}

func printVersion() {
	fmt.Printf("LiteSpeedTest %s %s %s with %s %s\n", C.Version, runtime.GOOS, runtime.GOARCH, runtime.Version(), C.BuildTime)
}

func detectInputLink(args []string) string {
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if _, err := utils.CheckLink(arg); err == nil {
			return arg
		}
	}
	return ""
}

func resolveRunMode(link string) runMode {
	if *test != "" {
		return runModeTest
	}
	if *grpc {
		return runModeGRPC
	}
	if link == "" {
		return runModeWeb
	}
	return runModeInstance
}

func runTestMode() error {
	return webServer.TestFromCMD(*test, conf)
}

func runGRPCMode() error {
	return grpcServer.StartServer(uint16(*port))
}

func runWebMode() error {
	if len(os.Args) < 2 {
		*port = 10888
	}
	return webServer.ServeFile(*port)
}

func runInstanceMode(link string) error {
	p, err := core.StartInstance(buildCoreOptions(link))
	if err != nil {
		return err
	}
	watchShutdown(p)
	return p.Run()
}

func buildCoreOptions(link string) core.Options {
	return core.Options{
		LocalHost:      "0.0.0.0",
		LocalPort:      *port,
		Link:           link,
		Ping:           *ping,
		Engine:         *engineName,
		SingboxBin:     *singboxBin,
		SingboxWorkDir: *singboxWorkDir,
		KeepTempFile:   *keepTemp,
	}
}

func watchShutdown(p closableRunner) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer signal.Stop(sigs)
		<-sigs
		_ = p.Close()
	}()
}
