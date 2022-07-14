package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
)

//permission middleware that check for user role
func CheckUserType(c *gin.Context, role string) (err error) {
	userType := c.GetString("user_type")
	err = nil
	//check for usertype
	if userType != role {
		err = errors.New("Unauthorized to access this resource")
		return err
	}
	return err
}

//permision middleware to ensure only admin or owner of resouce gets access
func MatchUserTypeToUid(c *gin.Context, userId string) (err error) {
	//get value associated with the key as a string
	userType := c.GetString("user_type")
	uid := c.GetString("uid")
	err = nil
	//ensuring only admin and owner of resource gets access
	if userType != "Admin" && uid != userId {
		err = errors.New("Not permitted to access this resource")
		return err
	}
	//check for user role
	err = CheckUserType(c, userType)
	return err
}
