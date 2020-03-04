package fixitycache

import (
	"fmt"
	"time"

	"github.com/leeola/fixity"
	"github.com/leeola/fixity/autoload"
)

const (
	defaultStoragePath string = "./shared/data/cache/.fixity"
)

/*
	Refs:
	- https://github.com/leeola/fixity
	- https://github.com/leeola/fixity/blob/master/_examples/local.toml
*/

type Cache struct {
	store            *fixity.Fixity
	storagePath      string
	snailIndexPath   string
	diskStorePath    string
	creatDirectories bool
}

type Config struct {
	StoragePath      string // An optional directory to base all relative paths in this config off of.
	SnailIndexPath   string // Where to write the index. If relative, this is a subdirectory of the rootPath
	DiskStorePath    string // Where to write the store. If relative, this is a subdirectory of the rootPath
	CreatDirectories bool   // Create any missing paths for the various stores/databases that may need to.
}

func New(config *Config) *Cache {

	if config == nil {
		config.StoragePath = defaultStoragePath
	}

	if config.StoragePath == "" {
		config.StoragePath = defaultStoragePath
	}

	db, err := autoload.LoadFixity(config.StoragePath)
	if err != nil {
		fmt.Println("Error occured while loading fixity")
	}
	return &Cache{
		store: db,
	}
}

// Mount returns a new Cache using the provided Fixity as underlying storage.
func Mount(store *fixity.Fixity) *Cache {
	return &Cache{store}
}

func (c *Cache) Get(key string) (resp []byte, ok bool) {
	return nil, false
}

func (c *Cache) Delete(key string) {
	return
}

func (c *Cache) Flush() {
	return
}

func (c *Cache) Close() {
	return
}
