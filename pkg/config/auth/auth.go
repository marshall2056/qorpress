package auth

import (
	"time"

	"github.com/qorpress/qorpress/core/auth"
	"github.com/qorpress/qorpress/core/auth/authority"
	"github.com/qorpress/qorpress/core/auth/providers/facebook"
	"github.com/qorpress/qorpress/core/auth/providers/github"
	"github.com/qorpress/qorpress/core/auth/providers/google"
	"github.com/qorpress/qorpress/core/auth/providers/twitter"
	"github.com/qorpress/qorpress/core/auth_themes/clean"
	"github.com/qorpress/qorpress/core/render"
	"github.com/qorpress/qorpress/pkg/config"
	"github.com/qorpress/qorpress/pkg/config/bindatafs"
	"github.com/qorpress/qorpress/pkg/config/db"
	"github.com/qorpress/qorpress/pkg/models/users"
)

var (
	// Auth initialize Auth for Authentication
	Auth = clean.New(&auth.Config{
		DB:         db.DB,
		Mailer:     config.Mailer,
		Render:     render.New(&render.Config{AssetFileSystem: bindatafs.AssetFS.NameSpace("auth")}),
		UserModel:  users.User{},
		Redirector: auth.Redirector{RedirectBack: config.RedirectBack},
	})

	// Authority initialize Authority for Authorization
	Authority = authority.New(&authority.Config{
		Auth: Auth,
	})
)

func init() {

	Auth.RegisterProvider(github.New(&config.Config.Oauth.Github))
	Auth.RegisterProvider(google.New(&config.Config.Oauth.Google))
	Auth.RegisterProvider(facebook.New(&config.Config.Oauth.Facebook))
	Auth.RegisterProvider(twitter.New(&config.Config.Oauth.Twitter))

	Authority.Register("logged_in_half_hour", authority.Rule{TimeoutSinceLastLogin: time.Minute * 30})
}
