package azurestorage

import (
	"github.com/jinzhu/configor"
	"github.com/k0kubun/pp"
	"github.com/sniperkit/config"
	"github.com/sniperkit/httpcache/constants"
	"github.com/sniperkit/vipertags"
)

type azurestorageConfig struct {
	Provider       string        `json:"provider" config:"database.provider"`
	Endpoints      []string      `json:"endpoints" config:"database.endpoints"`
	MaxConnections int           `json:"max_connections" config:"database.max_connections" default:"0"`
	BucketName     string        `json:"bucket_name" yaml:"bucket_name" config:"cache.http.provider"`
	StoragePath    string        `json:"storage_path" yaml:"storage_path" config:"cache.http.storage_path"`
	done           chan struct{} `json:"-" config:"-"`
}

// Config ...
var (
	PluginConfig = &azurestorageConfig{
		done: make(chan struct{}),
	}
)

// ConfigName ...
func (azurestorageConfig) ConfigName() string {
	return "AzureStorage"
}

// SetDefaults ...
func (a *azurestorageConfig) SetDefaults() {
	vipertags.SetDefaults(a)
}

// Read ...
func (a *azurestorageConfig) Read() {
	defer close(a.done)
	vipertags.Fill(a)
	if a.Provider == "" {
		a.Provider = a.ConfigName()
	}
	if a.MaxConnections == 0 {
		a.MaxConnections = constants.DefaultMaxConnections
	}
}

// Read several config files (yaml, json or env variables)
func (a *azurestorageConfig) Configor(files []string) {
	configor.Load(&PluginConfig, files...)
}

// Wait ...
func (c azurestorageConfig) Wait() {
	<-c.done
}

// String ...
func (c azurestorageConfig) String() string {
	return pp.Sprintln(c)
}

// Debug ...
func (c azurestorageConfig) Debug() {
	// log.Debug("AzureStorage PluginConfig = ", c)
}

func init() {
	config.Register(PluginConfig)
}
