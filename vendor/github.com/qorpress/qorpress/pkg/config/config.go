package config

import (
	"os"
	"strconv"

	"github.com/go-gomail/gomail"
	"github.com/jinzhu/configor"
	"github.com/k0kubun/pp"
	"github.com/qorpress/auth/providers/facebook"
	"github.com/qorpress/auth/providers/github"
	"github.com/qorpress/auth/providers/google"
	"github.com/qorpress/auth/providers/twitter"
	"github.com/qorpress/location"
	"github.com/qorpress/mailer"
	"github.com/qorpress/mailer/gomailer"
	"github.com/qorpress/mailer/logger"
	"github.com/qorpress/media/oss"
	"github.com/qorpress/oss/s3"
	"github.com/qorpress/redirect_back"
	"github.com/qorpress/session/manager"
	"github.com/unrolled/render"

	"github.com/qorpress/qorpress/pkg/storage/gorm/generator"
)

type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

type AuthConfig struct {
	Github   github.Config
	Google   google.Config
	Facebook facebook.Config
	Twitter  twitter.Config
}

type ApiKeyConfig struct {
	GoogleAPIKey string `env:"QORPRESS_GOOGLE_API_KEY"`
	BaiduAPIKey  string `env:"QORPRESS_BAIDU_API_KEY"`
	Twitter      TwitterApiConfig
}

type TwitterApiConfig struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

type BucketConfig struct {
	S3 struct {
		AccessKeyID     string `env:"QORPRESS_AWS_ACCESS_KEY_ID"`
		SecretAccessKey string `env:"QORPRESS_AWS_SECRET_ACCESS_KEY"`
		Region          string `env:"QORPRESS_AWS_REGION"`
		S3Bucket        string `env:"QORPRESS_AWS_BUCKET"`
	}
}

var Config = struct {
	HTTPS bool `default:"false" env:"QORPRESS_HTTPS"`
	Port  uint `default:"4000" env:"QORPRESS_PORT"`
	Theme string `env:"QORPRESS_THEME" default:"gopress"`
	DB    struct {
		Name     string `env:"QORPRESS_DB_NAME" default:"gopress"`
		Adapter  string `env:"QORPRESS_DB_ADAPTER" default:"mysql"`
		Host     string `env:"QORPRESS_DB_HOST" default:"localhost"`
		Port     string `env:"QORPRESS_DB_PORT" default:"3306"`
		User     string `env:"QORPRESS_DB_USER"`
		Password string `env:"QORPRESS_DB_PASSWORD"`
	}
	Bucket BucketConfig
	ApiKey ApiKeyConfig
	Auth   AuthConfig
	SMTP   SMTPConfig
	Schema []generator.Model
}{}

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
	if err := configor.Load(&Config, ".config/gopress.yml", ".config/gopress.yaml"); err != nil {
		panic(err)
	}

	pp.Println(Config)

	location.GoogleAPIKey = Config.ApiKey.GoogleAPIKey
	location.BaiduAPIKey = Config.ApiKey.BaiduAPIKey

	if Config.Bucket.S3.AccessKeyID != "" {
		oss.Storage = s3.New(&s3.Config{
			AccessID:  Config.Bucket.S3.AccessKeyID,
			AccessKey: Config.Bucket.S3.SecretAccessKey,
			Region:    Config.Bucket.S3.Region,
			Bucket:    Config.Bucket.S3.S3Bucket,
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
