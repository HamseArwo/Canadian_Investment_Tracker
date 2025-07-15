package controllers

import (
	"fmt"
	db "investment_tracker/database"
	"investment_tracker/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateContribution(c *gin.Context) {
	var contri = new(models.Contribution)
	accountId := c.Param("id")
	err := c.BindJSON(contri)
	fmt.Println(accountId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
	}
	statement, _ := db.DB.Prepare("INSERT INTO contributions (account_id, amount, year) VALUES (?, ?, ?)")
	_, err = statement.Exec(accountId, contri.Amount, contri.Year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
	c.JSON(http.StatusOK, gin.H{"message": "Contribution created successfully"})
}

func GetContributions(c *gin.Context) {
	accountId := c.Param("id")
	var contributionList []models.Contribution

	statement, _ := db.DB.Prepare("SELECT * FROM contributions WHERE account_id = ?")
	rows, err := statement.Query(accountId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed"})
		return
	}

	for rows.Next() {
		var contribution models.Contribution
		rows.Scan(&contribution.Id, &contribution.Account_id, &contribution.Amount, &contribution.Year)
		contributionList = append(contributionList, contribution)
	}

	if len(contributionList) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No contributions found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Contributions": contributionList})
}

func DeleteContribution(c *gin.Context) {
}

func UpdateContribution(c *gin.Context) {
}
