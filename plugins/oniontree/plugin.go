package main

import (
	"context"
	"fmt"

	plug "github.com/qorpress/qorpress/pkg/plugins"

	"github.com/qorpress/qorpress-contrib/oniontree/models"
)

var Tables = []interface{}{
	&models.PublicKey{}, 
	&models.Service{}, 
	&models.URL{},
	&models.Tag{},
}

var Resources = []interface{}{
	&models.Service{}, 
	&models.Tag{},
}

type onionTreePlugin string

func (o onionTreePlugin) Name() string      { return string(o) }
func (o onionTreePlugin) Section() string      { return `OnionTree` }
func (o onionTreePlugin) Usage() string     { return `hello` }
func (o onionTreePlugin) ShortDesc() string { return `prints greeting "hello there"` }
func (o onionTreePlugin) LongDesc() string  { return o.ShortDesc() }
func (o onionTreePlugin) Migrate() []interface{} {
	return Tables
}

func (o onionTreePlugin) Resources() []interface{} {
	return Resources
}

/*
func (o onionTreePlugin) Routes() map[string]http.HandlerFunc {
	h := make(map[string]http.Handler, 0)
	h["/test"] = 
	return h
}
*/

type onionTreeCommands struct{}

func (t *onionTreeCommands) Init(ctx context.Context) error {
	// to set your splash, modify the text in the println statement below, multiline is supported
	fmt.Println(`
:'#######:::'#######::'########::'########::'########::'########::'######:::'######::
'##.... ##:'##.... ##: ##.... ##: ##.... ##: ##.... ##: ##.....::'##... ##:'##... ##:
 ##:::: ##: ##:::: ##: ##:::: ##: ##:::: ##: ##:::: ##: ##::::::: ##:::..:: ##:::..::
 ##:::: ##: ##:::: ##: ########:: ########:: ########:: ######:::. ######::. ######::
 ##:'## ##: ##:::: ##: ##.. ##::: ##.....::: ##.. ##::: ##...:::::..... ##::..... ##:
 ##:.. ##:: ##:::: ##: ##::. ##:: ##:::::::: ##::. ##:: ##:::::::'##::: ##:'##::: ##:
: ##### ##:. #######:: ##:::. ##: ##:::::::: ##:::. ##: ########:. ######::. ######::
:.....:..:::.......:::..:::::..::..:::::::::..:::::..::........:::......::::......:::
--------------------------------------------------------------------------------------------
:'#######::'##::: ##:'####::'#######::'##::: ##::::'########:'########::'########:'########:
'##.... ##: ###:: ##:. ##::'##.... ##: ###:: ##::::... ##..:: ##.... ##: ##.....:: ##.....::
 ##:::: ##: ####: ##:: ##:: ##:::: ##: ####: ##::::::: ##:::: ##:::: ##: ##::::::: ##:::::::
 ##:::: ##: ## ## ##:: ##:: ##:::: ##: ## ## ##::::::: ##:::: ########:: ######::: ######:::
 ##:::: ##: ##. ####:: ##:: ##:::: ##: ##. ####::::::: ##:::: ##.. ##::: ##...:::: ##...::::
 ##:::: ##: ##:. ###:: ##:: ##:::: ##: ##:. ###::::::: ##:::: ##::. ##:: ##::::::: ##:::::::
. #######:: ##::. ##:'####:. #######:: ##::. ##::::::: ##:::: ##:::. ##: ########: ########:
:.......:::..::::..::....:::.......:::..::::..::::::::..:::::..:::::..::........::........::
`)

	return nil
}

func (t *onionTreeCommands) Registry() map[string]plug.Plugin {
	return map[string]plug.Plugin{
		"oniontree": onionTreePlugin("oniontree"), //OP
	}
}

var Plugins onionTreeCommands