package repository

import (
	"blog-system/internal/models"
	"database/sql"
	"time"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository() *PostRepository {
	return &PostRepository{db: DB}
}

func (r *PostRepository) Create(post *models.Post) error {
	query := `
		INSERT INTO posts (title, content, author_id, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, post.Title, post.Content, post.AuthorID,
		post.Status, time.Now(), time.Now())
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	post.ID = id
	return nil
}

func (r *PostRepository) FindByID(id int64) (*models.Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.author_id, p.status, p.created_at, p.updated_at,
			   u.id, u.username, u.email, u.role, u.status
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		WHERE p.id = ?
	`
	row := r.db.QueryRow(query, id)

	var post models.Post
	var user models.User
	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID,
		&post.Status, &post.CreatedAt, &post.UpdatedAt,
		&user.ID, &user.Username, &user.Email, &user.Role, &user.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	post.Author = user
	return &post, nil
}

func (r *PostRepository) List(limit, offset int, filterAuthorID *int64, filterStatus *string) ([]models.Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.author_id, p.status, p.created_at, p.updated_at,
			   u.id, u.username, u.email, u.role, u.status
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		WHERE 1=1
	`
	args := []interface{}{}

	if filterAuthorID != nil {
		query += " AND p.author_id = ?"
		args = append(args, *filterAuthorID)
	}

	if filterStatus != nil {
		query += " AND p.status = ?"
		args = append(args, *filterStatus)
	}

	query += " ORDER BY p.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var user models.User
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID,
			&post.Status, &post.CreatedAt, &post.UpdatedAt,
			&user.ID, &user.Username, &user.Email, &user.Role, &user.Status)
		if err != nil {
			return nil, err
		}
		post.Author = user
		posts = append(posts, post)
	}
	return posts, nil
}

func (r *PostRepository) Count() (int, error) {
	query := `SELECT COUNT(*) FROM posts`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

func (r *PostRepository) Update(post *models.Post) error {
	query := `UPDATE posts SET title = ?, content = ?, status = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, post.Title, post.Content, post.Status, time.Now(), post.ID)
	return err
}

func (r *PostRepository) Delete(id int64) error {
	query := `DELETE FROM posts WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *PostRepository) ListByAuthor(authorID int64, limit, offset int) ([]models.Post, error) {
	query := `
		SELECT p.id, p.title, p.content, p.author_id, p.status, p.created_at, p.updated_at,
			   u.id, u.username, u.email, u.role, u.status
		FROM posts p
		LEFT JOIN users u ON p.author_id = u.id
		WHERE p.author_id = ?
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.Query(query, authorID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		var user models.User
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID,
			&post.Status, &post.CreatedAt, &post.UpdatedAt,
			&user.ID, &user.Username, &user.Email, &user.Role, &user.Status)
		if err != nil {
			return nil, err
		}
		post.Author = user
		posts = append(posts, post)
	}
	return posts, nil
}
