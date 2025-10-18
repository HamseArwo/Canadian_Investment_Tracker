package main

import (
	"investment_tracker/initializer"
	// "runtime/trace"

	db "investment_tracker/database"

	"investment_tracker/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	db.InitDB()
	initializer.LoadEnvVariables()

}

func main() {

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // React app
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: true,
	}))

	routes.SetupRoutes(router)

	router.Run("localhost:8080")
}
