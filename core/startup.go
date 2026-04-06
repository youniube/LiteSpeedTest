package core

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/xxf098/lite-proxy/config"
	"github.com/xxf098/lite-proxy/engine"
	"github.com/xxf098/lite-proxy/request"
)

func startStartupPing(c Config) {
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
}
