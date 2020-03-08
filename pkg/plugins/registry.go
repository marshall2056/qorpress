package plugins

import (
	"context"
)

type QorPlugin struct {
	Ctx      context.Context
	Commands map[string]Plugin
	Closed   chan struct{}
}

func New() *QorPlugin {
	return &QorPlugin{
		// pluginsDir: plug.PluginsDir,
		Ctx:      context.Background(),
		Commands: make(map[string]Plugin),
		Closed:   make(chan struct{}),
	}
}
