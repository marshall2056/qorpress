package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AboutUs(c *gin.Context) {
	payload := make(map[string]interface{})
	payload["active"] = "aboutus"
	c.HTML(http.StatusOK, "about-us", payload)
}
