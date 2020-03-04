package migrations

import (
	"github.com/qorpress/activity"
	"github.com/qorpress/auth/auth_identity"
	"github.com/qorpress/banner_editor"
	"github.com/qorpress/help"
	"github.com/qorpress/media/asset_manager"
	"github.com/qorpress/transition"

	"github.com/qorpress/qorpress-example/pkg/app/admin"
	"github.com/qorpress/qorpress-example/pkg/config/db"
	"github.com/qorpress/qorpress-example/pkg/models/blogs"

	// "github.com/qorpress/qorpress-example/pkg/models/orders"
	"github.com/qorpress/qorpress-example/pkg/models/posts"
	"github.com/qorpress/qorpress-example/pkg/models/seo"
	"github.com/qorpress/qorpress-example/pkg/models/settings"

	// "github.com/qorpress/qorpress-example/pkg/models/stores"
	"github.com/qorpress/qorpress-example/pkg/models/users"
)

func init() {
	AutoMigrate(&asset_manager.AssetManager{})

	AutoMigrate(&posts.Post{}, &posts.PostVariation{}, &posts.PostImage{})
	AutoMigrate(&posts.Category{}, &posts.Tag{}, &posts.Collection{}, &posts.Comment{})

	AutoMigrate(&users.User{}, &users.Address{})

	AutoMigrate(&settings.Setting{}, &settings.MediaLibrary{})

	AutoMigrate(&transition.StateChangeLog{})

	AutoMigrate(&activity.QorActivity{})

	AutoMigrate(&admin.QorWidgetSetting{})

	AutoMigrate(&blogs.Page{}, &blogs.Article{})

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
