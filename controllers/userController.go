package controllers

import (
	"context"
	"log"
	"net/http"
	"restaurant-order-management/database"
	"restaurant-order-management/helpers"
	"restaurant-order-management/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)


var UserCollection *mongo.Collection = database.OpenCollection(database.Client,"user")


func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)
		recordPerPage,err := strconv.Atoi(c.Query("recordPerPage"))
		if err!=nil||recordPerPage<0{
			recordPerPage = 10
		}
		page,err1 := strconv.Atoi(c.Query("page"))

		if err1!=nil||page<0{
			page = 1
		}
		startIndex,err2:= strconv.Atoi(c.Query("startIndex"))
		if err2!=nil||startIndex<0{
			startIndex = (page-1)*recordPerPage
		}

		matchStage := bson.D{{"$match",bson.D{{}}}}
		projectStage:= bson.D{{"$project",bson.D{{"_id",0},{"total_count",1},{"user_items",bson.D{{"$slice",[]interface{}{"$data",startIndex,recordPerPage}}}}}}}

		result,err := UserCollection.Aggregate(ctx,mongo.Pipeline{
			matchStage,projectStage,
		})
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occured while listening to the user items"})
			return
		}

		var allUser []bson.M

		err = result.All(ctx,&allUser)
		if err!=nil{
			log.Fatal(err)
		}

		c.JSON(http.StatusOK,allUser)

	}
}

func GetUser() gin.HandlerFunc{
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)

		UserId:=c.Param("user_id")
		
		var User models.User

		err := UserCollection.FindOne(ctx,bson.M{"user_id":UserId}).Decode(&User)

		if err !=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occured while listening to User"})
			return
		}

		defer cancel()

		c.JSON(http.StatusOK,User)
		
	}
}

func Signup() gin.HandlerFunc  {
	return func(c *gin.Context) {
		ctx,cancel :=  context.WithTimeout(context.Background(),100*time.Second)

		var User models.User

		err := c.BindJSON(&User)
		if err!=nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
			return
		}

		validateErr := validate.Struct(User)
		if validateErr != nil{
			c.JSON(http.StatusBadRequest,gin.H{"error":validateErr.Error()})
			return
		}

		count,err := UserCollection.CountDocuments(ctx,bson.M{"email":User.Email})
		defer cancel()
		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occured while checking for the email"})
			log.Panic(err)
		}

		if count>0{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"user with this email already exists"})
			return
		}

		password := HashPasswaord(*User.Password)
		User.Password = &password

		count,err = UserCollection.CountDocuments(ctx,bson.M{"phone":User.Phone})
		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occured while checking for the phone"})
			log.Panic(err)
		}

		if count>0{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"user with this phone number already exists"})
			return
		}

		User.Created_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		User.Updated_at,_ = time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		User.ID = primitive.NewObjectID()
		User.User_id = User.ID.Hex()

		token,refresh_token,_ := helpers.GenerateAllTokens(*User.Email,*User.First_name,*User.Last_name,*&User.User_id)

		User.Token = &token
		User.Refresh_token = &refresh_token

		result,err := UserCollection.InsertOne(ctx,User)
		if err!=nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"error occured while creating the user"})
			return
		}
		c.JSON(http.StatusOK,result)

	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx,cancel := context.WithTimeout(context.Background(),100*time.Second)

		var user models.User
		var founduser models.User

		err := c.BindJSON(&user)
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
			return
		}

		err = UserCollection.FindOne(ctx,bson.M{"email":user.Email}).Decode(&founduser)
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError,gin.H{"error":"user not found, login seems to be incorrect"})
			return
		}

		passwordisValid, msg := VerifyPassword(*user.Password,*founduser.Password)

		if passwordisValid!= true{
			c.JSON(http.StatusInternalServerError,gin.H{"error":msg})
			return
		}

		token,refresh_token,_ := helpers.GenerateAllTokens(*founduser.Email,*founduser.First_name,*founduser.Last_name,*&founduser.User_id)

		helpers.UpdateAllTokens(token,refresh_token,founduser.User_id)

		c.JSON(http.StatusOK,founduser)



	}
}

func HashPasswaord(password string) string{
	bytes, err := bcrypt.GenerateFromPassword([]byte(password),14)
	if err!=nil{
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool,string){
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword),[]byte(userPassword))
	check := true
	msg := ""
	if err!=nil{
		msg = "incorrect password"
		check = false
	}
	return check,msg
}
