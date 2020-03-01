package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/corpix/uarand"
	badger "github.com/dgraph-io/badger"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
	"github.com/golang/snappy"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/parser"
	"github.com/google/go-github/v29/github"
	"github.com/h2non/filetype"
	"github.com/iancoleman/strcase"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/joho/godotenv"
	"github.com/k0kubun/pp"
	"github.com/nozzle/throttler"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/qorpress/auth/auth_identity"
	i18n_database "github.com/qorpress/i18n/backends/database"
	"github.com/qorpress/l10n"
	"github.com/qorpress/media"
	"github.com/qorpress/media/asset_manager"
	"github.com/qorpress/media/media_library"
	"github.com/qorpress/oss/filesystem"
	"github.com/qorpress/publish2"
	"github.com/qorpress/seo"
	"github.com/qorpress/slug"
	"github.com/qorpress/sorting"
	"github.com/qorpress/validations"
	log "github.com/sirupsen/logrus"
	"github.com/x0rzkov/go-vcsurl"

	ghclient "github.com/qorpress/qorpress/pkg/client"
	"github.com/qorpress/qorpress/pkg/models"
)

var (
	clientManager *ghclient.ClientManager
	clientGH      *ghclient.GHClient
	store         *badger.DB
	DB            *gorm.DB
	storage       *filesystem.FileSystem
	cachePath     = "./shared/data/httpcache"
	storagePath   = "./shared/data/badger"
	debug         = false
	isSelenium    = false
	isTwitter     = false
	isFollow      = true
	isStar        = true
	isReadme      = true
	logLevelStr   = "info"
	maxTweetLen   = 280
	addMedia      = false
)

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

func main() {

	DB = InitDB()
	m := cmap.New()

	err := removeContents("./public/content/images/")
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
		// knownUrls = append(knownUrls, e.Text)
		q.AddURL(e.Text)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
		// knownUrls = append(knownUrls, e.Text)
		q.AddURL(e.Text)
	})

	c.OnHTML("div.blog-posts.hfeed", func(e *colly.HTMLElement) {
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
	// log.Println("Collected", len(githubUrls), "URLs")
	log.Println("Collected cmap: ", m.Count(), "URLs")

	t := throttler.New(2, m.Count())

	// storage = filesystem.New("./public")

	createSeo(DB)
	createCategories(DB)

	// counter := 1
	// counterFollow := 1
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

				// pp.Println(repoInfo)
				// os.Exit(1)

				readme, err := getReadme(clientGH.Client, info.Username, info.Name)
				if err != nil {
					log.Warnln(err)
					return err
				}

				// youtube
				// `^((?:https?:)?\/\/)?((?:www|m)\.)?((?:youtube\.com|youtu.be))(\/(?:[\w\-]+\?v=|embed\/|v\/)?)([\w\-]+)(\S+)?$`

				videoPatternRegexp, err := regexp.Compile(`^(http:\/\/|https:\/\/)(vimeo\.com|youtu\.be|www\.youtube\.com)\/([\w\/]+)([\?].*)?$`)
				if err != nil {
					log.Warnln(err)
					return err
				}
				vidLinks := videoPatternRegexp.FindAllString(readme, -1)
				videoLinks = append(videoLinks, vidLinks...)

				// pp.Println(readme)
				imgPatternRegexp, err := regexp.Compile(`(http(s?):)([/|.|\w|\s|-])*\.(?:jpg|gif|GIF|png|PNG|jpeg|JPG|JPEG)`)
				// imgPatternRegexp, err := regexp.Compile(`(http(s?):)([/|.|\w|\s|-])*\.(?:gif|GIF)`)
				if err != nil {
					log.Warnln(err)
					return err
				}
				imgLinks = imgPatternRegexp.FindAllString(readme, -1)
				imgRelRegexp, err := regexp.Compile(`([/|.|\w|\s|-])*\.(?:jpg|gif|png|PNG|GIF|jpeg|JPG|JPEG)`)
				if err != nil {
					log.Warnln(err)
					return err
				}
				imgLinksRel := imgRelRegexp.FindAllString(readme, -1)
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
				}
				imgLinks = append(imgLinks, imgLinksRel...)
				imgLinks = removeDuplicates(imgLinks)
				// pp.Println(imgLinksRel)
				// pp.Println(imgLinks)

				var title, desc string
				if repoInfo.Description != nil {
					desc = strings.TrimSpace(*repoInfo.Description)
					if len(desc) > 255 {
						title = desc[0:255]
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

				// pp.Println("extTopics: ", extTopics)

				// unsafe := blackfriday.Run([]byte(readme))
				// html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

				// extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.Tables | parser.FencedCode | parser.Mmark
				parser := parser.NewWithExtensions(parser.CommonExtensions)

				if readme == "" {
					return nil
				}

				// md := []byte("## markdown document")
				html := markdown.ToHTML([]byte(readme), parser, nil)

				start := time.Now().AddDate(0, 0, 0)
				end := time.Now().AddDate(12, 0, 0)

				p := &models.Post{
					Title:   *repoInfo.Name,
					UUID:    "github-" + info.Username + "-" + info.Name,
					Body:    string(html),
					Summary: desc,
					Type:    "article",
					Links: []models.Link{
						models.Link{
							Url:   key,
							Title: desc,
						},
					},
					NameWithSlug: slug.Slug{p.NameWithSlug},
				}

				p.LanguageCode = "en-US"
				p.SetPublishReady(true)
				p.SetVersionName("v1")
				p.SetScheduledStartAt(&start)
				p.SetScheduledEndAt(&end)

				var tags []*models.Tag
				for _, extTopic := range extTopics {
					tag := &models.Tag{
						Name: extTopic,
						// LanguageCode: "en-US",
					}
					tags = append(tags, tag)
				}

				var post *models.Post
				post, err = createOrUpdatePost(DB, p)
				if err != nil {
					log.Warnln(err)
				}

				postId := post.ID

				for _, tag := range tags {
					t, err := createOrUpdateTag(DB, tag)
					if err != nil {
						log.Fatalln(err)
					}
					post.Tags = append(post.Tags, *t)
				}

				// for _, video := range videoLinks {
				// }

				if len(videoLinks) > 0 {
					pp.Println(videoLinks)
					os.Exit(1)
				}

				for _, img := range imgLinks {
					var image models.Image
					file, size, err := openFileByURL(img)
					if err != nil {
						file.Close()
						continue
					}

					head := make([]byte, 261)
					file.Read(head)

					if filetype.IsImage(head) {
						log.Println("File is an image")
					} else {
						log.Println("Not an image")
						continue
					}
					image.File.Scan(file)
					image.CreatedAt = time.Now()
					image.UpdatedAt = time.Now()
					image.PostID = postId
					err = DB.Create(&image).Error
					if err != nil {
						file.Close()
						continue
					}

					// pp.Println("image.ID: ", strconv.Itoa(int(image.ID)), "size=", size)
					// newPath := "./content/images/" + strconv.Itoa(int(image.ID)) + "/" + image.File.FileName
					// storage.Put(newPath, file)
					// image.File.Url = "/public/content/images/" + strconv.Itoa(int(image.ID)) + "/" + image.File.FileName
					// err = DB.Save(&image).Error
					// if err != nil {
					//	log.Fatalln(err)
					//	file.Close()
					//	continue
					// }

					if size > 20000 {
						post.Images = append(post.Images, image)
					}
					file.Close()
				}

				/*
					var keys []int
					var mainKey int
					for k := range imgKeys {
						keys = append(keys, k)
					}
					// sort.Ints(keys)
					sort.Sort(sort.Reverse(sort.IntSlice(keys)))
					// fmt.Println("Value:", imgKeys[keys[0]])
					if len(keys) > 0 {
						mainKey = imgKeys[keys[0]]
					}
				*/

				if len(post.Images) > 0 {
					post.MainImage.Files = []media_library.File{{
						ID:  json.Number(fmt.Sprint(post.Images[0].ID)),
						Url: post.Images[0].File.URL(),
					}}
				}
				err = DB.Save(post).Error
				if err != nil {
					log.Warnln("save: ", err)
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
	mysqlString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=True&loc=Local&charset=utf8mb4,utf8", "root", os.Getenv("DB_PASSWORD"), "127.0.0.1", "3306", "gopress")

	//psqlInfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable", host, port, user, dbname, password)
	db, err := gorm.Open("mysql", mysqlString)
	if err != nil {
		panic(err)
	}
	db.LogMode(false)
	DB = db

	query := `SELECT Concat("TRUNCATE TABLE ", table_name) as query FROM INFORMATION_SCHEMA.TABLES WHERE table_schema = "gopress" AND TABLE_NAME!="";`
	rows, err := db.Raw(query).Rows()
	if err != nil {
		log.Fatalln("could not fetch all tables: ", err)
	}
	defer rows.Close()
	for rows.Next() {
		var query string
		rows.Scan(&query)
		pp.Println(query)
		db.Raw(query)
	}

	var post models.Post
	var video []models.Video
	var image []models.Image
	var link []models.Link
	var documents []models.Document

	// TruncateTables(&models.Post{}, &models.Tag{}, &models.Category{})
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&models.Event{})
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&models.Category{})
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&models.Post{})
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&models.Tag{})
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&models.Document{})
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&models.Video{})
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&models.Image{})
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&models.Link{})
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&models.MediaLibrary{})
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&asset_manager.AssetManager{})
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&i18n_database.Translation{})
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&auth_identity.AuthIdentity{})
	DB.Set("gorm:table_options", "CHARSET=utf8mb4").AutoMigrate(&media_library.MediaLibrary{})

	DB.Model(&post).Related(&video)
	DB.Model(&post).Related(&image)
	DB.Model(&post).Related(&link)
	DB.Model(&post).Related(&documents)

	media.RegisterCallbacks(DB)
	l10n.RegisterCallbacks(DB)
	sorting.RegisterCallbacks(DB)
	validations.RegisterCallbacks(DB)
	media.RegisterCallbacks(DB)
	publish2.RegisterCallbacks(DB)

	return DB
}

func createCategories(db *gorm.DB) error {
	categories := []string{"article", "publication", "blog", "video", "press_release", "event", "news"}
	for _, category := range categories {
		c := &models.Category{
			Name: category,
		}
		if _, err := createOrUpdateCategory(db, c); err != nil {
			return err
		}
	}
	return nil
}

func createOrUpdateCategory(db *gorm.DB, category *models.Category) (*models.Category, error) {
	var existingCategory models.Category
	if db.Where("name = ?", category.Name).First(&existingCategory).RecordNotFound() {
		err := db.Set("l10n:locale", "en-US").Create(category).Error
		return category, err
	}
	category.ID = existingCategory.ID
	return category, db.Set("l10n:locale", "en-US").Save(category).Error
}

func createSeo(db *gorm.DB) error {
	globalSeoSetting := models.SEOSetting{}
	globalSetting := make(map[string]string)
	globalSetting["SiteName"] = "Qor Demo"
	globalSeoSetting.Setting = seo.Setting{
		GlobalSetting: globalSetting,
	}
	globalSeoSetting.Name = "QorSeoGlobalSettings"
	globalSeoSetting.LanguageCode = "en-US"
	globalSeoSetting.QorSEOSetting.SetIsGlobalSEO(true)
	if err := db.Create(&globalSeoSetting).Error; err != nil {
		return err
		// log.Fatalf("create seo (%v) failure, got err %v", globalSeoSetting, err)
	}

	defaultSeo := models.SEOSetting{}
	defaultSeo.Setting = seo.Setting{
		Title:       "{{SiteName}}",
		Description: "{{SiteName}} - Default Description",
		Keywords:    "{{SiteName}} - Default Keywords",
		Type:        "Default Page",
	}
	defaultSeo.Name = "Default Page"
	defaultSeo.LanguageCode = "en-US"
	if err := db.Create(&defaultSeo).Error; err != nil {
		return err
		// log.Fatalf("create seo (%v) failure, got err %v", defaultSeo, err)
	}

	postSeo := models.SEOSetting{}
	postSeo.Setting = seo.Setting{
		Title:       "{{SiteName}}",
		Description: "{{SiteName}} - {{Name}} - {{Code}}",
		Keywords:    "{{SiteName}},{{Name}},{{Code}}",
		Type:        "Post",
	}
	postSeo.Name = "Post"
	postSeo.LanguageCode = "en-US"
	if err := db.Create(&postSeo).Error; err != nil {
		return err
		// log.Fatalf("create seo (%v) failure, got err %v", postSeo, err)
	}
	return nil
}

func openFileByURL(rawURL string) (*os.File, int64, error) {
	if fileURL, err := url.Parse(rawURL); err != nil {
		return nil, 0, err
	} else {
		path := fileURL.Path
		segments := strings.Split(path, "/")
		fileName := segments[len(segments)-1]

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

func TruncateTables(tables ...interface{}) {
	for _, table := range tables {
		if err := DB.DropTableIfExists(table).Error; err != nil {
			panic(err)
		}
		// DB.AutoMigrate(table)
	}
}

func GetDB() *gorm.DB {
	return DB
}

func createOrUpdatePost(db *gorm.DB, post *models.Post) (*models.Post, error) {
	// post.Images = images
	var existingPost models.Post
	if db.Where("uuid = ?", post.UUID).First(&existingPost).RecordNotFound() {
		err := db.Set("l10n:locale", "en-US").Create(post).Error
		return post, err
	}
	post.ID = existingPost.ID
	return post, db.Set("l10n:locale", "en-US").Save(post).Error
}

func createOrUpdateTag(db *gorm.DB, tag *models.Tag) (*models.Tag, error) {
	// post.Images = images
	var existingTag models.Tag
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
		elements[v] = strings.ToLower(elements[v])
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
