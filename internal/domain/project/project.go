package domain_project

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	OwnerID     uuid.UUID `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateInput struct {
	Name        string
	Description string
	OwnerID     uuid.UUID
}

type UpdateInput struct {
	Name        *string
	Description *string
}

type Pagination struct {
	Page     int
	PageSize int
}

type Repository interface {
	Create(ctx context.Context, project *Project) error
	// GetAllForUser returns projects where user is owner or assignee on any task.
	GetAllForUser(ctx context.Context, userID uuid.UUID, p Pagination) ([]*Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (*Project, error)
	List(ctx context.Context, userID uuid.UUID, p Pagination) ([]*Project, error)
	Get(ctx context.Context, id uuid.UUID) (*Project, error)
	Update(ctx context.Context, id uuid.UUID, callerID uuid.UUID, input UpdateInput) (*Project, error)
	Delete(ctx context.Context, id uuid.UUID, callerID uuid.UUID) error
}
