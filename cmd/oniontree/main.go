package main

import (
	"fmt"
	stdioutil "io/ioutil"
	"log"
	"os"
	"strings"

	// log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/goccy/go-yaml"
	"github.com/gosimple/slug"
	"github.com/jinzhu/gorm"
	"github.com/k0kubun/pp"
	"github.com/karrick/godirwalk"
	"github.com/onionltd/oniontree-tools/pkg/types/service"

	"github.com/qorpress/qorpress/pkg/config/db"
	"github.com/qorpress/qorpress/cmd/oniontree/models"
)

var (
	truncate bool
	displayHelp     bool
	dirname string
	debugMode  = true
	debugMode2 = true
	isTruncate = true
	DB         *gorm.DB
	tables     = []interface{}{
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
					Code: 		 slug.Make(t.Name),
				}

				if err := DB.Create(m).Error; err != nil {
					fmt.Println(err)
					os.Exit(1)
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
					if _, err := createOrUpdatePublicKey(DB, m, pubKey); err != nil {
						fmt.Println(err)
						os.Exit(1)
					}
				}

				// add urls
				for _, url := range t.URLs {
					u := &models.OnionLink{
						URL: url,
					}
					if _, err := createOrUpdateURL(DB, m, u); err != nil {
						fmt.Println(err)
						// os.Exit(1)
					}

				}

				// add tags
				// check if tag already exists
				tag := &models.OnionTag{Name: parts[4]}
				if _, err := createOrUpdateTag(DB, m, tag); err != nil {
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

func createOrUpdateTag(DB *gorm.DB, svc *models.OnionService, tag *models.OnionTag) (bool, error) {
	var existingSvc models.OnionService
	if DB.Where("slug = ?", svc.Slug).First(&existingSvc).RecordNotFound() {
		err := DB.Create(svc).Error
		return err == nil, err
	}
	var existingTag models.OnionTag
	if DB.Where("name = ?", tag.Name).First(&existingTag).RecordNotFound() {
		err := DB.Create(tag).Error
		return err == nil, err
	}
	svc.ID = existingSvc.ID
	svc.CreatedAt = existingSvc.CreatedAt
	svc.Tags = append(svc.Tags, &existingTag)
	return false, DB.Save(svc).Error
}

func findPublicKeyByUID(DB *gorm.DB, uid string) *models.OnionPublicKey {
	pubKey := &models.OnionPublicKey{}
	if err := DB.Where(&models.OnionPublicKey{UID: uid}).First(pubKey).Error; err != nil {
		log.Fatalf("can't find public_key with uid = %q, got err %v", uid, err)
	}
	return pubKey
}

func createOrUpdatePublicKey(DB *gorm.DB, svc *models.OnionService, pubKey *models.OnionPublicKey) (bool, error) {
	var existingSvc models.OnionService
	if DB.Where("slug = ?", svc.Slug).First(&existingSvc).RecordNotFound() {
		err := DB.Create(svc).Error
		return err == nil, err
	}
	var existingPublicKey models.OnionPublicKey
	if DB.Where("uid = ?", pubKey.UID).First(&existingPublicKey).RecordNotFound() {
		err := DB.Create(pubKey).Error
		return err == nil, err
	}
	svc.ID = existingSvc.ID
	svc.CreatedAt = existingSvc.CreatedAt
	svc.PublicKeys = append(svc.PublicKeys, &existingPublicKey)
	return false, DB.Save(svc).Error
}

func createOrUpdateURL(DB *gorm.DB, svc *models.OnionService, url *models.OnionLink) (bool, error) {
	var existingSvc models.OnionService
	if DB.Where("slug = ?", svc.Slug).First(&existingSvc).RecordNotFound() {
		err := DB.Create(svc).Error
		return err == nil, err
	}
	var existingURL models.OnionLink
	if DB.Where("url = ?", url.Name).First(&existingURL).RecordNotFound() {
		err := DB.Create(url).Error
		return err == nil, err
	}
	svc.ID = existingSvc.ID
	svc.CreatedAt = existingSvc.CreatedAt
	svc.Links = append(svc.Links, &existingURL)
	return false, DB.Save(svc).Error
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
