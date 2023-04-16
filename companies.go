package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *RouteHandler) CollectAndVerifyIndustryRequested(c *gin.Context) {
	var body SampleRequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		log.Println("error parsing request body: ", err)
		c.AbortWithError(http.StatusBadRequest, fmt.Errorf("error parsing request body"))
		return
	}

	if len(body.SicDescription) == 0 || len(body.SicDescription) > 200 {
		log.Println("Invalid industry request: ", body.SicDescription)
		c.String(http.StatusBadRequest, fmt.Sprintf("Invalid industry: %s", body.SicDescription))
		return
	}

	c.Set("Industry", &body.SicDescription)

	c.Next()
}

type SampleRequestBody struct {
	SicDescription string
}

func (h *RouteHandler) CompanySample(c *gin.Context) {
	industry := c.MustGet("Industry").(*string)

	companies, err := h.Database.GetSampleListOfCompaniesForIndustry(industry)
	if err != nil {
		log.Printf("Failed to get database sample for sic %s. Error: %s", *industry, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	log.Printf("returning %d companies for sic \"%s\"", len(*companies), *industry)

	c.JSON(http.StatusOK, companies)
}