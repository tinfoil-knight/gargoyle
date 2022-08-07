package loadbalancer

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"syscall"
	"time"
)

var ErrInvalidAlgorithm = errors.New("ERR invalid load balancing algorithm provided")

type LoadBalancer struct {
	services        []*Service
	activeServices  []*Service
	healthCheckTick *time.Ticker
	// lastIndex stores index of the last selected target for some load balancing algorithms
	lastIndex int
	// possible values: "random", "round-robin"
	algorithm string
}

type Service struct {
	url     *url.URL
	proxy   *httputil.ReverseProxy
	healthy bool
}

func NewLoadBalancer(algorithm string, targetUrls []string) (*LoadBalancer, error) {
	services := make([]*Service, len(targetUrls))
	for idx, targetUrl := range targetUrls {
		url, err := url.Parse(targetUrl)
		if err != nil {
			return nil, err
		}
		proxy := httputil.NewSingleHostReverseProxy(url)
		services[idx] = &Service{url: url, proxy: proxy, healthy: false}
	}
	return &LoadBalancer{
		services:       services,
		activeServices: services,
		algorithm:      algorithm,
	}, nil
}

func (lb *LoadBalancer) RunHealthChecks(interval time.Duration, timeout time.Duration, path string) {
	ticker := time.NewTicker(interval)
	lb.healthCheckTick = ticker
	client := http.Client{
		Timeout: timeout,
	}
	for ; true; <-ticker.C {
		var wg sync.WaitGroup
		// TODO: inspect thread safety here
		for _, service := range lb.services {
			service := service
			wg.Add(1)

			go func() {
				defer wg.Done()
				url := fmt.Sprintf("%s%s", service.url.String(), path)
				res, err := client.Get(url)
				if err != nil {
					if errors.Is(err, syscall.ECONNREFUSED) || os.IsTimeout(err) {
						service.healthy = false
						return
					}
					// TODO: find more client errors which can occur
					panic(err)
				}
				service.healthy = res.StatusCode == http.StatusOK
			}()
		}

		wg.Wait()

		var activeServices []*Service

		for _, service := range lb.services {
			if service.healthy {
				activeServices = append(activeServices, service)
			}
		}

		lb.activeServices = activeServices
	}
}

func (lb *LoadBalancer) StopHealthChecks() {
	lb.healthCheckTick.Stop()
}

func (lb *LoadBalancer) GetSelectedProxy() *httputil.ReverseProxy {
	var idx int
	switch lb.algorithm {
	case "round-robin":
		idx = (lb.lastIndex + 1) % len(lb.services)
		lb.lastIndex = idx
	case "random":
		idx = rand.Int() % len(lb.services)
	default:
		panic(ErrInvalidAlgorithm)
	}
	target := lb.services[idx]
	return target.proxy
}
