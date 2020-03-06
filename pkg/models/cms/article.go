package cms

import (
	"github.com/jinzhu/gorm"
	"github.com/qorpress/qorpress/internal/publish2"

	"github.com/qorpress/qorpress/pkg/models/users"
)

type Article struct {
	gorm.Model
	Author   users.User
	AuthorID uint
	Title    string
	Content  string `gorm:"type:text"`
	publish2.Version
	publish2.Schedule
	publish2.Visible
}
