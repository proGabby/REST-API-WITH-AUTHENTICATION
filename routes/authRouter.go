package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/willie/AuthApp/controllers"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	//gin.Engine is an instance of the gin Engine. it contains the muxer, middleware and configuration settings.

	//handling post request on routes
	incomingRoutes.POST("users/signup", controller.Signup())
	incomingRoutes.POST("users/login", controller.Login())

}
