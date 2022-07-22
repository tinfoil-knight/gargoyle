package server

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

func NewHTTPServer() {
	addr := ":8080"

	mux := http.NewServeMux()

	mux.HandleFunc("/", handleAllRequests)

	log.Printf("INFO: Starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, logHTTPRequest(mux)))
}

func handleAllRequests(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %s!", r.UserAgent())
}

func logHTTPRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		dmp, _ := httputil.DumpRequest(r, true)
		log.Printf("%s", string(dmp))
	})
}
