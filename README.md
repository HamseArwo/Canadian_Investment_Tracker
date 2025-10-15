

# 💰 Trackvestment Backend

### Overview

Trackvestment is a backend system that helps users track their **TFSA**, **RRSP**, and **RESP** investments.
It makes sure users **don’t go over their contribution limits** each year.

The backend is built with **Go (Golang)** using the **Gin** web framework and uses **SQLite** for storing data.

---

## ⚙️ Tech Stack

* **Language:** Go
* **Framework:** Gin
* **Database:** SQLite
* **CORS:** gin-contrib/cors

---

## 🧩 Features

* Create and manage users
* Track TFSA, RRSP, and RESP accounts
* Record and check yearly contributions
* Detect over-contributions
* Track RESP grants and unused amounts
* Store and use salary data for contribution calculations

---



## 🗃️ Database Tables

* **users** – user info (name, email, password, birthyear)
* **accounts** – user accounts (TFSA, RRSP, RESP)
* **contributions** – user yearly contributions
* **cumulative_contributions** – total contributions per account
* **cumulative_grants** – RESP grants tracking
* **salary** – user salary info

---

## 🚀 Getting Started

### 1. Requirements

* [Go 1.20+](https://go.dev/dl/)
* [SQLite](https://www.sqlite.org/download.html)

### 2. Clone the Project

```bash
git clone https://github.com/<your-username>/trackvestment-backend.git
cd trackvestment-backend
```

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Run the Server

```bash
go run main.go
```

The server will run at: **[http://localhost:8080](http://localhost:8080)**

---



---

## Future Ideas

* JWT user authentication
* Charts and data visualization
* Email alerts for over-contribution
* Historical investment tracking

---

Made by
Hamse Arwo
arwohamse@gmail.com

---
