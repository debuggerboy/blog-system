package main

import (
	"blog-system/internal/repository"
	"log"
)

func main() {
	if err := repository.InitDB(); err != nil {
		log.Fatal(err)
	}
	defer repository.CloseDB()

	// Add pinned column
	_, err := repository.DB.Exec("ALTER TABLE posts ADD COLUMN pinned BOOLEAN DEFAULT 0")
	if err != nil {
		log.Printf("Note: %v (column may already exist)", err)
	}

	// Create index
	_, err = repository.DB.Exec("CREATE INDEX IF NOT EXISTS idx_posts_pinned ON posts(pinned)")
	if err != nil {
		log.Printf("Note: %v (index may already exist)", err)
	}

	log.Println("Migration completed successfully!")
}
