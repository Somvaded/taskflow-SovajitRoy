package project_repository

import (
	"context"
	"errors"
	"time"

	domain_project "taskflow/internal/domain/project"
	db "taskflow/internal/repository/project/driver/postgres"
	postgres "taskflow/utils/database/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type repository struct {
	q *db.Queries
}

func New(conn *postgres.DBConnector) domain_project.Repository {
	return &repository{q: db.New(conn.Pool)}
}

func (r *repository) Create(ctx context.Context, project *domain_project.Project) error {
	return r.q.CreateProject(ctx, db.CreateProjectParams{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		OwnerID:     project.OwnerID,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	})
}

func (r *repository) GetAllForUser(ctx context.Context, userID uuid.UUID, p domain_project.Pagination) ([]*domain_project.Project, error) {
	rows, err := r.q.GetProjectsForUser(ctx, db.GetProjectsForUserParams{
		OwnerID: userID,
		Limit:   int32(p.PageSize),
		Offset:  int32((p.Page - 1) * p.PageSize),
	})
	if err != nil {
		return nil, err
	}
	projects := make([]*domain_project.Project, len(rows))
	for i, row := range rows {
		projects[i] = toProject(row)
	}
	return projects, nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*domain_project.Project, error) {
	row, err := r.q.GetProjectByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return toProject(row), nil
}

func (r *repository) Update(ctx context.Context, id uuid.UUID, input domain_project.UpdateInput) (*domain_project.Project, error) {
	row, err := r.q.UpdateProject(ctx, db.UpdateProjectParams{
		ID:          id,
		UpdatedAt:   time.Now(),
		Name:        input.Name,
		Description: input.Description,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return toProject(row), nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteProject(ctx, id)
}

func toProject(row db.Project) *domain_project.Project {
	return &domain_project.Project{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		OwnerID:     row.OwnerID,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
