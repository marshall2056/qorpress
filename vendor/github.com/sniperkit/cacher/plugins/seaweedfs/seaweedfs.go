// Package seaweedfscache provides an implementation of httpcache.Cache that stores and
// retrieves data using Seaweedfs.
package seaweedfscache

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	seaweedfs "github.com/linxGnu/goseaweedfs"
	"github.com/sniperkit/httpcache/helpers"
)

/*
	Refs:
	- github.com/silentsharer/weedo
	- github.com/ginuerzh/weedo
	- github.com/linxGnu/goseaweedfs
	- github.com/ChristianNorbertBraun/Weedharvester
	- https://github.com/ChristianNorbertBraun/Weedharvester/blob/master/client_test.go
*/

const (
	defaultMaster string = "localhost:8898"
	defaultScheme string = "http"
)

var (
	defaultCacheDuration          = 10 * time.Minute
	defaultChunkSize              = 2 * 1024 * 1024 // 2MB
	defaultFilers        []string = []string{"localhost:7788"}
	validSchemes         []string = []string{"http", "https"}
)

// Cache objects store and retrieve data using Memcached.
type Cache struct {
	store *seaweedfs.Seaweed
}

type Config struct {
	Env       bool
	Master    string
	Scheme    string
	Filers    []string
	Timeout   time.Duration
	ChunkSize int64
}

// New returns a new Cache
func New(config *Config) *Cache {

	if config == nil {
		config.Scheme = defaultScheme
		config.Master = defaultMaster
		config.Filers = defaultFilers
		config.ChunkSize = defaultChunkSize
		config.Timeout = time.Duration(2) * time.Second
	}

	if config.Env {
		config.Scheme = os.Getenv("GOSWFS_SCHEME")
		config.Master = os.Getenv("GOSWFS_MASTER_URL")
		config.Filers = os.Getenv("GOSWFS_FILER_URL")
		// config.ChunkSize = os.Getenv("GOSWFS_MASTER_URL")
		// config.Timeout = os.Getenv("GOSWFS_MASTER_URL")
	}

	return &Cache{store: seaweedfs.NewSeaweed(
		config.Scheme,
		config.Master,
		config.Filers,
		config.ChunkSize,
		config.Timeout,
	)}
}

/*
	Ref:
	- https://github.com/linxGnu/goseaweedfs/blob/master/libs/httpClient.go
		- Grow(count int, collection, replication, dataCenter string) error
		- GrowArgs(args url.Values) (err error)
		- Lookup(volID string, args url.Values) (result *model.LookupResult, err error)
		- LookupNoCache(volID string, args url.Values) (result *model.LookupResult, err error)
		- LookupServerByFileID(fileID string, args url.Values, readonly bool) (server string, err error)
		- LookupFileID(fileID string, args url.Values, readonly bool) (fullURL string, err error)
		- LookupVolumeIDs(volIDs []string) (result map[string]*model.LookupResult, err error)
		- GC(threshold float64) (err error)
		- Status() (result *model.SystemStatus, err error)
		- ClusterStatus() (result *model.ClusterStatus, err error)
		- Assign(args url.Values) (result *model.AssignResult, err error)
		- Submit(filePath string, collection, ttl string) (result *model.SubmitResult, err error)
		- SubmitFilePart(f *model.FilePart, args url.Values) (result *model.SubmitResult, err error)
		- Upload(fileReader io.Reader, fileName string, size int64, collection, ttl string) (fp *model.FilePart, fileID string, err error)
		- UploadFile(filePath string, collection, ttl string) (cm *model.ChunkManifest, fp *model.FilePart, fileID string, err error)
		- UploadFilePart(f *model.FilePart) (cm *model.ChunkManifest, fileID string, err error)
		- BatchUploadFiles(files []string, collection, ttl string) ([]*model.SubmitResult, error)
		- BatchUploadFileParts(files []*model.FilePart, collection string, ttl string) ([]*model.SubmitResult, error)
		- Replace(fileID string, fileReader io.Reader, fileName string, size int64, collection, ttl string, deleteFirst bool) (err error)
		- ReplaceFile(fileID, filePath string, deleteFirst bool) error
		- ReplaceFilePart(f *model.FilePart, deleteFirst bool) (fileID string, err error)
		- DeleteChunks(cm *model.ChunkManifest, args url.Values) (err error)
		- DeleteFile(fileID string, args url.Values) (err error)
*/

func (c *Cache) lookup(key string, cache bool) (resp []byte, ok bool) {
	switch cache {
	case false:
		_, err := c.store.LookupNoCache(key, nil)
		if err != nil {
			return nil, false
		}
	default:
		_, err := c.store.Lookup(key, nil)
		if err != nil {
			return nil, false
		}
	}
	return nil, false
}

// to do, parse the json response
func (c *Cache) health(endpoint string) (status bool, err error) {
	return false, nil
}

func (c *Cache) status(component string) (status bool, err error) {
	switch component {
	case "cluster":
		status, err = c.store.ClusterStatus()
		if err != nil {
			fmt.Println("error: ", err)
		}
		return
	case "master":
		fallthrough
	case "default":
		status, err = c.store.Status()
		if err != nil {
			fmt.Println("error: ", err)
		}
		return
	}
	return false, errors.New("Unkown component type to check")
}

func (c *Cache) submit(filepath string) (status bool, err error) {
	// extract protocol from filepath
	protocol := "http"
	// c.store.status("cluster")
	// c.store.status("master")
	switch protocol {
	case "http":
		fallthrough
	case "https":
		if res, err := c.store.Submit(filepath, "", ""); err != nil {
			fmt.Println("error: ", err)
			return false, nil
		} else {
			fmt.Println("Submission result: ", res)
		}
	case "ipfs":
		fmt.Println("Not implemented yet")
		return false, nil
	case "default":
		if res, err := c.store.Submit(filepath, "", ""); err != nil {
			fmt.Println("error: ", err)
			return false, nil
		} else {
			fmt.Println("Submission result: ", res)
		}
	}
	return false, nil
}

func (c *Cache) replace(key string, localFile string) (status bool, err error) {
	if err := c.store.ReplaceFile(key, localFile, false); err != nil {
		t.Fatal(err)
		return
	}
}

func (c *Cache) batchUpload(files []string) (status bool, err error) {
	if res, err := c.store.BatchUploadFiles(files, "", ""); err != nil {
		return false, nil
	} else {
		fmt.Println(res)
	}
	return false, nil
}

func (c *Cache) upload(localpath string, remotePath string, filerEndpoint string) (status bool, err error) {
	var hasFiler bool
	if filerEndpoint != "" {
		// c.store.health(filerEndpoint)
		// check if filer is valid
		hasFiler = true
	}
	switch hasFiler {
	case true:
		if res, err := filer.UploadFile(localpath, remotePath, "", ""); err != nil {
			fmt.Println("error: ", err)
			return false, nil
		} else {
			fmt.Println(res)
		}
	default:
		if res, _, _, err := c.store.UploadFile(localpath, "", ""); err != nil {
			return false, nil
		} else {
			fmt.Println(res)
		}
	}
	return false, nil
}

// Read reads file with a given fileId
func (c *Cache) Get(key string) (resp []byte, ok bool) {
	location, err := c.store.lookup(key, false)
	if err != nil {
		return nil, err
	}
	rdata, err := http.Get(helpers.AddSlashIfNeeded(location.PublicURL) + key)

	if err != nil {
		log.Printf("Failure while sending get to %s/%s", location.PublicURL, key)
		log.Printf("Error: %s", err)
		return nil, false
	}

	if rdata.StatusCode >= 300 {
		log.Printf("Status %d while reading from %s/%s", rdata.StatusCode, location.PublicURL, key)
		log.Printf("Error: %s", errors.New("Bad StatusCode"))
		return nil, false
	}

	resp, err = ioutil.ReadAll(rdata)
	if err != nil {
		log.Printf("seaweedfscache.Read failed: %s", err)
		return nil, false
	}

	rdata.Close()
	return
}

func (c *Cache) Set(key string, content []byte) {
	/*
		// test upload file
		fh, err := os.Open(MediumFile)
		if err != nil {
			t.Fatal(err)
		}
		defer fh.Close()

		var size int64
		if fi, fiErr := fh.Stat(); fiErr != nil {
			t.Fatal(fiErr)
		} else {
			size = fi.Size()
		}

		if _, fID, err = c.store.Upload(fh, "test.txt", size, "col", ""); err != nil {
			t.Fatal(err)
		}

		// Replace with small file
		fs, err := os.Open(SmallFile)
		if err != nil {
			t.Fatal(err)
		}
		defer fs.Close()
		if fi, fiErr := fs.Stat(); fiErr != nil {
			t.Fatal(fiErr)
		} else {
			size = fi.Size()
		}
	*/
	return
}

func (c *Cache) Delete(key string) { // (status bool, err error) {
	if key == "" {
		fmt.Println("empty key, skipping...")
		return
	}
	err := c.store.DeleteFile(key, nil)
	if err != nil {
		fmt.Println("error occured while deleting a file: ", err)
	}
}

func (c *Cache) Indexes() {}

/*

// UploadResult contains upload result after put file to SeaweedFS
// Raw response: {"name":"go1.8.3.linux-amd64.tar.gz","size":82565628,"error":""}
type UploadResult struct {
	Name  string `json:"name,omitempty"`
	Size  int64  `json:"size,omitempty"`
	Error string `json:"error,omitempty"`
}

// AssignResult contains assign result.
// Raw response: {"fid":"1,0a1653fd0f","url":"localhost:8899","publicUrl":"localhost:8899","count":1,"error":""}
type AssignResult struct {
	FileID    string `json:"fid,omitempty"`
	URL       string `json:"url,omitempty"`
	PublicURL string `json:"publicUrl,omitempty"`
	Count     uint64 `json:"count,omitempty"`
	Error     string `json:"error,omitempty"`
}

// SubmitResult result of submit operation.
type SubmitResult struct {
	FileName string `json:"fileName,omitempty"`
	FileURL  string `json:"fileUrl,omitempty"`
	FileID   string `json:"fid,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Error    string `json:"error,omitempty"`
}

// ClusterStatus result of getting status of cluster
type ClusterStatus struct {
	IsLeader bool
	Leader   string
	Peers    []string
}

// SystemStatus result of getting status of system
type SystemStatus struct {
	Topology Topology
	Version  string
	Error    string
}

// Topology result of topology stats request
type Topology struct {
	DataCenters []*DataCenter
	Free        int
	Max         int
	Layouts     []*Layout
}

// DataCenter stats of a datacenter
type DataCenter struct {
	Free  int
	Max   int
	Racks []*Rack
}

// Rack stats of racks
type Rack struct {
	DataNodes []*DataNode
	Free      int
	Max       int
}

// DataNode stats of data node
type DataNode struct {
	Free      int
	Max       int
	PublicURL string `json:"PublicUrl"`
	URL       string `json:"Url"`
	Volumes   int
}

// Layout of replication/collection stats. According to https://github.com/chrislusf/seaweedfs/wiki/Master-Server-API
type Layout struct {
	Replication string
	Writables   []uint64
}

// ChunkInfo chunk information. According to https://github.com/chrislusf/seaweedfs/wiki/Large-File-Handling.
type ChunkInfo struct {
	Fid    string `json:"fid"`
	Offset int64  `json:"offset"`
	Size   int64  `json:"size"`
}

// ChunkManifest chunk manifest. According to https://github.com/chrislusf/seaweedfs/wiki/Large-File-Handling.
type ChunkManifest struct {
	Name   string       `json:"name,omitempty"`
	Mime   string       `json:"mime,omitempty"`
	Size   int64        `json:"size,omitempty"`
	Chunks []*ChunkInfo `json:"chunks,omitempty"`
}

// FilePart file wrapper with reader and some metadata
type FilePart struct {
	Reader     io.Reader
	FileName   string
	FileSize   int64
	IsGzipped  bool
	MimeType   string
	ModTime    int64 //in seconds
	Collection string

	// TTL Time to live.
	// 3m: 3 minutes
	// 4h: 4 hours
	// 5d: 5 days
	// 6w: 6 weeks
	// 7M: 7 months
	// 8y: 8 years
	TTL string

	Server string
	FileID string
}

// File structure according to filer API at https://github.com/chrislusf/seaweedfs/wiki/Filer-Server-API.
type File struct {
	FileID string `json:"fid"`
	Name   string `json:"name"`
}

// Dir directory of filer. According to https://github.com/chrislusf/seaweedfs/wiki/Filer-Server-API.
type Dir struct {
	Path    string `json:"Directory"`
	Files   []*File
	Subdirs []*File `json:"Subdirectories"`
}

// Filer client
type Filer struct {
	URL        string `json:"url"`
	HTTPClient *libs.HTTPClient
}

// FilerUploadResult upload result which responsed from filer server. According to https://github.com/chrislusf/seaweedfs/wiki/Filer-Server-API.
type FilerUploadResult struct {
	Name    string `json:"name,omitempty"`
	FileURL string `json:"url,omitempty"`
	FileID  string `json:"fid,omitempty"`
	Size    int64  `json:"size,omitempty"`
	Error   string `json:"error,omitempty"`
}

// VolumeLocation location of volume responsed from master API. According to https://github.com/chrislusf/seaweedfs/wiki/Master-Server-API
type VolumeLocation struct {
	URL       string `json:"url,omitempty"`
	PublicURL string `json:"publicUrl,omitempty"`
}

// VolumeLocations returned VolumeLocations (volumes)
type VolumeLocations []*VolumeLocation


// LookupResult the result of looking up volume. According to https://github.com/chrislusf/seaweedfs/wiki/Master-Server-API
type LookupResult struct {
	VolumeID        string          `json:"volumeId,omitempty"`
	VolumeLocations VolumeLocations `json:"locations,omitempty"`
	Error           string          `json:"error,omitempty"`
}



*/
