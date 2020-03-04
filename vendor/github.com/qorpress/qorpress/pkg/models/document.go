package models

import (
	"github.com/jinzhu/gorm"
	"github.com/qorpress/l10n"
	"github.com/qorpress/media/oss"
	"github.com/qorpress/sorting"
)

//go:generate gp-extender -structs Document -output document-funcs.go
type Document struct {
	gorm.Model
	l10n.Locale
	sorting.Sorting
	File   oss.OSS `gorm:"type:longtext" sql:"size:4294967295;" media_library:"url:/public/content/publications/{{basename}}.{{extension}};path:./public"`
	PostID uint
}
