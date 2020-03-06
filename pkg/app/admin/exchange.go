package admin

import (
	"github.com/gopress/internal/exchange"
	"github.com/gopress/internal/qor"
	"github.com/gopress/internal/qor/resource"

	"github.com/gopress/qorpress/pkg/models/posts"
)

// PostExchange post exchange
var PostExchange = exchange.NewResource(&posts.Post{}, exchange.Config{PrimaryField: "Code"})

func init() {
	PostExchange.Meta(&exchange.Meta{Name: "Code"})
	PostExchange.Meta(&exchange.Meta{Name: "Name"})

	PostExchange.AddValidator(&resource.Validator{
		Handler: func(record interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
			return nil
		},
	})
}
