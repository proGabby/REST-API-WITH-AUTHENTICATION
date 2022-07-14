package helper

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/willie/AuthApp/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email              string
	First_name         string
	Last_name          string
	Uid                string
	User_type          string
	jwt.StandardClaims //session cliams
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string = os.Getenv("SECRET_KEY")

//generate token with claims for the user
func GenerateAllTokens(email string, firstName string, lastName string, userType string, uid string) (signedToken string, signedRefreshToken string, err error) {
	//create session claims
	claims := &SignedDetails{
		Email:      email,
		First_name: firstName,
		Last_name:  lastName,
		Uid:        uid,
		User_type:  userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(), //unix changes the time to an int64 type
		},
	}

	//create a refresh claim
	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}

	//create a new token with the claim
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY)) //.SignedString gets the complete, signed token
	//create a new refresh token with the refresh claim
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, err
}

//validate token provided by the client and return token with claims
func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	//add claims to the token
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)
	//check for error while adding claims
	if err != nil {
		msg = err.Error()
		return
	}
	//valid token
	claims, ok := token.Claims.(*SignedDetails)
	//check for validation
	if !ok {
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}
	//check token expired time
	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("token is expired")
		msg = err.Error()
		return
	}

	return claims, msg
}

//update all tokens with the new available tokens
func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	//create a context instance with an elapses
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var updateObj primitive.D //ordered representation of a BSON document
	//add the signedToken and signedRefreshToken to the BSON document
	updateObj = append(updateObj, bson.E{"token", signedToken})
	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})
	//create an updated_at variable holding the time of update
	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	//add the updated_at to the BSON doc
	updateObj = append(updateObj, bson.E{"updated_at", Updated_at})

	upsert := true
	filter := bson.M{"user_id": userId}

	// A set of filters specifying to which array elements an update should apply. This option is only valid for MongoDB
	opt := options.UpdateOptions{
		//If true, a new document will be inserted if the filter does not match any documents in the collection
		Upsert: &upsert,
	}

	//update user doc in db
	_, err := userCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			//update db document with updageObj
			{"$set", updateObj},
		},
		&opt,
	)

	defer cancel()
	//check for error while updating
	if err != nil {
		log.Panic(err)
		return
	}
	return
}
