package models

import (
	"github.com/jinzhu/gorm"
)

//go:generate gp-extender -structs Video -output video-funcs.go
type Video struct {
	gorm.Model
	Url         string
	Value       string `gorm:"type:longtext"`
	Description string `gorm:"type:longtext"`
	PostID      uint
}
