package seo

import (
	"github.com/gopress/internal/l10n"
	"github.com/gopress/internal/seo"
)

type MySEOSetting struct {
	seo.QorSEOSetting
	l10n.Locale
}

type SEOGlobalSetting struct {
	SiteName string
}

var SEOCollection *seo.Collection
