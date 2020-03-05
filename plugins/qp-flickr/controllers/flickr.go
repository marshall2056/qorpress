package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qorpress/qorpress/pkg/services"
)

func GetFlickr(c *gin.Context) {
	payload, e := services.GetFlickrImages(9)
	services.GetFlickrAlbums()
	if e != nil {
		c.JSON(http.StatusInternalServerError, nil)
	} else {
		c.JSON(http.StatusOK, payload)
	}
}
