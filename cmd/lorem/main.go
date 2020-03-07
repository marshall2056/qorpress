// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	slugger "github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/now"
	"github.com/nozzle/throttler"
	loremipsum "gopkg.in/loremipsum.v1"

	"github.com/qorpress/qorpress/core/help"
	qoradmin "github.com/qorpress/qorpress/core/admin"
	"github.com/qorpress/qorpress/core/auth/auth_identity"
	"github.com/qorpress/qorpress/core/auth/providers/password"
	"github.com/qorpress/qorpress/core/banner_editor"
	i18n_database "github.com/qorpress/qorpress/core/i18n/backends/database"
	"github.com/qorpress/qorpress/core/media"
	"github.com/qorpress/qorpress/core/media/asset_manager"
	"github.com/qorpress/qorpress/core/media/media_library"
	"github.com/qorpress/qorpress/core/media/oss"
	"github.com/qorpress/qorpress/core/notification"
	"github.com/qorpress/qorpress/core/notification/channels/database"
	"github.com/qorpress/qorpress/core/publish2"
	"github.com/qorpress/qorpress/core/qor"
	"github.com/qorpress/qorpress/core/seo"
	"github.com/qorpress/qorpress/core/slug"
	"github.com/qorpress/qorpress/core/sorting"
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
	loremIpsumGenerator = loremipsum.NewWithSeed(1234)
	AdminUser           *users.User
	Notification        = notification.New(&notification.Config{})
	Tables              = []interface{}{
		&auth_identity.AuthIdentity{},
		&users.User{}, &users.Address{},
		&posts.Category{}, &posts.Collection{}, &posts.Tag{},
		&posts.Post{}, &posts.PostImage{},
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
	Notification.RegisterChannel(database.New(&database.Config{}))
	TruncateTables(Tables...)
	createRecords()
}

func createRecords() {
	fmt.Println("Start create sample data...")

	createSetting()
	fmt.Println("--> Created setting.")

	createSeo()
	fmt.Println("--> Created seo.")

	createAdminUsers()
	fmt.Println("--> Created admin users.")

	createUsers()
	fmt.Println("--> Created users.")

	createCategories()
	fmt.Println("--> Created categories.")

	createCollections()
	fmt.Println("--> Created collections.")

	createMediaLibraries()
	fmt.Println("--> Created medialibraries.")

	createPosts()
	fmt.Println("--> Created posts.")

	createArticles()
	fmt.Println("--> Created articles.")

	createHelps()
	fmt.Println("--> Created helps.")

	fmt.Println("--> Done!")
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

func createSetting() {
	setting := settings.Setting{}

	setting.ShippingFee = Seeds.Setting.ShippingFee
	setting.GiftWrappingFee = Seeds.Setting.GiftWrappingFee
	setting.CODFee = Seeds.Setting.CODFee
	setting.TaxRate = Seeds.Setting.TaxRate
	setting.Address = Seeds.Setting.Address
	setting.Region = Seeds.Setting.Region
	setting.City = Seeds.Setting.City
	setting.Country = Seeds.Setting.Country
	setting.Zip = Seeds.Setting.Zip
	setting.Latitude = Seeds.Setting.Latitude
	setting.Longitude = Seeds.Setting.Longitude

	if err := DraftDB.Create(&setting).Error; err != nil {
		log.Fatalf("create setting (%v) failure, got err %v", setting, err)
	}
}

func createSeo() {
	globalSeoSetting := adminseo.MySEOSetting{}
	globalSetting := make(map[string]string)
	globalSetting["SiteName"] = "Qor Demo"
	globalSeoSetting.Setting = seo.Setting{GlobalSetting: globalSetting}
	globalSeoSetting.Name = "QorSeoGlobalSettings"
	globalSeoSetting.LanguageCode = "en-US"
	globalSeoSetting.QorSEOSetting.SetIsGlobalSEO(true)

	if err := db.DB.Create(&globalSeoSetting).Error; err != nil {
		log.Fatalf("create seo (%v) failure, got err %v", globalSeoSetting, err)
	}

	defaultSeo := adminseo.MySEOSetting{}
	defaultSeo.Setting = seo.Setting{Title: "{{SiteName}}", Description: "{{SiteName}} - Default Description", Keywords: "{{SiteName}} - Default Keywords", Type: "Default Page"}
	defaultSeo.Name = "Default Page"
	defaultSeo.LanguageCode = "en-US"
	if err := db.DB.Create(&defaultSeo).Error; err != nil {
		log.Fatalf("create seo (%v) failure, got err %v", defaultSeo, err)
	}

	postSeo := adminseo.MySEOSetting{}
	postSeo.Setting = seo.Setting{Title: "{{SiteName}}", Description: "{{SiteName}} - {{Name}} - {{Code}}", Keywords: "{{SiteName}},{{Name}},{{Code}}", Type: "Post Page"}
	postSeo.Name = "Post Page"
	postSeo.LanguageCode = "en-US"
	if err := db.DB.Create(&postSeo).Error; err != nil {
		log.Fatalf("create seo (%v) failure, got err %v", postSeo, err)
	}

	// seoSetting := models.SEOSetting{}
	// seoSetting.SiteName = Seeds.Seo.SiteName
	// seoSetting.DefaultPage = seo.Setting{Title: Seeds.Seo.DefaultPage.Title, Description: Seeds.Seo.DefaultPage.Description, Keywords: Seeds.Seo.DefaultPage.Keywords}
	// seoSetting.HomePage = seo.Setting{Title: Seeds.Seo.HomePage.Title, Description: Seeds.Seo.HomePage.Description, Keywords: Seeds.Seo.HomePage.Keywords}
	// seoSetting.PostPage = seo.Setting{Title: Seeds.Seo.PostPage.Title, Description: Seeds.Seo.PostPage.Description, Keywords: Seeds.Seo.PostPage.Keywords}

	// if err := DraftDB.Create(&seoSetting).Error; err != nil {
	// 	log.Fatalf("create seo (%v) failure, got err %v", seoSetting, err)
	// }
}

func createAdminUsers() {
	AdminUser = &users.User{}
	AdminUser.Email = "dev@getqor.com"
	AdminUser.Confirmed = true
	AdminUser.Name = "QOR Admin"
	AdminUser.Role = "Admin"
	DraftDB.Create(AdminUser)

	provider := auth.Auth.GetProvider("password").(*password.Provider)
	hashedPassword, _ := provider.Encryptor.Digest("testing")
	now := time.Now()

	authIdentity := &auth_identity.AuthIdentity{}
	authIdentity.Provider = "password"
	authIdentity.UID = AdminUser.Email
	authIdentity.EncryptedPassword = hashedPassword
	authIdentity.UserID = fmt.Sprint(AdminUser.ID)
	authIdentity.ConfirmedAt = &now

	DraftDB.Create(authIdentity)

	// Send welcome notification
	Notification.Send(&notification.Message{
		From:        AdminUser,
		To:          AdminUser,
		Title:       "Welcome To QOR Admin",
		Body:        "Welcome To QOR Admin",
		MessageType: "info",
	}, &qor.Context{DB: DraftDB})
}

func createUsers() {
	emailRegexp := regexp.MustCompile(".*(@.*)")
	totalCount := 600
	t := throttler.New(5, totalCount)

	for i := 0; i < totalCount; i++ {
		go func() error {
			defer t.Done(nil)

			user := users.User{}
			user.Name = Fake.Name()
			user.Email = emailRegexp.ReplaceAllString(Fake.Email(), strings.Replace(strings.ToLower(user.Name), " ", "_", -1)+"@example.com")
			user.Gender = []string{"Female", "Male"}[i%2]
			if err := DraftDB.Create(&user).Error; err != nil {
				log.Fatalf("create user (%v) failure, got err %v", user, err)
			}

			day := (-14 + i/45)
			user.CreatedAt = now.EndOfDay().Add(time.Duration(day*rand.Intn(24)) * time.Hour)
			if user.CreatedAt.After(time.Now()) {
				user.CreatedAt = time.Now()
			}

			now := time.Now()
			unique := fmt.Sprintf("%v", now.Unix())

			if file, err := openFileByURL("https://i.pravatar.cc/150?u=" + unique); err != nil {
				fmt.Printf("open file failure, got err %v", err)
			} else {
				defer file.Close()
				user.Avatar.Scan(file)
			}

			if err := DraftDB.Save(&user).Error; err != nil {
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

			DraftDB.Create(authIdentity)
			return nil
		}()
		t.Throttle()
	}

	// throttler errors iteration
	if t.Err() != nil {
		// Loop through the errors to see the details
		for i, err := range t.Errs() {
			log.Printf("error #%d: %s", i, err)
		}
		log.Fatal(t.Err())
	}
}

func createCategories() {
	for _, c := range Seeds.Categories {
		category := posts.Category{}
		category.Name = c.Name
		category.Code = strings.ToLower(c.Name)
		if err := DraftDB.Create(&category).Error; err != nil {
			log.Fatalf("create category (%v) failure, got err %v", category, err)
		}
	}
}

func createCollections() {
	for _, c := range Seeds.Collections {
		collection := posts.Collection{}
		collection.Name = c.Name
		if err := DraftDB.Create(&collection).Error; err != nil {
			log.Fatalf("create collection (%v) failure, got err %v", collection, err)
		}
	}
}

func createPosts() {
	numberPosts := 200
	// numberSubImgs := 4
	minTags := 1
	maxTags := 10

	for i := 0; i < numberPosts; i++ {
		category := findCategoryByName("News")

		post := posts.Post{}
		post.CategoryID = category.ID

		postName := loremIpsumGenerator.Words(20)

		post.Name = postName
		post.NameWithSlug = slug.Slug{createUniqueSlug(postName)}
		post.Code = createUniqueSlug(postName)

		post.Description = strings.Replace(loremIpsumGenerator.Paragraphs(10), "\n", "<br/><br/>", -1)
		post.PublishReady = true

		if err := DraftDB.Create(&post).Error; err != nil {
			log.Fatalf("create post (%v) failure, got err %v", post, err)
		}

		Admin := qoradmin.New(&qoradmin.AdminConfig{
			SiteName: "QORPRESS DEMO",
			Auth:     auth.AdminAuth{},
			DB:       db.DB.Set(publish2.VisibleMode, publish2.ModeOff).Set(publish2.ScheduleMode, publish2.ModeOff),
		})
		image := posts.PostImage{Title: postName, SelectedType: "image"}

		//for j := 0; j < numberSubImgs; j++ {
		if file, err := openFileByURL("https://dummyimage.com/600x400/000/fff.png&text=" + loremIpsumGenerator.Words(2)); err != nil {
			fmt.Printf("open file failure, got err %v", err)
		} else {
			defer file.Close()
			image.File.Scan(file)
		}
		if err := DraftDB.Create(&image).Error; err != nil {
			log.Fatalf("create variation_image (%v) failure, got err %v", image, err)
		}

		post.Images.Crop(Admin.NewResource(&posts.PostImage{}), DraftDB, media_library.MediaOption{
			Sizes: map[string]*media.Size{
				"main":    {Width: 560, Height: 700},
				"icon":    {Width: 50, Height: 50},
				"preview": {Width: 300, Height: 300},
				"listing": {Width: 640, Height: 640},
			},
		})
		DraftDB.Save(&post)
		//}

		if len(post.MainImage.Files) == 0 {
			post.MainImage.Files = []media_library.File{{
				ID:  json.Number(fmt.Sprint(image.ID)),
				Url: image.File.URL(),
			}}
			post.MainImage.Crop(Admin.NewResource(&posts.PostImage{}), DraftDB, media_library.MediaOption{
				Sizes: map[string]*media.Size{
					"main":    {Width: 560, Height: 700},
					"icon":    {Width: 50, Height: 50},
					"preview": {Width: 300, Height: 300},
					"listing": {Width: 640, Height: 640},
				},
			})
			DraftDB.Save(&post)
		}

		// add random tags
		countTags := rand.Intn(maxTags-minTags) + minTags
		for i := 0; i < countTags; i++ {
			word := loremIpsumGenerator.Word()
			t := &posts.Tag{
				Name: word,
			}
			tag, err := createOrUpdateTag(DraftDB, t)
			if err != nil {
				panic(err)
			}
			post.Tags = append(post.Tags, *tag)
		}
		DraftDB.Save(&post)

		if i%3 == 0 {
			start := time.Now().AddDate(0, 0, i-7)
			end := time.Now().AddDate(0, 0, i-4)
			post.SetVersionName("v1")
			post.Name = postName + " - v1"
			post.Description = strings.Replace(loremIpsumGenerator.Paragraphs(10), "\n", "<br/><br/>", -1)
			post.SetScheduledStartAt(&start)
			post.SetScheduledEndAt(&end)
			DraftDB.Save(&post)
		}

	}
}

func createOrUpdateTag(db *gorm.DB, tag *posts.Tag) (*posts.Tag, error) {
	// post.Images = images
	var existingTag posts.Tag
	if db.Where("name = ?", tag.Name).First(&existingTag).RecordNotFound() {
		err := db.Set("l10n:locale", "en-US").Create(tag).Error
		return tag, err
	}
	tag.ID = existingTag.ID
	return tag, db.Set("l10n:locale", "en-US").Save(tag).Error
}

func createMediaLibraries() {
	numberMedia := 100
	for i := 0; i < numberMedia; i++ {
		medialibrary := settings.MediaLibrary{}
		medialibrary.Title = loremIpsumGenerator.Words(10)

		if file, err := openFileByURL("https://loremflickr.com/640/360"); err != nil {
			fmt.Printf("open file failure, got err %v", err)
		} else {
			defer file.Close()
			medialibrary.File.Scan(file)
		}

		if err := DraftDB.Create(&medialibrary).Error; err != nil {
			log.Fatalf("create medialibrary (%v) failure, got err %v", medialibrary, err)
		}
	}
}

func createWidgets() {
	// home page banner
	type ImageStorage struct{ oss.OSS }
	topBannerSetting := admin.QorWidgetSetting{}
	topBannerSetting.Name = "home page banner"
	topBannerSetting.Description = "This is a top banner"
	topBannerSetting.WidgetType = "NormalBanner"
	topBannerSetting.GroupName = "Banner"
	topBannerSetting.Scope = "from_google"
	topBannerSetting.Shared = true
	topBannerValue := &struct {
		Title           string
		ButtonTitle     string
		Link            string
		BackgroundImage ImageStorage `sql:"type:varchar(4096)"`
		Logo            ImageStorage `sql:"type:varchar(4096)"`
	}{
		Title:       "Welcome Googlistas!",
		ButtonTitle: "LEARN MORE",
		Link:        "http://getqor.com",
	}
	if file, err := openFileByURL("http://qor3.s3.amazonaws.com/slide01.jpg"); err == nil {
		defer file.Close()
		topBannerValue.BackgroundImage.Scan(file)
	} else {
		fmt.Printf("open file (%q) failure, got err %v", "banner", err)
	}

	if file, err := openFileByURL("http://qor3.s3.amazonaws.com/qor_logo.png"); err == nil {
		defer file.Close()
		topBannerValue.Logo.Scan(file)
	} else {
		fmt.Printf("open file (%q) failure, got err %v", "qor_logo", err)
	}

	topBannerSetting.SetSerializableArgumentValue(topBannerValue)
	if err := DraftDB.Create(&topBannerSetting).Error; err != nil {
		log.Fatalf("Save widget (%v) failure, got err %v", topBannerSetting, err)
	}

	// SlideShow banner
	type slideImage struct {
		Title    string
		SubTitle string
		Button   string
		Link     string
		Image    oss.OSS
	}
	slideshowSetting := admin.QorWidgetSetting{}
	slideshowSetting.Name = "home page banner"
	slideshowSetting.GroupName = "Banner"
	slideshowSetting.WidgetType = "SlideShow"
	slideshowSetting.Scope = "default"
	slideshowValue := &struct {
		SlideImages []slideImage
	}{}

	for _, s := range Seeds.Slides {
		slide := slideImage{Title: s.Title, SubTitle: s.SubTitle, Button: s.Button, Link: s.Link}
		if file, err := openFileByURL(s.Image); err == nil {
			defer file.Close()
			slide.Image.Scan(file)
		} else {
			fmt.Printf("open file (%q) failure, got err %v", "banner", err)
		}
		slideshowValue.SlideImages = append(slideshowValue.SlideImages, slide)
	}
	slideshowSetting.SetSerializableArgumentValue(slideshowValue)
	if err := DraftDB.Create(&slideshowSetting).Error; err != nil {
		fmt.Printf("Save widget (%v) failure, got err %v", slideshowSetting, err)
	}

	// Featured Posts
	featurePosts := admin.QorWidgetSetting{}
	featurePosts.Name = "featured posts"
	featurePosts.Description = "featured post list"
	featurePosts.WidgetType = "Posts"
	featurePosts.SetSerializableArgumentValue(&struct {
		Posts       []string
		PostsSorter sorting.SortableCollection
	}{
		Posts:       []string{"1", "2", "3", "4", "5", "6", "7", "8"},
		PostsSorter: sorting.SortableCollection{PrimaryKeys: []string{"1", "2", "3", "4", "5", "6", "7", "8"}},
	})
	if err := DraftDB.Create(&featurePosts).Error; err != nil {
		log.Fatalf("Save widget (%v) failure, got err %v", featurePosts, err)
	}

	// Banner edit items
	for _, s := range Seeds.BannerEditorSettings {
		setting := banner_editor.QorBannerEditorSetting{}
		id, _ := strconv.Atoi(s.ID)
		setting.ID = uint(id)
		setting.Kind = s.Kind
		setting.Value.SerializedValue = s.Value
		if err := DraftDB.Create(&setting).Error; err != nil {
			log.Fatalf("Save QorBannerEditorSetting (%v) failure, got err %v", setting, err)
		}
	}

	// Model posts
	modelCollectionWidget := admin.QorWidgetSetting{}
	modelCollectionWidget.Name = "model posts"
	modelCollectionWidget.Description = "Model posts banner"
	modelCollectionWidget.WidgetType = "FullWidthBannerEditor"
	modelCollectionWidget.Value.SerializedValue = `{"Value":"%3Cdiv%20class%3D%22qor-bannereditor__html%22%20style%3D%22position%3A%20relative%3B%20height%3A%20100%25%3B%22%20data-image-width%3D%221100%22%20data-image-height%3D%221200%22%3E%3Cspan%20class%3D%22qor-bannereditor-image%22%3E%3Cimg%20src%3D%22%2Fsystem%2Fmedia_libraries%2F4%2Ffile.jpg%22%3E%3C%2Fspan%3E%3Cspan%20class%3D%22qor-bannereditor__draggable%22%20data-edit-id%3D%2249%22%20style%3D%22position%3A%20absolute%3B%20left%3A%2026.4545%25%3B%20top%3A%204.41667%25%3B%20right%3A%20auto%3B%20bottom%3A%20auto%3B%22%20data-position-left%3D%22291%22%20data-position-top%3D%2253%22%3E%3Ch1%20class%3D%22banner-title%22%20style%3D%22color%3A%20%3B%22%3EENJOY%20THE%20NEW%20FASHION%20EXPERIENCE%3C%2Fh1%3E%3C%2Fspan%3E%3Cspan%20class%3D%22qor-bannereditor__draggable%22%20data-edit-id%3D%2242%22%20style%3D%22position%3A%20absolute%3B%20left%3A%2043.2727%25%3B%20top%3A%208.41667%25%3B%20right%3A%20auto%3B%20bottom%3A%20auto%3B%22%20data-position-left%3D%22476%22%20data-position-top%3D%22101%22%3E%3Cp%20class%3D%22banner-text%22%20style%3D%22color%3A%20%3B%22%3ENew%20look%20of%202017%3C%2Fp%3E%3C%2Fspan%3E%3Cspan%20class%3D%22qor-bannereditor__draggable%20qor-bannereditor__draggable-left%22%20data-edit-id%3D%2243%22%20style%3D%22position%3A%20absolute%3B%20left%3A%205.45455%25%3B%20top%3A%2044.25%25%3B%20right%3A%20auto%3B%20bottom%3A%20auto%3B%22%20data-position-left%3D%2260%22%20data-position-top%3D%22531%22%3E%3Cdiv%20class%3D%22model-buy-block%22%3E%3Ch2%20class%3D%22banner-sub-title%22%3ETOP%3C%2Fh2%3E%3Cp%20class%3D%22banner-text%22%3E%2429.99%3C%2Fp%3E%3Ca%20class%3D%22button%20button__primary%20banner-button%22%20href%3D%22%23%22%3EVIEW%20DETAILS%3C%2Fa%3E%3C%2Fdiv%3E%3C%2Fspan%3E%3Cspan%20class%3D%22qor-bannereditor__draggable%22%20data-edit-id%3D%2244%22%20style%3D%22position%3A%20absolute%3B%20left%3A%20auto%3B%20top%3A%2050.8333%25%3B%20right%3A%209.58527%25%3B%20bottom%3A%20auto%3B%22%20data-position-left%3D%22841%22%20data-position-top%3D%22610%22%3E%3Cdiv%20class%3D%22model-buy-block%22%3E%3Ch2%20class%3D%22banner-sub-title%22%3EPINK%20JACKET%3C%2Fh2%3E%3Cp%20class%3D%22banner-text%22%3E%2469.99%3C%2Fp%3E%3Ca%20class%3D%22button%20button__primary%20banner-button%22%20href%3D%22%23%22%3EVIEW%20DETAILS%3C%2Fa%3E%3C%2Fdiv%3E%3C%2Fspan%3E%3Cspan%20class%3D%22qor-bannereditor__draggable%20qor-bannereditor__draggable-left%22%20data-edit-id%3D%2247%22%20style%3D%22position%3A%20absolute%3B%20left%3A%2012.3636%25%3B%20top%3A%20auto%3B%20right%3A%20auto%3B%20bottom%3A%2014.2032%25%3B%22%20data-position-left%3D%22136%22%20data-position-top%3D%22903%22%3E%3Cdiv%20class%3D%22model-buy-block%22%3E%3Ch2%20class%3D%22banner-sub-title%22%3EBOTTOM%3C%2Fh2%3E%3Cp%20class%3D%22banner-text%22%3E%2432.99%3C%2Fp%3E%3Ca%20class%3D%22button%20button__primary%20banner-button%22%20href%3D%22%23%22%3EVIEW%20DETAILS%3C%2Fa%3E%3C%2Fdiv%3E%3C%2Fspan%3E%3Cspan%20class%3D%22qor-bannereditor__draggable%22%20data-edit-id%3D%2245%22%20style%3D%22position%3A%20absolute%3B%20left%3A%2053.2727%25%3B%20top%3A%2048.5%25%3B%20right%3A%20auto%3B%20bottom%3A%20auto%3B%22%20data-position-left%3D%22586%22%20data-position-top%3D%22582%22%3E%3Cimg%20src%3D%22%2F%2Fqor3.s3.amazonaws.com%2Fmedialibrary%2Farrow-left.png%22%20class%3D%22banner-image%22%3E%3C%2Fspan%3E%3Cspan%20class%3D%22qor-bannereditor__draggable%22%20data-edit-id%3D%2246%22%20style%3D%22position%3A%20absolute%3B%20left%3A%2015.5455%25%3B%20top%3A%2043.0833%25%3B%20right%3A%20auto%3B%20bottom%3A%20auto%3B%22%20data-position-left%3D%22171%22%20data-position-top%3D%22517%22%3E%3Cimg%20src%3D%22%2F%2Fqor3.s3.amazonaws.com%2Fmedialibrary%2Farrow-right.png%22%20class%3D%22banner-image%22%3E%3C%2Fspan%3E%3Cspan%20class%3D%22qor-bannereditor__draggable%22%20data-edit-id%3D%2248%22%20style%3D%22position%3A%20absolute%3B%20left%3A%2019.2727%25%3B%20top%3A%20auto%3B%20right%3A%20auto%3B%20bottom%3A%2024.8333%25%3B%22%20data-position-left%3D%22212%22%20data-position-top%3D%22879%22%3E%3Cimg%20src%3D%22%2F%2Fqor3.s3.amazonaws.com%2Fmedialibrary%2Farrow-right.png%22%20class%3D%22banner-image%22%3E%3C%2Fspan%3E%3C%2Fdiv%3E"}`
	if err := DraftDB.Create(&modelCollectionWidget).Error; err != nil {
		log.Fatalf("Save widget (%v) failure, got err %v", modelCollectionWidget, err)
	}
}

func createHelps() {
	helps := map[string][]string{
		"How to setup a microsite":        []string{"micro_sites"},
		"How to create a user":            []string{"users"},
		"How to create an admin user":     []string{"users"},
		"How to handle abandoned order":   []string{"abandoned_orders", "orders"},
		"How to cancel a order":           []string{"orders"},
		"How to create a order":           []string{"orders"},
		"How to upload post images":       []string{"posts", "post_images"},
		"How to create a post":            []string{"posts"},
		"How to create a discounted post": []string{"posts"},
		"How to create a store":           []string{"stores"},
		"How shop setting works":          []string{"shop_settings"},
		"How to setup seo settings":       []string{"seo_settings"},
		"How to setup seo for blog":       []string{"seo_settings"},
		"How to setup seo for post":       []string{"seo_settings"},
		"How to setup seo for microsites": []string{"micro_sites", "seo_settings"},
		"How to setup promotions":         []string{"promotions"},
		"How to publish a promotion":      []string{"schedules", "promotions"},
		"How to create a publish event":   []string{"schedules", "scheduled_events"},
		"How to publish a post":           []string{"schedules", "posts"},
		"How to publish a microsite":      []string{"schedules", "micro_sites"},
		"How to create a scheduled data":  []string{"schedules"},
		"How to take something offline":   []string{"schedules"},
	}

	for key, value := range helps {
		helpEntry := help.QorHelpEntry{
			Title: key,
			Body:  "Content of " + key,
			Categories: help.Categories{
				Categories: value,
			},
		}
		DraftDB.Create(&helpEntry)
	}
}

func createArticles() {
	for idx := 1; idx <= 10; idx++ {
		title := fmt.Sprintf("Article %v", idx)
		article := cms.Article{Title: title}
		article.PublishReady = true
		DraftDB.Create(&article)

		for i := 1; i <= idx-5; i++ {
			article.SetVersionName(fmt.Sprintf("v%v", i))
			start := time.Now().AddDate(0, 0, i*2-3)
			end := time.Now().AddDate(0, 0, i*2-1)
			article.SetScheduledStartAt(&start)
			article.SetScheduledEndAt(&end)
			DraftDB.Save(&article)
		}
	}
}

func findCategoryByName(name string) *posts.Category {
	category := &posts.Category{}
	if err := DraftDB.Where(&posts.Category{Name: name}).First(category).Error; err != nil {
		log.Fatalf("can't find category with name = %q, got err %v", name, err)
	}
	return category
}

func findCollectionByName(name string) *posts.Collection {
	collection := &posts.Collection{}
	if err := DraftDB.Where(&posts.Collection{Name: name}).First(collection).Error; err != nil {
		log.Fatalf("can't find collection with name = %q, got err %v", name, err)
	}
	return collection
}

func randTime() time.Time {
	num := rand.Intn(10)
	return time.Now().Add(-time.Duration(num*24) * time.Hour)
}

func openFileByURL(rawURL string) (*os.File, error) {
	if _, err := url.Parse(rawURL); err != nil {
		return nil, err
	} else {
		// path := fileURL.Path
		// segments := strings.Split(path, "/")
		fileName := Fake.UserName() + ".png" // segments[len(segments)-1]

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
