package models

import (
	"github.com/qorpress/l10n"
	"github.com/qorpress/widget"
)

type WidgetSetting struct {
	widget.QorWidgetSetting
	// publish2.Version
	// publish2.Schedule
	// publish2.Visible
	l10n.Locale
}