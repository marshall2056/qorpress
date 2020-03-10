package models

import (
	"github.com/jinzhu/gorm"
)

type TwitterSetting struct {
	gorm.Model
	Enabled bool
	ScreenName string
	Count int
	// TweetMode string // default: extended
	TwitterAPISetting
}

type TwitterAPISetting struct {
	ConsumerKey   string
	ConsumerSecret string
	AccessToken string
	AccessSecret string
}
