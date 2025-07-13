package controllers

// import (
// 	"unicode/utf8"

// 	db "investment_tracker/database"
// 	"investment_tracker/models"

// 	"github.com/gin-gonic/gin"
// )

// func PostSignup(c *gin.Context) {

// 	var user = new(models.User)

// 	c.BindJSON(user)
// 	if utf8.RuneCountInString(user.Password) < 8 {
// 		c.JSON(400, gin.H{"error": "Password must be at least 8 characters long"})
// 		return
// 	}

// 	statement, _ := db.DB.Prepare("INSERT INTO users (name, email, birthyear, password) VALUES (?, ?, ?, ?)")
// 	statement.Exec(user.Name, user.Email, user.Birthyear, user.Password)

// }
