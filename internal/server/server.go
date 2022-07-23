package server

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"

	"github.com/tinfoil-knight/gargoyle/internal/loadbalancer"
)

type Config struct {
	ReverseProxies []ReverseProxy `json:"reverse_proxy"`
}

type ReverseProxy struct {
	Source    string   `json:"source"`
	Algorithm string   `json:"algorithm"`
	Targets   []string `json:"targets"`
}

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

func loadConfig(filePath string) *Config {
	var config Config
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	if err = json.NewDecoder(f).Decode(&config); err != nil {
		panic(err)
	}
	return &config
}
