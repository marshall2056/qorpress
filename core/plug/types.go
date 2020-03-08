package plug

import (
	"context"
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
}

// Plugins a plugin that contains one or more command
type Plugins interface {
	Module
	Registry() map[string]Plugin
}
