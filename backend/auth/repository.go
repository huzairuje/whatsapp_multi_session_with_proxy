package auth

import (
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUsersTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) CreateUsersTablePostgres() error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *Repository) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, created_at, updated_at FROM users WHERE username = ?`
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) GetUserByUsernamePostgres(username string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, password, created_at, updated_at FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *Repository) CreateUser(username, hashedPassword string) error {
	query := `INSERT INTO users (username, password, created_at, updated_at) VALUES (?, ?, ?, ?)`
	_, err := r.db.Exec(query, username, hashedPassword, time.Now(), time.Now())
	return err
}

func (r *Repository) CreateUserPostgres(username, hashedPassword string) error {
	query := `INSERT INTO users (username, password, created_at, updated_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, username, hashedPassword, time.Now(), time.Now())
	return err
}

func (r *Repository) UpdatePassword(username, hashedPassword string) error {
	query := `UPDATE users SET password = ?, updated_at = ? WHERE username = ?`
	_, err := r.db.Exec(query, hashedPassword, time.Now(), username)
	return err
}

func (r *Repository) UpdatePasswordPostgres(username, hashedPassword string) error {
	query := `UPDATE users SET password = $1, updated_at = $2 WHERE username = $3`
	_, err := r.db.Exec(query, hashedPassword, time.Now(), username)
	return err
}

func (r *Repository) CountUsers() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users`
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}
