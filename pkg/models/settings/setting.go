package settings

import (
	"github.com/jinzhu/gorm"
	"github.com/gopress/internal/l10n"
	"github.com/gopress/internal/location"
)

// to do
type FeeSetting struct {
	ShippingFee     uint
	GiftWrappingFee uint
	CODFee          uint `gorm:"column:cod_fee"`
	TaxRate         int
}

type Setting struct {
	gorm.Model
	FeeSetting
	location.Location `location:"name:Company Address"`
	l10n.Locale
}
