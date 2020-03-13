package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	slugger "github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"github.com/k0kubun/pp"
	"github.com/qorpress/go-wordpress"
	"github.com/spf13/pflag"

	"github.com/qorpress/qorpress/core/auth/auth_identity"
	"github.com/qorpress/qorpress/core/auth/providers/password"
	"github.com/qorpress/qorpress/core/banner_editor"
	"github.com/qorpress/qorpress/core/help"
	i18n_database "github.com/qorpress/qorpress/core/i18n/backends/database"
	"github.com/qorpress/qorpress/core/media/asset_manager"
	"github.com/qorpress/qorpress/core/notification"
	"github.com/qorpress/qorpress/core/qor"
	"github.com/qorpress/qorpress/core/slug"
	"github.com/qorpress/qorpress/pkg/app/admin"
	"github.com/qorpress/qorpress/pkg/config/auth"
	"github.com/qorpress/qorpress/pkg/config/db"
	_ "github.com/qorpress/qorpress/pkg/config/db/migrations"
	"github.com/qorpress/qorpress/pkg/models/cms"
	"github.com/qorpress/qorpress/pkg/models/posts"
	adminseo "github.com/qorpress/qorpress/pkg/models/seo"
	"github.com/qorpress/qorpress/pkg/models/settings"
	"github.com/qorpress/qorpress/pkg/models/users"
)

var (
	wpUsername   string
	wpPassword   string
	endpoint     string
	truncate     bool
	displayHelp  bool
	DB           *gorm.DB
	AdminUser    *users.User
	Notification = notification.New(&notification.Config{})
	Tables       = []interface{}{
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
	mapTags        = make(map[int]string, 0)
	mapCategories  = make(map[int]string, 0)
	mapMedia2Posts = make(map[int]PostMedia, 0)
	mapPosts = make(map[int]PostData, 0)

)

type PostData struct {
	Tags  PostTaxonomy
	Media PostMedia
}

type PostTaxonomy struct {
	Name string
	ID string
}

type PostMedia struct {
	SourceURL string
	// Post int
	Date time.Time
	Modified time.Time
	Author uint
	MediaType string
	Description struct {
		Rendered string
	}
	ImageMeta map[string]interface{}
}

func main() {

	// read .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	pflag.StringVarP(&wpUsername, "username", "", os.Getenv("WORDPRESS_USERNAME"), "wordpress' username.")
	pflag.StringVarP(&wpPassword, "password", "", os.Getenv("WORDPRESS_PASSWORD"), "wordpress' password.")
	pflag.StringVarP(&endpoint, "endpoint", "", os.Getenv("WORDPRESS_API_ENDPOINT"), "wordpress api endpoint (eg. https://domain.com/wp-json).")
	pflag.BoolVarP(&truncate, "truncate", "t", false, "truncate tables")
	pflag.BoolVarP(&displayHelp, "help", "h", false, "help info")
	pflag.Parse()
	if displayHelp {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	// init database, cleanup
	DB = db.DB

	if truncate {
		TruncateTables(Tables...)
	}

	// create wp-api client
	tp := wordpress.BasicAuthTransport{
		Username: wpUsername,
		Password: wpPassword,
	}
	client, err := wordpress.NewClient(endpoint, tp.Client())
	if err != nil {
		log.Fatal("Error while creating wp-api client.")
	}

	ctx := context.Background()

	// get the currently authenticated users details
	authenticatedUser, _, err := client.Users.Me(ctx, "context=edit")
	if err != nil {
		log.Fatalln(err)
	}
	// pp.Printf("resp %+v\n", resp)
	pp.Printf("Authenticated user %+v\n", authenticatedUser)

	createAdminUsers()
	importUsers(ctx, client)
	importCategories(ctx, client)
	importTags(ctx, client)
	importPages(ctx, client)
	importPosts(ctx, client)
	importMedias(ctx, client)
	os.Exit(1)
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
	userOpts.Context = "edit"
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
	// pp.Println(allUsers)
	// os.Exit(1)

	for _, wpUser := range allUsers {
		user := users.User{}
		user.Name = wpUser.Name
		user.Email = wpUser.Email

		user.CreatedAt = wpUser.RegisteredDate.Time

		if file, err := openFileByURL(wpUser.AvatarURLs.Size96); err != nil {
			fmt.Printf("open file failure, got err %v", err)
		} else {
			defer file.Close()
			user.Avatar.Scan(file)
		}

		if err := DB.Save(&user).Error; err != nil {
			log.Fatalf("Save user (%v) failure, got err %v", user, err)
		}

		provider := auth.Auth.GetProvider("password").(*password.Provider)
		hashedPassword, _ := provider.Encryptor.Digest("testing")
		authIdentity := &auth_identity.AuthIdentity{}
		authIdentity.Provider = "password"
		authIdentity.UID = user.Email
		authIdentity.EncryptedPassword = hashedPassword
		authIdentity.UserID = fmt.Sprint(user.ID)
		authIdentity.ConfirmedAt = &user.CreatedAt

		DB.Create(authIdentity)
	}

	return nil
}

func createAdminUsers() {
	AdminUser = &users.User{}
	AdminUser.Email = "dev@getqor.com"
	AdminUser.Confirmed = true
	AdminUser.Name = "QOR Admin"
	AdminUser.Role = "Admin"
	DB.Create(AdminUser)

	provider := auth.Auth.GetProvider("password").(*password.Provider)
	hashedPassword, _ := provider.Encryptor.Digest("testing")
	now := time.Now()

	authIdentity := &auth_identity.AuthIdentity{}
	authIdentity.Provider = "password"
	authIdentity.UID = AdminUser.Email
	authIdentity.EncryptedPassword = hashedPassword
	authIdentity.UserID = fmt.Sprint(AdminUser.ID)
	authIdentity.ConfirmedAt = &now

	DB.Create(authIdentity)

	// Send welcome notification
	Notification.Send(&notification.Message{
		From:        AdminUser,
		To:          AdminUser,
		Title:       "Welcome To QOR Admin",
		Body:        "Welcome To QOR Admin",
		MessageType: "info",
	}, &qor.Context{DB: DB})
}

func importMedias(ctx context.Context, client *wordpress.Client) error {
	// Import medias
	mediaOpts := &wordpress.MediaListOptions{
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	var wpMedias []*wordpress.Media
	for {
		medias, resp, err := client.Media.List(ctx, mediaOpts)
		if err != nil {
			return err
		}
		wpMedias = append(wpMedias, medias...)
		if resp.NextPage == 0 {
			break
		}
		mediaOpts.Page = resp.NextPage
	}
	pp.Println(wpMedias)
	// for _, wpMedia := range wpMedias {
	//	media := posts.Tag{}	
	// }
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
	var wpTaxonomies []*wordpress.Tag
	for {
		tags, resp, err := client.Tags.List(ctx, tagOpts)
		if err != nil {
			return err
		}
		wpTaxonomies = append(wpTaxonomies, tags...)
		if resp.NextPage == 0 {
			break
		}
		tagOpts.Page = resp.NextPage
	}
	for _, wpTaxonomy := range wpTaxonomies {
		mapCategories[wpTaxonomy.ID] = wpTaxonomy.Name
		taxonomy := posts.Tag{}
		taxonomy.Name = wpTaxonomy.Name
		taxonomy.NameWithSlug = slug.Slug{createUniqueSlug(wpTaxonomy.Name)}
		if err := DB.Create(&taxonomy).Error; err != nil {
			log.Fatalf("create taxonomy (%v) failure, got err %v", taxonomy, err)
		}
	}
	return nil
}

func importPages(ctx context.Context, client *wordpress.Client) error {
	// Import pages
	pageOpts := &wordpress.PageListOptions{
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	pageOpts.Context = "edit"
	var wpPages []*wordpress.Page
	for {
		pages, resp, err := client.Pages.List(ctx, pageOpts)
		if err != nil {
			return err
		}
		wpPages = append(wpPages, pages...)
		if resp.NextPage == 0 {
			break
		}
		pageOpts.Page = resp.NextPage
	}
	// pp.Println(wpPages)
	for _, wpPage := range wpPages {
		page := cms.Article{}
		page.Title = wpPage.Title.Rendered
		page.Slug = wpPage.Slug
		page.Content = wpPage.Content.Rendered
		if wpPage.Status == "publish" {
			page.PublishReady = true
			// start := time.Now().AddDate(0, 0, i-7)
			// end := time.Now().AddDate(0, 0, i-4)
			// page.SetScheduledStartAt(&start)
			// page.SetScheduledEndAt(&end)
		}
		if err := DB.Create(&page).Error; err != nil {
			log.Fatalf("create page (%v) failure, got err %v", page, err)
		}
	}

	// Slug
	// Status=="publish"
	// Content.Rendered

	return nil
}

func importPosts(ctx context.Context, client *wordpress.Client) error {
	// Import posts
	postOpts := &wordpress.PostListOptions{
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	postOpts.Context = "edit"
	var wpPosts []*wordpress.Post
	for {
		posts, resp, err := client.Posts.List(ctx, postOpts)
		if err != nil {
			return err
		}
		wpPosts = append(wpPosts, posts...)
		if resp.NextPage == 0 {
			break
		}
		postOpts.Page = resp.NextPage
	}
	pp.Println(wpPosts)
	/*
	for _, wpPost := range wpPosts {
		post := posts.Post{}
		post.Name = wpPost
		// post.Name = 
		// post.Name = 
		// post.Name = 
		// post.Name = 
		if err := DB.Create(&post).Error; err != nil {
			log.Fatalf("create post (%v) failure, got err %v", post, err)
		}
	}
	*/
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
	var wpCategories []*wordpress.Category
	for {
		categories, resp, err := client.Categories.List(ctx, catOpts)
		if err != nil {
			return err
		}
		wpCategories = append(wpCategories, categories...)
		if resp.NextPage == 0 {
			break
		}
		catOpts.Page = resp.NextPage
	}
	for _, wpCategory := range wpCategories {
		mapCategories[wpCategory.ID] = wpCategory.Name
		category := posts.Category{}
		category.Name = wpCategory.Name
		category.Code = strings.ToLower(wpCategory.Name)
		if err := DB.Create(&category).Error; err != nil {
			log.Fatalf("create category (%v) failure, got err %v", category, err)
		}
	}

	return nil
}

func openFileByURL(rawURL string) (*os.File, error) {
	if fileURL, err := url.Parse(rawURL); err != nil {
		return nil, err
	} else {
		path := fileURL.Path
		segments := strings.Split(path, "/")
		fileName := segments[len(segments)-1]

		filePath := filepath.Join(os.TempDir(), fileName)

		if _, err := os.Stat(filePath); err == nil {
			return os.Open(filePath)
		}

		file, err := os.Create(filePath)
		if err != nil {
			return file, err
		}

		check := http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}
		resp, err := check.Get(rawURL) // add a filter to check redirect
		if err != nil {
			return file, err
		}
		defer resp.Body.Close()
		fmt.Printf("----> Downloaded %v\n", rawURL)

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return file, err
		}
		return file, nil
	}
}

func createUniqueSlug(title string) string {
	slugTitle := slugger.Make(title)
	if len(slugTitle) > 128 {
		slugTitle = slugTitle[:128]
		if slugTitle[len(slugTitle)-1:] == "-" {
			slugTitle = slugTitle[:len(slugTitle)-1]
		}
	}
	return slugTitle
}
