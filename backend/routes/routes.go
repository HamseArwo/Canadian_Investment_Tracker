package routes

import (
	"investment_tracker/controllers"
	"investment_tracker/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	router.POST("/signup", controllers.Signup)
	router.POST("/login", controllers.Login)
	router.DELETE("/logout", controllers.Logout)
	router.GET("/validate", middleware.Authentication, controllers.Validate)
}
