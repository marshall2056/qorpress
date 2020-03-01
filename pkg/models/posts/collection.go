package posts

import (
	"github.com/jinzhu/gorm"
	"github.com/qorpress/l10n"
)

type Collection struct {
	gorm.Model
	Name string
	l10n.LocaleCreatable
}
