package constants

import (
	"fmt"
	"path/filepath"
	"time"

	homedir "github.com/mitchellh/go-homedir"
)

const (
	DefaultMaxConnections int    = 2000 // throttle cache requests to 2000 QPS
	DefaultCacheMode      string = "rw" // rw (Read/Write), w (Ready Only)
	DefaultCachePath      string = "sniperkit/shared/data/cache/"
)

var (
	DefaultCacheDir          string        = "~"
	DefaultCacheDuration     time.Duration = 30 * 24 * time.Hour  // 30 Day
	DefaultCacheDurationLong time.Duration = 365 * 24 * time.Hour // 1 Year
)

func init() {
	homeDir, err := homedir.Dir()
	if err != nil {
		fmt.Println("Could not find/expand the home directory")
	}
	DefaultCacheDir = filepath.Join(homeDir, DefaultCachePath)
}
