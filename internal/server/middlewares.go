package server

import (
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/tinfoil-knight/gargoyle/internal/config"
)

func logHTTPRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		dmp, _ := httputil.DumpRequest(r, true)
		log.Printf("%s", string(dmp))
	})
}

func useCustomRewriter(handler http.Handler, service config.ServiceCfg) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &customResponseWriter{w, false, service.Header}
		handler.ServeHTTP(rw, r)
	})
}

func applyMiddlewares(handler http.Handler, service config.ServiceCfg) http.Handler {
	return logHTTPRequest(useCustomRewriter(handler, service))
}
