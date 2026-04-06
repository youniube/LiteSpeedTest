package web

import "net/http"

const (
	routeHome                = "/"
	routeTest                = "/test"
	routeGetSubscriptionLink = "/getSubscriptionLink"
	routeGetSubscription     = "/getSubscription"
	routeGenerateResult      = "/generateResult"
	routeRenameNodes         = "/renameNodes"
)

func newServerMux() *http.ServeMux {
	mux := http.NewServeMux()
	registerRoutes(mux)
	return mux
}

func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc(routeHome, serverFile)
	mux.HandleFunc(routeTest, updateTest)
	mux.HandleFunc(routeGetSubscriptionLink, getSubscriptionLink)
	mux.HandleFunc(routeGetSubscription, getSubscription)
	mux.HandleFunc(routeGenerateResult, generateResult)
	mux.HandleFunc(routeRenameNodes, renameNodesHandler)
}
