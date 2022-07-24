package server

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/tinfoil-knight/gargoyle/internal/loadbalancer"
)

func NewHTTPServer() {
	config := loadConfig("./config.json")

	var wg sync.WaitGroup

	for _, rp := range config.ReverseProxies {
		rp := rp
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("INFO: Starting listening on %s", rp.Source)
			log.Fatal(NewReverseProxy(rp))
		}()
	}

	wg.Wait()
}

func NewReverseProxy(rp ReverseProxy) error {
	if len(rp.Targets) == 0 {
		panic("no targets specified")
	}
	mux := http.NewServeMux()
	if len(rp.Targets) == 1 {
		url, err := url.Parse(rp.Targets[0])
		if err != nil {
			return err
		}
		proxy := httputil.NewSingleHostReverseProxy(url)
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})
	} else {
		lb, err := loadbalancer.NewLoadBalancer(rp.Algorithm, rp.Targets)
		if err != nil {
			return err
		}
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			proxy := lb.GetSelectedProxy()
			proxy.ServeHTTP(w, r)
		})
	}
	return http.ListenAndServe(rp.Source, logHTTPRequest(mux))
}

func logHTTPRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		dmp, _ := httputil.DumpRequest(r, true)
		log.Printf("%s", string(dmp))
	})
}
