package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/qorpress/qorpress/core/media/media_library"
	qor_seo "github.com/qorpress/qorpress/core/seo"
	"github.com/qorpress/qorpress/core/slug"
	"github.com/qorpress/qorpress/core/validations"
	"github.com/qorpress/qorpress/pkg/models/seo"
)

type OnionService struct {
	gorm.Model
	Name              string    `json:"name" yaml:"name"`
	NameWithSlug      slug.Slug `l10n:"sync"`
	MainImage         media_library.MediaBox
	Featured          bool
	CategoryID        uint              `l10n:"sync" gorm:"index:category_id"`
	Category          OnionCategory     `l10n:"sync"`
	Code              string            `l10n:"sync" gorm:"index:code"`
	Slug              string            `json:"slug,omitempty" yaml:"slug,omitempty"`
	Description       string            `gorm:"type:mediumtext" json:"description,omitempty" yaml:"description,omitempty"`
	Summary           string            `gorm:"type:mediumtext"`
	ServiceProperties ServiceProperties `sql:"type:text"`
	Links             []*OnionLink      `l10n:"sync" json:"urls,omitempty" yaml:"urls,omitempty"`
	PublicKeys        []*OnionPublicKey `l10n:"sync" json:"public_keys,omitempty" yaml:"public_keys,omitempty"`
	Tags              []*OnionTag       `l10n:"sync" gorm:"many2many:service_tags;" json:"tags,omitempty" yaml:"tags,omitempty"`
	Seo               qor_seo.Setting
}

func (o OnionService) GetID() uint {
	return o.ID
}

func (o OnionService) GetSEO() *qor_seo.SEO {
	return seo.SEOCollection.GetSEO("OnionService Page")
}

func (o OnionService) DefaultPath() string {
	defaultPath := "/"
	return defaultPath
}

func (o OnionService) Validate(db *gorm.DB) {
	if strings.TrimSpace(o.Name) == "" {
		db.AddError(validations.NewError(o, "Name", "Name can not be empty"))
	}

	if strings.TrimSpace(o.Code) == "" {
		db.AddError(validations.NewError(o, "Code", "Code can not be empty"))
	}
}

type ServiceProperties []ServiceProperty

type ServiceProperty struct {
	Name  string
	Value string
}

func (serviceProperties *ServiceProperties) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, serviceProperties)
	case string:
		if v != "" {
			return serviceProperties.Scan([]byte(v))
		}
	default:
		return errors.New("not supported")
	}
	return nil
}

func (serviceProperties ServiceProperties) Value() (driver.Value, error) {
	if len(serviceProperties) == 0 {
		return nil, nil
	}
	return json.Marshal(serviceProperties)
}
