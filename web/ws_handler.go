package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/xxf098/lite-proxy/config"
)

var upgrader = websocket.Upgrader{}

type safeWSWriter struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *safeWSWriter) WriteMessage(messageType int, data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteMessage(messageType, data)
}

func updateTest(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	writer := &safeWSWriter{conn: c}
	var running atomic.Bool

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		if !running.CompareAndSwap(false, true) {
			_ = writer.WriteMessage(mt, []byte(`{"info": "error", "reason": "busy"}`))
			continue
		}

		links, options, err := parseMessage(message)
		if err != nil {
			_ = writer.WriteMessage(mt, []byte(`{"info": "error", "reason": "invalidsub"}`))
			running.Store(false)
			continue
		}

		if options.Unique {
			links = uniqueLinks(links)
		}

		p := ProfileTest{
			Writer:      writer,
			MessageType: mt,
			Links:       links,
			Options:     options,
		}
		go func() {
			defer running.Store(false)
			_, _ = p.testAll(ctx)
		}()
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
