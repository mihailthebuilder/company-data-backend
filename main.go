package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	if isRunningLocally() {
		loadEnvironmentVariablesFromDotEnvFile()
	}

	c := &RouteHandler{
		Emailer: &Emailer{
			EmailApiUrl: getEnv("EMAIL_API_URL"),
		},
		JwtTokenLifespanInMinutes: getEnv("TOKEN_MINUTE_LIFESPAN"),
		ApiSecret:                 getEnv("API_SECRET"),
	}

	r := createRouter(c)

	r.Run()
}

type RouteHandler struct {
	Emailer                   IEmailer
	JwtTokenLifespanInMinutes string
	ApiSecret                 string
	Database                  IDatabase
}

func isRunningLocally() bool {
	return os.Getenv("GIN_MODE") == ""
}

func loadEnvironmentVariablesFromDotEnvFile() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func createRouter(handler *RouteHandler) *gin.Engine {
	r := gin.Default()

	serverRecoversFromAnyPanicAndWrites500(r)
	allowAllOriginsForCORS(r)

	r.POST("/register", handler.Registration)

	companies := r.Group("/companies", handler.CollectAndVerifyIndustryRequested)
	{
		companies.POST("/sample", handler.CompanySample)
	}

	authorised := r.Group("/authorized", verifyAuthorization)
	{
		authorised.POST("/companies", handleRequestForEntireList)
	}

	return r
}

func serverRecoversFromAnyPanicAndWrites500(engine *gin.Engine) {
	engine.Use(gin.Recovery())
}

func allowAllOriginsForCORS(engine *gin.Engine) {
	engine.Use(cors.Default())
}
