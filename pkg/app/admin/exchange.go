package admin

import (
	"github.com/qorpress/exchange"
	"github.com/qorpress/qor"
	"github.com/qorpress/qor/resource"

	"github.com/qorpress/qorpress/pkg/models/posts"
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
