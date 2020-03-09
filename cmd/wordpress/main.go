package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/k0kubun/pp"
	"github.com/spf13/pflag"

	"github.com/qorpress/go-wordpress"
)

var (
	username string
	password string
	endpoint string
	help     bool
)

func main() {

	// read .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	pflag.StringVarP(&username, "username", "", os.Getenv("WORDPRESS_USERNAME"), "wordpress' username.")
	pflag.StringVarP(&password, "password", "", os.Getenv("WORDPRESS_PASSWORD"), "wordpress' password.")
	pflag.StringVarP(&endpoint, "endpoint", "", os.Getenv("WORDPRESS_API_ENDPOINT"), "wordpress api endpoint (eg. https://domain.com/wp-json).")
	pflag.BoolVarP(&help, "help", "h", false, "help info")
	pflag.Parse()
	if help {
		pflag.PrintDefaults()
		os.Exit(1)
	}

	tp := wordpress.BasicAuthTransport{
		Username: username,
		Password: password,
	}

	// create wp-api client
	client, _ := wordpress.NewClient(endpoint, tp.Client())

	ctx := context.Background()

	// get the currently authenticated users details
	authenticatedUser, _, err := client.Users.Me(ctx, nil)
	if err != nil {
		log.Fatalln(err)
	}
	pp.Printf("Authenticated user %+v\n", authenticatedUser)

	// Import users

	// Import media

	// Import categories
	catOpts := &wordpress.CategoryListOptions{
		HideEmpty: true,
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	var allCategories []*wordpress.Category
	for {
		categories, resp, err := client.Categories.List(ctx, catOpts)
		if err != nil {
			log.Fatalln(err)
		}
		allCategories = append(allCategories, categories...)
		if resp.NextPage == 0 {
		  	break
		}
		catOpts.Page = resp.NextPage
	}
	// pp.Println(allCategories)

	// Import tags
	tagOpts := &wordpress.TagListOptions{
		HideEmpty: true,
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	var allTags []*wordpress.Tag
	for {
		tags, resp, err := client.Tags.List(ctx, tagOpts)
		if err != nil {
			log.Fatalln(err)
		}
		allTags = append(allTags, tags...)
		if resp.NextPage == 0 {
		  	break
		}
		tagOpts.Page = resp.NextPage
	}	
	// pp.Println(allTags)

	// Import medias
	mediaOpts := &wordpress.MediaListOptions{
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	var allMedias []*wordpress.Media
	for {
		medias, resp, err := client.Media.List(ctx, mediaOpts)
		if err != nil {
			log.Fatalln(err)
		}
		allMedias = append(allMedias, medias...)
		if resp.NextPage == 0 {
		  	break
		}
		mediaOpts.Page = resp.NextPage
	}
	pp.Println(allMedias)

	// Import pages
	pageOpts := &wordpress.PageListOptions{
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	var allPages []*wordpress.Page
	for {
		pages, resp, err := client.Pages.List(ctx, pageOpts)
		if err != nil {
			log.Fatalln(err)
		}
		allPages = append(allPages, pages...)
		if resp.NextPage == 0 {
		  	break
		}
		pageOpts.Page = resp.NextPage
	}
	// pp.Println(allPages)

	// Import posts
	postOpts := &wordpress.PostListOptions{
		ListOptions: wordpress.ListOptions{
			PerPage: 10,
		},
	}
	var allPosts []*wordpress.Post
	for {
		posts, resp, err := client.Posts.List(ctx, postOpts)
		if err != nil {
			log.Fatalln(err)
		}
		allPosts = append(allPosts, posts...)
		if resp.NextPage == 0 {
		  	break
		}
		postOpts.Page = resp.NextPage
	}
	// pp.Println(allPosts)

}
