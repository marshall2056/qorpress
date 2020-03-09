package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/k0kubun/pp"
	"github.com/qorpress/go-wordpress"
	"github.com/jinzhu/gorm"
	"github.com/spf13/pflag"

	"github.com/qorpress/qorpress/core/help"
	adminseo "github.com/qorpress/qorpress/pkg/models/seo"
	i18n_database "github.com/qorpress/qorpress/core/i18n/backends/database"
	"github.com/qorpress/qorpress/core/auth/auth_identity"
	"github.com/qorpress/qorpress/core/banner_editor"
	"github.com/qorpress/qorpress/core/media/asset_manager"
	"github.com/qorpress/qorpress/core/notification"
	"github.com/qorpress/qorpress/pkg/app/admin"
	"github.com/qorpress/qorpress/pkg/models/cms"
	"github.com/qorpress/qorpress/pkg/models/posts"
	"github.com/qorpress/qorpress/pkg/models/settings"
	"github.com/qorpress/qorpress/pkg/models/users"
)

var (
	username string
	password string
	endpoint string
	truncate bool
	displayHelp     bool
	DB       *gorm.DB
	Tables   = []interface{}{
		&auth_identity.AuthIdentity{},
		&users.User{},
		&posts.Category{}, &posts.Collection{}, &posts.Tag{},
		&posts.Post{}, &posts.PostImage{}, &posts.Link{}, &posts.Comment{},
		&settings.Setting{},
		&adminseo.MySEOSetting{},
		&cms.Article{},
		&settings.MediaLibrary{},
		&banner_editor.QorBannerEditorSetting{},
		&asset_manager.AssetManager{},
		&i18n_database.Translation{},
		&notification.QorNotification{},
		&admin.QorWidgetSetting{},
		&help.QorHelpEntry{},
	}
)

func main() {

	// read .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	pflag.StringVarP(&username, "username", "", os.Getenv("WORDPRESS_USERNAME"), "wordpress' username.")
	pflag.StringVarP(&password, "password", "", os.Getenv("WORDPRESS_PASSWORD"), "wordpress' password.")
	pflag.StringVarP(&endpoint, "endpoint", "", os.Getenv("WORDPRESS_API_ENDPOINT"), "wordpress api endpoint (eg. https://domain.com/wp-json).")
	pflag.BoolVarP(&truncate, "truncate", "t", false, "truncate tables")
	pflag.BoolVarP(&displayHelp, "help", "h", false, "help info")
	pflag.Parse()
	if displayHelp {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	// init database, cleanup
	DB = InitDB()
	if truncate {
		TruncateTables(Tables...)
	}

	// create wp-api client
	tp := wordpress.BasicAuthTransport{
		Username: username,
		Password: password,
	}
	client, err := wordpress.NewClient(endpoint, tp.Client())
	if err != nil {
		log.Fatal("Error while creating wp-api client.")
	}

	ctx := context.Background()

	// get the currently authenticated users details
	authenticatedUser, _, err := client.Users.Me(ctx, nil)
	if err != nil {
		log.Fatalln(err)
	}
	pp.Printf("Authenticated user %+v\n", authenticatedUser)

	importUsers(ctx, client)
	importCategories(ctx, client)
	importTags(ctx, client)
	importMedias(ctx, client)
	importPages(ctx, client)
	importPosts(ctx, client)

}

func InitDB() *gorm.DB {
	mysqlString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=True&loc=Local&charset=utf8mb4,utf8", "root", os.Getenv("DB_PASSWORD"), "127.0.0.1", "3306", "qorpress")
	//psqlInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", host, port, user, dbname, password)
	db, err := gorm.Open("mysql", mysqlString)
	if err != nil {
		panic(err)
	}
	db.LogMode(false)
	DB = db
	return DB
}

func TruncateTables(tables ...interface{}) {
	for _, table := range tables {
		if err := DB.DropTableIfExists(table).Error; err != nil {
			panic(err)
		}
		DB.AutoMigrate(table)
	}
}

func importUsers(ctx context.Context, client *wordpress.Client) error {
	// Import users
	userOpts := &wordpress.UserListOptions{
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	var allUsers []*wordpress.User
	for {
		users, resp, err := client.Users.List(ctx, userOpts)
		if err != nil {
			return err
		}
		allUsers = append(allUsers, users...)
		if resp.NextPage == 0 {
			break
		}
		userOpts.Page = resp.NextPage
	}
	pp.Println(allUsers)
	return nil
}

func importMedias(ctx context.Context, client *wordpress.Client) error {
	// Import medias
	mediaOpts := &wordpress.MediaListOptions{
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	var allMedias []*wordpress.Media
	for {
		medias, resp, err := client.Media.List(ctx, mediaOpts)
		if err != nil {
			return err
		}
		allMedias = append(allMedias, medias...)
		if resp.NextPage == 0 {
			break
		}
		mediaOpts.Page = resp.NextPage
	}
	// pp.Println(allMedias)
	return nil
}

func importTags(ctx context.Context, client *wordpress.Client) error {
	// Import tags
	tagOpts := &wordpress.TagListOptions{
		HideEmpty: true,
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	var allTags []*wordpress.Tag
	for {
		tags, resp, err := client.Tags.List(ctx, tagOpts)
		if err != nil {
			return err
		}
		allTags = append(allTags, tags...)
		if resp.NextPage == 0 {
			break
		}
		tagOpts.Page = resp.NextPage
	}
	// pp.Println(allTags)
	return nil
}

func importPages(ctx context.Context, client *wordpress.Client) error {
	// Import pages
	pageOpts := &wordpress.PageListOptions{
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	var allPages []*wordpress.Page
	for {
		pages, resp, err := client.Pages.List(ctx, pageOpts)
		if err != nil {
			return err
		}
		allPages = append(allPages, pages...)
		if resp.NextPage == 0 {
			break
		}
		pageOpts.Page = resp.NextPage
	}
	// pp.Println(allPages)
	return nil
}

func importPosts(ctx context.Context, client *wordpress.Client) error {
	// Import posts
	postOpts := &wordpress.PostListOptions{
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	var allPosts []*wordpress.Post
	for {
		posts, resp, err := client.Posts.List(ctx, postOpts)
		if err != nil {
			return err
		}
		allPosts = append(allPosts, posts...)
		if resp.NextPage == 0 {
			break
		}
		postOpts.Page = resp.NextPage
	}
	// pp.Println(allPosts)
	return nil
}

func importCategories(ctx context.Context, client *wordpress.Client) error {
	// Import categories
	catOpts := &wordpress.CategoryListOptions{
		HideEmpty: true,
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	var allCategories []*wordpress.Category
	for {
		categories, resp, err := client.Categories.List(ctx, catOpts)
		if err != nil {
			return err
		}
		allCategories = append(allCategories, categories...)
		if resp.NextPage == 0 {
			break
		}
		catOpts.Page = resp.NextPage
	}
	// pp.Println(allCategories)
	return nil
}
