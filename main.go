package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	routes "github.com/willie/AuthApp/routes"
)

func main() {
	//load .env file
	err := godotenv.Load(".env")
	//check for env loading error
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port, isAvail := os.LookupEnv("PORT")

	if !isAvail {
		log.Fatal("PORT key is not an env variable")
	}

	//create a gin router without middleware
	router := gin.New()

	// Global middleware
	// Logger middleware will write the logs to gin.DefaultWriter
	router.Use(gin.Logger())

	//Handling routes
	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	//Handle routes
	router.GET("/api-1", func(c *gin.Context) {
		//gin.context allows us to pass variables between middleware, manage the flow,
		//validate the JSON of a request and render a JSON response

		//JSON serializes the given struct as JSON into the response body. It also sets the Content-Type as "application/json".
		//gin.H parse its input into a map
		c.JSON(200, gin.H{"success": "Access granted for api-1"})
	})

	router.GET("/api-2", func(c *gin.Context) {
		//gin.context allows us to pass variables between middleware, manage the flow,
		//validate the JSON of a request and render a JSON response

		//JSON serializes the given struct as JSON into the response body. It also sets the Content-Type as "application/json".
		//gin.H parse its input into a map
		c.JSON(200, gin.H{"success": "Access granted for api-2"})
	})

	//starts server
	router.Run(":" + port)
}
