package controllers

import (
	"net/http"
	"os"
	"time"
	"unicode/utf8"

	db "investment_tracker/database"
	"investment_tracker/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func Signup(c *gin.Context) {

	var user = new(models.User)

	// Get the user credentials from the request body
	if c.BindJSON(user) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	if utf8.RuneCountInString(user.Password) < 8 {
		c.JSON(400, gin.H{"error": "Password must be at least 8 characters long"})
		return
	}
	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = string(hash)

	// Create the user
	// Response
	result, err := db.DB.Exec("INSERT INTO users (name, email,password ,birthyear) VALUES (?, ?, ?, ?)", user.Name, user.Email, user.Password, user.Birthyear)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to create user"})
		return
	}
	userId, _ := result.LastInsertId()

	err = CreateSalary(userId, user.Birthyear)

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})

}

func Login(c *gin.Context) {
	// Get email and password from body
	var user = new(models.User)

	if c.BindJSON(user) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	// Look up requested user

	rows, err := db.DB.Query("SELECT * FROM users WHERE email = ?", user.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exist"})
		return
	}
	defer rows.Close()
	var user2 = new(models.User)

	for rows.Next() {
		rows.Scan(&user2.Id, &user2.Name, &user2.Email, &user2.Password, &user2.Birthyear)
	}

	// Compare
	if err := bcrypt.CompareHashAndPassword([]byte(user2.Password), []byte(user.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password"})
		return
	}

	// generate jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user2.Id,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to sign token"})
		return
	}

	// Send it back
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged in successfully"})

}

func Logout(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	// Deletes cookie by setting expiration time to -1
	c.SetCookie("Authorization", "", -1, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
func GetUser(c *gin.Context) {
	user, _ := c.Get("user")
	userName := user.(*models.User).Name
	c.JSON(http.StatusOK, gin.H{"name": userName})

}
func Validate(c *gin.Context) {
	user, _ := c.Get("user")

	c.JSON(http.StatusOK, gin.H{"Message": user})

}
