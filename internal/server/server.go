package server

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/tinfoil-knight/gargoyle/internal/loadbalancer"
)

func NewHTTPServer() {
	config := loadConfig("./config.json")

	var wg sync.WaitGroup

	for _, service := range config.Services {
		service := service
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("INFO: Starting listening on %s", service.Source)
			mux, _ := NewReverseProxy(service)
			handler := logHTTPRequest(mux)

			log.Fatal(http.ListenAndServe(service.Source, handler))
		}()
	}

	wg.Wait()
}

func NewReverseProxy(service Service) (*http.ServeMux, error) {
	rp := service.ReverseProxy
	if len(rp.Targets) == 0 {
		panic("no targets specified")
	}
	mux := http.NewServeMux()
	if len(rp.Targets) == 1 {
		url, err := url.Parse(rp.Targets[0])
		if err != nil {
			return nil, err
		}
		proxy := httputil.NewSingleHostReverseProxy(url)
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			rw := &customResponseWriter{w, false, service.Header}
			proxy.ServeHTTP(rw, r)
		})
	} else {
		lb, err := loadbalancer.NewLoadBalancer(rp.Algorithm, rp.Targets)
		if err != nil {
			return nil, err
		}
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			rw := &customResponseWriter{w, false, service.Header}
			proxy := lb.GetSelectedProxy()
			proxy.ServeHTTP(rw, r)
		})
		if rp.HealthCheck.Enabled {
			go lb.RunHealthChecks(
				time.Duration(rp.HealthCheck.Interval)*time.Second,
				time.Duration(rp.HealthCheck.Timeout)*time.Second,
				rp.HealthCheck.Path,
			)
		}
	}
	return mux, nil
}

func logHTTPRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		dmp, _ := httputil.DumpRequest(r, true)
		log.Printf("%s", string(dmp))
	})
}
