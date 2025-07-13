package main

import (
	"investment_tracker/initializer"

	db "investment_tracker/database"

	"investment_tracker/routes"

	"github.com/gin-gonic/gin"
)

func init() {
	db.InitDB()
	initializer.LoadEnvVariables()

}

func main() {
	router := gin.Default()
	routes.SetupRoutes(router)

	router.Run("localhost:8080")
}
