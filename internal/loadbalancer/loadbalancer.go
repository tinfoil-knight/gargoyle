package loadbalancer

import (
	"errors"
	"math/rand"
	"net/http/httputil"
	"net/url"
)

var ErrInvalidAlgorithm = errors.New("ERR invalid load balancing algorithm provided")

type LoadBalancer struct {
	services []*Service
	// lastIndex stores index of the last selected target for some load balancing algorithms
	lastIndex int
	// possible values: "random", "round-robin"
	algorithm string
}

type Service struct {
	url   *url.URL
	proxy *httputil.ReverseProxy
}

func NewLoadBalancer(algorithm string, targetUrls []string) (*LoadBalancer, error) {
	services := make([]*Service, len(targetUrls))
	for idx, targetUrl := range targetUrls {
		url, err := url.Parse(targetUrl)
		if err != nil {
			return nil, err
		}
		proxy := httputil.NewSingleHostReverseProxy(url)
		services[idx] = &Service{url: url, proxy: proxy}
	}
	return &LoadBalancer{
		services:  services,
		algorithm: algorithm,
	}, nil
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
