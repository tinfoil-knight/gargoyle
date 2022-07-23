package server

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewHTTPServer() {
	addr := ":8080"

	proxy, err := newProxy("http://localhost:3030")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", handleAllRequests(proxy))

	log.Printf("INFO: Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, logHTTPRequest(mux)))
}

func handleAllRequests(proxy *httputil.ReverseProxy) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

func newProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(url), err
}

func logHTTPRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		dmp, _ := httputil.DumpRequest(r, true)
		log.Printf("%s", string(dmp))
	})
}
