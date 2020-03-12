package main

import (
	"context"
	"fmt"

	// move to core
	"github.com/qorpress/qorpress-contrib/flickr/controllers"
	"github.com/qorpress/qorpress-contrib/flickr/models"
	"github.com/qorpress/qorpress-contrib/flickr/utils/funcmapmaker"
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

type flickrPlugin string

func (o flickrPlugin) Name() string      { return string(o) }
func (o flickrPlugin) Section() string   { return `Flickr` }
func (o flickrPlugin) Usage() string     { return `hello` }
func (o flickrPlugin) ShortDesc() string { return `prints greeting "hello there"` }
func (o flickrPlugin) LongDesc() string  { return o.ShortDesc() }

func (o flickrPlugin) Migrate() []interface{} {
	return Tables
}

func (o flickrPlugin) Resources() []interface{} {
	return Resources
}

func (o flickrPlugin) Application() application.MicroAppInterface {
	return controllers.New(&controllers.Config{})
}

func (o flickrPlugin) FuncMapMaker(view *render.Render) *render.Render {
	return funcmapmaker.AddFuncMapMaker(view)
}

type flickrCommands struct{}

func (t *flickrCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
---------------------------------------------------------------------------------------------
'########:'##:::::::'####::'######::'##:::'##:'########::::::::::::::'###::::'########::'####:
 ##.....:: ##:::::::. ##::'##... ##: ##::'##:: ##.... ##::::::::::::'## ##::: ##.... ##:. ##::
 ##::::::: ##:::::::: ##:: ##:::..:: ##:'##::: ##:::: ##:::::::::::'##:. ##:: ##:::: ##:: ##::
 ######::: ##:::::::: ##:: ##::::::: #####:::: ########::'#######:'##:::. ##: ########::: ##::
 ##...:::: ##:::::::: ##:: ##::::::: ##. ##::: ##.. ##:::........: #########: ##.....:::: ##::
 ##::::::: ##:::::::: ##:: ##::: ##: ##:. ##:: ##::. ##::::::::::: ##.... ##: ##::::::::: ##::
 ##::::::: ########:'####:. ######:: ##::. ##: ##:::. ##:::::::::: ##:::: ##: ##::::::::'####:
..::::::::........::....:::......:::..::::..::..:::::..:::::::::::..:::::..::..:::::::::....::
`)

	return nil
}

func (t *flickrCommands) Registry() map[string]plug.Plugin {
	return map[string]plug.Plugin{
		"flickr": flickrPlugin("flickr"), //OP
	}
}

var Plugins flickrCommands
