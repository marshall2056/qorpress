package config

import (
	"os"
	"strconv"

	"github.com/go-gomail/gomail"
	"github.com/jinzhu/configor"
	"github.com/unrolled/render"

	"github.com/qorpress/qorpress/internal/auth/providers/facebook"
	"github.com/qorpress/qorpress/internal/auth/providers/github"
	"github.com/qorpress/qorpress/internal/auth/providers/google"
	"github.com/qorpress/qorpress/internal/auth/providers/twitter"
	"github.com/qorpress/qorpress/internal/mailer"
	"github.com/qorpress/qorpress/internal/mailer/gomailer"
	"github.com/qorpress/qorpress/internal/mailer/logger"
	"github.com/qorpress/qorpress/internal/media/oss"
	"github.com/qorpress/qorpress/internal/oss/s3"
	"github.com/qorpress/qorpress/internal/redirect_back"
	"github.com/qorpress/qorpress/internal/session/manager"
)

var Config = struct {

	App struct {
		Port  uint `default:"7000" env:"QORPRESS_PORT"`
		HTTPS struct {
			Enabled bool `default:"false" env:"QORPRESS_HTTPS"`
			Local bool `default:"false" env:"QORPRESS_HTTPS_LOCAL"`
			Email string `env:"QORPRESS_HTTPS_EMAIL"`
			Domains string `env:"QORPRESS_HTTPS_DOMAINS"`
		}
		Location struct {
			BaiduAPI string
			GoogleAPI string
		}
	}

	DB    struct {
		Name     string `env:"QORPRESS_DB_NAME" default:"qor_example"`
		Adapter  string `env:"QORPRESS_DB_ADAPTER" default:"mysql"`
		Host     string `env:"QORPRESS_DB_HOST" default:"localhost"`
		Port     string `env:"QORPRESS_DB_PORT" default:"3306"`
		User     string `env:"QORPRESS_DB_USER"`
		Password string `env:"QORPRESS_DB_PASSWORD"`
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

	SMTP  SMTPConfig

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
)

func init() {
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
