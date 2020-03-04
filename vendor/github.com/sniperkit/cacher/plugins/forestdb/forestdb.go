package forestdb

import (
	"github.com/dgraph-io/badger"
)

// Cache stores and retrieves data using Badger KV.
type Cache struct {
	db *badger.DB
}
