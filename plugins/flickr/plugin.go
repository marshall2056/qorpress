package main

/*
import (
	"fmt"
	"strings"
	"sync/atomic"
	"github.com/ncarlier/feedpushr/v2/pkg/expr"
	"github.com/ncarlier/feedpushr/v2/pkg/model"
)

var spec = model.Spec{
	Name: "telegram-bot",
	Desc: "Send new articles to a telegram bot.",
	PropsSpec: []model.PropSpec{
		{
			Name: "telegramBot",
			Desc: "Telegram bot name",
			Type: model.Text,
		},
		{
			Name: "telegramApiKey",
			Desc: "Telegram API Key",
			Type: model.Password,
		},
	},
}

// TelegramOutputPlugin is the Twitter output plugin
type TelegramOutputPlugin struct{}

// Spec returns plugin spec
func (p *TelegramOutputPlugin) Spec() model.Spec {
	return spec
}

// Build creates Twitter output provider instance
func (p *TelegramOutputPlugin) Build(output *model.OutputDef) (model.OutputProvider, error) {
	condition, err := expr.NewConditionalExpression(output.Condition)
	if err != nil {
		return nil, err
	}
	telegramApiKey := output.Props.Get("telegramApiKey")
	if telegramApiKey == "" {
		return nil, fmt.Errorf("missing telegram api key property")
	}
	telegramBot := output.Props.Get("telegramBot")
	if telegramBot == "" {
		return nil, fmt.Errorf("missing telegram bot name property")
	}

	bot, err := tgbotapi.NewBotAPI(telegramApiKey)
	bot.Debug = false

	return &TelegramOutputProvider{
		id:             output.ID,
		alias:          output.Alias,
		spec:           spec,
		condition:      condition,
		enabled:        output.Enabled,
		telegramApiKey: telegramApiKey,
		telegramBot:    telegramBot,
		bot:            bot,
	}, nil
}

// TwitterOutputProvider output provider to send articles to Twitter
type TelegramOutputProvider struct {
	id             int
	alias          string
	spec           model.Spec
	condition      *expr.ConditionalExpression
	enabled        bool
	nbError        uint64
	nbSuccess      uint64
	telegramApiKey string
	telegramBot    string
	bot            *tgbotapi.BotAPI
}

// Send sent an article as Tweet to a Twitter timeline
func (op *TelegramOutputProvider) Send(article *model.Article) error {
	if !op.enabled || !op.condition.Match(article) {
		// Ignore if disabled or if the article doesn't match the condition
		return nil
	}
	msg := tgbotapi.NewMessage(795118556, article.Link)
	_, err := op.bot.Send(msg)
	if err != nil {
		fmt.Println("err.Error(): ", err.Error())
		// Ignore error due to duplicate status
		if strings.Contains(err.Error(), "\"code\":187") {
			return nil
		}
		atomic.AddUint64(&op.nbError, 1)
	} else {
		atomic.AddUint64(&op.nbSuccess, 1)
	}
	return err
}

// GetDef return filter definition
func (op *TelegramOutputProvider) GetDef() model.OutputDef {
	result := model.OutputDef{
		ID:        op.id,
		Alias:     op.alias,
		Spec:      op.spec,
		Condition: op.condition.String(),
		Enabled:   op.enabled,
	}
	result.Props = map[string]interface{}{
		"telegramApiKey": op.telegramApiKey,
		"telegramBot":    op.telegramBot,
		"nbError":        op.nbError,
		"nbSuccess":      op.nbSuccess,
	}
	return result
}

// GetPluginSpec returns plugin spec
func GetPluginSpec() model.PluginSpec {
	return model.PluginSpec{
		Spec: spec,
		Type: model.OUTPUT_PLUGIN,
	}
}

// GetOutputPlugin returns output plugin
func GetOutputPlugin() (op model.OutputPlugin, err error) {
	return &TelegramOutputPlugin{}, nil
}
*/
