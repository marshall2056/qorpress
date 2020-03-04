package seaweedfscache

/*
import (
	"github.com/k0kubun/pp"
	"github.com/sniperkit/config"
	"github.com/sniperkit/vipertags"
	"github.com/roscopecoltran/database"
)

type seaweedfscacheConfig struct {
	Provider       string        `json:"provider" config:"database.provider"`
	Endpoints      []string      `json:"endpoints" config:"database.endpoints"`
	MaxConnections int           `json:"max_connections" config:"database.max_connections" default:"0"`
	done           chan struct{} `json:"-" config:"-"`
}

// Config ...
var (
	Config = &seaweedfscacheConfig{
		done: make(chan struct{}),
	}
)

// ConfigName ...
func (seaweedfscacheConfig) ConfigName() string {
	return "BBoltDB"
}

// SetDefaults ...
func (a *seaweedfscacheConfig) SetDefaults() {
	vipertags.SetDefaults(a)
}

// Read ...
func (a *seaweedfscacheConfig) Read() {
	defer close(a.done)
	vipertags.Fill(a)
	if a.MaxConnections == 0 {
		a.MaxConnections = database.DefaultMaxConnections
	}
}

// Wait ...
func (c seaweedfscacheConfig) Wait() {
	<-c.done
}

// String ...
func (c seaweedfscacheConfig) String() string {
	return pp.Sprintln(c)
}

// Debug ...
func (c seaweedfscacheConfig) Debug() {
	log.Debug("BBoltDB Config = ", c)
}

func init() {
	config.Register(Config)
}
*/
