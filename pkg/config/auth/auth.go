package auth

import (
	"time"

	"github.com/gopress/internal/auth"
	"github.com/gopress/internal/auth/authority"
	"github.com/gopress/internal/auth/providers/facebook"
	"github.com/gopress/internal/auth/providers/github"
	"github.com/gopress/internal/auth/providers/google"
	"github.com/gopress/internal/auth/providers/password"
	"github.com/gopress/internal/auth/providers/twitter"
	"github.com/gopress/internal/auth_themes/clean"
	"github.com/gopress/internal/render"

	"github.com/gopress/qorpress/pkg/config"
	"github.com/gopress/qorpress/pkg/config/bindatafs"
	"github.com/gopress/qorpress/pkg/config/db"
	"github.com/gopress/qorpress/pkg/models/users"
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

	Auth.RegisterProvider(password.New(&password.Config{}))
	Auth.RegisterProvider(github.New(&config.Config.Oauth.Github))
	Auth.RegisterProvider(google.New(&config.Config.Oauth.Google))
	Auth.RegisterProvider(facebook.New(&config.Config.Oauth.Facebook))
	Auth.RegisterProvider(twitter.New(&config.Config.Oauth.Twitter))

	Authority.Register("logged_in_half_hour", authority.Rule{TimeoutSinceLastLogin: time.Minute * 30})
}
