package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleAuthentication(c *gin.Context) {
	c.JSON(http.StatusOK, "")
}
