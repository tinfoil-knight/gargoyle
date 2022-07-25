package server

import (
	"encoding/json"
	"errors"
	"os"
)

var ErrInvalidConfig = errors.New("ERR invalid config")

type Config struct {
	ReverseProxies []ReverseProxy `json:"reverse_proxy"`
}

type ReverseProxy struct {
	Source      string   `json:"source"`
	Algorithm   string   `json:"algorithm"`
	Targets     []string `json:"targets"`
	HealthCheck struct {
		Enabled  bool   `json:"enabled"`
		Path     string `json:"path"`
		Interval int    `json:"interval"` // unit: seconds
		Timeout  int    `json:"timeout"`  // unit: seconds
	} `json:"health_check"`
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
	if err = dec.Decode(&config); err != nil {
		panic(err)
	}
	// Setting Defaults
	for _, rp := range config.ReverseProxies {
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
	return &config
}
