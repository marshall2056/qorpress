package cms

import (
	"github.com/qorpress/qorpress/core/page_builder"
	"github.com/qorpress/qorpress/core/publish2"
)

type Page struct {
	page_builder.Page

	publish2.Version
	publish2.Schedule
	publish2.Visible
}
