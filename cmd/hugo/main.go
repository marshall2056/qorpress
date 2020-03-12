package main

import (
	"os"
	"log"

	"github.com/jinzhu/gorm"
	"github.com/k0kubun/pp"
	"github.com/spf13/pflag"

	// "github.com/jinzhu/configor"
	// "github.com/karrick/godirwalk"
	// "github.com/gohugoio/hugo/hugolib"
	// "github.com/k0kubun/pp"
	// "github.com/pkg/errors"
	// "github.com/qorpress/qorpress/cmd/hugo/models"
	"github.com/qorpress/qorpress/pkg/config/db"
	"github.com/qorpress/qorpress/pkg/models/posts"
)

var (
	truncate    bool
	displayHelp bool
	dirname     string
	debugMode   bool
	isTruncate  = true
	DB          *gorm.DB
	tables      = []interface{}{
		&posts.Post{},
		&posts.Tag{},
	}
)

func main() {
	pflag.StringVarP(&dirname, "source", "s", "./examples/blog", "directory with hugo files.")
	pflag.BoolVarP(&truncate, "truncate", "t", false, "truncate tables")
	pflag.BoolVarP(&debugMode, "debug", "d", false, "truncate tables")
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
	scanFromSite(dirname)
}

// builds the search index by passing all pages of hugo site that have a title to the indexer
func scanFromSite(theHugoPath string) {
	pages := readSitePages(theHugoPath)
	for _, page := range pages {
		// TODO: home page has no title, are we properly reading the config file ?
		if pageHasTitle(page) && pageHasValidContent(page) {
			entry := newEntry(page)
			pp.Println(entry)
		}
	}
}

func checkFatal(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}

func TruncateTables(DB *gorm.DB, tables ...interface{}) {
	for _, table := range tables {
		if debugMode {
			pp.Println(table)
		}
		if err := DB.DropTableIfExists(table).Error; err != nil {
			panic(err)
		}
		// DB.AutoMigrate(table)
	}
}
