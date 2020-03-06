package seo

import (
	"github.com/qorpress/qorpress/internal/l10n"
	"github.com/qorpress/qorpress/internal/seo"
)

type MySEOSetting struct {
	seo.QorSEOSetting
	l10n.Locale
}

type SEOGlobalSetting struct {
	SiteName string
}

var SEOCollection *seo.Collection
