package services

import (
	"errors"
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/qorpress/auth/auth_identity"
	i18n_database "github.com/qorpress/i18n/backends/database"
	"github.com/qorpress/l10n"
	"github.com/qorpress/media"
	"github.com/qorpress/media/asset_manager"
	"github.com/qorpress/media/media_library"
	"github.com/qorpress/publish2"
	"github.com/qorpress/sorting"
	"github.com/qorpress/validations"

	"github.com/qorpress/qorpress/pkg/config"
	"github.com/qorpress/qorpress/pkg/models"
)

var (
	DB *gorm.DB
)

func Init() *gorm.DB {
	var err error

	dbConfig := config.Config.DB
	if config.Config.DB.Adapter == "mysql" {
		DB, err = gorm.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?parseTime=True&loc=Local&charset=utf8mb4,utf8", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Name))
		// DB = DB.Set("gorm:table_options", "CHARSET=utf8")
	} else if config.Config.DB.Adapter == "postgres" {
		DB, err = gorm.Open("postgres", fmt.Sprintf("postgres://%v:%v@%v/%v?sslmode=disable", dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Name))
	} else if config.Config.DB.Adapter == "sqlite" {
		// change path to ./shared/database, not in temp directory
		DB, err = gorm.Open("sqlite3", fmt.Sprintf("%v/%v", os.TempDir(), dbConfig.Name))
	} else {
		panic(errors.New("not supported database adapter"))
	}

	if os.Getenv("DEBUG") != "" {
		DB.LogMode(true)
	}

	if err == nil {
		var post models.Post
		// var page models.Page
		var video []models.Video
		var image []models.Image
		var link []models.Link
		var documents []models.Document

		DB.AutoMigrate(
			&models.Event{},
			&models.Category{},
			&models.Post{},
			&models.Page{},
			&models.Document{},
			&models.Video{},
			&models.Image{},
			&models.Link{},
			&models.Tag{},
			&asset_manager.AssetManager{},
			&i18n_database.Translation{},
			&auth_identity.AuthIdentity{},
			&media_library.MediaLibrary{},
		)

		DB.Model(&post).Related(&video)
		DB.Model(&post).Related(&image)
		DB.Model(&post).Related(&link)
		DB.Model(&post).Related(&documents)
		/*
			DB.Model(&page).Related(&video)
			DB.Model(&page).Related(&image)
			DB.Model(&page).Related(&link)
			DB.Model(&page).Related(&documents)
		*/
		l10n.RegisterCallbacks(DB)
		sorting.RegisterCallbacks(DB)
		validations.RegisterCallbacks(DB)
		media.RegisterCallbacks(DB)
		publish2.RegisterCallbacks(DB)

		DB.Set(publish2.VisibleMode, publish2.ModeOff).Set(publish2.ScheduleMode, publish2.ModeOff)

	} else {
		panic(err)
	}

	return DB
}

func GetDB() *gorm.DB {
	return DB
}

func getEnvOrDefault(variable string, defaultValue string) string {
	thevar := os.Getenv(variable)

	if len(thevar) > 0 {
		return thevar
	}
	return defaultValue
}
