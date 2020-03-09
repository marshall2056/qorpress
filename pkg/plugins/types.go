package plugins

import (
	"context"

	"github.com/qorpress/qorpress/core/render"
	"github.com/qorpress/qorpress/pkg/config/application"
)

// Module a plugin that can be initialized
type Module interface {
	Init(context.Context) error
}

type Plugin interface {
	Name() string
	Usage() string
	Section() string
	ShortDesc() string
	LongDesc() string
	Migrate() []interface{}
	Resources() []interface{}
	Application() application.MicroAppInterface
	FuncMapMaker(view *render.Render) *render.Render
}

// Plugins a plugin that contains one or more command
type Plugins interface {
	Module
	Registry() map[string]Plugin
}
