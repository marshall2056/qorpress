package services

import (
	"log"

	"github.com/dgraph-io/badger"
	"github.com/golang/snappy"

	"github.com/qorpress/qorpress/pkg/utils"
)

type StoreKV struct {
	db          *badger.DB
	storagePath string
}

func New(storagePath string) (*StoreKV, error) {
	err = utils.EnsureDir(storagePath)
	if err != nil {
		return nil, err
	}
	store, err = badger.Open(badger.DefaultOptions(storagePath))
	if err != nil {
		return nil, err
	}
	s := &StoreKV{
		db:          store,
		storagePath: storagePath,
	}
	return s, nil
}

func (kv *StoreKV) Get(key string) (resp []byte, ok bool) {
	err := kv.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) (err error) {
			resp, err = decompress(val)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return resp, err == nil
}

func (kv *StoreKV) Set(key, value string) error {
	err := kv.db.Update(func(txn *badger.Txn) error {
		if debug {
			log.Println("indexing: ", key)
		}
		cnt, err := compress([]byte(value))
		if err != nil {
			return err
		}
		err = txn.Set([]byte(key), cnt)
		return err
	})
	return err
}

func compress(data []byte) ([]byte, error) {
	return snappy.Encode([]byte{}, data), nil
}

func decompress(data []byte) ([]byte, error) {
	return snappy.Decode([]byte{}, data)
}
