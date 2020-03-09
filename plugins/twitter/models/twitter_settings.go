package models

import (
	"github.com/jinzhu/gorm"
)

type TwitterSetting struct {
	gorm.Model
	Enabled bool
	Limit uint
	TwitterAPISetting
}

type TwitterAPISetting struct {
	ConsumerKey   string
	ConsumerSecret string
	AccessToken string
	AccessSecret string
}
