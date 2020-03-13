package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/cep21/xdgbasedir"
	"github.com/go-gomail/gomail"
	"github.com/jinzhu/configor"
	"github.com/unrolled/render"

	"github.com/qorpress/qorpress/core/auth/providers/facebook"
	"github.com/qorpress/qorpress/core/auth/providers/github"
	"github.com/qorpress/qorpress/core/auth/providers/google"
	"github.com/qorpress/qorpress/core/auth/providers/twitter"
	"github.com/qorpress/qorpress/core/mailer"
	"github.com/qorpress/qorpress/core/mailer/gomailer"
	"github.com/qorpress/qorpress/core/mailer/logger"
	"github.com/qorpress/qorpress/core/media/oss"
	"github.com/qorpress/qorpress/core/oss/s3"
	"github.com/qorpress/qorpress/core/redirect_back"
	"github.com/qorpress/qorpress/core/session/manager"
	plug "github.com/qorpress/qorpress/pkg/plugins"
)

var Config = struct {
	App struct {
		Port     uint   `default:"7000" env:"QORPRESS_PORT"`
		SiteName string `default:"QorPress Demo" env:"QORPRESS_SITENAME"`
		HTTPS    struct {
			Enabled bool   `default:"false" env:"QORPRESS_HTTPS"`
			Local   bool   `default:"false" env:"QORPRESS_HTTPS_LOCAL"`
			Email   string `env:"QORPRESS_HTTPS_EMAIL"`
			Domains string `env:"QORPRESS_HTTPS_DOMAINS"`
		}
		Location struct {
			BaiduAPI  string `env:"QORPRESS_BAIDU_API"`
			GoogleAPI string `env:"QORPRESS_GOOGLE_MAP_API"`
		}
		Theme  string `json:"theme" yaml:"theme"`
		Plugin struct {
			Filter bool   `json:"filter" yaml:"filter"`
			Dir    string `json:"dir" yaml:"dir"`
		}
		Cors struct {
			AccessControlAllowOrigin string `default:"*" json:"access-control-allow-origin" yaml:"access-control-allow-origin"`
		} `json:"cors" yaml:"cors"`
	}

	DB struct {
		Name     string `env:"QORPRESS_DB_NAME" default:"qor_example"`
		Adapter  string `env:"QORPRESS_DB_ADAPTER" default:"mysql"`
		Host     string `env:"QORPRESS_DB_HOST" default:"localhost"`
		Port     string `env:"QORPRESS_DB_PORT" default:"3306"`
		User     string `env:"QORPRESS_DB_USER"`
		Password string `env:"QORPRESS_DB_PASSWORD"`
	}

	Search struct {
		Adapter  string `env:"QORPRESS_SEARCH_ADAPTER" default:"manticore"`
		Host     string `env:"QORPRESS_SEARCH_HOST" default:"localhost"`
		Port     string `env:"QORPRESS_SEARCH_PORT" default:"3306"`		
	}

	Cloud struct {
		AWS struct {
			S3 struct {
				AccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
				SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
				Region          string `env:"AWS_REGION"`
				S3Bucket        string `env:"AWS_BUCKET"`
			}
		}
	}

	Oauth struct {
		Github   github.Config
		Google   google.Config
		Facebook facebook.Config
		Twitter  twitter.Config
	}

	SMTP SMTPConfig
}{}

type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

var (
	Root         = os.Getenv("GOPATH") + "/src/github.com/qorpress/qorpress"
	Mailer       *mailer.Mailer
	Render       = render.New()
	RedirectBack = redirect_back.New(&redirect_back.Config{
		SessionManager:  manager.SessionManager,
		IgnoredPrefixes: []string{"/auth"},
	})
	QorPlugins *plug.QorPlugin
)

func init() {

	baseDir, err := xdgbasedir.ConfigHomeDirectory()
	if err != nil {
		log.Fatal("Can't find XDG BaseDirectory")
	}
	// to do, add it for docker config path
	fmt.Println("baseDir:", baseDir)

	if err := configor.Load(&Config, ".config/qorpress.yml", ".config/database.yml", ".config/smtp.yml", ".config/application.yml"); err != nil {
		panic(err)
	}

	if Config.Cloud.AWS.S3.AccessKeyID != "" {
		oss.Storage = s3.New(&s3.Config{
			AccessID:  Config.Cloud.AWS.S3.AccessKeyID,
			AccessKey: Config.Cloud.AWS.S3.SecretAccessKey,
			Region:    Config.Cloud.AWS.S3.Region,
			Bucket:    Config.Cloud.AWS.S3.S3Bucket,
		})
	}

	portSmtp, err := strconv.Atoi(Config.SMTP.Port)
	if err != nil {
		panic(err)
	}

	dialer := gomail.NewDialer(Config.SMTP.Host, portSmtp, Config.SMTP.User, Config.SMTP.Password)
	sender, err := dialer.Dial()
	if err != nil {
		Mailer = mailer.New(&mailer.Config{
			Sender: logger.New(&logger.Config{}),
		})
	} else {
		Mailer = mailer.New(&mailer.Config{
			Sender: gomailer.New(&gomailer.Config{Sender: sender}),
		})
	}
}
