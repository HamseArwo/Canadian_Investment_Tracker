package calculator

import (
	"errors"
	"fmt"
	db "investment_tracker/database"
	"investment_tracker/models"
	"math"
	"sync"
)

func ValidateContribution(contribution *models.Contribution, account *models.Account) error {

	//  TFSA Calculations
	if account.AccountTypeId == 1 || account.AccountTypeId == 3 {
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
	} else if account.AccountTypeId == 2 {
		if account.Total+contribution.Amount > 50000 {
			err := errors.New("Contribution limit exceeded")
			return err
		}

		// RRSP CALCULATION
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

func CalculateCumulativeContribution(accountID string, accountTypeID int, userID int, birthYear int) error {
	errChan := make(chan error, 4)
	salary := make(map[int]float64)
	contributionLimits := make(map[int]float64)
	contributions := make(map[int]float64)
	var wg sync.WaitGroup
	salaryRows, err := db.DB.Query(`SELECT year, amount FROM salary WHERE user_id = ?`, userID)
	if err != nil {
		return err
	}
	defer salaryRows.Close()
	for salaryRows.Next() {
		var year int
		var amount float64
		err := salaryRows.Scan(&year, &amount)
		if err != nil {
			return err
		}
		salary[year] = amount
	}

	wg.Add(2)

	go func() {
		defer wg.Done()
		if accountTypeID == 1 { // TFSA
			rows, err := db.DB.Query(`SELECT year, contribute_limit FROM contribution_limit WHERE account_type_id = 1`)
			if err != nil {
				errChan <- err
				return
			}
			defer rows.Close()
			for rows.Next() {
				var year int
				var limit float64
				err := rows.Scan(&year, &limit)
				if err != nil {
					errChan <- err
					return
				}
				contributionLimits[year] = limit
			}
		} else if accountTypeID == 3 { // RRSP
			rows, err := db.DB.Query(`SELECT year, rrsp_limit FROM rrsp_cap WHERE account_type_id = 3`)
			if err != nil {
				errChan <- err
				return
			}
			defer rows.Close()
			for rows.Next() {
				var year int
				var cap float64
				err := rows.Scan(&year, &cap)
				if err != nil {
					errChan <- err
					return
				}
				salaryAmount := salary[year]
				contributionLimits[year] = math.Min(cap, salaryAmount*0.18)
			}
		}
	}()

	go func() {
		contributionRows, err := db.DB.Query(`SELECT year, amount FROM contributions WHERE account_id = ?`, accountID)
		if err != nil {
			errChan <- err
			return
		}
		defer contributionRows.Close()
		defer wg.Done()

		for contributionRows.Next() {
			var year int
			var amount float64
			err := contributionRows.Scan(&year, &amount)
			if err != nil {
				errChan <- err
				return
			}
			contributions[year] = amount
		}
	}()
	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	cumulative := make(map[int]float64)
	overContributions := make(map[int]float64)
	var previous float64 = 0

	startYear := GetStartYear(accountTypeID, birthYear)

	for year := startYear; year <= 2025; year++ {
		limit := contributionLimits[year]
		actual := contributions[year]

		newRoom := previous + limit - actual

		if newRoom < 0 {
			overContributions[year] = -newRoom
			cumulative[year] = 0
		} else {
			overContributions[year] = 0
			cumulative[year] = newRoom
		}
		// fmt.Println(limit, actual, cumulative[year], year, accountID, accountTypeID)

		previous = cumulative[year]
	}

	// Step 6: Upsert cumulative_contributions
	tx, err := db.DB.Begin()
	if err != nil {

		return err
	}
	stmt, err := tx.Prepare(`
		UPDATE cumulative_contributions
			SET amount = ?, over_contribution_amount = ?
			WHERE account_id = ? AND year = ?
	`)

	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for year := startYear; year <= 2025; year++ {
		fmt.Println(accountID, "<--")
		fmt.Println(cumulative[year], overContributions[year], accountID, year)
		res, err := stmt.Exec(cumulative[year], overContributions[year], accountID, year)
		if err != nil {
			tx.Rollback()
			return err
		}
		affected, _ := res.RowsAffected()
		fmt.Printf("Year %d: updated %d rows\n", year, affected)
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println("THIS FUCKING RETARED ERROR")
		return err
	}
	return nil
}

func CalculateGrantContribution(accountID string, oldValue float64, newValue float64) error {
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

func GetStartYear(accountTypeID int, birthYear int) int {
	const tfsaStart = 2009
	userAdultYear := birthYear + 18

	if accountTypeID == 1 { // TFSA
		if userAdultYear > tfsaStart {
			return userAdultYear
		}
		return tfsaStart
	}

	return userAdultYear // RRSP (or RESP/others if needed)
}
