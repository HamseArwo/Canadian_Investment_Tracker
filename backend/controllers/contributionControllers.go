package controllers

import (
	"errors"
	"fmt"
	calculator "investment_tracker/Calculator"
	db "investment_tracker/database"
	"investment_tracker/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateContribution(account models.Account, user models.User) error {

	var start_year int
	if account.AccountTypeId == 2 {
		if account.ChildYear == 0 {
			return errors.New("Child year not set")
		}
		start_year = account.ChildYear

	} else {
		start_year = calculator.GetStartYear(account.AccountTypeId, user.Birthyear)
	}

	for start_year < 2026 {
		fmt.Println(account.Id, user.Id, start_year)
		_, err := db.DB.Exec("INSERT INTO contributions (account_id,user_id,amount, year) VALUES (?, ?, ?, ?)", account.Id, user.Id, 0, start_year)

		if err != nil {
			fmt.Println("HERE?")

			return err
		}

		if account.AccountTypeId == 1 || account.AccountTypeId == 3 {
			_, err = db.DB.Exec("INSERT INTO cumulative_contributions (account_id, amount, year) VALUES (?, ?, ?)", account.Id, 0, start_year)
			if err != nil {
				fmt.Println(account.AccountTypeId)

				return err
			}
		} else if account.AccountTypeId == 2 {
			_, err = db.DB.Exec("INSERT INTO cumulative_grants (account_id, grant_earned, grant_unused, year) VALUES (?, ?, ?, ?)", account.Id, 0, 0, start_year)
			if err != nil {

				return err
			}
		}
		start_year++

	}
	if account.AccountTypeId == 1 || account.AccountTypeId == 3 {
		err := calculator.CalculateCumulativeContribution(strconv.Itoa(account.Id), account.AccountTypeId, user.Id, user.Birthyear)
		if err != nil {
			return err
		}

	}
	return nil
}

func GetContributions(c *gin.Context) {
	accountId := c.Param("id")
	user, _ := c.Get("user")
	userID := user.(*models.User).Id
	var contributionList []models.Contribution

	rows, err := db.DB.Query("SELECT * FROM contributions WHERE account_id = ? AND user_id = ?", accountId, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Query failed"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var contribution models.Contribution
		rows.Scan(&contribution.Id, &contribution.UserId, &contribution.AccountId, &contribution.Amount, &contribution.Year)
		contributionList = append(contributionList, contribution)
	}

	if len(contributionList) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No contributions found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Contributions": contributionList})
}

func UpdateContribution(c *gin.Context) {
	var contri = new(models.Contribution)
	var account models.Account
	var oldValue float64 = 1
	var oldTotal float64
	user, _ := c.Get("user")
	Birthyear := user.(*models.User).Birthyear
	userID := user.(*models.User).Id

	accountId := c.Param("id")
	err := c.BindJSON(contri)

	rows, err := db.DB.Query("SELECT * FROM accounts WHERE id = ?", accountId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&account.Id, &account.UserId, &account.AccountTypeId, &account.Total, &account.ChildYear)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	validate_error := calculator.ValidateContribution(contri, &account)
	if validate_error != nil {

		c.JSON(http.StatusBadRequest, gin.H{"error": "Error validating contribution"})
		return
	}

	err = db.DB.QueryRow("SELECT amount FROM contributions WHERE year = ? AND account_id = ?", contri.Year, accountId).Scan(&oldValue)
	if err != nil {
		fmt.Println(":Hello")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}
	if oldValue != contri.Amount {
		err = db.DB.QueryRow("SELECT total FROM accounts WHERE id = ?", accountId).Scan(&oldTotal)
		newTotal := oldTotal + (contri.Amount - oldValue)
		_, err = db.DB.Exec("UPDATE accounts SET total = ? WHERE id = ?", newTotal, accountId)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

	}

	_, err = db.DB.Exec("UPDATE contributions SET amount = ? WHERE year = ?", contri.Amount, contri.Year)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if account.AccountTypeId == 1 || account.AccountTypeId == 3 {
		err := calculator.CalculateCumulativeContribution(accountId, account.AccountTypeId, userID, Birthyear)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal this one ?server error"})
			return
		}
	} else {
		err := calculator.CalculateGrantContribution(accountId, contri.Amount, contri.Year, account.ChildYear)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

	}

	c.JSON(http.StatusOK, gin.H{"message": "Contribution created successfully"})
}
