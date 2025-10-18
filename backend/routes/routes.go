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
	router.GET("/user", middleware.Authentication, controllers.GetUser)
	// Account routes
	router.GET("/accounts", middleware.Authentication, controllers.GetAccounts)
	router.GET("/accounts/:id", middleware.Authentication, controllers.GetAccount)
	router.POST("/accounts", middleware.Authentication, controllers.CreateAccount)
	router.DELETE("/accounts/:id", middleware.Authentication, controllers.DeleteAccount)
	// Contribution routes
	router.POST("accounts/contribution/:id", middleware.Authentication, controllers.UpdateContribution)
	router.GET("accounts/contribution/:id", middleware.Authentication, controllers.GetContributions)
	router.GET("accounts/culumative/:id/:type", middleware.Authentication, controllers.GetCumulativeContribution)
	router.GET("accounts/limit", middleware.Authentication, controllers.GetContributionsLimit)
	router.GET("accounts/rrsplimit", middleware.Authentication, controllers.GetRRSPLimit)
	// Salary routes
	router.POST("/salary", middleware.Authentication, controllers.UpdateSalary)
	router.GET("/salary", middleware.Authentication, controllers.GetSalaries)

}
