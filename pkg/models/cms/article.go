package cms

import (
	"github.com/jinzhu/gorm"

	"github.com/qorpress/qorpress/core/publish2"
	"github.com/qorpress/qorpress/pkg/models/users"
)

type Article struct {
	gorm.Model
	Author   users.User
	AuthorID uint
	Title    string `gorm:"type:mediumtext"`
	Content  string `gorm:"type:longtext"`
	Slug 	 string 
	publish2.Version
	publish2.Schedule
	publish2.Visible
}
