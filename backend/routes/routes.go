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
	// Account routes
	router.GET("/accounts", middleware.Authentication, controllers.GetAccounts)
	router.GET("/accounts/:id", middleware.Authentication, controllers.GetAccount)
	router.POST("/accounts", middleware.Authentication, controllers.CreateAccount)
	router.DELETE("/accounts/:id", middleware.Authentication, controllers.DeleteAccount)
	// Contribution routes
	router.POST("accounts/contribution/:id", middleware.Authentication, controllers.CreateContribution)
	router.GET("accounts/contribution/:id", middleware.Authentication, controllers.GetContributions)

}
