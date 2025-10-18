package controllers

import (
	calculator "investment_tracker/Calculator"
	db "investment_tracker/database"
	"net/http"
	"strconv"

	"investment_tracker/models"

	"github.com/gin-gonic/gin"
)

func CreateSalary(userID int64, userBirthYear int) error {
	for i := userBirthYear + 18; i <= 2025; i++ {

		_, err := db.DB.Exec("INSERT INTO salary (user_id, amount , year) VALUES (?, ?, ?)", userID, 0, i)
		if err != nil {
			return err
		}

	}
	return nil

}

func UpdateSalary(c *gin.Context) {
	// salaryID := c.Param("id")
	user, _ := c.Get("user")
	userID := user.(*models.User).Id
	birthYear := user.(*models.User).Birthyear
	var salary models.Salary
	err := c.BindJSON(&salary)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Failed to update salary")
		return
	}

	_, err = db.DB.Exec("UPDATE salary SET amount = ? WHERE year = ?", salary.Amount, salary.Year)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Failed to update salary")
		return
	}
	var accountIDs []int
	rows, err := db.DB.Query("SELECT id FROM accounts WHERE account_type_id = 3 AND user_id = ?", userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, "Failed to fetch account IDs")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			c.JSON(http.StatusBadRequest, "Failed to scan account ID")
			return
		}
		accountIDs = append(accountIDs, id)
	}

	for _, accountID := range accountIDs {
		err := calculator.CalculateCumulativeContribution(strconv.Itoa(accountID), 3, userID, birthYear)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Failed to update cumulative contributions")
			return
		}
	}

	c.JSON(http.StatusOK, "Salary updated successfully")

}

func GetSalaries(c *gin.Context) {
	user, _ := c.Get("user")
	userID := user.(*models.User).Id
	var salaryList []models.Salary
	var salary models.Salary
	rows, _ := db.DB.Query("SELECT * FROM salary WHERE user_id = ?", userID)
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&salary.Id, &salary.UserId, &salary.Amount, &salary.Year)
		if err != nil {
			c.JSON(http.StatusBadRequest, "Failed to fetch salaries")
		}
		salaryList = append(salaryList, salary)
	}
	c.JSON(http.StatusOK, gin.H{"salaries": salaryList})

}
