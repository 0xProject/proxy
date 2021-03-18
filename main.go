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

	// The name of the header to add to requests
	HeaderName string `env:"HEADER_NAME"`

	// The value of the header to add to requests
	HeaderValue string `env:"HEADER_VALUE"`

	// Optional response timeout in seconds for proxy requests
	ResponseTimeout time.Duration `env:"RESPONSE_TIMEOUT" envDefault:"30s"`

	// LogLevel is the logging verbosity: 0=panic, 1=fatal, 2=error, 3=warn, 4=info, 5=debug 6=trace
	LogLevel int `env:"LOG_LEVEL" envDefault:"5"`
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	cfg := ProxyConfig{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("could not parse config: %s", err)
	}
	cacheCfg := CacheConfig{}
	if err := env.Parse(&cacheCfg); err != nil {
		log.Fatalf("could not parse cache config: %s", err)
	}

	log.SetLevel(log.Level(cfg.LogLevel))
	log.WithFields(log.Fields{
		"Port":            cfg.Port,
		"TargetURL":       cfg.TargetURL,
		"QueryParamName":  cfg.QueryParamName,
		"QueryParamValue": cfg.QueryParamValue,
		"CacheExpiration": cacheCfg.CacheExpiration,
		"HeaderName":      cfg.HeaderName,
		"HeaderValue":     cfg.HeaderValue,
	}).Info("parsed config successfully")

	var proxy http.Handler
	var err error
	if cacheCfg.CacheExpiration.Seconds() == 0 {
		log.Info("configuring standard proxy")
		proxy, err = NewProxy(&cfg)
		if err != nil {
			log.Fatalf("failed to create a new proxy: %s", err)
		}
	} else {
		log.Info("configuring cached proxy")
		proxy, err = NewCachedProxy(&cfg, &cacheCfg)
		if err != nil {
			log.Fatalf("failed to create a new cached proxy: %s", err)
		}
	}

	listenString := fmt.Sprintf(":%d", cfg.Port)
	log.Infof("starting proxy on %s", listenString)
	if err := http.ListenAndServe(listenString, proxy); err != nil {
		log.Panic(err)
	}
}
