package models

import (
	"github.com/jinzhu/gorm"
)

type OnionPublicKey struct {
	gorm.Model
	UID         string `gorm:"primary_key" json:"id,omitempty" yaml:"id,omitempty"`
	UserID      string `json:"user_id,omitempty" yaml:"user_id,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty" yaml:"fingerprint,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Value       string `json:"value" yaml:"value"`
	ServiceID   uint   `json:"-" yaml:"-"`
}
