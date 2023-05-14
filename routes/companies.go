package routes

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *RouteHandler) CompanyFullList(c *gin.Context) {
	// industry := c.MustGet("Industry").(string)

	// companies, err := h.Database.GetListOfCompanies(&industry, false)
	// if err != nil {
	// 	log.Printf("Failed to get full list for sic %s. Error: %s", industry, err)
	// 	c.AbortWithStatus(http.StatusInternalServerError)
	// 	return
	// }

	// log.Printf("returning %d companies for sic \"%s\"", len(companies), industry)

	c.JSON(http.StatusOK, nil)
}

func (h *RouteHandler) CompanySampleV2(c *gin.Context) {
	industry := c.MustGet("Industry").(string)

	results, err := h.Database.GetCompaniesAndOwnershipForIndustry(&industry, true)
	if err != nil {
		log.Printf("Failed to get sample for sic %s. Error: %s", industry, err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	log.Printf("returning %d companies and %d PSCs for sic \"%s\"", len(results.Companies), len(results.PSCs), industry)

	c.JSON(http.StatusOK, results)
}
