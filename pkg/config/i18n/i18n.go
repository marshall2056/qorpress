package i18n

import (
	"path/filepath"

	"github.com/qorpress/i18n"
	"github.com/qorpress/i18n/backends/database"
	"github.com/qorpress/i18n/backends/yaml"

	"github.com/qorpress/qorpress-example/pkg/config"
	"github.com/qorpress/qorpress-example/pkg/config/db"
)

var I18n *i18n.I18n

func init() {
	I18n = i18n.New(database.New(db.DB), yaml.New(filepath.Join(config.Root, "pkg/config/locales")))
}
