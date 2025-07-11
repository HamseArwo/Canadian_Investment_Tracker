package db

import (
	// sql pacakge

	// Importing the sqlite3 driver
	"database/sql"
	"investment_tracker/models"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// initialize database
var DB *sql.DB

func InitDB() {
	DB, err := sql.Open("sqlite3", "database/tracker.db")
	if err != nil {
		log.Fatal(err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatal(err)
	}

	// models.CreateUserTable(DB)
	models.CreateAccountTable(DB)
	models.CreateContributionTable(DB)
	models.CreateCumulativeContributionTable(DB)
	models.CreateCumulativeGrantTable(DB)
	models.CreateSalaryTable(DB)

	// fmt.Println("Hey this from db")
}
