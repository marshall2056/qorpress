package models

import (
	"github.com/jinzhu/gorm"

	"github.com/qorpress/qorpress/core/l10n"
)

// to do
type PaginationSetting struct {
	Limit   uint
	PerPage uint
}

type Setting struct {
	gorm.Model
	PaginationSetting
	l10n.Locale
}
