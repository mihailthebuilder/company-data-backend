package main

import "github.com/gin-gonic/gin"

func createRouter(handler *RouteHandler) *gin.Engine {
	r := gin.Default()

	serverRecoversFromAnyPanicAndWrites500(r)
	allowAllOriginsForCORS(r)

	v2 := r.Group("/v2")
	{
		v2.POST("/companies/sample", handler.CompanySample)
	}

	return r
}
