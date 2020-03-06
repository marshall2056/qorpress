package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Depado/bfchroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/corpix/uarand"
	badger "github.com/dgraph-io/badger"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/golang/snappy"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/google/go-github/v29/github"
	slugger "github.com/gosimple/slug"
	"github.com/h2non/filetype"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jinzhu/now"
	"github.com/joho/godotenv"
	"github.com/k0kubun/pp"
	"github.com/nozzle/throttler"
	cmap "github.com/orcaman/concurrent-map"
	qoradmin "github.com/qorpress/qorpress/internal/admin"
	"github.com/qorpress/qorpress/internal/auth/auth_identity"
	"github.com/qorpress/qorpress/internal/auth/providers/password"
	"github.com/qorpress/qorpress/internal/banner_editor"
	"github.com/qorpress/grab"
	"github.com/qorpress/qorpress/internal/help"
	i18n_database "github.com/qorpress/qorpress/internal/i18n/backends/database"
	"github.com/qorpress/qorpress/internal/media"
	"github.com/qorpress/qorpress/internal/media/asset_manager"
	"github.com/qorpress/qorpress/internal/media/media_library"
	"github.com/qorpress/qorpress/internal/media/oss"
	"github.com/qorpress/qorpress/internal/notification"
	"github.com/qorpress/qorpress/internal/oss/filesystem"
	"github.com/qorpress/qorpress/internal/publish2"
	"github.com/qorpress/qorpress/internal/qor"
	"github.com/qorpress/qorpress/internal/seo"
	"github.com/qorpress/qorpress/internal/slug"
	"github.com/qorpress/qorpress/internal/sorting"
	bf "github.com/russross/blackfriday/v2"
	log "github.com/sirupsen/logrus"
	"github.com/x0rzkov/go-vcsurl"
	loremipsum "gopkg.in/loremipsum.v1"

	ghclient "github.com/qorpress/qorpress/internal/qorpress-test/pkg/client"
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
	clientManager       *ghclient.ClientManager
	clientGH            *ghclient.GHClient
	store               *badger.DB
	DB                  *gorm.DB
	clientGrab          = grab.NewClient()
	storage             *filesystem.FileSystem
	cachePath           = "./shared/data/httpcache"
	storagePath         = "./shared/data/badger"
	debug               = false
	isChroma            = false
	logLevelStr         = "info"
	addMedia            = false
	minComments         = 1
	maxComments         = 5
	Tables              = []interface{}{
		&auth_identity.AuthIdentity{},
		&users.User{}, &users.Address{},
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

	// Defines the extensions that are used
	exts = bf.NoIntraEmphasis | bf.Tables | bf.FencedCode | bf.Autolink |
		bf.Strikethrough | bf.SpaceHeadings | bf.BackslashLineBreak |
		bf.DefinitionLists | bf.Footnotes

	// Defines the HTML rendering flags that are used
	flags = bf.UseXHTML | bf.Smartypants | bf.SmartypantsFractions |
		bf.SmartypantsDashes | bf.SmartypantsLatexDashes | bf.TOC
)

func init() {
	// log.SetReportCaller(true)
}

func main() {

	DB = InitDB()
	m := cmap.New()

	TruncateTables(Tables...)

	// add indexes
	if err := DB.Table("post_tags").AddIndex("idx_post_id", "post_id").Error; err != nil {
		log.Fatalln("Error adding index: ", err)
	}
	if err := DB.Table("post_tags").AddIndex("idx_tag_id", "tag_id").Error; err != nil {
		log.Fatalln("Error adding index: ", err)
	}

	if err := DB.Table("post_links").AddIndex("idx_post_id", "post_id").Error; err != nil {
		log.Fatalln("Error adding index: ", err)
	}
	if err := DB.Table("post_links").AddIndex("idx_link_id", "link_id").Error; err != nil {
		log.Fatalln("Error adding index: ", err)
	}

	err := removeContents("./public/system/")
	if err != nil {
		log.Warn("Error deleting old images")
	}

	// read .env file
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	err = ensureDir(storagePath)
	if err != nil {
		log.Fatal(err)
	}
	store, err = badger.Open(badger.DefaultOptions(storagePath))
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	createRecords()

	// github client init
	clientManager = ghclient.NewManager(cachePath, []string{os.Getenv("GITHUB_TOKEN")})
	defer clientManager.Shutdown()
	clientGH = clientManager.Fetch()

	// Create a Collector specifically for Shopify
	c := colly.NewCollector(
		colly.UserAgent(uarand.GetRandom()),
		colly.AllowedDomains("www.kitploit.com"),
		colly.CacheDir("./shared/data/cache"),
	)

	// create a request queue with 2 consumer threads
	q, _ := queue.New(
		20, // Number of consumer threads
		&queue.InMemoryQueueStorage{
			MaxSize: 100000,
		}, // Use default queue storage
	)

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//sitemap/loc", func(e *colly.XMLElement) {
		q.AddURL(e.Text)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
		q.AddURL(e.Text)
	})

	c.OnHTML("div.blog-posts.hfeed", func(e *colly.HTMLElement) {
		// pp.Println(e)
		// os.Exit(1)
		e.ForEach("a[href]", func(_ int, eli *colly.HTMLElement) {
			if strings.HasPrefix(eli.Attr("href"), "https://github.com") {
				var vcsUrl string
				if info, err := vcsurl.Parse(eli.Attr("href")); err == nil {
					vcsUrl = fmt.Sprintf("https://github.com/%s/%s", info.Username, info.Name)
					// githubUrls = append(githubUrls, vcsUrl)
					// githubUrls[vcsUrl] = true
					log.Println("found href=", vcsUrl)
				}
				var topics []string
				e.ForEach(".label-head > a", func(_ int, eli *colly.HTMLElement) {
					topic := strcase.ToCamel(fmt.Sprintf("%s", eli.Attr("title")))
					topic = strings.Replace(topic, " ", "", -1)
					topic = strings.Replace(topic, "!", "", -1)
					topic = strings.Replace(topic, "/", "", -1)
					topic = strings.Replace(topic, "'", "", -1)
					// log.Println("topic: ", fmt.Sprintf("#%s", topic))
					topics = append(topics, fmt.Sprintf("%s", topic))
				})
				if vcsUrl != "" {
					m.Set(vcsUrl, strings.Join(topics, ","))
				}
			}
		})
	})

	q.AddURL("https://www.kitploit.com/sitemap.xml")

	// Consume URLs
	q.Run(c)

	log.Println("All github URLs:")
	log.Println("Collected cmap: ", m.Count(), "URLs")

	// be carefull, higher values implies potential DEADLOCKs fro the datadase
	// at least for now, until I solve this issue
	t := throttler.New(1, m.Count())

	m.IterCb(func(key string, v interface{}) {
		var topics string
		_, ok := v.(string)
		if ok {
			topics = v.(string)
		}

		go func(key, topics string) error {
			// Let Throttler know when the goroutine completes
			// so it can dispatch another worker
			defer t.Done(nil)

			var imgLinks []string
			var videoLinks []string
			if info, err := vcsurl.Parse(key); err == nil {

				repoInfo, err := getInfo(clientGH.Client, info.Username, info.Name)
				if err != nil {
					log.Warnln(err)
					return err
				}

				readme, err := getReadme(clientGH.Client, info.Username, info.Name)
				if err != nil {
					log.Warnln(err)
					return err
				}

				parser := parser.NewWithExtensions(parser.CommonExtensions)
				if readme == "" {
					return nil
				}

				var html []byte
				if isChroma {
					html = render([]byte(readme))
				} else {
					html = markdown.ToHTML([]byte(readme), parser, nil)
				}

				// youtube
				// `^((?:https?:)?\/\/)?((?:www|m)\.)?((?:youtube\.com|youtu.be))(\/(?:[\w\-]+\?v=|embed\/|v\/)?)([\w\-]+)(\S+)?$`

				videoPatternRegexp, err := regexp.Compile(`^(http:\/\/|https:\/\/)(vimeo\.com|youtu\.be|www\.youtube\.com)\/([\w\/]+)([\?].*)?$`)
				if err != nil {
					log.Warnln(err)
					return err
				}
				vidLinks := videoPatternRegexp.FindAllString(string(html), -1)
				videoLinks = append(videoLinks, vidLinks...)

				// pp.Println(readme)

				imgPatternRegexp, err := regexp.Compile(`(http(s?):)([/|.|\w|\s|-])*\.(?:jpg|gif|GIF|png|PNG|jpeg|JPG|JPEG)`)
				// imgPatternRegexp, err := regexp.Compile(`(http(s?):)([/|.|\w|\s|-])*\.(?:gif|GIF)`)
				if err != nil {
					log.Warnln(err)
					return err
				}
				imgLinks = imgPatternRegexp.FindAllString(string(html), -1)
				imgRelRegexp, err := regexp.Compile(`([/|.|\w|\s|-])*\.(?:jpg|gif|png|PNG|GIF|jpeg|JPG|JPEG)`)
				if err != nil {
					log.Warnln(err)
					return err
				}
				imgLinksRel := imgRelRegexp.FindAllString(string(html), -1)
				for i, imgRel := range imgLinksRel {
					if strings.HasPrefix(imgRel, "//") {
						imgLinksRel[i] = "https:" + imgRel
					} else {
						// https://raw.githubusercontent.com/TheSph1nx/AbsoluteZero/master/screenshots/AbsoluteZero.png
						// https://github.com/TheSph1nx/absolutezero/raw/master/screenshots/AbsoluteZero.png
						// prefixUrl := strings.Replace(key, "https://github.com/", "https://raw.githubusercontent.com/", -1)
						rawURL := "https://raw.githubusercontent.com/" + info.Username + "/" + info.Name + "/" + *repoInfo.DefaultBranch + "/" + imgRel
						if _, err := url.Parse(rawURL); err != nil {
							continue
						} else {
							imgLinksRel[i] = rawURL
						}
					}
					imgLinksRel[i] = strings.Replace(imgLinksRel[i], "/blob/", "/raw/", -1)
				}

				imgLinks = append(imgLinks, imgLinksRel...)
				imgLinks = removeDuplicates(imgLinks)
				// pp.Println(imgLinksRel)
				pp.Println(imgLinks)

				var title, desc string
				if repoInfo.Description != nil {
					desc = strings.TrimSpace(*repoInfo.Description)
					if len(desc) > 512 {
						title = desc[0:512]
					} else {
						title = desc
					}
					if title == "" {
						title = *repoInfo.Name
					}
				}

				var extTopics []string
				extTopics = append(extTopics, repoInfo.Topics...)
				kitTopics := strings.Split(topics, ",")
				extTopics = append(extTopics, kitTopics...)
				extTopics = removeDuplicates(extTopics)

				start := time.Now().AddDate(0, 0, 0)
				end := time.Now().AddDate(12, 0, 0)

				category := findCategoryByName("News")

				p := &posts.Post{
					Name:        *repoInfo.Name,
					Code:        "github-" + info.Username + "-" + info.Name,
					Description: string(html),
					Summary:     desc,
					PostProperties: []posts.PostProperty{
						posts.PostProperty{
							Name:  "UpdatedAt",
							Value: repoInfo.UpdatedAt.String(),
						},
						posts.PostProperty{
							Name:  "CreatedAt",
							Value: repoInfo.CreatedAt.String(),
						},
					},
					NameWithSlug: slug.Slug{"github-" + info.Username + "-" + info.Name},
				}

				p.CategoryID = category.ID

				p.LanguageCode = "en-US"
				p.SetPublishReady(true)
				p.SetVersionName("v1")
				p.SetScheduledStartAt(&start)
				p.SetScheduledEndAt(&end)

				var tags []*posts.Tag
				for _, extTopic := range extTopics {
					tag := &posts.Tag{
						Name: extTopic,
					}
					tag.SetLanguageCode("en-US")
					tags = append(tags, tag)
				}

				var post *posts.Post
				post, err = createOrUpdatePost(DB, p)
				if err != nil {
					log.Fatalln("createOrUpdatePost: ", err)
				}

				link := &posts.Link{
					URL:   key,
					Name:  "Download "+ *repoInfo.Name,
					Title: desc,
					PostID: post.ID,
				}
				l, err := createOrUpdateLink(DB, link)
				if err != nil {
					log.Fatalln("createOrUpdateLink: ", err)
				}

				pp.Println("new link: ", l)
				p.Links = append(p.Links, *l)


				for _, tag := range tags {
					t, err := createOrUpdateTag(DB, tag)
					if err != nil {
						log.Fatalln("createOrUpdateTag: ", err)
					}
					post.Tags = append(post.Tags, *t)
				}

				countComments := rand.Intn(maxComments-minComments) + minComments
				for i := 0; i < countComments; i++ {
					content := loremIpsumGenerator.Words(20)
					c := &posts.Comment{
						Content: content,
					}
					comment, err := createComment(DraftDB, c)
					if err != nil {
						panic(err)
					}
					post.Comments = append(post.Comments, *comment)
				}

				// for _, video := range videoLinks {
				// }

				if len(videoLinks) > 0 {
					pp.Println("videoLinks:", videoLinks)
					os.Exit(1)
				}

				Admin := qoradmin.New(&qoradmin.AdminConfig{
					SiteName: "QORPRESS DEMO",
					Auth:     auth.AdminAuth{},
					DB:       db.DB.Set(publish2.VisibleMode, publish2.ModeOff).Set(publish2.ScheduleMode, publish2.ModeOff),
				})

				for _, img := range imgLinks {
					file, size, err := openFileByURL(img)
					if err != nil {
						fmt.Printf("open file failure, got err %v\n", err)
						continue
					}

					head := make([]byte, 261)
					file.Read(head)

					if filetype.IsImage(head) {
						log.Println("File is an image: ", img)
					} else {
						log.Println("Not an image", img)
						continue
					}

					if size < 5000 {
						continue
					}

					image := posts.PostImage{Title: *repoInfo.Name, SelectedType: "image"}
					image.File.Scan(file)

					if err := DraftDB.Create(&image).Error; err != nil {
						log.Warnf("create variation_image (%v) failure, got err %v\n", image, err)
						continue
					}

					post.Images.Files = append(post.Images.Files, media_library.File{
						ID:  json.Number(fmt.Sprint(image.ID)),
						Url: image.File.URL(),
					})

					post.Images.Crop(Admin.NewResource(&posts.PostImage{}), DraftDB, media_library.MediaOption{
						Sizes: map[string]*media.Size{
							"main":    {Width: 560, Height: 700},
							"icon":    {Width: 50, Height: 50},
							"preview": {Width: 300, Height: 300},
							"listing": {Width: 640, Height: 640},
						},
					})
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
					}
					//if err := DraftDB.Save(&post).Error; err != nil {
					// 	log.Fatalln("Save.post #1: ", err)
					//}

					file.Close()
				}

				/*
					if len(imgLinks) == 0 {
						image := posts.PostImage{Title: "default image", SelectedType: "image"}
						if file, _, err := openFileByURL("https://dummyimage.com/700/09f/fff.png"); err != nil {
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
								// "preview": {Width: 300, Height: 300},
								// "listing": {Width: 640, Height: 640},
							},
						})
						if len(post.MainImage.Files) == 0 {
							post.MainImage.Files = []media_library.File{{
								ID:  json.Number(fmt.Sprint(image.ID)),
								Url: image.File.URL(),
							}}
							post.MainImage.Crop(Admin.NewResource(&posts.PostImage{}), DraftDB, media_library.MediaOption{
								Sizes: map[string]*media.Size{
									"main":    {Width: 560, Height: 700},
									"icon":    {Width: 50, Height: 50},
									// "preview": {Width: 300, Height: 300},
									// "listing": {Width: 640, Height: 640},
								},
							})
						}
					}
				*/

				if err := DraftDB.Save(&post).Error; err != nil {
					log.Fatalln("Save.post #2: ", err)
				}

				return nil
			}
			return nil

		}(key, topics)

		t.Throttle()

	})

	// throttler errors iteration
	if t.Err() != nil {
		// Loop through the errors to see the details
		for i, err := range t.Errs() {
			log.Printf("error #%d: %s", i, err)
		}
		log.Fatal(t.Err())
	}

}

func InitDB() *gorm.DB {
	mysqlString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=True&loc=Local&charset=utf8mb4,utf8", "root", os.Getenv("DB_PASSWORD"), "127.0.0.1", "3306", "qorpress_example")

	//psqlInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", host, port, user, dbname, password)
	db, err := gorm.Open("mysql", mysqlString)
	if err != nil {
		panic(err)
	}
	db.LogMode(false)
	DB = db

	return DB
}

// render will take a []byte input and will render it using a new renderer each
// time because reusing the same can mess with TOC and header IDs
func render(input []byte) []byte {
	return bf.Run(
		input,
		bf.WithRenderer(
			bfchroma.NewRenderer(
				bfchroma.WithoutAutodetect(),
				bfchroma.ChromaOptions(
					html.WithLineNumbers(false),
					html.WithClasses(true),
				),
				bfchroma.Extend(
					bf.NewHTMLRenderer(bf.HTMLRendererParameters{
						Flags: flags,
					}),
				),
				bfchroma.Style("monokai"),
			),
		),
		bf.WithExtensions(exts),
	)
}

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func createComment(db *gorm.DB, comment *posts.Comment) (*posts.Comment, error) {
	err := db.Set("l10n:locale", "en-US").Create(comment).Error
	return comment, err
}

func createOrUpdateCategory(db *gorm.DB, category *posts.Category) (*posts.Category, error) {
	var existingCategory posts.Category
	if db.Where("name = ?", category.Name).First(&existingCategory).RecordNotFound() {
		err := db.Set("l10n:locale", "en-US").Create(category).Error
		return category, err
	}
	category.ID = existingCategory.ID
	return category, db.Set("l10n:locale", "en-US").Save(category).Error
}

func openFileByURL(rawURL string) (*os.File, int64, error) {
	req, _ := grab.NewRequest(os.TempDir(), rawURL)
	if req == nil {
		return nil, 0, errors.New("----> could not make request.\n")
	}

	// start download
	log.Printf("----> Downloading %v...\n", req.URL())
	resp := clientGrab.Do(req)
	// pp.Println(resp)
	// fmt.Printf("  %v\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			log.Printf("---->  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size,
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		log.Printf("----> Download failed: %v\n", err)
		// os.Exit(1)
		return nil, 0, err
	}

	// fmt.Printf("----> Downloaded %v\n", rawURL)
	log.Printf("----> Download saved to %v \n", resp.Filename)
	fi, err := os.Stat(resp.Filename)
	if err != nil {
		return nil, 0, err
	}
	file, _ := os.Open(resp.Filename)

	return file, fi.Size(), nil

}

func openFileByURL2(rawURL string) (*os.File, int64, error) {
	if fileURL, err := url.Parse(rawURL); err != nil {
		return nil, 0, err
	} else {
		path := fileURL.Path
		segments := strings.Split(path, "/")
		extension := filepath.Ext(path)
		var fileName string
		if extension != "" {
			fileName = segments[len(segments)-1]
		} else {
			fileName = Fake.UserName() + ".png"
		}

		filePath := filepath.Join(os.TempDir(), fileName)

		if fi, err := os.Stat(filePath); err == nil {
			file, err := os.Open(filePath)
			return file, fi.Size(), err
		}

		file, err := os.Create(filePath)
		if err != nil {
			return file, 0, err
		}

		check := http.Client{
			CheckRedirect: func(r *http.Request, via []*http.Request) error {
				r.URL.Opaque = r.URL.Path
				return nil
			},
		}
		resp, err := check.Get(rawURL) // add a filter to check redirect
		if err != nil {
			return file, 0, err
		}
		defer resp.Body.Close()
		fmt.Printf("----> Downloaded %v\n", rawURL)

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			return file, 0, err
		}

		fi, err := file.Stat()
		if err != nil {
			return file, 0, err
		}

		return file, fi.Size(), nil
	}
}

func GetDB() *gorm.DB {
	return DB
}

func createOrUpdatePost(db *gorm.DB, post *posts.Post) (*posts.Post, error) {
	var existingPost posts.Post
	if db.Where("code = ?", post.Code).First(&existingPost).RecordNotFound() {
		err := db.Set("l10n:locale", "en-US").Create(post).Error
		return post, err
	}
	post.ID = existingPost.ID
	return post, db.Set("l10n:locale", "en-US").Save(post).Error
}

func createOrUpdateLink(db *gorm.DB, link *posts.Link) (*posts.Link, error) {
	var existingLink posts.Link
	if db.Where("href = ?", link.URL).First(&existingLink).RecordNotFound() {
		err := db.Set("l10n:locale", "en-US").Create(link).Error
		return link, err
	}
	link.ID = existingLink.ID
	return link, db.Set("l10n:locale", "en-US").Save(link).Error
}

func createOrUpdateTag(db *gorm.DB, tag *posts.Tag) (*posts.Tag, error) {
	var existingTag posts.Tag
	if db.Where("name = ?", tag.Name).First(&existingTag).RecordNotFound() {
		err := db.Set("l10n:locale", "en-US").Create(tag).Error
		return tag, err
	}
	tag.ID = existingTag.ID
	return tag, db.Set("l10n:locale", "en-US").Save(tag).Error
}

func removeDuplicates(elements []string) []string {
	// Use map to record duplicates as we find them.
	encountered := map[string]bool{}
	result := []string{}

	for v := range elements {
		// elements[v] = strings.ToLower(elements[v])
		if encountered[elements[v]] == true {
			// Do not add duplicate.
		} else {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}

func addslashes(str string) string {
	var tmpRune []rune
	strRune := []rune(str)
	for _, ch := range strRune {
		switch ch {
		case []rune{'\\'}[0], []rune{'"'}[0]:
			tmpRune = append(tmpRune, []rune{'\\'}[0])
			tmpRune = append(tmpRune, ch)
		default:
			tmpRune = append(tmpRune, ch)
		}
	}
	return string(tmpRune)
}

func escape(sql string) string {
	dest := make([]byte, 0, 2*len(sql))
	var escape byte
	for i := 0; i < len(sql); i++ {
		c := sql[i]

		escape = 0

		switch c {
		case 0: /* Must be escaped for 'mysql' */
			escape = '0'
			break
		case '\n': /* Must be escaped for logs */
			escape = 'n'
			break
		case '\r':
			escape = 'r'
			break
		case '\\':
			escape = '\\'
			break
		case '\'':
			escape = '\''
			break
		case '"': /* Better safe than sorry */
			escape = '"'
			break
		case '\032': /* This gives problems on Win32 */
			escape = 'Z'
		}

		if escape != 0 {
			dest = append(dest, '\\', escape)
		} else {
			dest = append(dest, c)
		}
	}

	return string(dest)
}

// HTTP GET timeout
const TIMEOUT = 20

func downloadAsOne(url, out string) (int64, error) {
	var client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 30,
		},
		Timeout: TIMEOUT * time.Second,
	}

	resp, err := client.Get(url)

	if err != nil {
		log.Println("Trouble making GET photo request!")
		return 0, err
	}

	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Trouble reading response body!")
		return 0, err
	}

	err = ioutil.WriteFile(out, contents, 0644)
	if err != nil {
		log.Println("Trouble creating file!")
		return 0, err
	}

	fi, err := os.Stat(out)
	if err != nil {
		return 0, err
	}
	// get the size
	size := fi.Size()

	fmt.Printf("The file is %d bytes long", fi.Size())
	return size, nil
}

func getFromBadger(key string) (resp []byte, ok bool) {
	err := store.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			// This func with val would only be called if item.Value encounters no error.
			// Accessing val here is valid.
			// fmt.Printf("The answer is: %s\n", val)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return resp, err == nil
}

func addToBadger(key, value string) error {
	err := store.Update(func(txn *badger.Txn) error {
		if debug {
			log.Println("indexing: ", key)
		}
		cnt, err := compress([]byte(value))
		if err != nil {
			return err
		}
		err = txn.Set([]byte(key), cnt)
		return err
	})
	return err
}

func compress(data []byte) ([]byte, error) {
	return snappy.Encode([]byte{}, data), nil
}

func decompress(data []byte) ([]byte, error) {
	return snappy.Decode([]byte{}, data)
}

func ensureDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		os.MkdirAll(path, os.FileMode(0755))
	} else {
		return err
	}
	d.Close()
	return nil
}

func followUsername(client *github.Client, username string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	waitForRemainingLimit(client, true, 10)
	resp, err := client.Users.Follow(ctx, username)
	if err != nil {
		bs, _ := ioutil.ReadAll(resp.Body)
		log.Errorf("follow %s err: %s [%s]", username, bs, err)
		return false, err
	}
	return true, nil
}

func starRepo(client *github.Client, owner, name string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	waitForRemainingLimit(client, true, 10)
	resp, err := client.Activity.Star(ctx, owner, name)
	if err != nil {
		bs, _ := ioutil.ReadAll(resp.Body)
		log.Errorf("star %s/%s err: %s [%s]", owner, name, bs, err)
		return false, err
	}
	return true, nil
}

func getInfo(client *github.Client, owner, name string) (*github.Repository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	waitForRemainingLimit(client, true, 10)
	info, _, err := client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func getTopics(client *github.Client, owner, name string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	waitForRemainingLimit(client, true, 10)
	topics, _, err := client.Repositories.ListAllTopics(ctx, owner, name)
	if err != nil {
		return nil, err
	}
	return topics, nil
}

func getReadme(client *github.Client, owner, name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	waitForRemainingLimit(client, true, 10)
	readme, _, err := client.Repositories.GetReadme(ctx, owner, name, nil)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	content, err := readme.GetContent()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return content, nil
}

func waitForRemainingLimit(cl *github.Client, isCore bool, minLimit int) {
	for {
		rateLimits, _, err := cl.RateLimits(context.Background())
		if err != nil {
			if debug {
				log.Printf("could not access rate limit information: %s\n", err)
			}
			<-time.After(time.Second * 1)
			continue
		}

		var rate int
		var limit int
		if isCore {
			rate = rateLimits.GetCore().Remaining
			limit = rateLimits.GetCore().Limit
		} else {
			rate = rateLimits.GetSearch().Remaining
			limit = rateLimits.GetSearch().Limit
		}

		if rate < minLimit {
			if debug {
				log.Printf("Not enough rate limit: %d/%d/%d\n", rate, minLimit, limit)
			}
			<-time.After(time.Second * 60)
			continue
		}
		if debug {
			log.Printf("Rate limit: %d/%d\n", rate, limit)
		}
		break
	}
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

	// createMediaLibraries()
	// fmt.Println("--> Created medialibraries.")

	// createPosts()
	// fmt.Println("--> Created posts.")

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
	globalSetting["SiteName"] = "QorPress Demo"
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
	totalCount := 100
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

			if file, _, err := openFileByURL2("https://i.pravatar.cc/150?u=" + unique); err != nil {
				fmt.Printf("open file failure, got err %v", err)
			} else {
				defer file.Close()
				user.Avatar.Scan(file)
			}

			if err := DraftDB.Save(&user).Error; err != nil {
				log.Warnf("Save user (%v) failure, got err %v", user, err)
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
	pp.Println("Seeds.Categories: ", Seeds.Categories)
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

func createMediaLibraries() {
	numberMedia := 100
	for i := 0; i < numberMedia; i++ {
		medialibrary := settings.MediaLibrary{}
		medialibrary.Title = loremIpsumGenerator.Words(10)

		if file, _, err := openFileByURL("https://loremflickr.com/640/360"); err != nil {
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
	if file, _, err := openFileByURL("http://qor3.s3.amazonaws.com/slide01.jpg"); err == nil {
		defer file.Close()
		topBannerValue.BackgroundImage.Scan(file)
	} else {
		fmt.Printf("open file (%q) failure, got err %v", "banner", err)
	}

	if file, _, err := openFileByURL("http://qor3.s3.amazonaws.com/qor_logo.png"); err == nil {
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
		if file, _, err := openFileByURL(s.Image); err == nil {
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
