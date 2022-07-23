package server

import (
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/tinfoil-knight/gargoyle/internal/loadbalancer"
)

func NewHTTPServer() {
	addr := ":8080"

	lb, err := loadbalancer.NewLoadBalancer([]string{"http://localhost:3040", "http://localhost:3030"})
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", handleAllRequests(lb))

	log.Printf("INFO: Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, logHTTPRequest(mux)))
}

func handleAllRequests(lb *loadbalancer.LoadBalancer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy := lb.GetSelectedProxy()
		proxy.ServeHTTP(w, r)
	}
}

func logHTTPRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		dmp, _ := httputil.DumpRequest(r, true)
		log.Printf("%s", string(dmp))
	})
}
