package boltdbttlcache

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/k0kubun/pp"
	log "github.com/sirupsen/logrus"
)

const (
	bktGroupsName               string = "__groups"
	bktDefaultGroupName         string = "default"
	bktName                     string = "httpcache"
	bktDir                      string = "./shared/data/cache/.bolt-ttl"
	defaultStorageFileExtension string = ".bttl"
)

var (
	defaultCacheTTL    time.Duration = time.Duration(24) * time.Hour // 1 Day
	defaultStorageDir  string        = filepath.Join(bktDir, bktName)
	defaultStorageFile string        = fmt.Sprintf("%s/%s%s", defaultStorageDir, bktName, defaultStorageFileExtension)
)

/*
	Refs:
	- https://github.com/br0xen/boltbrowser
*/

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register(map[string]string{})
	gob.Register([]interface{}{})
}

type Cache struct {
	debug            bool
	db               *bolt.DB
	bucket           *bolt.Bucket
	group            *Group
	storagePath      string
	bucketName       string
	groupName        string
	groupsName       string
	defaultGroupName string
	cacheTTL         time.Duration
	stop             chan bool
	ticker           *time.Ticker
	Groups           map[string]*Group
}

type Config struct {
	Debug            bool
	GroupsName       string
	GroupName        string
	BucketName       string
	StoragePath      string
	DefaultGroupName string
	CacheTTL         time.Duration
}

func New(config *Config) (*Cache, error) {
	if config == nil {
		log.Infoln("boltdbttlcache.New(): config is nil")
		config.CacheTTL = defaultCacheTTL
		config.StoragePath = defaultStorageFile
		config.BucketName = bktName
		config.GroupName = bktDefaultGroupName
		config.GroupsName = bktGroupsName
	}

	if config.StoragePath == "" {
		config.StoragePath = defaultStorageFile
		// return nil, errors.New("boltdbcache.New(): Storage path is not defined.")
	}

	if config.BucketName == "" {
		config.BucketName = bktName
	}

	if config.GroupName == "" {
		config.GroupName = bktDefaultGroupName
	}

	if config.GroupsName == "" {
		config.GroupsName = bktGroupsName
	}

	if config.CacheTTL < time.Millisecond {
		config.CacheTTL = defaultCacheTTL
	}

	if config.Debug {
		log.WithFields(log.Fields{
			"config": config,
		}).Warnf("boltdbttlcache.New()")
	}

	// Open the database
	db, err := bolt.Open(config.StoragePath, 0755, nil)
	if err != nil {
		if config.Debug {
			log.WithFields(log.Fields{
				"config": config,
			}).Fatalln("boltdbttlcache.New() error while creating boltdb bucket: ", err)
		}
		return nil, err
	}

	ttlc := Cache{
		db:          db,
		storagePath: config.StoragePath,
		bucketName:  config.BucketName,
		cacheTTL:    config.CacheTTL,
		groupName:   config.GroupName,
		groupsName:  config.GroupsName,
		debug:       config.Debug,
	}
	ttlc.Groups = make(map[string]*Group)

	// Load up existing groups
	err = db.Update(func(tx *bolt.Tx) error {
		// Ensure the groups bucket exists
		gps, err := tx.CreateBucketIfNotExists([]byte(config.BucketName))
		if err != nil {
			return err
		}
		// Get any existing groups
		err = gps.ForEach(func(k, v []byte) error {
			g := Group{
				Key:      k,
				ttlIndex: []IndexEntry{},
				ttlCache: &ttlc,
			}
			gbkt, err := tx.CreateBucketIfNotExists(k)
			if err != nil {
				return err
			}
			g.Lock()
			defer g.Unlock()
			g.ttlIndex = []IndexEntry{}
			err = gbkt.ForEach(func(k, v []byte) error {
				e := CacheEntry{}
				r := bytes.NewBuffer(v)
				decoder := gob.NewDecoder(r)
				err := decoder.Decode(&e)
				if err != nil {
					return err
				}
				g.ttlIndex = append(g.ttlIndex, IndexEntry{Key: k, ExpiresAt: e.ExpiresAt})
				return nil
			})
			if err != nil {
				return err
			}
			sort.Sort(ByIndexEntryExpiry(g.ttlIndex))
			ttlc.Groups[string(k)] = &g
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if config.Debug {
			log.WithFields(log.Fields{
				"config": config,
			}).Fatalln("boltdbttlcache.New() db.Update error: ", err)
		}
		return nil, err
	}

	ttlc.ticker = time.NewTicker(config.CacheTTL)
	ttlc.stop = make(chan bool, 1)

	go func() {
		for {
			select {
			case <-ttlc.ticker.C:
				ttlc.reapExpired()
			case <-ttlc.stop:
				return
			}
		}
	}()

	group, err := ttlc.CreateGroupIfNotExists(config.GroupName)
	if err != nil {
		if config.Debug {
			log.WithFields(log.Fields{
				"config": config,
			}).Fatalln("boltdbttlcache.New() CreateGroupIfNotExists error: ", err)
		}
	}

	ttlc.group = group

	return &ttlc, nil

}

// Mount returns a new Cache using the provided (and opened) bolt database.
func Mount(db *bolt.DB) *Cache {
	return &Cache{db: db}
}

// Get retrieves the response corresponding to the given key if present.
func (c *Cache) Get(key string) (resp []byte, ok bool) {
	var err error
	entry := &CacheEntry{}
	entry, err = c.group.Get([]byte(key))
	if c.debug && err != nil {
		log.Fatalln("boltdbttlcache.Get(): error: ", err)
	}
	if entry != nil && err == nil {
		if c.debug {
			pp.Println("Key: ", string(entry.Key))
			pp.Println("CreatedAt: ", entry.CreatedAt)
			pp.Println("ExpiresAt: ", entry.ExpiresAt)
		}
		resp, err = c.getBytes(entry.Value)
		if err != nil {
			if c.debug {
				log.Fatalln("boltdbttlcache.Get(): error: ", err)
			}
		}
	}
	return resp, resp != nil
}

func (c *Cache) getBytes(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (c *Cache) getInterface(bts []byte, data interface{}) error {
	buf := bytes.NewBuffer(bts)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(data)
	if err != nil {
		return err
	}
	return nil
}

// Set stores a response to the cache at the given key.
func (c *Cache) Set(key string, resp []byte) {
	if err := c.group.Put([]byte(key), resp, c.cacheTTL); err != nil {
		if c.debug {
			log.Fatalln("boltdbttlcache.Set(): error: ", err)
		}
	}
}

// Delete removes the response with the given key from the cache.
func (c *Cache) Delete(key string) {
	if err := c.group.Delete([]byte(key)); err != nil {
		if c.debug {
			log.Fatalln("boltdbttlcache.Delete(): error: ", err)
		}
	}
}

func (c *Cache) reapExpired() {
	// Loop over each group
	for _, g := range c.Groups {
		go func(g *Group) {
			g.ttlCache.db.Update(func(tx *bolt.Tx) error {
				now := time.Now()
				newIndex := []IndexEntry{}
				g.Lock()
				defer g.Unlock()
				for _, e := range g.ttlIndex {
					if e.ExpiresAt.Before(now) {

						gbkt := tx.Bucket(g.Key)
						if gbkt == nil {
							if c.debug {
								log.Fatal("boltdbttlcache.reapExpired(): bucket is nil")
							}
							return fmt.Errorf("Bucket does not exist")
						}
						err := gbkt.Delete(e.Key)
						if err != nil {
							return err
						}
					} else {
						newIndex = append(newIndex, e)
					}
				}
				// Ensure the groups bucket exists
				g.ttlIndex = newIndex
				return nil
			})
		}(g)
	}
}

type Group struct {
	sync.Mutex
	Key      []byte
	ttlIndex []IndexEntry
	ttlCache *Cache
}

func (g *Group) Count() int {
	return len(g.ttlIndex)
}

// Set stores a response to the cache at the given key.
func (g *Group) Put(key []byte, value interface{}, ttl time.Duration) error {
	err := g.ttlCache.db.Update(func(tx *bolt.Tx) error {
		// Ensure the groups bucket exists
		gbkt, err := tx.CreateBucketIfNotExists(g.Key)
		if err != nil {
			return err
		}
		existing := gbkt.Get(key)
		if existing != nil {
			// We are doing this primarily to ensure that the index stays in sync with the entries
			gbkt.Delete(key)
		}
		b := new(bytes.Buffer)
		now := time.Now()
		e := CacheEntry{
			Key:       key,
			Value:     value,
			CreatedAt: now,
			ExpiresAt: now.Add(ttl),
		}
		encoder := gob.NewEncoder(b)
		err = encoder.Encode(&e)
		if err != nil {
			return err
		}
		// Add the item to the group bucket
		err = gbkt.Put(key, b.Bytes())
		if err != nil {
			return err
		}
		g.Lock()
		defer g.Unlock()
		g.ttlIndex = appendIndex(g.ttlIndex, IndexEntry{Key: e.Key, ExpiresAt: e.ExpiresAt})
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *Group) Get(key []byte) (*CacheEntry, error) {
	var val []byte
	e := CacheEntry{}
	err := g.ttlCache.db.View(func(tx *bolt.Tx) error {
		// Ensure the groups bucket exists
		gbkt := tx.Bucket(g.Key)
		if gbkt == nil {
			return nil
		}
		val = gbkt.Get(key)
		if val == nil {
			return nil
		}
		r := bytes.NewBuffer(val)
		decoder := gob.NewDecoder(r)
		err := decoder.Decode(&e)
		if err != nil {
			return err
		}
		return nil

	})
	if err != nil {
		return nil, err
	}
	if e.Value == nil {
		return nil, nil
	}
	return &e, nil
}

func (g *Group) Delete(key []byte) error {
	g.Lock()
	defer g.Unlock()
	err := g.ttlCache.db.Update(func(tx *bolt.Tx) error {
		// Ensure the groups bucket exists
		gbkt := tx.Bucket(g.Key)
		if gbkt == nil {
			return fmt.Errorf("boltdbttlcache.Group.Delete(): Bucket does not exist")
		}
		return gbkt.Delete(key)

	})
	if err != nil {
		return err
	}

	ttlIndex := []IndexEntry{}

	for _, e := range g.ttlIndex {
		if !bytes.Equal(e.Key, key) {
			ttlIndex = append(ttlIndex, e)
		}
	}
	g.ttlIndex = ttlIndex
	return nil
}

// CreateGroupIfNotExists gets a group by key.  If the grouiop does not exist, it will be created.
func (c *Cache) CreateGroupIfNotExists(key string) (*Group, error) {
	// Check for an existing group.
	g := c.Groups[key]
	if g != nil {
		return g, nil
	}
	// New Group
	g = &Group{
		Key:      []byte(key),
		ttlIndex: []IndexEntry{},
	}
	g.Lock()
	defer g.Unlock()
	// If the group does not already exist, create it and add it to the database
	err := c.db.Update(func(tx *bolt.Tx) error {
		// get the bucket for the groups
		gps, err := tx.CreateBucketIfNotExists([]byte(c.groupsName))
		if err != nil {
			return err
		}
		// Add the group to the bucket
		err = gps.Put(g.Key, g.Key)
		if err != nil {
			return err
		}
		// Now add the group bucket
		_, err = tx.CreateBucketIfNotExists([]byte(key))
		if err != nil {
			return err
		}
		g.ttlCache = c
		c.Groups[key] = g
		return nil
	})
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (c *Cache) DeleteGroup(key string) error {
	// Check for an existing group.
	g := c.Groups[key]
	if len(g.Key) == 0 {
		return fmt.Errorf("Group does not exist")
	}
	g.Lock()
	defer g.Unlock()
	// If the group does not already exist, create it and add it to the database
	err := c.db.Update(func(tx *bolt.Tx) error {
		// get the bucket for the groups
		gbkt := tx.Bucket([]byte(bktGroupsName))
		err := gbkt.Delete(g.Key)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) Close() {
	c.stop <- true
	c.db.Close()
}

func (c *Cache) findAll() map[string][]byte {
	m := make(map[string][]byte)
	c.db.View(func(tx *bolt.Tx) error {
		namebytes := []byte(c.bucketName)
		b := tx.Bucket(namebytes)
		if b == nil {
			return errors.New("can't get [default] bucket in boltdb")
		}
		b.ForEach(func(k, v []byte) error {
			if c.debug {
				log.Println(string(k))
			}
			// must copy key, value to out varialbe (out of this transaction scope)
			// to use the value, or creat a new one and use it immediately
			if len(v) == 0 {
				m[string(k)] = nil
			} else {
				value := make([]byte, len(v))
				copy(value, v)
				m[string(k)] = value
			}
			return nil
		})

		return nil
	})
	return m
}

// EmptyBoltBucket -- empty a bucket if not exist, create one
func (c *Cache) truncate() error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		namebytes := []byte(c.bucketName)
		bucket := tx.Bucket(namebytes)
		if bucket != nil {
			err := tx.DeleteBucket(namebytes)
			if err != nil {
				if c.debug {
					log.Printf("empty bucket - [%s] error\n", c.bucketName)
				}
				return err
			}
		}
		_, err := tx.CreateBucket(namebytes)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}
