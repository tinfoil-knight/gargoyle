package server

import (
	"encoding/json"
	"errors"
	"os"
)

var ErrInvalidConfig = errors.New("ERR invalid config")

type Config struct {
	Services []Service
}

type Service struct {
	Source       string          `json:"source"`
	ReverseProxy ReverseProxyCfg `json:"reverse_proxy"`
	Header       HeaderCfg       `json:"header"`
}

type ReverseProxyCfg struct {
	Targets     []string `json:"targets"`
	Algorithm   string   `json:"lb_algorithm"`
	HealthCheck struct {
		Enabled  bool   `json:"enabled"`
		Path     string `json:"path"`
		Interval int    `json:"interval"` // unit: seconds
		Timeout  int    `json:"timeout"`  // unit: seconds
	} `json:"health_check"`
}

type HeaderCfg struct {
	Add    map[string]string `json:"add"`
	Remove []string          `json:"remove"`
}

func loadConfig(filePath string) *Config {
	f, err := os.Open(filePath)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()

	var config Config
	if err = dec.Decode(&config.Services); err != nil {
		panic(err)
	}
	// Setting Defaults
	for _, service := range config.Services {
		rp := service.ReverseProxy
		if len(rp.Targets) > 0 {
			if rp.Algorithm == "" {
				rp.Algorithm = "random"
			}
			if rp.HealthCheck.Enabled {
				if rp.HealthCheck.Interval == 0 {
					panic(ErrInvalidConfig)
				}
				if rp.HealthCheck.Timeout == 0 {
					rp.HealthCheck.Timeout = 5
				}
			}
		}
	}
	return &config
}
