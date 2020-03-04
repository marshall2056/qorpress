package diskvcache

import (
	"github.com/jinzhu/configor"
	"github.com/k0kubun/pp"
	"github.com/sniperkit/config"
	"github.com/sniperkit/httpcache/constants"
	"github.com/sniperkit/vipertags"
)

type diskvcacheConfig struct {
	Provider       string        `json:"provider" yaml:"provider" config:"http.cache.provider"`
	BasePath       string        `json:"base_path" yaml:"base_path" config:"http.cache.base_path"`
	CacheSizeMax   uint64        `json:"cache_size_max" yaml:"cache_size_max" config:"http.cache.cache_size_max"`
	Transform      string        `json:"transform" yaml:"transform" config:"http.cache.transform"`
	MaxConnections int           `json:"max_connections" yaml:"max_connections" config:"http.cache.max_connections" default:"0"`
	done           chan struct{} `json:"-" yaml:"-" config:"-"`
}

// Config ...
var (
	PluginConfig = &diskvcacheConfig{
		done: make(chan struct{}),
	}
)

// ConfigName ...
func (diskvcacheConfig) ConfigName() string {
	return "DiskV"
}

// SetDefaults ...
func (a *diskvcacheConfig) SetDefaults() {
	vipertags.SetDefaults(a)
}

// Read ...
func (a *diskvcacheConfig) Read() {
	defer close(a.done)
	vipertags.Fill(a)
	if a.MaxConnections == 0 {
		a.MaxConnections = constants.DefaultMaxConnections
	}
}

// Read several config files (yaml, json or env variables)
func (a *diskvcacheConfig) Configor(files []string) {
	configor.Load(&PluginConfig, files...)
}

// Wait ...
func (c diskvcacheConfig) Wait() {
	<-c.done
}

// String ...
func (c diskvcacheConfig) String() string {
	return pp.Sprintln(c)
}

// Debug ...
func (c diskvcacheConfig) Debug() {
	// log.Debug("DiskV PluginConfig = ", c)
}

func init() {
	config.Register(PluginConfig)
}
