package main

import (
	"company-data-backend/integrations"
	"company-data-backend/routes"
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

	c := &routes.RouteHandler{
		EmailAPI:                  &integrations.EmailAPI{URL: getEnv("EMAIL_API_URL")},
		JwtTokenLifespanInMinutes: getEnv("TOKEN_MINUTE_LIFESPAN"),
		ApiSecret:                 getEnv("API_SECRET"),
		Database: &integrations.Database{
			Host:     getEnv("DB_HOST"),
			Port:     getEnv("DB_PORT"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME"),
		},
	}

	r := createRouter(c)

	r.Run()
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

func createRouter(handler *routes.RouteHandler) *gin.Engine {
	engine := gin.Default()

	serverRecoversFromAnyPanicAndWrites500(engine)
	allowAllOriginsForCORS(engine)

	v2 := engine.Group("/v2")
	{
		v2.POST("/register", handler.Register)

		companies := v2.Group("/companies", handler.CollectAndVerifyIndustryRequested)
		{
			companies.POST("/sample", handler.CompanySampleV2)

			authorised := companies.Group("/authorized", handler.VerifyAuthorization)
			{
				authorised.POST("/full", handler.CompanyFullList)
			}
		}
	}

	return engine
}

func serverRecoversFromAnyPanicAndWrites500(engine *gin.Engine) {
	engine.Use(gin.Recovery())
}

func allowAllOriginsForCORS(engine *gin.Engine) {
	engine.Use(cors.Default())
}
