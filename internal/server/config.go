package server

import (
	"encoding/json"
	"os"
)

type Config struct {
	ReverseProxies []ReverseProxy `json:"reverse_proxy"`
}

type ReverseProxy struct {
	Source    string   `json:"source"`
	Algorithm string   `json:"algorithm"`
	Targets   []string `json:"targets"`
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
	}
	return &config
}
