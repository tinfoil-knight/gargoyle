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

	for _, serviceCfg := range config.Services {
		serviceCfg := serviceCfg
		wg.Add(1)
		go func() {
			defer wg.Done()
			NewServiceController(serviceCfg)
		}()
	}

	wg.Wait()
}

func NewServiceController(service ServiceCfg) {
	if service.ReverseProxy != nil {
		rp := service.ReverseProxy
		mux := http.NewServeMux()
		if len(rp.Targets) == 0 {
			panic("no targets specified")
		}

		if len(rp.Targets) == 1 {
			url, err := url.Parse(rp.Targets[0])
			if err != nil {
				panic(err)
			}
			proxy := httputil.NewSingleHostReverseProxy(url)
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				rw := &customResponseWriter{w, false, service.Header}
				proxy.ServeHTTP(rw, r)
			})
		} else {
			lb, err := loadbalancer.NewLoadBalancer(rp.Algorithm, rp.Targets)
			if err != nil {
				panic(err)
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

		handler := logHTTPRequest(mux)
		log.Printf("INFO: Starting listening on %s", service.Source)
		log.Fatal(http.ListenAndServe(service.Source, handler))
	}
	log.Printf("Service with port %s didn't have any required config", service.Source)
}

func logHTTPRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		dmp, _ := httputil.DumpRequest(r, true)
		log.Printf("%s", string(dmp))
	})
}
