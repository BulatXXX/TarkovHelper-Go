package repo

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRow struct {
	ID           string
	Email        string
	Name         string
	AvatarURL    *string
	PasswordHash string
}

var ErrEmailTaken = errors.New("email taken")
var ErrUserNotFound = errors.New("user not found")

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo { return &UserRepo{db: db} }

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (r *UserRepo) Create(ctx context.Context, email, name, passwordHash string) (UserRow, error) {
	email = normalizeEmail(email)

	var u UserRow
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (email, name, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id::text, email, name, avatar_url, password_hash
	`, email, name, passwordHash).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.PasswordHash)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // unique_violation
				return UserRow{}, ErrEmailTaken
			}
		}
		return UserRow{}, err
	}
	return u, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (UserRow, error) {
	email = normalizeEmail(email)

	var u UserRow
	err := r.db.QueryRow(ctx, `
		SELECT id::text, email, name, avatar_url, password_hash
		FROM users
		WHERE email = $1
	`, email).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.PasswordHash)

	if errors.Is(err, pgx.ErrNoRows) {
		return UserRow{}, ErrUserNotFound
	}
	return u, err
}

func (r *UserRepo) FindByID(ctx context.Context, userID string) (UserRow, error) {
	var u UserRow
	err := r.db.QueryRow(ctx, `
		SELECT id::text, email, name, avatar_url, password_hash
		FROM users
		WHERE id = $1::uuid
	`, userID).Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.PasswordHash)

	if errors.Is(err, pgx.ErrNoRows) {
		return UserRow{}, ErrUserNotFound
	}
	return u, err
}
