package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Thedrogon/blogbish/post-service/internal/models"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	GetByID(ctx context.Context, id int64) (*models.Category, error)
	GetBySlug(ctx context.Context, slug string) (*models.Category, error)
	List(ctx context.Context) ([]*models.Category, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, id int64) error
}

type PostgresCategoryRepository struct {
	db *sql.DB
}

func NewPostgresCategoryRepository(db *sql.DB) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{db: db}
}

func (r *PostgresCategoryRepository) Create(ctx context.Context, category *models.Category) error {
	query := `
		INSERT INTO categories (name, slug, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx,
		query,
		category.Name,
		category.Slug,
		category.Description,
		category.CreatedAt,
		category.UpdatedAt,
	).Scan(&category.ID)

	if err != nil {
		return fmt.Errorf("error creating category: %w", err)
	}

	return nil
}

func (r *PostgresCategoryRepository) GetByID(ctx context.Context, id int64) (*models.Category, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM categories
		WHERE id = $1`

	category := &models.Category{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category not found")
	}

	if err != nil {
		return nil, fmt.Errorf("error getting category: %w", err)
	}

	return category, nil
}

func (r *PostgresCategoryRepository) GetBySlug(ctx context.Context, slug string) (*models.Category, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM categories
		WHERE slug = $1`

	category := &models.Category{}
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category not found")
	}

	if err != nil {
		return nil, fmt.Errorf("error getting category: %w", err)
	}

	return category, nil
}

func (r *PostgresCategoryRepository) List(ctx context.Context) ([]*models.Category, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM categories
		ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error listing categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		category := &models.Category{}
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning category: %w", err)
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

func (r *PostgresCategoryRepository) Update(ctx context.Context, category *models.Category) error {
	query := `
		UPDATE categories
		SET name = $1, slug = $2, description = $3, updated_at = $4
		WHERE id = $5`

	category.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		category.Name,
		category.Slug,
		category.Description,
		category.UpdatedAt,
		category.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

func (r *PostgresCategoryRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM categories WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting category: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}
