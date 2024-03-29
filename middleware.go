package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *RouteHandler) CollectAndVerifyIndustryRequested(c *gin.Context) {
	var body SampleRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("Error parsing request body: ", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if len(body.SicDescription) == 0 || len(body.SicDescription) > 200 {
		log.Println("Invalid industry request: ", body.SicDescription)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	log.Println("Industry requested: ", body.SicDescription)

	c.Set("Industry", body.SicDescription)

	c.Next()
}

type SampleRequestBody struct {
	SicDescription string
}
