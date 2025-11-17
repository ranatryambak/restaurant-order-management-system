package helpers

import (
	"context"
	"log"
	"os"
	"restaurant-order-management/database"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Uid        string
	jwt.StandardClaims
}

var UserCollection *mongo.Collection = database.OpenCollection(database.Client,"user")

var SECRET_KEY = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string,first_name string,last_name string,uid string)(signedToken string, refresh_Token string, err error) {
	claims := &SignedDetails{
		Email: email,
		First_name: first_name,
		Last_name: last_name,
		Uid: uid,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour*time.Duration(24)).Unix(),
		},
	}
	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour*time.Duration(168)).Unix(),
		},
	}
	signedToken,err = jwt.NewWithClaims(jwt.SigningMethodHS256,claims).SignedString([]byte(SECRET_KEY))
	if err!=nil{
		log.Panic(err)
	}
	refresh_Token,err = jwt.NewWithClaims(jwt.SigningMethodHS256,refreshClaims).SignedString([]byte(SECRET_KEY))

	if err!=nil{
		log.Panic(err)
	}

	return signedToken, refresh_Token,err

}

func UpdateAllTokens(signedToken string,signedrefresh_Token string,uid string) {
	ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)

	var updateObj primitive.D

	updateObj = append(updateObj, bson.E{"token",signedToken})
	updateObj = append(updateObj, bson.E{"refresh_token",signedrefresh_Token})

	Updated_at, _ := time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at",Updated_at})

	filter := bson.M{"user_id":uid}
	upsert := true

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}
	_,err := UserCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set",updateObj},
		},
		&opt,
	)

	defer cancel()

	if err!=nil{
		log.Panic(err)
	}

}

func ValidateAllTokens(signedToken string)(claims *SignedDetails, msg string) {

	token,err :=jwt.ParseWithClaims(
		signedToken,
		SignedDetails{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY),nil
		},
	)

	claims,ok := token.Claims.(*SignedDetails)
	if !ok{
		msg = "token is invalid"
		msg = err.Error()
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix(){
		msg = "token is expired"
		msg = err.Error()
		return
	}

	return claims,msg

}