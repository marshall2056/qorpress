package i18n

import (
	"path/filepath"

	"github.com/qorpress/qorpress/core/i18n"
	"github.com/qorpress/qorpress/core/i18n/backends/database"
	"github.com/qorpress/qorpress/core/i18n/backends/yaml"
	"github.com/qorpress/qorpress/pkg/config"
	"github.com/qorpress/qorpress/pkg/config/db"
)

var I18n *i18n.I18n

func init() {
	localesDir := filepath.Join(config.Root, "themes", "qorpress", "locales")
	// check if exists
	I18n = i18n.New(database.New(db.DB), yaml.New(localesDir))
}
