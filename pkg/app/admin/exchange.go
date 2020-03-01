package admin

import (
	"github.com/qorpress/exchange"
	"github.com/qorpress/qor"
	"github.com/qorpress/qor/resource"
	"github.com/qorpress/qor/utils"
	"github.com/qorpress/validations"

	"github.com/qorpress/qorpress-example/pkg/models/posts"
)

// PostExchange post exchange
var PostExchange = exchange.NewResource(&posts.Post{}, exchange.Config{PrimaryField: "Code"})

func init() {
	PostExchange.Meta(&exchange.Meta{Name: "Code"})
	PostExchange.Meta(&exchange.Meta{Name: "Name"})
	PostExchange.Meta(&exchange.Meta{Name: "Price"})

	PostExchange.AddValidator(&resource.Validator{
		Handler: func(record interface{}, metaValues *resource.MetaValues, context *qor.Context) error {
			if utils.ToInt(metaValues.Get("Price").Value) < 100 {
				return validations.NewError(record, "Price", "price can't less than 100")
			}
			return nil
		},
	})
}
