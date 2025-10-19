package repository

import (
	"database/sql"
	"fmt"

	"github.com/polyakovaa/grpcproxy/auth_service/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}
func (r *UserRepository) CreateUser(u *model.User) (*model.User, error) {
	query := `INSERT INTO users (user_name, email, password_hash) VALUES ($1, $2, $3) RETURNING id`

	if err := r.db.QueryRow(
		query,
		u.UserName,
		u.Email,
		u.PasswordHash,
	).Scan(&u.ID); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) FindByID(id string) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, user_name, email, password_hash FROM users WHERE id = $1`
	if err := r.db.QueryRow(query, id).Scan(
		&u.ID,
		&u.UserName,
		&u.Email,
		&u.PasswordHash,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with id '%s' not found", id)
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	u := &model.User{}
	query := `SELECT id, user_name, email, password_hash FROM users WHERE email = $1`
	if err := r.db.QueryRow(query, email).Scan(
		&u.ID,
		&u.UserName,
		&u.Email,
		&u.PasswordHash,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email '%s' not found", email)
		}
		return nil, err
	}
	return u, nil
}
