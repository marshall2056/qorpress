package models

import (
	"github.com/jinzhu/gorm"
)

type Tag struct {
	gorm.Model
	Name string `gorm:"size:32;unique" json:"name" yaml:"name"`
}