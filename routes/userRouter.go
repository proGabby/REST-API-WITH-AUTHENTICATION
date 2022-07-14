package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/willie/AuthApp/controllers"
	"github.com/willie/AuthApp/middleware"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	//gin.Engine is an instance of the gin Engine. it contains the muxer, middleware and configuration settings.

	//.Use attaches a global middleware to the router
	//protecting routes with a auth middleware
	incomingRoutes.Use(middleware.Authenticate())

	//handling post request on routes
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.GET("/users/:userId", controller.GetUser())
}
