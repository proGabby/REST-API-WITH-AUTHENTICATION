package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	helper "github.com/willie/AuthApp/helpers"
)

//authentication middleware that ensure routes are provided from public access
func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		//get the token provided by the request header
		clientToken := c.Request.Header.Get("token")
		//abort if token is empty
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("No Authorization header provided")})
			c.Abort()
			return
		}
		//validate token
		claims, err := helper.ValidateToken(clientToken)
		//check for validation error
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}
		//update the context with the details
		c.Set("email", claims.Email)
		c.Set("first_name", claims.First_name)
		c.Set("last_name", claims.Last_name)
		c.Set("uid", claims.Uid)
		c.Set("user_type", claims.User_type)
		//execute the next middleware
		c.Next()
	}
}
