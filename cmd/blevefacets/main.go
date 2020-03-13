package main

import (
	"encoding/json"
	"fmt"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/analysis/analyzer/keyword"
)

type Blog struct {
	Title  string "json:title"
	Author string "json:author"
	Body   string "json:body"
}

func main() {
	enFieldMapping := bleve.NewTextFieldMapping()
	enFieldMapping.Analyzer = "en"

	kwFieldMapping := bleve.NewTextFieldMapping()
	kwFieldMapping.Analyzer = keyword.Name

	blogMapping := bleve.NewDocumentMapping()
	blogMapping.AddFieldMappingsAt("Title", enFieldMapping)
	blogMapping.AddFieldMappingsAt("Body", enFieldMapping)
	blogMapping.AddFieldMappingsAt("Author", kwFieldMapping)

	mapping := bleve.NewIndexMapping()
	mapping.DefaultMapping = blogMapping

	index, err := bleve.NewMemOnly(mapping)
	if err != nil {
		panic(err)
	}

	blogs := []Blog{
		{Title: "Hello World", Author: "Sangeet Kumar", Body: "Hello world post"},
		{Title: "Hello World Again", Author: "Mayan Sangeet", Body: "Hello world post again"},
		{Title: "Hello World Again 2", Author: "Mayan Sangeet", Body: "Hello world post again one more time"},
	}

	// Index blogs
	for _, b := range blogs {
		index.Index(b.Title, b)
	}

	// Query string
	query := bleve.NewQueryStringQuery("hello")

	// Simple text search
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"*"}
	searchResult, err := index.Search(searchRequest)
	if err != nil || searchResult.Total == 0 {
		fmt.Println("Not found")
		return
	}

	fmt.Println("Simple search result:")

	for i, hit := range searchResult.Hits {
		jsonstr, _ := json.Marshal(hit.Fields)
		fmt.Printf("Hit[%d]: %v %v\n", i, hit.ID, string(jsonstr))
	}

	// Facets search
	facet := bleve.NewFacetRequest("Author", 10)
	searchRequest.AddFacet("Author", facet)
	searchResult, err = index.Search(searchRequest)
	if err != nil || searchResult.Total == 0 {
		fmt.Println("Facets Not found")
		return
	}

	fmt.Println("\nFacets search result:")

	for i, hit := range searchResult.Hits {
		jsonstr, _ := json.Marshal(hit.Fields)
		fmt.Printf("Hit[%d]: %v %v\n", i, hit.ID, string(jsonstr))
	}

	for fname, fresult := range searchResult.Facets {
		jsonstr, _ := json.Marshal(fresult)
		fmt.Println("Facets:", fname, string(jsonstr))
		fmt.Println("Authors:")
		for _, tfacet := range fresult.Terms {
			fmt.Printf("\t%s (%d)\n", tfacet.Term, tfacet.Count)
		}
	}
}
