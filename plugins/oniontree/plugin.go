package main

import (
	"github.com/qorpress/qorpress/pkg/models/"

	"github.com/qorpress/qorpress-contrib/oniontree/models"

)

// TwitterOutputPlugin is the Twitter output plugin
type OnionTreePlugin struct{}

// Spec returns plugin spec
//func (o *OnionTreePlugin) Spec() model.Spec {
//	return spec
//}

var Tables = []interface{}{&models.PublicKey{}, &models.Service{}, &models.URL{}, &models.Tag{}}

func Migrate() []interface{} {
	return Tables
}

//func (p *TwitterOutputPlugin) Build(output *model.OutputDef) (model.OutputProvider, error) {
//}

// TwitterOutputProvider output provider to send articles to Twitter
type OnionTreeProvider struct {
	id             int
	alias          string
	//spec           model.Spec
	//condition      *expr.ConditionalExpression
	//formatter      format.Formatter
	enabled        bool
	nbError        uint64
	nbSuccess      uint64
	// consumerKey    string
	// consumerSecret string
	//api            *anaconda.TwitterApi
}

//func (ot *OnionTreeProvider) Send(article *model.Article) error {
//}

//func (ot *OnionTreeProvider) GetDef() model.OutputDef {
//}

func Routes() {

}

func Render() {
}

func Views() {
}

func Api() []interface{} {
	return Tables
}

// GetPluginSpec returns plugin spec
//func GetPluginSpec() model.PluginSpec {
//	return model.PluginSpec{
//		Spec: spec,
//		Type: model.OUTPUT_PLUGIN,
//	}
//}

// GetOutputPlugin returns output plugin
//func GetOutputPlugin() (op model.OutputPlugin, err error) {
//	return &OnionTreePlugin{}, nil
//}

