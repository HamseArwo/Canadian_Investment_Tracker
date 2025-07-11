package main

import (
	db "investment_tracker/database"
	// "github.com/gin-gonic/gin"
)

func main() {
	// fmt.Println("Database connection established")

	// r := gin.Default()
	db.InitDB()

	// r.Run("8080")

}
