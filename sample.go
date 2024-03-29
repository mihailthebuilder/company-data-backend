package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *RouteHandler) CompanySample(c *gin.Context) {
	industry := c.MustGet("Industry").(string)

	companies, err := h.Datastore.GetListOfCompanies(industry, true)
	if err != nil {
		log.Printf("Failed to get company sample for sic %s. Error: %s", industry, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	log.Printf("returning %d companies for %s", len(companies), industry)

	c.JSON(http.StatusOK, SampleRouteResponse{
		Companies: companies,
	})
}

type SampleRouteResponse struct {
	Companies []Company `json:"companies"`
}
