package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func handleAuthorisation(c *gin.Context) {
	log.Println("handled authorisation")
	c.AbortWithStatus(401)
}
