package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qorpress/media"
	"github.com/qorpress/media/media_library"
)

//go:generate gp-extender -structs Image -output image-funcs.go
type Image struct {
	gorm.Model
	//  media_library:"url:/public/content/{{class}}/{{primary_key}}/{{column}}.{{extension}};path:./testmedia"
	File   media_library.MediaLibraryStorage `gorm:"type:longtext" sql:"size:4294967295;"`
	// media_library.MediaBox
	PostID uint
}

func (i *Image) BeforeCreate() (err error) {
	if i.CreatedAt.String() == "0000-00-00" {
		i.CreatedAt = time.Now()
	}
	if i.UpdatedAt.String() == "0000-00-00" {
		i.UpdatedAt = time.Now()
	}
	return nil
}

func (Image) GetSizes() map[string]*media.Size {
	return map[string]*media.Size{
		"small":           {Width: 320, Height: 320},
		"middle":          {Width: 640, Height: 640},
		"big":             {Width: 1024, Height: 720},
		"article_preview": {Width: 390, Height: 300},
		"preview":         {Width: 200, Height: 200},
	}
}
