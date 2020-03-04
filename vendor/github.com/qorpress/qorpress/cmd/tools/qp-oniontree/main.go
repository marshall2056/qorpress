package main

import (
	"fmt"
	stdioutil "io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/k0kubun/pp"
	"github.com/karrick/godirwalk"
	"github.com/onionltd/oniontree-tools/pkg/types/service"

	// "github.com/qorpress/qorpress/pkg/config"
	"github.com/qorpress/qorpress/pkg/models"
	"github.com/qorpress/qorpress/pkg/services"
)

var (
	db        *gorm.DB
	debugMode bool = true
)

func main() {
	db = services.Init()
	db.LogMode(true)
	dirWalkServices(db, "./shared/datasets/oniontree/tagged")
}

func dirWalkServices(db *gorm.DB, dirname string) {
	err := godirwalk.Walk(dirname, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if !de.IsDir() {
				parts := strings.Split(osPathname, "/")
				if debugMode {
					pp.Println(parts)
					// os.Exit(1)
					fmt.Printf("Type:%s osPathname:%s tag:%s\n", de.ModeType(), osPathname, parts[4])
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
				m := &models.Service{
					Name:        t.Name,
					Description: t.Description,
					Slug:        slug.Make(t.Name),
				}

				if err := db.Create(m).Error; err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				// add public keys
				for _, publicKey := range t.PublicKeys {
					pubKey := &models.PublicKey{
						UID:         publicKey.ID,
						UserID:      publicKey.UserID,
						Fingerprint: publicKey.Fingerprint,
						Description: publicKey.Description,
						Value:       publicKey.Value,
					}
					if _, err := createOrUpdatePublicKey(db, m, pubKey); err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
				}

				// add urls
				for _, url := range t.URLs {
					var urlExists models.URL
					u := &models.URL{Name: url}
					if db.Where("name = ?", url).First(&urlExists).RecordNotFound() {
						db.Create(&u)
						if debugMode {
							pp.Println(u)
						}
					}
					if _, err := createOrUpdateURL(db, m, u); err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

				}

				// add tags
				// check if tag already exists
				tag := &models.Tag{Name: parts[4]}
				var tagExists models.Tag
				if db.Where("name = ?", parts[4]).First(&tagExists).RecordNotFound() {
					db.Create(&tag)
					if debugMode {
						pp.Println(tag)
					}
				}

				if _, err := createOrUpdateTag(db, m, tag); err != nil {
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

func createOrUpdateTag(db *gorm.DB, svc *models.Service, tag *models.Tag) (bool, error) {
	var existingSvc models.Service
	if db.Where("slug = ?", svc.Slug).First(&existingSvc).RecordNotFound() {
		err := db.Create(svc).Error
		return err == nil, err
	}
	var existingTag models.Tag
	if db.Where("name = ?", tag.Name).First(&existingTag).RecordNotFound() {
		err := db.Create(tag).Error
		return err == nil, err
	}
	svc.ID = existingSvc.ID
	svc.CreatedAt = existingSvc.CreatedAt
	svc.Tags = append(svc.Tags, &existingTag)
	return false, db.Save(svc).Error
}

func findPublicKeyByUID(db *gorm.DB, uid string) *models.PublicKey {
	pubKey := &models.PublicKey{}
	if err := db.Where(&models.PublicKey{UID: uid}).First(pubKey).Error; err != nil {
		log.Fatalf("can't find public_key with uid = %q, got err %v", uid, err)
	}
	return pubKey
}

func createOrUpdatePublicKey(db *gorm.DB, svc *models.Service, pubKey *models.PublicKey) (bool, error) {
	var existingSvc models.Service
	if db.Where("slug = ?", svc.Slug).First(&existingSvc).RecordNotFound() {
		err := db.Create(svc).Error
		return err == nil, err
	}
	var existingPublicKey models.PublicKey
	if db.Where("uid = ?", pubKey.UID).First(&existingPublicKey).RecordNotFound() {
		err := db.Create(pubKey).Error
		return err == nil, err
	}
	svc.ID = existingSvc.ID
	svc.CreatedAt = existingSvc.CreatedAt
	svc.PublicKeys = append(svc.PublicKeys, &existingPublicKey)
	return false, db.Save(svc).Error
}

func createOrUpdateURL(db *gorm.DB, svc *models.Service, url *models.URL) (bool, error) {
	var existingSvc models.Service
	if db.Where("slug = ?", svc.Slug).First(&existingSvc).RecordNotFound() {
		err := db.Create(svc).Error
		return err == nil, err
	}
	var existingURL models.URL
	if db.Where("name = ?", url.Name).First(&existingURL).RecordNotFound() {
		err := db.Create(url).Error
		return err == nil, err
	}
	svc.ID = existingSvc.ID
	svc.CreatedAt = existingSvc.CreatedAt
	svc.URLs = append(svc.URLs, &existingURL)
	return false, db.Save(svc).Error
}
