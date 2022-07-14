package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/willie/AuthApp/database"
	helper "github.com/willie/AuthApp/helpers"
	"github.com/willie/AuthApp/models"
	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

//create a user collection
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

//create a validator instance
var validate = validator.New()

//hashed password
func HashPassword(password string) string {
	//hash the provided password
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

//verify the password provided by the user
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	//compare user provided with the hashed password on db
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""
	//check for error when comparing password
	if err != nil {
		msg = fmt.Sprint("email of password is incorrect")
		check = false
	}

	return check, msg
}

//signup a User
func Signup() gin.HandlerFunc {

	return func(c *gin.Context) {
		//create a new context instance with a timeout
		var ctx, cancelFunc = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancelFunc()

		//user instance
		var user models.User

		//bind json data from context into user ref. instance and check for error during binding
		if err := c.BindJSON(&user); err != nil {
			//send to data to response body
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//validates a structs exposed fields
		validationErr := validate.Struct(user)

		//check for error during validation
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			//defer cancelFunc()
			return
		}

		//get count of user with same email in the collection
		emailCount, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})

		//releases resources if signup completes before timeout elapses
		defer cancelFunc()

		//check for error during couting
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
		}
		//count user with same phone document in the collection
		count, err := userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancelFunc()

		//check for error while counting
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone number"})
		}

		//handle email or phone number duplication
		if count > 0 || emailCount > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number already exists"})
			return
		}

		//hashed password
		password := HashPassword(*user.Password)
		//set user password to the hashed password
		user.Password = &password

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		//generates a new ObjectID to userID
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex() //.hex change the primitive object id to a string
		//get the user a token and a refresh token
		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, *&user.User_id)
		user.Token = &token
		user.Refresh_token = &refreshToken

		//insert user into db
		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)

		//closes and releases resources if process is complete
		defer cancelFunc()
		//check if insertion error
		if insertErr != nil {
			msg := fmt.Sprint("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		//send back respond as json
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

//login user
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancelFunc = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancelFunc()

		var user models.User
		var foundUser models.User
		//bind the context data into user ref.
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		//find the user with the email on the db
		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancelFunc()
		//check for error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
			return
		}
		//verify password provided by user
		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancelFunc()

		//check if provided password is valid
		if !passwordIsValid {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		//ensure foundUser email isn't empty
		if foundUser.Email == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		}

		//generate token for the user
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)
		//update in the db all token with new available tokens
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		//find user on db using id
		err = userCollection.FindOne(ctx, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)
		//check for error finding user
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		//return response to client
		c.JSON(http.StatusOK, foundUser)
	}
}

//fetch users.  Only admin can call this func.
func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		//check for user role. To ensure only admin can access
		if err := helper.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//create context
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1 {
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		//runnng mongo aggregation
		//stage 1
		matchStage := bson.D{{"$match", bson.D{{}}}}
		//stage 2
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}},
			{"total_count", bson.D{{"$sum", 1}}},
			{"data", bson.D{{"$push", "$$ROOT"}}}}}}
		//stage 3
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}}}}}

		//execute the aggregation
		result, err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, groupStage, projectStage})
		defer cancel()
		//check for aggregation error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items"})
		}

		var allusers []bson.M
		//iterate through the aggreegation result and decode each item into allusers
		if err = result.All(ctx, &allusers); err != nil {
			log.Fatal(err)
		}
		//send back respond
		c.JSON(http.StatusOK, allusers[0])
	}
}

//fetch a single user
func GetUser() gin.HandlerFunc {

	return func(c *gin.Context) { //Note gin.context passes value between middlewares
		//get the userid from the request param
		userId := c.Param("userId")
		//ensuring only admin and owner of resource can access
		if err := helper.MatchUserTypeToUid(c, userId); err != nil {
			log.Print("error here")
			//serializes the value as JSON into the response body
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) //gin.H returns its input as a map
			return
		}

		//create a new context instance with a timeout
		var ctx, cancelFunc = context.WithTimeout(context.Background(), 100*time.Second)

		//a user model instance
		var user models.User

		//querying the db for 1 document using the userid
		//Decode will unmarshal the document represented by this SingleResult into user ref.
		err := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		//ensure connection is close after every process
		defer cancelFunc()

		if err != nil {
			log.Print("error2 here")
			//serializes the value as JSON into the response body
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		//send to data to response body
		c.JSON(http.StatusOK, user)
	}
}
