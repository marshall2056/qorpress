package posts

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/qorpress/qorpress/core/l10n"
	"github.com/qorpress/qorpress/core/media/media_library"
	"github.com/qorpress/qorpress/core/publish2"
	qor_seo "github.com/qorpress/qorpress/core/seo"
	"github.com/qorpress/qorpress/core/slug"
	"github.com/qorpress/qorpress/core/sorting"
	"github.com/qorpress/qorpress/core/validations"
	"github.com/qorpress/qorpress/pkg/models/seo"
)

type Post struct {
	gorm.Model
	l10n.Locale
	sorting.SortingDESC
	Name         string    `gorm:"index:name"`
	NameWithSlug slug.Slug `l10n:"sync"`
	Featured     bool
	Code         string   `l10n:"sync" gorm:"index:code"`
	CategoryID   uint     `l10n:"sync" gorm:"index:category_id"`
	Category     Category `l10n:"sync"`
	// Categories []Category `gorm:"many2many:post_categories" l10n:"sync"`
	Collections    []Collection `l10n:"sync" gorm:"many2many:post_collections;"`
	Tags           []Tag        `l10n:"sync" gorm:"many2many:post_tags"`
	Comments       []Comment    `l10n:"sync"`
	Links          []Link       `l10n:"sync"` // gorm:"many2many:post_links"`
	MainImage      media_library.MediaBox
	Images         media_library.MediaBox
	Description    string         `gorm:"type:longtext"`
	Summary        string         `gorm:"type:mediumtext"`
	PostProperties PostProperties `sql:"type:text"`
	Seo            qor_seo.Setting
	publish2.Version
	publish2.Schedule
	publish2.Visible
}

type PostVariation struct {
	gorm.Model
	PostID   *uint
	Post     Post
	Featured bool
	Images   media_library.MediaBox
}

func (post Post) GetID() uint {
	return post.ID
}

func (post Post) GetSEO() *qor_seo.SEO {
	return seo.SEOCollection.GetSEO("Post Page")
}

func (post Post) DefaultPath() string {
	defaultPath := "/"
	return defaultPath
}

func (post Post) MainImageURL(styles ...string) string {
	style := "main"
	if len(styles) > 0 {
		style = styles[0]
	}

	if len(post.MainImage.Files) > 0 {
		return post.MainImage.URL(style)
	}

	return "/images/default_post.png"
}

func (post Post) Validate(db *gorm.DB) {
	if strings.TrimSpace(post.Name) == "" {
		db.AddError(validations.NewError(post, "Name", "Name can not be empty"))
	}

	if strings.TrimSpace(post.Code) == "" {
		db.AddError(validations.NewError(post, "Code", "Code can not be empty"))
	}
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

func (postImage *PostImage) GetMediaOption() (mediaOption media_library.MediaOption) {
	mediaOption.Video = postImage.File.Video
	mediaOption.FileName = postImage.File.FileName
	mediaOption.URL = postImage.File.URL()
	mediaOption.OriginalURL = postImage.File.URL("original")
	mediaOption.CropOptions = postImage.File.CropOptions
	mediaOption.Sizes = postImage.File.GetSizes()
	mediaOption.Description = postImage.File.Description
	return
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
