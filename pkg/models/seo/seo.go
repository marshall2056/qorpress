package seo

import (
	"github.com/qorpress/l10n"
	"github.com/qorpress/seo"
)

type MySEOSetting struct {
	seo.QorSEOSetting
	l10n.Locale
}

type SEOGlobalSetting struct {
	SiteName string
}

var SEOCollection *seo.Collection
