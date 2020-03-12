package main

import (
	"fmt"
	stdioutil "io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"

	// log "github.com/sirupsen/logrus"
	"github.com/goccy/go-yaml"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/k0kubun/pp"
	"github.com/karrick/godirwalk"
	"github.com/onionltd/oniontree-tools/pkg/types/service"
	"github.com/spf13/pflag"

	"github.com/qorpress/qorpress/cmd/oniontree/models"
	"github.com/qorpress/qorpress/pkg/config/db"
)

var (
	truncate    bool
	displayHelp bool
	dirname     string
	debugMode   = true
	isTruncate  = true
	DB          *gorm.DB
	tables      = []interface{}{
		&models.OnionTag{},
		&models.OnionService{},
		&models.OnionPublicKey{},
		&models.OnionLink{},
	}
)

/*
go run cmd/oniontree/main.go --dirname ./cmd/oniontree/data
*/

func main() {

	pflag.StringVarP(&dirname, "dirname", "d", "./data/tagged", "directory with onion yaml files.")
	pflag.BoolVarP(&truncate, "truncate", "t", false, "truncate tables")
	pflag.BoolVarP(&displayHelp, "help", "h", false, "help info")
	pflag.Parse()
	if displayHelp {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	DB = db.DB

	if truncate {
		TruncateTables(DB, tables...)
	}

	// getWorkTree(db)
	// os.Exit(1)
	dirWalkServices(DB, dirname)
}

func dirWalkServices(DB *gorm.DB, dirname string) {
	err := godirwalk.Walk(dirname, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if !de.IsDir() {
				parts := strings.Split(osPathname, "/")
				if debugMode {
					fmt.Printf("Type:%s osPathname:%s tag:%s\n", de.ModeType(), osPathname, parts[1])
					pp.Println(parts)
					// os.Exit(1)
				}
				bytes, err := stdioutil.ReadFile(osPathname)
				if err != nil {
					return err
				}
				t := service.Service{}
				yaml.Unmarshal(bytes, &t)
				if debugMode {
					pp.Println(t)
				}

				// add service
				m := &models.OnionService{
					Name:        t.Name,
					Description: t.Description,
					Slug:        slug.Make(t.Name),
					Code:        slug.Make(t.Name),
				}

				if err := DB.Create(m).Error; err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				// add urls
				for _, url := range t.URLs {
					u := &models.OnionLink{URL: url}
					if l, err := createOrUpdateLink(DB, u); err != nil {
						fmt.Println(err)
						os.Exit(1)
					} else {
						m.Links = append(m.Links, *l)
					}
				}

				// add tags
				// check if tag already exists
				tag := &models.OnionTag{Name: parts[4]}
				if t, err := createOrUpdateTag(DB, tag); err != nil {
					fmt.Println(err)
					os.Exit(1)
				} else {
					m.Tags = append(m.Tags, *t)
				}

				// add public keys
				for _, publicKey := range t.PublicKeys {
					pubKey := &models.OnionPublicKey{
						UID:         publicKey.ID,
						UserID:      publicKey.UserID,
						Fingerprint: publicKey.Fingerprint,
						Description: publicKey.Description,
						Value:       publicKey.Value,
					}
					if k, err := createOrUpdatePublicKey(DB, pubKey); err != nil {
						fmt.Println(err)
						os.Exit(1)
					} else {
						m.PublicKeys = append(m.PublicKeys, *k)
					}
				}

				if err := DB.Save(m).Error; err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

			}
			return nil
		},
		Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
	})
	if err != nil {
		log.Fatal(err)
	}
}

func createOrUpdateTag(DB *gorm.DB, tag *models.OnionTag) (*models.OnionTag, error) {
	var existingTag models.OnionTag
	if DB.Where("name = ?", tag.Name).First(&existingTag).RecordNotFound() {
		err := DB.Create(tag).Error
		return tag, errors.Wrap(err, "create tag failed")
	}
	tag.ID = existingTag.ID
	tag.CreatedAt = existingTag.CreatedAt
	return tag, DB.Save(tag).Error
}

func createOrUpdatePublicKey(DB *gorm.DB, pubKey *models.OnionPublicKey) (*models.OnionPublicKey, error) {
	var existingPublicKey models.OnionPublicKey
	if DB.Where("uid = ?", pubKey.UID).First(&existingPublicKey).RecordNotFound() {
		err := DB.Create(pubKey).Error
		return pubKey, errors.Wrap(err, "create public key failed")
	}
	pubKey.ID = existingPublicKey.ID
	// pubKey.CreatedAt = existingPublicKey.CreatedAt
	return pubKey, DB.Save(pubKey).Error
}

func createOrUpdateLink(DB *gorm.DB, link *models.OnionLink) (*models.OnionLink, error) {
	var existingLink models.OnionLink
	if DB.Where("url = ?", link.URL).First(&existingLink).RecordNotFound() {
		err := DB.Create(link).Error
		return link, errors.Wrap(err, "create link failed")
	}
	link.ID = existingLink.ID
	link.CreatedAt = existingLink.CreatedAt
	return link, DB.Save(link).Error
}

func findPublicKeyByUID(DB *gorm.DB, uid string) *models.OnionPublicKey {
	pubKey := &models.OnionPublicKey{}
	if err := DB.Where(&models.OnionPublicKey{UID: uid}).First(pubKey).Error; err != nil {
		log.Fatalf("can't find public_key with uid = %q, got err %v", uid, err)
	}
	return pubKey
}

func TruncateTables(DB *gorm.DB, tables ...interface{}) {
	for _, table := range tables {
		if debugMode {
			pp.Println(table)
		}
		if err := DB.DropTableIfExists(table).Error; err != nil {
			panic(err)
		}
		DB.AutoMigrate(table)
	}
}
