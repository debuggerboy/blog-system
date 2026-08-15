package repository

import (
	"blog-system/internal/models"
	"database/sql"
	"errors"
	"time"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository() *UserRepository {
	return &UserRepository{db: DB}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, role, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, user.Username, user.Email, user.PasswordHash, 
		user.Role, user.Status, time.Now(), time.Now())
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	user.ID = id
	return nil
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, status, created_at, updated_at 
			  FROM users WHERE username = ?`
	row := r.db.QueryRow(query, username)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, 
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, status, created_at, updated_at 
			  FROM users WHERE email = ?`
	row := r.db.QueryRow(query, email)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, 
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id int64) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, role, status, created_at, updated_at 
			  FROM users WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, 
		&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) List() ([]models.User, error) {
	query := `SELECT id, username, email, role, status, created_at, updated_at 
			  FROM users ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Role, 
			&user.Status, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepository) Update(user *models.User) error {
	query := `UPDATE users SET username = ?, email = ?, role = ?, status = ?, updated_at = ? 
			  WHERE id = ?`
	_, err := r.db.Exec(query, user.Username, user.Email, user.Role, user.Status, 
		time.Now(), user.ID)
	return err
}

func (r *UserRepository) UpdatePassword(userID int64, passwordHash string) error {
	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, passwordHash, time.Now(), userID)
	return err
}

func (r *UserRepository) Delete(id int64) error {
	query := `DELETE FROM users WHERE id = ?`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *UserRepository) SaveRefreshToken(userID int64, token string, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, userID, token, expiresAt)
	return err
}

func (r *UserRepository) GetRefreshToken(token string) (int64, error) {
	query := `SELECT user_id FROM refresh_tokens WHERE token = ? AND expires_at > ?`
	row := r.db.QueryRow(query, token, time.Now())

	var userID int64
	err := row.Scan(&userID)
	if err == sql.ErrNoRows {
		return 0, errors.New("invalid or expired refresh token")
	}
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *UserRepository) DeleteRefreshToken(token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = ?`
	_, err := r.db.Exec(query, token)
	return err
}

func (r *UserRepository) DeleteUserRefreshTokens(userID int64) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = ?`
	_, err := r.db.Exec(query, userID)
	return err
}
