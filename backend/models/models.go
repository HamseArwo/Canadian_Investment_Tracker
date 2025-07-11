package models

import (
	"database/sql"
	"fmt"
)

type user struct {
	id        int    `json:"id"`
	name      string `json:"name"`
	email     string `json:"email"`
	password  string `json:"password"`
	birthdate string `json:"birthdate"`
}

type account struct {
	id              int     `json:"id"`
	user_id         int     `json:"user_id"`
	account_type_id int     `json:"account_type_id"`
	total           float64 `json:"total"`
}

type contribution struct {
	id         int     `json:"id"`
	account_id int     `json:"account_id"`
	amount     float64 `json:"amount"`
	year       int     `json:"date"`
}
type cumulative_contribution struct {
	id         int     `json:"id"`
	account_id int     `json:"account_id"`
	amount     float64 `json:"amount"`
}

type grant_cumulative struct {
	id           int `json:"id"`
	account_id   int `json:"account_id"`
	grant_earned int `json:"grant_earned"`
	grant_unused int `json:"grant_unused"`
}

type salary struct {
	id         int     `json:"id"`
	user_id    int     `json:"user_id"`
	amount     float64 `json:"amount"`
	start_year int     `json:"start_year"`
	end_year   int     `json:"end_year"`
}

func CreateUserTable(DB *sql.DB) {

	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			birthyear INTEGER NOT NULL
		);
	`)
	if err != nil {
		fmt.Println(" Failed to create users table")
		panic(err)
	}
	// fmt.Println("Sucessfully Added")
}

func CreateAccountTable(DB *sql.DB) {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL REFERENCES users(id),
			account_type_id INTEGER REFERENCES account_types(id),
			total REAL NOT NULL

		);
	`)
	if err != nil {
		fmt.Println(" Failed to create account table")

		panic(err)
	}
}

func CreateContributionTable(DB *sql.DB) {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS contributions (
			id SERIAL PRIMARY KEY,
			account_id INTEGER REFERENCES accounts(id),
			amount REAL NOT NULL,
			year INTEGER NOT NULL
		);
	`)
	if err != nil {
		panic(err)
	}
}

func CreateCumulativeContributionTable(DB *sql.DB) {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS cumulative_contributions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER REFERENCES accounts(id),
			amount REAL NOT NULL
		);
	`)
	if err != nil {
		panic(err)
	}
}

func CreateSalaryTable(DB *sql.DB) {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS salary (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER REFERENCES users(id),
			amount REAL NOT NULL,
			start_year INTEGER NOT NULL,
			end_year INTEGER
		);
	`)
	if err != nil {
		panic(err)
	}
}

func CreateCumulativeGrantTable(DB *sql.DB) {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS cumulative_grants (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER REFERENCES accounts(id),
			grant_earned REAL NOT NULL,
			grant_unused REAL NOT NULL
		);
	`)
	if err != nil {
		panic(err)
	}
}
