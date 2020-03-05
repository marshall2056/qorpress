package models

type Payload struct {
	Photos Photos `json:"photos"`
}

type Photos struct {
	PhotoItems []PhotoItem `json:"photo"`
}

type PhotoItem struct {
	Id           string `json:"id"`
	UrlMedium    string `json:"url_m"`
	UrlThumbnail string `json:"url_t"`
	UrlMedium2   string `json:"url_n"`
	UrlSquare    string `json:"url_sq"`
	UrlSmall     string `json:"url_s"`
	UrlLarge     string `json:"url_l"`
}

type AlbumPayload struct {
	Stat      string    `json:"stat"`
	PhotoSets PhotoSets `json:"photosets"`
}

type PhotoSets struct {
	PhotoAlbums []PhotoAlbum `json:"photoset"`
}

type PhotoAlbum struct {
	Id                string            `json:"id"`
	Primary           string            `json:"primary"`
	Photos            int               `json:"photos"`
	Title             Title             `json:"title"`
	Description       Description       `json:"description"`
	PrimaryPhotoExtra PrimaryPhotoExtra `json:"primary_photo_extras"`
}

type PrimaryPhotoExtra struct {
	Url       string `json:"url_n"`
	UrlMedium string `json:"url_m"`
}

type Title struct {
	Content string `json:"_content"`
}

type Description struct {
	Content string `json:"_content"`
}

type PhotoSet struct {
	PhotoItems []PhotoItem `json:"photo"`
}

type PhotosPayload struct {
	PhotoSet PhotoSet `json:"photoset"`
	Stat     string   `json:"stat"`
}
