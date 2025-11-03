package main

import (

	// "runtime/trace"

	db "investment_tracker/database"

	"investment_tracker/routes"

	"github.com/gin-gonic/gin"
)

func init() {
	db.InitDB()
	// initializer.LoadEnvVariables()

}

func main() {

	router := gin.Default()
	router.Use(CORSMiddleware())

	routes.SetupRoutes(router)

	router.Run("localhost:8080")
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
