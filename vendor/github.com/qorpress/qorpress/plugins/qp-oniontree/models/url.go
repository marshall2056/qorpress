package models

import (
	"github.com/jinzhu/gorm"
)

//go:generate gp-extender -structs URL -output url-funcs.go
type URL struct {
	gorm.Model
	Name      string `gorm:"size:255;unique" json:"href" yaml:"href"`
	Healthy   bool   `json:"healthy" yaml:"healthy"`
	ServiceID uint   `json:"-" yaml:"-"`
}
