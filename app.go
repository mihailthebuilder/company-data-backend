package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func runApplication() {
	r := gin.Default()

	serverRecoversFromAnyPanicAndWrites500(r)
	allowAllOriginsForCORS(r)

	r.POST("/companies/sample", handleRequestForCompaniesSample)
	r.POST("/authenticate", handleAuthentication)

	r.Run()
}

func serverRecoversFromAnyPanicAndWrites500(engine *gin.Engine) {
	engine.Use(gin.Recovery())
}

func allowAllOriginsForCORS(engine *gin.Engine) {
	engine.Use(cors.Default())
}
