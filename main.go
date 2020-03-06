package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/foomo/simplecert"
	"github.com/spf13/pflag"

	"github.com/qorpress/qorpress/internal/admin"
	"github.com/qorpress/qorpress/internal/publish2"
	"github.com/qorpress/qorpress/internal/qor"
	"github.com/qorpress/qorpress/internal/qor/utils"
	"github.com/qorpress/qorpress/pkg/app/account"
	adminapp "github.com/qorpress/qorpress/pkg/app/admin"
	"github.com/qorpress/qorpress/pkg/app/api"
	"github.com/qorpress/qorpress/pkg/app/home"
	"github.com/qorpress/qorpress/pkg/app/pages"
	"github.com/qorpress/qorpress/pkg/app/posts"
	"github.com/qorpress/qorpress/pkg/app/static"
	"github.com/qorpress/qorpress/pkg/config"
	"github.com/qorpress/qorpress/pkg/config/application"
	"github.com/qorpress/qorpress/pkg/config/auth"
	"github.com/qorpress/qorpress/pkg/config/bindatafs"
	"github.com/qorpress/qorpress/pkg/config/db"
	_ "github.com/qorpress/qorpress/pkg/config/db/migrations"
	"github.com/qorpress/qorpress/pkg/utils/funcmapmaker"
)

var (
	compileTemplate        bool
	help bool
)

func main() {

	pflag.BoolVarP(&compileTemplate, "compile-templates", "c", false, "Compile Templates.")
	pflag.Parse()
	if help {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	// cmdLine := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	// compileTemplate := cmdLine.Bool("compile-templates", false, "Compile Templates")
	// cmdLine.Parse(os.Args[1:])

	var (
		Router = chi.NewRouter()
		Admin  = admin.New(&admin.AdminConfig{
			SiteName: config.Config.App.SiteName,
			Auth:     auth.AdminAuth{},
			DB:       db.DB.Set(publish2.VisibleMode, publish2.ModeOff).Set(publish2.ScheduleMode, publish2.ModeOff),
		})
		Application = application.New(&application.Config{
			Router: Router,
			Admin:  Admin,
			DB:     db.DB,
		})
	)

	funcmapmaker.AddFuncMapMaker(auth.Auth.Config.Render)

	Router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// for demo, don't use this for your production site
			// to do: add to the yaml configuration file
			w.Header().Add("Access-Control-Allow-Origin", "*")
			handler.ServeHTTP(w, req)
		})
	})

	Router.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			req.Header.Del("Authorization")
			handler.ServeHTTP(w, req)
		})
	})

	Router.Use(middleware.RealIP)
	Router.Use(middleware.Logger)
	Router.Use(middleware.Recoverer)

	Router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var (
				tx         = db.DB
				qorContext = &qor.Context{Request: req, Writer: w}
			)
			if locale := utils.GetLocale(qorContext); locale != "" {
				tx = tx.Set("l10n:locale", locale)
			}
			ctx := context.WithValue(req.Context(), utils.ContextDBName, publish2.PreviewByDB(tx, qorContext))
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})

	Application.Use(api.New(&api.Config{}))
	Application.Use(adminapp.New(&adminapp.Config{}))
	Application.Use(home.New(&home.Config{}))
	Application.Use(posts.New(&posts.Config{}))
	Application.Use(account.New(&account.Config{}))
	Application.Use(pages.New(&pages.Config{}))

	Application.Use(static.New(&static.Config{
		Prefixs: []string{"/system"},
		Handler: utils.FileServer(http.Dir(filepath.Join(config.Root, "public"))),
	}))

	Application.Use(static.New(&static.Config{
		Prefixs: []string{"javascripts", "stylesheets", "images", "dist", "fonts", "vendors", "favicon.ico"},
		Handler: bindatafs.AssetFS.FileServer(http.Dir(filepath.Join("themes", "qorpress", "public")), "javascripts", "stylesheets", "images", "dist", "fonts", "vendors", "favicon.ico"),
	}))

	if compileTemplate {
		bindatafs.AssetFS.Compile()
	} else {
		fmt.Printf("Listening on: %v\n", config.Config.App.Port)
		if config.Config.App.HTTPS.Enabled {
			domains := strings.Split(config.Config.App.HTTPS.Domains, ",")
			if err := simplecert.ListenAndServeTLS(
				fmt.Sprintf(":%d", config.Config.App.Port), 
				Application.NewServeMux(), 
				config.Config.App.HTTPS.Email, 
				nil, 
				domains...); err != nil {
				panic(err)
			}
		} else {
			if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Config.App.Port), Application.NewServeMux()); err != nil {
				panic(err)
			}
		}
	}
}
