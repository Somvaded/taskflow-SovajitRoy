package auth_usecase

import (
	"context"
	"fmt"
	"time"

	domain_error "taskflow/internal/domain/errors"
	domain_user "taskflow/internal/domain/user"
	"taskflow/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type useCase struct {
	users  domain_user.Repository
	config *utils.Config
}

func New(config *utils.Config, users domain_user.Repository) domain_user.UseCase {
	return &useCase{users: users, config: config}
}

func (uc *useCase) Register(ctx context.Context, input domain_user.RegisterInput) (*domain_user.AuthResult, error) {
	existing, err := uc.users.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if existing != nil {
		return nil, domain_error.Raise(domain_error.CODE_EMAIL_TAKEN, "", nil)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", fmt.Errorf("hash password: %w", err))
	}

	user := &domain_user.User{
		ID:           uuid.New(),
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}

	token, err := uc.generateToken(user)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}

	return &domain_user.AuthResult{Token: token, User: user}, nil
}

func (uc *useCase) Login(ctx context.Context, input domain_user.LoginInput) (*domain_user.AuthResult, error) {
	user, err := uc.users.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}
	if user == nil {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_CREDENTIALS, "", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INVALID_CREDENTIALS, "", nil)
	}

	token, err := uc.generateToken(user)
	if err != nil {
		return nil, domain_error.Raise(domain_error.CODE_INTERNAL_ERROR, "", err)
	}

	return &domain_user.AuthResult{Token: token, User: user}, nil
}

func (uc *useCase) generateToken(user *domain_user.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"exp":     time.Now().Add(uc.config.JWTExpiration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.config.JWTSecret))
}
