package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"reflect"
)

var ErrInvalidConfig = errors.New("ERR invalid config")

type Config struct {
	Services []ServiceCfg
}

type ServiceCfg struct {
	Source       string           `json:"source"`
	ReverseProxy *ReverseProxyCfg `json:"reverse_proxy"`
	Header       *HeaderCfg       `json:"header"`
	Fs           *FsConfig        `json:"fs"`
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

type FsConfig struct {
	Path string `json:"path"`
}

func LoadConfig(filePath string) *Config {
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

	for _, service := range config.Services {
		// Check for conflicts
		conflicts := [][]interface{}{{service.ReverseProxy, service.Fs}}
		for _, list := range conflicts {
			nonNil := 0
			for _, item := range list {
				if !reflect.ValueOf(item).IsNil() {
					nonNil++
					if nonNil > 1 {
						panic(ErrInvalidConfig)
					}
				}
			}
		}

		// Validating & Setting Defaults

		if service.ReverseProxy != nil {
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
			} else {
				panic(ErrInvalidConfig)
			}
		}

		if service.Header != nil {
			header := service.Header
			for _, v := range header.Remove {
				if _, ok := header.Add[v]; ok {
					panic(ErrInvalidConfig)
				}
			}
		}

		if service.Fs != nil {
			dirPath := service.Fs.Path
			info, err := os.Stat(dirPath)
			dirExists := !errors.Is(err, fs.ErrNotExist) && info.IsDir()
			if !dirExists {
				panic(err)
			}
		}
	}
	return &config
}
