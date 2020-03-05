package utils

import (
	"github.com/jinzhu/copier"
	"github.com/koreset/go-twitter/twitter"
)

type User struct {
	Name                 string `json:"name"`
	ScreenName           string `json:"screen_name"`
	ProfileImageURL      string `json:"profile_image_url"`
	ProfileImageURLHttps string `json:"profile_image_url_https"`
}

type ShallowTweet struct {
	ID               int64                   `json:"id"`
	IDStr            string                  `json:"id_str"`
	Text             string                  `json:"text"`
	FullText         string                  `json:"full_text"`
	User             User                    `json:"user"`
	Entities         *twitter.Entities       `json:"entities"`
	RetweetedStatus  *twitter.Tweet          `json:"retweeted_status"`
	ExtendedEntities *twitter.ExtendedEntity `json:"extended_entities"`
	ExtendedTweet    *twitter.ExtendedTweet  `json:"extended_tweet"`
}

func GetShallowTweets(tweets []twitter.Tweet) (shallowTweets []ShallowTweet) {
	copier.Copy(&shallowTweets, &tweets)
	return
}
