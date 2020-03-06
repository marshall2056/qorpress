package migrations

import (
	"github.com/qorpress/qorpress/internal/activity"
	"github.com/qorpress/qorpress/internal/auth/auth_identity"
	"github.com/qorpress/qorpress/internal/banner_editor"
	"github.com/qorpress/help"
	"github.com/qorpress/qorpress/internal/media/asset_manager"
	"github.com/qorpress/qorpress/internal/transition"

	"github.com/qorpress/qorpress/pkg/app/admin"
	"github.com/qorpress/qorpress/pkg/config/db"
	"github.com/qorpress/qorpress/pkg/models/cms"
	"github.com/qorpress/qorpress/pkg/models/posts"
	"github.com/qorpress/qorpress/pkg/models/seo"
	"github.com/qorpress/qorpress/pkg/models/settings"
	"github.com/qorpress/qorpress/pkg/models/users"
)

func init() {
	AutoMigrate(&asset_manager.AssetManager{})

	AutoMigrate(&posts.Post{}, &posts.PostVariation{}, &posts.PostImage{})

	AutoMigrate(&posts.Category{}, &posts.Tag{}, &posts.Collection{}, &posts.Comment{}, &posts.Link{})

	AutoMigrate(&users.User{}, &users.Address{})

	AutoMigrate(&settings.Setting{}, &settings.MediaLibrary{})

	AutoMigrate(&transition.StateChangeLog{})

	AutoMigrate(&activity.QorActivity{})

	AutoMigrate(&admin.QorWidgetSetting{})

	AutoMigrate(&cms.Page{}, &cms.Article{})

	AutoMigrate(&seo.MySEOSetting{})

	AutoMigrate(&help.QorHelpEntry{})

	AutoMigrate(&auth_identity.AuthIdentity{})

	AutoMigrate(&banner_editor.QorBannerEditorSetting{})
}

// AutoMigrate run auto migration
func AutoMigrate(values ...interface{}) {
	for _, value := range values {
		db.DB.AutoMigrate(value)
	}
}
