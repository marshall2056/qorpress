package main

import (
	"context"
	"fmt"

	// move to core
	"github.com/qorpress/qorpress-contrib/twitter/controllers"
	"github.com/qorpress/qorpress-contrib/twitter/models"
	"github.com/qorpress/qorpress-contrib/twitter/utils/funcmapmaker"
	"github.com/qorpress/qorpress/core/render"
	"github.com/qorpress/qorpress/pkg/config/application"
	plug "github.com/qorpress/qorpress/pkg/plugins"
)

var Tables = []interface{}{
	&models.TwitterSetting{},
}

var Resources = []interface{}{
	&models.TwitterSetting{},
}

type twitterPlugin string

func (o twitterPlugin) Name() string      { return string(o) }
func (o twitterPlugin) Section() string   { return `Twitter` }
func (o twitterPlugin) Usage() string     { return `hello` }
func (o twitterPlugin) ShortDesc() string { return `prints greeting "hello there"` }
func (o twitterPlugin) LongDesc() string  { return o.ShortDesc() }

func (o twitterPlugin) Migrate() []interface{} {
	return Tables
}

func (o twitterPlugin) Resources() []interface{} {
	return Resources
}

func (o twitterPlugin) Application() application.MicroAppInterface {
	return controllers.New(&controllers.Config{})
}

func (o twitterPlugin) FuncMapMaker(view *render.Render) *render.Render {
	return funcmapmaker.AddFuncMapMaker(view)
}

type twitterCommands struct{}

func (t *twitterCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
--------------------------------------------------------------------------------------------
'########:'##:::::'##:'####:'########:'########:'########:'########::::::::::::::'###::::'########::'####:
... ##..:: ##:'##: ##:. ##::... ##..::... ##..:: ##.....:: ##.... ##::::::::::::'## ##::: ##.... ##:. ##::
::: ##:::: ##: ##: ##:: ##::::: ##::::::: ##:::: ##::::::: ##:::: ##:::::::::::'##:. ##:: ##:::: ##:: ##::
::: ##:::: ##: ##: ##:: ##::::: ##::::::: ##:::: ######::: ########::'#######:'##:::. ##: ########::: ##::
::: ##:::: ##: ##: ##:: ##::::: ##::::::: ##:::: ##...:::: ##.. ##:::........: #########: ##.....:::: ##::
::: ##:::: ##: ##: ##:: ##::::: ##::::::: ##:::: ##::::::: ##::. ##::::::::::: ##.... ##: ##::::::::: ##::
::: ##::::. ###. ###::'####:::: ##::::::: ##:::: ########: ##:::. ##:::::::::: ##:::: ##: ##::::::::'####:
:::..::::::...::...:::....:::::..::::::::..:::::........::..:::::..:::::::::::..:::::..::..:::::::::....::
`)

	return nil
}

func (t *twitterCommands) Registry() map[string]plug.Plugin {
	return map[string]plug.Plugin{
		"twitter": twitterPlugin("twitter"), //OP
	}
}

var Plugins twitterCommands
