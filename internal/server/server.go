package server

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"

	"github.com/tinfoil-knight/gargoyle/internal/loadbalancer"
)

type Config struct {
	ReverseProxy []struct {
		Source    string   `json:"source"`
		Algorithm string   `json:"algorithm"`
		Targets   []string `json:"targets"`
	} `json:"reverse_proxy"`
}

func NewHTTPServer() {
	config := loadConfig("./config.json")

	var wg sync.WaitGroup

	for _, rp := range config.ReverseProxy {
		wg.Add(1)

		lb, err := loadbalancer.NewLoadBalancer(rp.Algorithm, rp.Targets)
		if err != nil {
			panic(err)
		}

		mux := http.NewServeMux()
		mux.HandleFunc("/", handleAllRequests(lb))

		addr := rp.Source

		go func() {
			defer wg.Done()
			log.Printf("INFO: Started listening on %s", addr)
			log.Fatal(http.ListenAndServe(addr, logHTTPRequest(mux)))
		}()
	}

	wg.Wait()
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
