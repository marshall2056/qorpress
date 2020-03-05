package posts

import (
	"fmt"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/qorpress/l10n"
	"github.com/qorpress/sorting"
	"github.com/qorpress/validations"
)

type Category struct {
	gorm.Model
	l10n.Locale
	sorting.Sorting
	Name string `gorm:"index:name"`
	Code string `gorm:"index:code"`

	Categories []Category
	CategoryID uint `gorm:"index:category_id"`
}

func (category Category) Validate(db *gorm.DB) {
	if strings.TrimSpace(category.Name) == "" {
		db.AddError(validations.NewError(category, "Name", "Name can not be empty"))
	}
}

func (category Category) DefaultPath() string {
	if len(category.Code) > 0 {
		return fmt.Sprintf("/category/%s", category.Code)
	}
	return "/"
}
