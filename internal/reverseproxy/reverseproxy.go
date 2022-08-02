package reverseproxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/tinfoil-knight/gargoyle/internal/config"
	"github.com/tinfoil-knight/gargoyle/internal/loadbalancer"
)

func NewReverseProxy(service config.ServiceCfg) *http.ServeMux {
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
	return mux
}
