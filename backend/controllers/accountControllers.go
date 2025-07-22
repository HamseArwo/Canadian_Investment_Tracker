package controllers

import (
	"fmt"
	db "investment_tracker/database"
	"investment_tracker/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAccounts(c *gin.Context) {
	// Implementation goes here
	user, _ := c.Get("user")
	userID := user.(*models.User).Id
	statement, _ := db.DB.Prepare("SELECT * FROM accounts WHERE user_id = ?")
	rows, err := statement.Query(userID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to query accounts"})
		return
	}
	defer rows.Close()
	var accountList []models.Account

	for rows.Next() {
		var account models.Account
		err := rows.Scan(&account.Id, &account.UserId, &account.AccountTypeId, &account.Total, &account.ChildYear)
		if err != nil {
			c.JSON(500, gin.H{"error": account.ChildYear})
			return
		}
		accountList = append(accountList, account)
	}

	if len(accountList) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No accounts found"})
		return
	}

	c.JSON(200, gin.H{"accounts": accountList})
}

func GetAccount(c *gin.Context) {
	// Implementation goes here
	var account = new(models.Account)
	id := c.Param("id")
	statement, _ := db.DB.Prepare("SELECT * FROM accounts WHERE id = ?")
	rows, err := statement.Query(id)

	if err != nil {
		c.JSON(http.StatusNotFound, "Query failed")
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&account.Id, &account.UserId, &account.AccountTypeId, &account.Total, &account.ChildYear)
	}

	if account.Id == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"account": account})
}

func CreateAccount(c *gin.Context) {
	// Implementation goes here
	var account = new(models.Account)
	user, _ := c.Get("user")
	userID := user.(*models.User).Id
	fmt.Println(userID)

	err := c.BindJSON(account)

	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to receive account"})
		return
	}
	statement, _ := db.DB.Prepare("INSERT INTO accounts (user_id,account_type_id,total,child_year) VALUES (?, ?, ?, ?)")
	result, err := statement.Exec(userID, account.AccountTypeId, account.Total, account.ChildYear)
	accountID, _ := result.LastInsertId()
	account.Id = int(accountID)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create account"})
		return
	}
	err = CreateContribution(*account, *user.(*models.User))
	if err != nil {
		c.JSON(http.StatusInternalServerError, "Failed to create contributions")
		return
	}

	c.JSON(201, gin.H{"message": "Account created successfully"})

}

func DeleteAccount(c *gin.Context) {
	// Implementation goes here
	id := c.Param("id")
	var accountType int
	var err2 error = nil
	var err3 error = nil

	db.DB.QueryRow("SELECT account_type_id FROM accounts WHERE id = ?", id).Scan(&accountType)
	_, err1 := db.DB.Exec("DELETE FROM contributions WHERE account_id = ?", id)

	if accountType == 2 {
		_, err3 = db.DB.Exec("DELETE FROM cumulative_grants WHERE id = ?", id)
	} else {
		_, err2 = db.DB.Exec("DELETE FROM cumulative_contributions WHERE account_id = ?", id)

	}

	_, err4 := db.DB.Exec("DELETE FROM accounts WHERE id = ?", id)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		c.JSON(http.StatusNotFound, "Query failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Account deleted successfully"})
}
