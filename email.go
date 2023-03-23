package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleRequestToSendEmail(c *gin.Context) {
	c.Status(http.StatusOK)
}
