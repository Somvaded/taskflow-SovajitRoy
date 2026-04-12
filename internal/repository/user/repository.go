package user_repository

import (
	"context"
	"errors"

	domain_user "taskflow/internal/domain/user"
	db "taskflow/internal/repository/user/driver/postgres"
	postgres "taskflow/utils/database/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type repository struct {
	q *db.Queries
}

func New(conn *postgres.DBConnector) domain_user.Repository {
	return &repository{q: db.New(conn.Pool)}
}

func (r *repository) Create(ctx context.Context, user *domain_user.User) error {
	return r.q.CreateUser(ctx, db.CreateUserParams{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	})
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*domain_user.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return toUser(row), nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*domain_user.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return toUser(row), nil
}

func toUser(row db.User) *domain_user.User {
	return &domain_user.User{
		ID:           row.ID,
		Name:         row.Name,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
	}
}
