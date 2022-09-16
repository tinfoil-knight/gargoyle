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
	Source  string `json:"source"`
	Timeout struct {
		// unit: seconds, default: 0
		Read int `json:"read"`
		// unit: seconds, default: 0
		Write int `json:"write"`
		// unit: seconds, default: 0
		Idle int `json:"idle"`
	} `json:"timeout"`
	ReverseProxy *ReverseProxyCfg `json:"reverse_proxy"`
	Header       *HeaderCfg       `json:"header"`
	Rewrite      *RewriteCfg      `json:"rewrite"`
	Fs           *FsConfig        `json:"fs"`
	Auth         *AuthConfig      `json:"auth"`
	TLS          *struct {
		// default: false
		Enabled bool `json:"enabled"`
		// default: ./.gargoyle/cert.pem
		CertPath string `json:"cert_path"`
		// default: ./.gargoyle/key.pem
		KeyPath string `json:"key_path"`
	} `json:"tls"`
}

type ReverseProxyCfg struct {
	Targets []string `json:"targets"`
	// default: random
	Algorithm   string `json:"lb_algorithm"`
	HealthCheck struct {
		// default: false
		Enabled bool   `json:"enabled"`
		Path    string `json:"path"`
		// unit: seconds
		Interval int `json:"interval"`
		// unit: seconds, default: 5
		Timeout int `json:"timeout"`
	} `json:"health_check"`
}

type HeaderCfg struct {
	Add    map[string]string `json:"add"`
	Remove []string          `json:"remove"`
}

type RewriteCfg map[string]string // from : to

type FsConfig struct {
	Path string `json:"path"`
}

type AuthConfig struct {
	KeyAuth *struct {
		Key string `json:"key"`
		// default: X-Api-Key
		Header string `json:"header"`
	} `json:"key_auth"`
	BasicAuth map[string]([]byte) `json:"basic_auth"` // map[Username][PasswordHash]
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
		if err := checkForConflicts([]interface{}{service.ReverseProxy, service.Fs}); err != nil {
			panic(err)
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

		if service.Auth != nil {
			auth := service.Auth
			if err := checkForConflicts([]interface{}{auth.BasicAuth, auth.KeyAuth}); err != nil {
				panic(err)
			}

			if len(auth.BasicAuth) != 0 {
				for username, hash := range auth.BasicAuth {
					if username == "" || len(hash) == 0 {
						panic(ErrInvalidConfig)
					}
				}
			}

			if auth.KeyAuth != nil {
				if auth.KeyAuth.Key == "" {
					panic(ErrInvalidConfig)
				}
				if auth.KeyAuth.Header == "" {
					auth.KeyAuth.Header = "X-Api-Key"
				}
			}
		}

		if service.TLS != nil && service.TLS.Enabled {
			tlsConfig := service.TLS
			if tlsConfig.CertPath == "" {
				tlsConfig.CertPath = "./.gargoyle/cert.pem"
			}
			if tlsConfig.KeyPath == "" {
				tlsConfig.KeyPath = "./.gargoyle/key.pem"
			}
			// TODO: check if files exists
		}
	}
	return &config
}

func checkForConflicts(conflicts []interface{}) error {
	nonNil := 0
	for _, item := range conflicts {
		if !reflect.ValueOf(item).IsNil() {
			nonNil++
			if nonNil > 1 {
				return ErrInvalidConfig
			}
		}
	}
	return nil
}
