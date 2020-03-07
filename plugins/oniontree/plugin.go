package main

import (
	"github.com/qorpress/qorpress-contrib/oniontree/models"
)

var Tables = []interface{}{&models.PublicKey{}, &models.Service{}, &models.URL{}, &models.Tag{}}

func Migrate() []interface{} {
	return Tables
}