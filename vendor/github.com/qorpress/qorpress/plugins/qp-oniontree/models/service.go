package models

import (
	"github.com/jinzhu/gorm"
)

//go:generate gp-extender -structs Service -output service-funcs.go
type Service struct {
	gorm.Model
	Name        string       `json:"name" yaml:"name"`
	Slug        string       `json:"slug,omitempty" yaml:"slug,omitempty"`
	Description string       `json:"description,omitempty" yaml:"description,omitempty"`
	URLs        []*URL       `json:"urls,omitempty" yaml:"urls,omitempty"`
	PublicKeys  []*PublicKey `json:"public_keys,omitempty" yaml:"public_keys,omitempty"`
	Tags        []*Tag       `gorm:"many2many:service_tags;" json:"tags,omitempty" yaml:"tags,omitempty"`
}
