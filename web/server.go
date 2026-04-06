package web

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
)

func ServeFile(port int) error {
	mux := newServerMux()
	log.Printf("Start server at http://127.0.0.1:%d\n", port)
	if ipAddr, err := localIP(); err == nil {
		log.Printf("Start server at http://%s", net.JoinHostPort(ipAddr.String(), strconv.Itoa(port)))
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func serverFile(w http.ResponseWriter, r *http.Request) {
	h := http.FileServer(http.FS(guiStatic))
	r.URL.Path = "gui/dist" + r.URL.Path
	h.ServeHTTP(w, r)
}
