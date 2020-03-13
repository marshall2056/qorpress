package funcmapmaker

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	"github.com/qorpress/qorpress-contrib/flickr/config"
	"github.com/qorpress/qorpress-contrib/flickr/models"
	"github.com/qorpress/qorpress/core/render"
	"github.com/qorpress/qorpress/pkg/utils"
)

func AddFuncMapMaker(view *render.Render) *render.Render {
	oldFuncMapMaker := view.FuncMapMaker
	view.FuncMapMaker = func(render *render.Render, req *http.Request, w http.ResponseWriter) template.FuncMap {
		funcMap := template.FuncMap{}
		if oldFuncMapMaker != nil {
			funcMap = oldFuncMapMaker(render, req, w)
		}

		funcMap["get_flickr_images"] = func() (payload models.FlickrPayload) {
			var fs models.FlickrSetting
			utils.GetDB(req).Find(&fs)
			if fs.ApiKey != "" && fs.UserId != "" {
				// get flickr images
				photostreamUrl := fmt.Sprintf(config.PhotoStreamUrl, fs.ApiKey, fs.UserId, fs.PerPage)
				response, err := http.Get(photostreamUrl)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				defer response.Body.Close() //Response.Body is of type io.ReadCloser *Look this up later"
				body, err := ioutil.ReadAll(response.Body)
				if err != nil {
					fmt.Println(err.Error())
				}
				json.Unmarshal(body, &payload)
				return payload
			}
			return
		}

		funcMap["get_flickr_albums"] = func() []models.FlickrPhotoAlbum {
			var fs models.FlickrSetting
			utils.GetDB(req).Find(&fs)
			if fs.ApiKey != "" && fs.UserId != "" {
				// get flickr albums
				albumsUrl := fmt.Sprintf(config.AlbumsUrl, fs.ApiKey, fs.UserId)
				response, err := http.Get(albumsUrl)
				if err != nil {
					fmt.Println(err)
					return nil
				}
				defer response.Body.Close()
				body, err := ioutil.ReadAll(response.Body)
				var payload models.FlickrAlbumPayload
				json.Unmarshal(body, &payload)
				return payload.PhotoSets.PhotoAlbums
			}
			return nil
		}

		funcMap["get_flickr_photos_in_album"] = func(albumId int) (photos []models.FlickrPhotoItem) {
			var fs models.FlickrSetting
			utils.GetDB(req).Find(&fs)
			if fs.ApiKey != "" && fs.UserId != "" {
				// get flickr images from album
				albumUrl := fmt.Sprintf(config.AlbumsUrl, fs.ApiKey, albumId, fs.UserId)
				resp, err := http.Get(albumUrl)
				if err != nil {
					fmt.Println(err)
					return
				}
				defer resp.Body.Close()
				var photoAlbum models.FlickrPhotosPayload
				body, err := ioutil.ReadAll(resp.Body)
				jsonError := json.Unmarshal(body, &photoAlbum)
				if jsonError != nil {
					fmt.Println("Json marshal error: ", jsonError)
					return
				}
				return photoAlbum.PhotoSet.PhotoItems
			}
			return
		}

		return funcMap
	}

	return view
}
