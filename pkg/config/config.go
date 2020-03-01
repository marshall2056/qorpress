package config

import (
	"os"
	"strconv"

	"github.com/go-gomail/gomail"
	"github.com/jinzhu/configor"
	"github.com/qorpress/auth/providers/facebook"
	"github.com/qorpress/auth/providers/github"
	"github.com/qorpress/auth/providers/google"
	"github.com/qorpress/auth/providers/twitter"
	"github.com/qorpress/mailer"
	"github.com/qorpress/mailer/gomailer"
	"github.com/qorpress/mailer/logger"
	"github.com/qorpress/media/oss"
	"github.com/qorpress/oss/s3"
	"github.com/qorpress/redirect_back"
	"github.com/qorpress/session/manager"
	"github.com/unrolled/render"
)

type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

var Config = struct {
	HTTPS bool `default:"false" env:"HTTPS"`
	Port  uint `default:"7000" env:"PORT"`
	DB    struct {
		Name     string `env:"DBName" default:"qor_example"`
		Adapter  string `env:"DBAdapter" default:"mysql"`
		Host     string `env:"DBHost" default:"localhost"`
		Port     string `env:"DBPort" default:"3306"`
		User     string `env:"DBUser"`
		Password string `env:"DBPassword"`
	}
	S3 struct {
		AccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
		SecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
		Region          string `env:"AWS_Region"`
		S3Bucket        string `env:"AWS_Bucket"`
	}
	SMTP         SMTPConfig
	Oauth struct {
		Github       github.Config
		Google       google.Config
		Facebook     facebook.Config
		Twitter      twitter.Config
	}
}{}

var (
	Root           = os.Getenv("GOPATH") + "/src/github.com/qorpress/qorpress-example"
	Mailer         *mailer.Mailer
	Render         = render.New()
	RedirectBack   = redirect_back.New(&redirect_back.Config{
		SessionManager:  manager.SessionManager,
		IgnoredPrefixes: []string{"/auth"},
	})
)

func init() {
	if err := configor.Load(&Config, ".config/qorpress.yml", ".config/database.yml", ".config/smtp.yml", ".config/application.yml"); err != nil {
		panic(err)
	}

	if Config.S3.AccessKeyID != "" {
		oss.Storage = s3.New(&s3.Config{
			AccessID:  Config.S3.AccessKeyID,
			AccessKey: Config.S3.SecretAccessKey,
			Region:    Config.S3.Region,
			Bucket:    Config.S3.S3Bucket,
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
