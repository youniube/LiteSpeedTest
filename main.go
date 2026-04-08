package main

import (
	"flag"
	"log"
)

var (
	port           = flag.Int("p", 8090, "set port")
	test           = flag.String("test", "", "test from command line with subscription link or file")
	conf           = flag.String("config", "", "command line options")
	ping           = flag.Int("ping", 2, "retry times to ping link on startup")
	grpc           = flag.Bool("grpc", false, "start grpc server")
	version        = flag.Bool("v", false, "show LiteSpeedTest version")
	engineName     = flag.String("engine", "", "native | singbox")
	singboxBin     = flag.String("singbox-bin", "sing-box", "path to sing-box binary")
	singboxWorkDir = flag.String("singbox-workdir", ".lite-singbox", "sing-box temp work directory")
	keepTemp       = flag.Bool("keep-temp", false, "keep sing-box temp files")
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
