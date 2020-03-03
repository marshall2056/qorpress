package auth

import (
	"time"

	"github.com/qorpress/auth"
	"github.com/qorpress/auth/authority"
	"github.com/qorpress/auth/providers/facebook"
	"github.com/qorpress/auth/providers/github"
	"github.com/qorpress/auth/providers/google"
	"github.com/qorpress/auth/providers/password"
	"github.com/qorpress/auth/providers/twitter"
	"github.com/qorpress/auth_themes/clean"
	"github.com/qorpress/render"

	"github.com/qorpress/qorpress-example/pkg/config"
	"github.com/qorpress/qorpress-example/pkg/config/bindatafs"
	"github.com/qorpress/qorpress-example/pkg/config/db"
	"github.com/qorpress/qorpress-example/pkg/models/users"
)

var (
	// Auth initialize Auth for Authentication
	Auth = clean.New(&auth.Config{
		DB:         db.DB,
		Mailer:     config.Mailer,
		Render:     render.New(
			&render.Config{
				AssetFileSystem: bindatafs.AssetFS.NameSpace("auth"),
			}),
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
	Auth.RegisterProvider(password.New(&password.Config{}))

	Authority.Register("logged_in_half_hour", authority.Rule{TimeoutSinceLastLogin: time.Minute * 30})
}
