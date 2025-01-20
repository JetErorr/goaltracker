package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Name  string `bson:"name"`
	Email string `bson:"email"`
}

var collection *mongo.Collection

const secretKey = "my_secret_key"

func connectToMongoDB() (*mongo.Client, error) {
	credential := options.Credential{
		Username: "root",
		Password: "example",
	}

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017").SetAuth(credential)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the MongoDB server to ensure connectivity
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB!")
	return client, nil
}

func createUser(c *gin.Context) {
	result, err := collection.InsertOne(context.Background(), User{Name: "John Doe", Email: "john@example.com"})
	if err != nil {
		log.Println("createUser:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode post"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func getUsers(c *gin.Context) {
	cur, err := collection.Find(context.Background(), bson.M{"name": "John Doe"})
	if err != nil {
		log.Println("getUsers:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode post"})
		return
	}
	defer cur.Close(context.Background())

	var users []User
	for cur.Next(context.Background()) {
		var user User
		if err := cur.Decode(&user); err != nil {
			log.Println("getUsers:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode post"})
			return
		}
		users = append(users, user)
	}

	if err := cur.Err(); err != nil {
		log.Println("getUsers:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode post"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func updateUser(c *gin.Context) {
	result, err := collection.UpdateMany(context.Background(), bson.M{"name": "John Doe"}, bson.M{"$set": bson.M{"email": "newemail@example.com"}})
	if err != nil {
		log.Println("getUsers:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode post"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func deleteUser(c *gin.Context) {
	result, err := collection.DeleteMany(context.Background(), bson.M{"name": "John Doe"})
	if err != nil {
		log.Println("getUsers:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode post"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func handleLogin(c *gin.Context) {
	// In a real application, authenticate the user (this is just an example)
	username := c.PostForm("username")
	// password := c.PostForm("password")

	// Check user credentials
	var role string = `user`
	if !true {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Create a new token object, specifying signing method and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour * 1).Unix(), // Token expiration time: 1 hour
	})

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenJWT, _ := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")

		// Parse the token
		token, err := jwt.Parse(tokenJWT, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrAbortHandler
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Set the token claims to the context
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			fmt.Println(claims)
			c.Set("claims", claims)
		} else {
			fmt.Println(claims)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Next() // Proceed to the next handler if authorized
	}
}

func main() {
	client, err := connectToMongoDB()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	collection = client.Database("mydb").Collection("users")

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Caught error:", err)
			}
		}()
		c.Next()
	})

	router.GET("/users", getUsers)
	router.POST("/users/create", createUser)
	router.POST("/login", handleLogin)
	router.POST("/protected", authMiddleware(), getUsers)
	// TODO: API-ize the 2 missing function

	router.Run("localhost:8080")
}
