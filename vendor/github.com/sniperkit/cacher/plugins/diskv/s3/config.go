package diskvs3cache

import (
	"github.com/jinzhu/configor"
	"github.com/k0kubun/pp"
	"github.com/sniperkit/config"
	"github.com/sniperkit/httpcache/constants"
	"github.com/sniperkit/vipertags"
)

type diskvs3cacheConfig struct {
	Provider       string        `json:"provider" yaml:"provider" config:"http.cache.provider"`
	BasePath       string        `json:"base_path" yaml:"base_path" config:"http.cache.base_path"`
	CacheSizeMax   uint64        `json:"cache_size_max" yaml:"cache_size_max" config:"http.cache.cache_size_max"`
	Transform      string        `json:"transform" yaml:"transform" config:"http.cache.transform"`
	MaxConnections int           `json:"max_connections" yaml:"max_connections" config:"http.cache.max_connections" default:"0"`
	done           chan struct{} `json:"-" yaml:"-" config:"-"`
}

// Config ...
var (
	Config = &diskvs3cacheConfig{
		done: make(chan struct{}),
	}
)

// ConfigName ...
func (diskvs3cacheConfig) ConfigName() string {
	return "DiskvS3"
}

// SetDefaults ...
func (a *diskvs3cacheConfig) SetDefaults() {
	vipertags.SetDefaults(a)
}

// Read ...
func (a *diskvs3cacheConfig) Read() {
	defer close(a.done)
	vipertags.Fill(a)
	if a.MaxConnections == 0 {
		a.MaxConnections = constants.DefaultMaxConnections
	}
}

// Read several config files (yaml, json or env variables)
func (a *diskvs3cacheConfig) Configor(files []string) {
	configor.Load(&Config, filepath...)
}

// Wait ...
func (c diskvs3cacheConfig) Wait() {
	<-c.done
}

// String ...
func (c diskvs3cacheConfig) String() string {
	return pp.Sprintln(c)
}

// Debug ...
func (c diskvs3cacheConfig) Debug() {
	// log.Debug("DiskvS3 Config = ", c)
}

func init() {
	config.Register(Config)
}
