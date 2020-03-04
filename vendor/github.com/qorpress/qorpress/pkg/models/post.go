package models

import (
	"log"
	"strings"
	"time"
	"errors"
	"encoding/json"
	"database/sql/driver"

	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/qorpress/l10n"
	"github.com/qorpress/media/media_library"
	"github.com/qorpress/publish2"
	qor_seo "github.com/qorpress/seo"
	qor_slug "github.com/qorpress/slug"
	"github.com/qorpress/sorting"
	"github.com/qorpress/validations"
)

//go:generate gp-extender -structs Post -output post-funcs.go
type Post struct {
	// Categories []Category `gorm:"many2many:category_post"`
	ID           uint     `gorm:"primary_key"`
	// CategoryID   uint     `l10n:"sync"`
	// Category     Category `l10n:"sync"`
	Categories []Category `gorm:"many2many:category_post" l10n:"sync"`
	Tags         []Tag    `l10n:"sync" gorm:"many2many:tag_post"`
	Title        string   `l10n:"sync" gorm:"type:mediumtext"`
	UUID         string   `l10n:"sync" gorm:"unique"`
	MainImage    media_library.MediaBox
	Body         string     `gorm:"type:longtext"`
	Summary      string     `gorm:"type:longtext"`
	Images       []Image    `l10n:"sync"`
	Documents    []Document `l10n:"sync"`
	Videos       []Video    `l10n:"sync"`
	Links        []Link     `l10n:"sync"`
	Type         string     `l10n:"sync"`
	Created      int32
	Updated      int32
	Author       User
	AuthorID     uint
	PostProperties     PostProperties `sql:"type:text"`
	NameWithSlug qor_slug.Slug `l10n:"sync"`
	Seo          qor_seo.Setting
	publish2.Version
	publish2.Schedule
	sorting.SortingDESC
	publish2.Visible
	l10n.Locale
}

func (p Post) GetSEO() *qor_seo.SEO {
	return SEOCollection.GetSEO("Post")
}

func (p Post) DefaultPath() string {
	defaultPath := "/"
	//if len(product.ColorVariations) > 0 {
	//	defaultPath = fmt.Sprintf("/post/%s_%s", p.Code, product.ColorVariations[0].ColorCode)
	//}
	return defaultPath
}

func (p Post) MainImageURL(styles ...string) string {
	style := "main"
	if len(styles) > 0 {
		style = styles[0]
	}
	if len(p.MainImage.Files) > 0 {
		return p.MainImage.URL(style)
	}
	return "/images/default_post.png"
}

func (p Post) Validate(db *gorm.DB) {
	if strings.TrimSpace(p.Title) == "" {
		db.AddError(validations.NewError(p, "Title", "Title can not be empty"))
	}
}

func (p *Post) BeforeCreate() (err error) {
	if p.Created == 0 {
		p.Created = int32(time.Now().Unix())
	}
	if p.Updated == 0 {
		p.Updated = int32(time.Now().Unix())
	}

	p.LanguageCode = "en-US"
	p.PublishReady = true

	log.Printf("======> New post: %#v\n", p.Title)
	log.Printf("======> New post: %#v\n", p.Summary)
	log.Printf("======> New post: %#v\n", p.Images)
	if len(p.Images) > 0 {
		log.Printf("=======> Images: %#v\n", p.Images[0].File.FileName)
	}

	for i := range p.Images {
		p.Images[i].File.Sizes = p.Images[i].GetSizes()
		file, err := p.Images[i].File.Base.FileHeader.Open()
		if err != nil {
			log.Fatal(err)
		}
		p.Images[i].File.Scan(file)
	}
	return nil
}

type PostProperties []PostProperty

type PostProperty struct {
	Name  string
	Value string
}

func (postProperties *PostProperties) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, postProperties)
	case string:
		if v != "" {
			return postProperties.Scan([]byte(v))
		}
	default:
		return errors.New("not supported")
	}
	return nil
}

func (postProperties PostProperties) Value() (driver.Value, error) {
	if len(postProperties) == 0 {
		return nil, nil
	}
	return json.Marshal(postProperties)
}

type PostImage struct {
	gorm.Model
	Title        string
	Category     Category
	CategoryID   uint
	SelectedType string
	File         media_library.MediaLibraryStorage `sql:"size:4294967295;" media_library:"url:/system/{{class}}/{{primary_key}}/{{column}}.{{extension}}"`
}

func (postImage PostImage) Validate(db *gorm.DB) {
	if strings.TrimSpace(postImage.Title) == "" {
		db.AddError(validations.NewError(postImage, "Title", "Title can not be empty"))
	}
}

func (postImage *PostImage) SetSelectedType(typ string) {
	postImage.SelectedType = typ
}

func (postImage *PostImage) GetSelectedType() string {
	return postImage.SelectedType
}

func (postImage *PostImage) ScanMediaOptions(mediaOption media_library.MediaOption) error {
	if bytes, err := json.Marshal(mediaOption); err == nil {
		return postImage.File.Scan(bytes)
	} else {
		return err
	}
}

func createUniqueSlug(title string) string {
	slugTitle := slug.Make(title)
	if len(slugTitle) > 50 {
		slugTitle = slugTitle[:50]
		if slugTitle[len(slugTitle)-1:] == "-" {
			slugTitle = slugTitle[:len(slugTitle)-1]
		}
	}
	return slugTitle
}
