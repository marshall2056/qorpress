package services

import (
	"path/filepath"

	"github.com/qorpress/i18n"
	"github.com/qorpress/i18n/backends/database"
	"github.com/qorpress/i18n/backends/yaml"

	"github.com/qorpress/qorpress/pkg/config"
)

var I18n *i18n.I18n

func init() {
	if DB == nil {
		DB = Init()
	}
	I18n = i18n.New(database.New(DB), yaml.New(filepath.Join(config.Root, ".config/locales")))
}
