package constants

import (
	"errors"
)

var (
	// ErrCacheKeyNotFound is returned whenever there's a cache miss
	ErrCacheKeyNotFound = errors.New("Cache key not found, cache miss")
)
