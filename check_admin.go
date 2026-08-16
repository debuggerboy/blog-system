package main

import (
	"blog-system/internal/repository"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if err := repository.InitDB(); err != nil {
		log.Fatal(err)
	}
	defer repository.CloseDB()

	var username, passwordHash string
	err := repository.DB.QueryRow("SELECT username, password_hash FROM users WHERE username = 'admin'").Scan(&username, &passwordHash)
	if err == sql.ErrNoRows {
		fmt.Println("Admin user not found!")
		return
	}
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found user: %s\n", username)
	fmt.Printf("Password hash: %s\n", passwordHash)

	// Test the password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("admin123"))
	if err == nil {
		fmt.Println("✅ Password 'admin123' is CORRECT")
	} else {
		fmt.Println("❌ Password 'admin123' is INCORRECT")
		fmt.Printf("Error: %v\n", err)
	}
}
