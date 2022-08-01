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

func Start(configPath string) {
	config := loadConfig(configPath)

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
				proxy.ServeHTTP(w, r)
			})
		} else {
			lb, err := loadbalancer.NewLoadBalancer(rp.Algorithm, rp.Targets)
			if err != nil {
				panic(err)
			}
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				proxy := lb.GetSelectedProxy()
				proxy.ServeHTTP(w, r)
			})
			if rp.HealthCheck.Enabled {
				go lb.RunHealthChecks(
					time.Duration(rp.HealthCheck.Interval)*time.Second,
					time.Duration(rp.HealthCheck.Timeout)*time.Second,
					rp.HealthCheck.Path,
				)
			}
		}

		handler := applyMiddlewares(mux, service)
		log.Printf("INFO: Starting reverse proxy on %s", service.Source)
		log.Fatal(http.ListenAndServe(service.Source, handler))
	}

	if service.Fs != nil {
		fs := service.Fs
		handler := applyMiddlewares(http.FileServer(http.Dir(fs.Path)), service)
		http.Handle("/", handler)
		log.Printf("INFO: Starting file server on %s", service.Source)
		log.Fatal(http.ListenAndServe(service.Source, nil))
	}
	log.Printf("Port %s didn't have any required config for starting a service", service.Source)
}

func logHTTPRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
		dmp, _ := httputil.DumpRequest(r, true)
		log.Printf("%s", string(dmp))
	})
}

func useCustomRewriter(handler http.Handler, service ServiceCfg) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &customResponseWriter{w, false, service.Header}
		handler.ServeHTTP(rw, r)
	})
}

func applyMiddlewares(handler http.Handler, service ServiceCfg) http.Handler {
	return logHTTPRequest(useCustomRewriter(handler, service))
}
