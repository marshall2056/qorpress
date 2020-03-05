package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/qorpress/qorpress/plugins/qp-flickr/models"
)

func GetFlickrImages(number int) (payload models.Payload, err error) {
	photostreamUrl := "https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=b5fd7ac0bc2b2e1670312fa98fbe0ae8&user_id=100756072%40N02&extras=url_sq%2Curl_t%2Curl_m%2Curl_b%2Curl_l%2Curl_n&per_page=" + strconv.Itoa(number) + "&format=json&nojsoncallback=1"

	response, err := http.Get(photostreamUrl)
	if err != nil {
		fmt.Println(err.Error())
		return payload, err
	}

	defer response.Body.Close() //Response.Body is of type io.ReadCloser *Look this up later"
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err.Error())
	}

	json.Unmarshal(body, &payload)
	return payload, nil
}

func GetFlickrAlbums() []models.PhotoAlbum {
	albumsUrl := "https://api.flickr.com/services/rest/?method=flickr.photosets.getList&api_key=b5fd7ac0bc2b2e1670312fa98fbe0ae8&user_id=100756072%40N02&primary_photo_extras=url_n,url_m&format=json&nojsoncallback=1"

	response, err := http.Get(albumsUrl)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	var payload models.AlbumPayload
	json.Unmarshal(body, &payload)
	return payload.PhotoSets.PhotoAlbums
}

func GetFlickrPhotosInAlbum(albumId int) (photos []models.PhotoItem) {
	albumUrl := "https://api.flickr.com/services/rest/?method=flickr.photosets.getPhotos&api_key=b5fd7ac0bc2b2e1670312fa98fbe0ae8&photoset_id=" + strconv.Itoa(albumId) + "&user_id=100756072%40N02&extras=url_sq%2Curl_t%2Curl_s%2Curl_m%2Curl_l%2Curl_n&format=json&nojsoncallback=1"
	resp, err := http.Get(albumUrl)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	defer resp.Body.Close()

	var photoAlbum models.PhotosPayload
	body, err := ioutil.ReadAll(resp.Body)
	jsonError := json.Unmarshal(body, &photoAlbum)
	if jsonError != nil {
		fmt.Println("Json marshal error: ", jsonError)
		return nil
	}

	return photoAlbum.PhotoSet.PhotoItems
}
