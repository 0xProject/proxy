package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/caarlos0/env/v6"

	log "github.com/sirupsen/logrus"
)

type CacheConfig struct {
	CacheExpiration time.Duration `env:"CACHE_EXPIRATION" envDefault:"2m"`
}

type ProxyConfig struct {
	// Port on which the proxy is listening
	Port int `env:"PORT" envDefault:"3000"`

	// Proxy target
	TargetURL string `env:"TARGET_URL"`

	// The name of the query param to append to requests
	QueryParamName string `env:"QUERY_PARAM_NAME"`

	// The value of the query param to append to requests
	QueryParamValue string `env:"QUERY_PARAM_VALUE"`
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	cfg := ProxyConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("could not parse config: %s", err)
	}

	log.WithFields(log.Fields{
		"Port":            cfg.Port,
		"TargetURL":       cfg.TargetURL,
		"QueryParamName":  cfg.QueryParamName,
		"QueryParamValue": cfg.QueryParamValue,
	}).Info("parsed config successfully")

	proxy, err := NewProxy(&cfg)
	if err != nil {
		log.Fatalf("failed to create a new proxy: %s", err)
	}

	listenString := fmt.Sprintf(":%d", cfg.Port)
	log.Infof("starting proxy on %s", listenString)
	if err := http.ListenAndServe(listenString, proxy); err != nil {
		log.Panic(err)
	}
}
