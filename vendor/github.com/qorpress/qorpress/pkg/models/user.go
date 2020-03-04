package models

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/qorpress/media"
	"github.com/qorpress/media/oss"
	qor_seo "github.com/qorpress/seo"
	qor_slug "github.com/qorpress/slug"
)

//go:generate gp-extender -structs User -output user-funcs.go
type User struct {
	gorm.Model
	Email    string `form:"email"`
	Password string
	Name     string `form:"name"`
	Gender   string
	Role     string
	Birthday *time.Time
	Avatar   AvatarImageStorage

	// Confirm
	ConfirmToken string
	Confirmed    bool

	// Recover
	RecoverToken       string
	RecoverTokenExpiry *time.Time

	NameWithSlug qor_slug.Slug `l10n:"sync"`
	Seo          qor_seo.Setting

	// Accepts
	AcceptPrivate bool `form:"accept-private"`
	AcceptLicense bool `form:"accept-license"`
	AcceptNews    bool `form:"accept-news"`
}

func (user User) GetSEO() *qor_seo.SEO {
	return SEOCollection.GetSEO("User")
}

func (user User) DisplayName() string {
	return user.Email
}

func (user User) AvailableLocales() []string {
	return []string{"en-US", "zh-CN"}
}

type AvatarImageStorage struct{ oss.OSS }

func (AvatarImageStorage) GetSizes() map[string]*media.Size {
	return map[string]*media.Size{
		"small":  {Width: 50, Height: 50},
		"middle": {Width: 120, Height: 120},
		"big":    {Width: 320, Height: 320},
	}
}
