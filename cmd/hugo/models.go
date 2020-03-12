package main

import (
	"strings"
	"time"

	"github.com/k0kubun/pp"
	"github.com/gohugoio/hugo/hugolib"
)

// PageEntry maps the hugo internal page structure to a JSON structure
// that blevesearch can understand.
type PageEntry struct {
	Title        string    `json:"title"`
	Type         string    `json:"type"`
	Section      string    `json:"section"`
	Content      string    `json:"content"`
	UniqueID string `uniqueid`
	RelPermalink string `relpermalink`
	WordCount    float64   `json:"word_count"`
	ReadingTime  float64   `json:"reading_time"`
	Keywords     []string  `json:"keywords"`
	Tags     	 []string  `json:"tags"`
	Categories   []string  `json:"categories"`
	Videos []string `json:"images"`
	Images []string `json:"images"`
	IsPage bool  `json:"is_page"`
	IsHome bool `json:"is_home"`
	Date         time.Time `json:"date"`
	LastModified time.Time `json:"last_modified"`
	Author       string    `json:"author"`
}

func newEntry(page *hugolib.Page) *PageEntry {
	var author string

	// BUG: page.Author() and page.Authors() return empty values
	switch str := page.Params()["author"].(type) {
	case string:
		author = str
	case []string:
		author = strings.Join(str, ", ")
	}

	pp.Println("Categories", page.GetParam("tags"))
	pp.Println("Tags", page.GetParam("categories"))
	pp.Println("categories", page.Params()["categories"])
	pp.Println("tags", page.Params()["tags"])

	return &PageEntry{
		Title:        page.Title(),
		Type:         page.Type(),
		Section:      page.Section(),
		Content:      page.Plain(),
		WordCount:    float64(page.WordCount()),
		ReadingTime:  float64(page.ReadingTime()),
		Keywords:     page.Keywords,
		RelPermalink: page.RelPermalink(),
		//Categories:   page.Categories(),
		//Tags:         page.Tags(),
		UniqueID: page.UniqueID(),
		IsHome: page.IsHome(),
		IsPage: page.IsPage(),
		Date:         page.Date,
		LastModified: page.Lastmod,
		Author:       author,
	}
}