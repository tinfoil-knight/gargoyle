package loadbalancer

import (
	"math/rand"
	"net/http/httputil"
	"net/url"
)

type LoadBalancer struct {
	services []*Service
}

type Service struct {
	url   *url.URL
	proxy *httputil.ReverseProxy
}

func NewLoadBalancer(targetUrls []string) (*LoadBalancer, error) {
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
		services: services,
	}, nil
}

func (lb *LoadBalancer) GetSelectedProxy() *httputil.ReverseProxy {
	target := lb.selectTarget()
	return target.proxy
}

func (lb *LoadBalancer) selectTarget() *Service {
	return lb.services[rand.Int()%len(lb.services)]
}
