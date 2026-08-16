package main

import (
	"blog-system/internal/repository"
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

	// Generate new hash for "admin123"
	newPassword := "admin123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}

	// Update admin password
	result, err := repository.DB.Exec(
		"UPDATE users SET password_hash = ? WHERE username = 'admin'",
		string(hashedPassword),
	)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Admin doesn't exist, create it
		_, err = repository.DB.Exec(
			"INSERT INTO users (username, email, password_hash, role, status) VALUES (?, ?, ?, ?, ?)",
			"admin", "admin@blog.com", string(hashedPassword), "admin", "active",
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("✅ Admin user created with password: admin123")
	} else {
		fmt.Println("✅ Admin password reset to: admin123")
	}

	// Verify the password works
	var passwordHash string
	err = repository.DB.QueryRow("SELECT password_hash FROM users WHERE username = 'admin'").Scan(&passwordHash)
	if err != nil {
		log.Fatal(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("admin123"))
	if err == nil {
		fmt.Println("✅ Verification: Password 'admin123' is CORRECT")
	} else {
		fmt.Println("❌ Verification failed!")
	}
}
