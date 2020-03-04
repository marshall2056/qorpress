package pivotcache

import (
	"github.com/ghetzel/pivot"
	pivot_backends "github.com/ghetzel/pivot/backends"
	pivot_dal "github.com/ghetzel/pivot/dal"
	pivot_filter "github.com/ghetzel/pivot/filter"
	pivot_mapper "github.com/ghetzel/pivot/mapper"
	// "github.com/sniperkit/httpcache/helpers"
)

// Cache stores and retrieves data using Badger KV.
type Cache struct {
	db         *pivot_backends.Backend
	mapper     *pivot_mapper.Mapper
	collection *pivot_dal.Collection
	filters    *pivot_filter.Filter
	// connectOptions *pivot_backends.ConnectOptions
}

func Attach(backend *pivot_backends.Backend) *Cache {
	return &Cache{db: backend}
}

func New(dsn string) *Cache {
	return &Cache{
		db: pivot.NewDatabase(dsn),
	}
}

func NewWithOptions(dsn string, options backends.ConnectOptions) *Cache {
	return &Cache{
		db: pivot.NewDatabaseWithOptions(dsn, options),
		// connectOptions: options,
	}
}
