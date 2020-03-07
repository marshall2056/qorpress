package seo

import (
	"github.com/qorpress/qorpress/core/l10n"
	"github.com/qorpress/qorpress/core/seo"
)

type MySEOSetting struct {
	seo.QorSEOSetting
	l10n.Locale
}

type SEOGlobalSetting struct {
	SiteName string
}

var SEOCollection *seo.Collection
