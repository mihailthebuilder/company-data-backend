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

	c := &RouterConfig{
		Emailer: &Emailer{
			EmailApiUrl: getEnv("EMAIL_API_URL"),
		},
	}

	r := createRouter(c)

	r.Run()
}

type RouterConfig struct {
	Emailer IEmailer
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

func createRouter(config *RouterConfig) *gin.Engine {
	r := gin.Default()

	serverRecoversFromAnyPanicAndWrites500(r)
	allowAllOriginsForCORS(r)
	setCustomRouterConfig(r, config)

	r.POST("/companies/sample", handleRequestForCompaniesSample)
	r.POST("/register", handleRegistration)

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

func setCustomRouterConfig(engine *gin.Engine, config *RouterConfig) {
	engine.Use(func(c *gin.Context) {
		c.Set("config", config)
		c.Next()
	})
}
