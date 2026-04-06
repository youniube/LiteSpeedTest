package web

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/xxf098/lite-proxy/config"
)

var upgrader = websocket.Upgrader{}

func updateTest(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		links, options, err := parseMessage(message)
		if err != nil {
			msg := `{"info": "error", "reason": "invalidsub"}`
			c.WriteMessage(mt, []byte(msg))
			continue
		}

		if options.Unique {
			links = uniqueLinks(links)
		}

		p := ProfileTest{
			Writer:      c,
			MessageType: mt,
			Links:       links,
			Options:     options,
		}
		go p.testAll(ctx)
	}
}

func uniqueLinks(links []string) []string {
	uniqueLinks := make([]string, 0, len(links))
	uniqueMap := make(map[string]struct{}, len(links))
	for _, link := range links {
		cfg, err := config.Link2Config(link)
		if err != nil {
			continue
		}
		key := fmt.Sprintf("%s%d%s%s%s", cfg.Server, cfg.Port, cfg.Password, cfg.Protocol, cfg.SNI)
		if _, ok := uniqueMap[key]; ok {
			continue
		}
		uniqueLinks = append(uniqueLinks, link)
		uniqueMap[key] = struct{}{}
	}
	return uniqueLinks
}
