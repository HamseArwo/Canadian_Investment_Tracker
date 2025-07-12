package main

import (
	"fmt"
	db "investment_tracker/database"
	"investment_tracker/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// fmt.Println("Database connection established")

	r := gin.Default()
	db.InitDB()
	fmt.Println(db.DB)
	routes.SetupRoutes(r)
	r.Run("localhost:8080")

}
