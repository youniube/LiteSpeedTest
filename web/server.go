package web

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
)

func ServeFile(port int) error {
	registerRoutes()
	log.Printf("Start server at http://127.0.0.1:%d\n", port)
	if ipAddr, err := localIP(); err == nil {
		log.Printf("Start server at http://%s", net.JoinHostPort(ipAddr.String(), strconv.Itoa(port)))
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func registerRoutes() {
	http.HandleFunc("/", serverFile)
	http.HandleFunc("/test", updateTest)
	http.HandleFunc("/getSubscriptionLink", getSubscriptionLink)
	http.HandleFunc("/getSubscription", getSubscription)
	http.HandleFunc("/generateResult", generateResult)
	http.HandleFunc("/renameNodes", renameNodesHandler)
}

func serverFile(w http.ResponseWriter, r *http.Request) {
	h := http.FileServer(http.FS(guiStatic))
	r.URL.Path = "gui/dist" + r.URL.Path
	h.ServeHTTP(w, r)
}
