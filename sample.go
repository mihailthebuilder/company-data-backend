package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *RouteHandler) CompanySample(c *gin.Context) {
	industry := c.MustGet("Industry").(string)

	companies, err := h.Database.GetListOfCompanies(&industry, true)
	if err != nil {
		log.Printf("Failed to get company sample for sic %s. Error: %s", industry, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	psc, err := h.Database.GetListOfPersonsWithSignificantControl(&companies)
	if err != nil {
		log.Printf("Failed to get psc sample for sic %s. Error: %s", industry, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	log.Printf("returning %d companies and %d PSCs for sic \"%s\"", len(companies), len(psc), industry)

	c.JSON(http.StatusOK, SampleRouteResponse{
		Companies:                     companies,
		PersonsWithSignificantControl: psc,
	})
}

type SampleRouteResponse struct {
	Companies                     []Company                      `json:"companies"`
	PersonsWithSignificantControl []PersonWithSignificantControl `json:"personsWithSignificantControl"`
}
