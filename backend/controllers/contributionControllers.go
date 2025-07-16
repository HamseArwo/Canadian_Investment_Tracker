package controllers

import (
	"errors"
	"fmt"
	db "investment_tracker/database"
	"investment_tracker/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateContribution(account models.Account, user models.User) error {
	var start_year int
	if account.Account_type_id == 2 {
		if account.Child_year == 0 {
			return errors.New("Child year not set")
		}
		start_year = account.Child_year
		statement, _ := db.DB.Prepare("INSERT INTO cumulative_grants (account_id, grant_earned, grant_unused) VALUES (?, ?, ?)")
		_, err := statement.Exec(account.Id, 0, 500*(2025-start_year))
		if err != nil {
			return err
		}

	} else if account.Account_type_id == 1 {
		start_year = max(user.Birthyear+18, 2005)
	} else {
		start_year = user.Birthyear + 18
	}

	for start_year < 2026 {
		statement, _ := db.DB.Prepare("INSERT INTO contributions (account_id, amount, year) VALUES (?, ?, ?)")
		_, err := statement.Exec(account.Id, 0, start_year)
		if err != nil {
			return err
		}

		if account.Account_type_id == 1 || account.Account_type_id == 3 {
			statement, _ = db.DB.Prepare("INSERT INTO cumulative_contributions (account_id, amount, year) VALUES (?, ?, ?)")
			_, err = statement.Exec(account.Id, 0, start_year)
			if err != nil {
				return err
			}
		}
		start_year++

	}
	return nil
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
	var contri = new(models.Contribution)
	var account models.Account
	var oldValue float64 = 1
	var oldTotal float64

	accountId := c.Param("id")
	err := c.BindJSON(contri)

	statement, _ := db.DB.Prepare("SELECT * FROM accounts WHERE id = ?")
	rows, err := statement.Query(accountId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}
	for rows.Next() {
		rows.Scan(&account.Id, &account.User_id, &account.Account_type_id, &account.Total, &account.Child_year)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	validate_error := ValidateContribution(contri, &account)
	if validate_error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error validating contribution"})
		return
	}

	err = db.DB.QueryRow("SELECT amount FROM contributions WHERE year = ? AND account_id = ?", contri.Year, accountId).Scan(&oldValue)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}
	if oldValue != contri.Amount {
		err = db.DB.QueryRow("SELECT total FROM accounts WHERE id = ?", accountId).Scan(&oldTotal)
		newTotal := oldTotal + (contri.Amount - oldValue)
		statement, _ = db.DB.Prepare("UPDATE accounts SET total = ? WHERE id = ?")
		_, err = statement.Exec(newTotal, accountId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

	}

	statement, _ = db.DB.Prepare("UPDATE contributions SET amount = ? WHERE year = ?")
	_, err = statement.Exec(contri.Amount, contri.Year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if account.Account_type_id == 1 || account.Account_type_id == 3 {
		err := calculateCumulativeContribution(accountId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	} else {
		err := calculateGrantContribution(accountId, oldValue, contri.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

	}

	c.JSON(http.StatusOK, gin.H{"message": "Contribution created successfully"})
}

func ValidateContribution(contribution *models.Contribution, account *models.Account) error {

	//  TFSA Calculations
	if account.Account_type_id == 1 {
		oldValue := 0.0

		cumulative_contributions, err := getContributionCumulative(account.Id, contribution.Year)
		if err != nil {
			return err
		}
		_ = db.DB.QueryRow("SELECT amount FROM contributions WHERE year = ?", contribution.Year).Scan(&oldValue)

		if contribution.Amount > cumulative_contributions+oldValue {
			err := errors.New("Contribution limit exceeded")
			fmt.Print(cumulative_contributions)
			return err
		} else if contribution.Amount == cumulative_contributions {

			return nil
		}

		// RESP CALCULATION
	} else if account.Account_type_id == 2 {
		if account.Total+contribution.Amount > 50000 {
			err := errors.New("Contribution limit exceeded")
			return err
		}

	}
	return nil

}

func getContributionCumulative(accountId int, year int) (float64, error) {
	var amount float64
	err := db.DB.QueryRow("SELECT amount FROM cumulative_contributions WHERE account_id = ? AND year = ?", accountId, year).Scan(&amount)
	if err != nil {
		return amount, err
	}
	return amount, nil

}

func calculateCumulativeContribution(accountID string) error {
	// Step 1: Fetch all contribution limits

	limitQuery := "SELECT year, contribute_limit FROM contribution_limit WHERE account_type_id = 1"
	limitRows, err := db.DB.Query(limitQuery)
	if err != nil {
		return err
	}

	contributionLimits := make(map[int]float64)
	for limitRows.Next() {
		var year int
		var amount float64
		if err := limitRows.Scan(&year, &amount); err != nil {
			return err
		}
		contributionLimits[year] = amount
	}

	// Step 2: Fetch all contributions for account
	contributionQuery := "SELECT year, amount FROM contributions WHERE account_id = ?"
	contributionRows, err := db.DB.Query(contributionQuery, accountID)
	if err != nil {
		return err
	}
	contributions := make(map[int]float64)
	for contributionRows.Next() {
		var year int
		var amount float64
		if err := contributionRows.Scan(&year, &amount); err != nil {
			return err
		}
		contributions[year] = amount
	}

	// Step 3: Calculate cumulative and track over-contributions
	cumulative := make(map[int]float64)
	overContributions := make(map[int]float64)

	for year := 2024; year <= 2025; year++ {
		var prevCumulative float64
		if year > 2024 {
			prevCumulative = cumulative[year-1]
		}
		limit := contributionLimits[year]
		contribution := contributions[year]

		newCumulative := prevCumulative + (limit - contribution)
		cumulative[year] = newCumulative

		if contribution > prevCumulative+limit {
			overContributions[year] = contribution - (prevCumulative + limit)
		} else {
			overContributions[year] = 0
		}
		if cumulative[year] < 0 {
			cumulative[year] = contribution + cumulative[year]
		}
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("UPDATE cumulative_contributions SET amount = ?, over_contribution_amount = ? WHERE account_id = ? AND year = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for year := 2024; year <= 2025; year++ {
		_, err := stmt.Exec(cumulative[year], overContributions[year], accountID, year)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func calculateGrantContribution(accountID string, oldValue float64, newValue float64) error {
	// grantUsed := db.DB.QueryRow("SELECT grant_unused FROM cumulative_grants WHERE account_id = ?", accountID)
	var grantEarned float64
	var grantUnused float64
	oldGrantEarned := min(500, oldValue*0.20)
	newGrantEarned := min(500, newValue*0.20)
	fmt.Println(oldGrantEarned, newGrantEarned)

	oldGrantUnused := 500 - oldGrantEarned
	newGrantUnused := 500 - newGrantEarned
	fmt.Println(oldGrantUnused, newGrantUnused)

	err1 := db.DB.QueryRow("SELECT grant_earned FROM cumulative_grants WHERE account_id = ?", accountID).Scan(&grantEarned)
	err2 := db.DB.QueryRow("SELECT grant_unused FROM cumulative_grants WHERE account_id = ?", accountID).Scan(&grantUnused)
	if err1 != nil || err2 != nil {
		return err1
	}
	fmt.Println(grantEarned, grantUnused)
	grantEarned = grantEarned - oldGrantEarned + newGrantEarned
	grantUnused = grantUnused - oldGrantUnused + newGrantUnused

	_, err := db.DB.Exec("UPDATE cumulative_grants SET grant_earned = ?, grant_unused = ? WHERE account_id = ?", grantEarned, grantUnused, accountID)
	if err != nil {
		return err
	}

	return nil
}
