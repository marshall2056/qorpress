package models

import (
	"github.com/qorpress/media/media_library"
)

//go:generate gp-extender -structs MediaLibrary -output media-funcs.go
type MediaLibrary struct {
	Title string
	media_library.MediaLibrary
}
