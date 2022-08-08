package server

import (
	"log"
	"net/http"
	"os"
	"sync"

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
	switch true {
	case service.ReverseProxy != nil:
		mux := reverseproxy.NewReverseProxy(service)
		handler := applyMiddlewares(mux, service)
		log.Printf("INFO: Starting reverse proxy on %s", service.Source)
		if service.TLS != nil && service.TLS.Enabled {
			log.Fatal(
				http.ListenAndServeTLS(service.Source, service.TLS.CertPath, service.TLS.KeyPath, handler),
			)
		}
		log.Fatal(http.ListenAndServe(service.Source, handler))
	case service.Fs != nil:
		fs := service.Fs
		handler := applyMiddlewares(http.FileServer(http.Dir(fs.Path)), service)
		http.Handle("/", handler)
		log.Printf("INFO: Starting file server on %s", service.Source)
		if service.TLS != nil && service.TLS.Enabled {
			log.Fatal(
				http.ListenAndServeTLS(service.Source, service.TLS.CertPath, service.TLS.KeyPath, handler),
			)
		}
		log.Fatal(http.ListenAndServe(service.Source, nil))
	default:
		log.Printf("Port %s didn't have any required config for starting a service", service.Source)
		os.Exit(1)
	}
}
