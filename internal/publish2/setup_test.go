package publish2_test

import (
	"github.com/gopress/internal/l10n"
	"github.com/gopress/internal/publish2"
	"github.com/gopress/internal/qor/test/utils"
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
