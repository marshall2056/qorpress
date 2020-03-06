package cms

import (
	"github.com/gopress/internal/page_builder"
	"github.com/gopress/internal/publish2"
)

type Page struct {
	page_builder.Page

	publish2.Version
	publish2.Schedule
	publish2.Visible
}
