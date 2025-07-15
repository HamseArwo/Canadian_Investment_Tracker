package controllers

import (
	"errors"
	"fmt"
	db "investment_tracker/database"
	"investment_tracker/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateContribution(account models.Account, user models.User) {
	var year int
	if user.Birthyear+18 < 2005 {
		year = 2005
	} else {
		year = user.Birthyear + 18
	}
	for year < 2026 {
		statement, _ := db.DB.Prepare("INSERT INTO contributions (account_id, amount, year) VALUES (?, ?, ?)")
		_, err := statement.Exec(account.Id, 0, year)
		if err != nil {
			return
		}
		statement, _ = db.DB.Prepare("INSERT INTO cumulative_contributions (account_id, amount, year) VALUES (?, ?, ?)")
		_, err = statement.Exec(account.Id, 0, year)
		if err != nil {
			return
		}
		year++
	}

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
	var oldValue float64
	var oldTotal float64

	accountId := c.Param("id")
	err := c.BindJSON(contri)
	fmt.Println(accountId)
	statement, _ := db.DB.Prepare("SELECT * FROM accounts WHERE id = ?")
	rows, err := statement.Query(accountId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}
	for rows.Next() {
		rows.Scan(&account.Id, &account.User_id, &account.Account_type_id, &account.Total)
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
	err = db.DB.QueryRow("SELECT total FROM accounts WHERE id = ?", accountId).Scan(&oldTotal)
	fmt.Println(oldTotal, oldValue, contri.Amount)
	newTotal := oldTotal + (contri.Amount - oldValue)

	statement, _ = db.DB.Prepare("UPDATE contributions SET amount = ? WHERE year = ?")
	_, err = statement.Exec(contri.Amount, contri.Year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	statement, _ = db.DB.Prepare("UPDATE accounts SET total = ? WHERE id = ?")
	_, err = statement.Exec(newTotal, accountId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	err = calculateCumulativeContribution(accountId, contri.Year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Intsdsadernal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contribution created successfully"})
}

func ValidateContribution(contribution *models.Contribution, account *models.Account) error {

	//  TFSA Calculations
	if account.Account_type_id == 1 {
		cumulative_contributions, err := getContributionCumulative(account.Id)
		if err != nil {
			return err
		}
		// contrtibution limit of that year
		// compare contrbutions if greater than limit cancel if equal to the limit add if less then limit added to culnmative
		contribution_limit, err := getContributionLimit(contribution.Year)
		if err != nil {
			return err
		}

		if contribution.Amount > contribution_limit.Amount+cumulative_contributions.Amount {
			err := errors.New("Contribution limit exceeded")
			fmt.Print(cumulative_contributions.Amount)
			return err
		} else if contribution.Amount == contribution_limit.Amount+cumulative_contributions.Amount {
			fmt.Print("Bye")

			return nil
		}

	}
	return nil

}
func getContributionLimit(year int) (models.ContributionLimit, error) {
	var contributionLimit models.ContributionLimit

	statement, _ := db.DB.Prepare("SELECT * FROM contribution_limit WHERE year = ?")
	rows, err := statement.Query(year)
	if err != nil {
		return contributionLimit, err
	}
	for rows.Next() {
		rows.Scan(&contributionLimit.Id, &contributionLimit.Account_type_id, &contributionLimit.Amount, &contributionLimit.Year)
	}
	return contributionLimit, nil
}
func getContributionCumulative(accountId int) (models.CumulativeContribution, error) {
	var contributionCumulative models.CumulativeContribution

	statement, _ := db.DB.Prepare("SELECT * FROM cumulative_contributions WHERE account_id = ?")
	rows, err := statement.Query(accountId)
	if err != nil {
		return contributionCumulative, err
	}
	for rows.Next() {
		rows.Scan(&contributionCumulative.Id, &contributionCumulative.Account_id, &contributionCumulative.Amount, &contributionCumulative.Year)
	}
	return contributionCumulative, nil
}

func calculateCumulativeContribution(accountID string, contributionYear int) error {
	limitQuery := "SELECT year, contribute_limit FROM contribution_limit WHERE account_type_id = 1"
	limitRows, err := db.DB.Query(limitQuery)
	if err != nil {
		fmt.Println("hello world")
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

	cumulative := make(map[int]float64)
	for year := contributionYear; year <= 2025; year++ {
		contribution := contributions[year]
		limit := contributionLimits[year]
		cumulative[year+1] = cumulative[year] + (limit - contribution)
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("UPDATE cumulative_contributions SET amount = ? WHERE account_id = ? AND year = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for year, amount := range cumulative {
		_, err := stmt.Exec(amount, accountID, year)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
