package publish2_test

import (
	"github.com/qorpress/qorpress/internal/l10n"
	"github.com/qorpress/qorpress/internal/publish2"
	"github.com/qorpress/qorpress/internal/qor/test/utils"
)

var DB = utils.TestDB()

func init() {
	models := []interface{}{
		&Wiki{}, &Post{}, &Article{}, &Discount{}, &User{}, &Campaign{},
		&Product{}, &L10nProduct{}, &SharedVersionProduct{}, &SharedVersionColorVariation{}, &SharedVersionSizeVariation{},
	}

	DB.DropTableIfExists(models...)
	DB.AutoMigrate(models...)
	publish2.RegisterCallbacks(DB)
	l10n.RegisterCallbacks(DB)
}
