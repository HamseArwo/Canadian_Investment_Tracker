package middleware

import (
	"log"
	"net/http"
	"os"
	"time"

	db "investment_tracker/database"
	"investment_tracker/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Authentication(c *gin.Context) {
	// Get cookie off the request
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parsing token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		log.Fatal(err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Check expiration
		if claims["exp"].(float64) < float64(time.Now().Unix()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
			return
		}
		// Finds user
		statement, _ := db.DB.Prepare("SELECT * FROM USERS WHERE id = ?")
		rows, err := statement.Query(claims["sub"])
		// If user does not exist
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		var user = new(models.User)
		// get the user
		for rows.Next() {
			err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.Password, &user.Birthyear)
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			c.Set("user", user)
		}

	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
	}

	c.Next()
}
