package server

import (
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/tinfoil-knight/gargoyle/internal/config"
	"github.com/tinfoil-knight/gargoyle/internal/reverseproxy"
)

func Start(configPath string) {
	config := config.LoadConfig(configPath)

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

func NewServiceController(service config.ServiceCfg) {
	var handler http.Handler
	switch true {
	case service.ReverseProxy != nil:
		mux := reverseproxy.NewReverseProxy(service)
		handler = applyMiddlewares(mux, service)
		log.Printf("INFO: Starting reverse proxy on %s", service.Source)
	case service.Fs != nil:
		fs := service.Fs
		handler = applyMiddlewares(http.FileServer(http.Dir(fs.Path)), service)
		http.Handle("/", handler)
		log.Printf("INFO: Starting file server on %s", service.Source)
	default:
		log.Printf("Port %s didn't have any required config for starting a service", service.Source)
		os.Exit(1)
	}

	srv := httpServer(handler, &service)
	if service.TLS != nil && service.TLS.Enabled {
		log.Fatal(
			srv.ListenAndServeTLS(service.TLS.CertPath, service.TLS.KeyPath),
		)
	}
	log.Fatal(srv.ListenAndServe())
}

func httpServer(handler http.Handler, serviceCfg *config.ServiceCfg) *http.Server {
	return &http.Server{
		Addr:         serviceCfg.Source,
		ReadTimeout:  time.Duration(serviceCfg.Timeout.Read) * time.Second,
		WriteTimeout: time.Duration(serviceCfg.Timeout.Write) * time.Second,
		IdleTimeout:  time.Duration(serviceCfg.Timeout.Idle) * time.Second,
		Handler:      handler,
	}
}
