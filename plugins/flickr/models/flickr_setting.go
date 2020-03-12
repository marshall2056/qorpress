package models

import (
	"github.com/jinzhu/gorm"
)

type FlickrSetting struct {
	gorm.Model
	Enabled    	bool
	ApiKey    	string
	UserId   	string
	PerPage     int
}
