package models

type FlickrPayload struct {
	Photos FlickrPhotos `json:"photos"`
}

type FlickrPhotos struct {
	PhotoItems []FlickrPhotoItem `json:"photo"`
}

type FlickrPhotoItem struct {
	Id           string `json:"id"`
	UrlMedium    string `json:"url_m"`
	UrlThumbnail string `json:"url_t"`
	UrlMedium2   string `json:"url_n"`
	UrlSquare    string `json:"url_sq"`
	UrlSmall     string `json:"url_s"`
	UrlLarge     string `json:"url_l"`
}

type FlickrAlbumPayload struct {
	Stat      string    `json:"stat"`
	PhotoSets FlickrPhotoSets `json:"photosets"`
}

type FlickrPhotoSets struct {
	PhotoAlbums []FlickrPhotoAlbum `json:"photoset"`
}

type FlickrPhotoAlbum struct {
	Id                string            `json:"id"`
	Primary           string            `json:"primary"`
	Photos            int               `json:"photos"`
	Title             FlickrTitle             `json:"title"`
	Description       FlickrDescription       `json:"description"`
	PrimaryPhotoExtra FlickrPrimaryPhotoExtra `json:"primary_photo_extras"`
}

type FlickrPrimaryPhotoExtra struct {
	Url       string `json:"url_n"`
	UrlMedium string `json:"url_m"`
}

type FlickrTitle struct {
	Content string `json:"_content"`
}

type FlickrDescription struct {
	Content string `json:"_content"`
}

type FlickrPhotoSet struct {
	PhotoItems []FlickrPhotoItem `json:"photo"`
}

type FlickrPhotosPayload struct {
	PhotoSet FlickrPhotoSet `json:"photoset"`
	Stat     string   `json:"stat"`
}
