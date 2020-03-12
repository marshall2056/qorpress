package config

const (
	PhotoStreamUrl = `https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=%s&user_id=%s&extras=url_sq%2Curl_t%2Curl_m%2Curl_b%2Curl_l%2Curl_n&per_page=%d&format=json&nojsoncallback=1`
	AlbumsUrl = `https://api.flickr.com/services/rest/?method=flickr.photosets.getList&api_key=%s&user_id=%s&primary_photo_extras=url_n,url_m&format=json&nojsoncallback=1`
	AlbumUrl = `https://api.flickr.com/services/rest/?method=flickr.photosets.getPhotos&api_key=%s&photoset_id=%d&user_id=%s&extras=url_sq%2Curl_t%2Curl_s%2Curl_m%2Curl_l%2Curl_n&format=json&nojsoncallback=1`
)