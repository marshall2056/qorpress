package posts

import (
	"github.com/jinzhu/gorm"

	"github.com/qorpress/qorpress/core/l10n"
)

type Collection struct {
	gorm.Model
	Name string `gorm:"index:name"`
	l10n.LocaleCreatable
}
