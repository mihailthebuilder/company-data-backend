package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleAuthentication(c *gin.Context) {
	token := generateToken()

	c.JSON(http.StatusOK, gin.H{"token": *token})
}

func generateToken() *string {
	out := "helloworld"
	return &out
}
