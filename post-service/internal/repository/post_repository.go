package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Thedrogon/blogbish/post-service/internal/models"
	"github.com/lib/pq"
)

type PostRepository interface {
	Create(ctx context.Context, post *models.Post) error
	GetByID(ctx context.Context, id int64) (*models.Post, error)
	GetBySlug(ctx context.Context, slug string) (*models.Post, error)
	Update(ctx context.Context, post *models.Post) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter *models.PostFilter) ([]*models.Post, error)
	IncrementViewCount(ctx context.Context, id int64) error
}

type PostgresPostRepository struct {
	db *sql.DB
}

func NewPostgresPostRepository(db *sql.DB) *PostgresPostRepository {
	return &PostgresPostRepository{db: db}
}

func (r *PostgresPostRepository) Create(ctx context.Context, post *models.Post) error {
	query := `
		INSERT INTO posts (title, content, slug, author_id, category_id, status, tags, published_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.Slug,
		post.AuthorID,
		post.CategoryID,
		post.Status,
		pq.Array(post.Tags),
		post.PublishedAt,
		post.CreatedAt,
		post.UpdatedAt,
	).Scan(&post.ID)

	if err != nil {
		return fmt.Errorf("error creating post: %w", err)
	}

	return nil
}

func (r *PostgresPostRepository) GetByID(ctx context.Context, id int64) (*models.Post, error) {
	query := `
		SELECT id, title, content, slug, author_id, category_id, status, tags, view_count, published_at, created_at, updated_at
		FROM posts
		WHERE id = $1`

	post := &models.Post{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Slug,
		&post.AuthorID,
		&post.CategoryID,
		&post.Status,
		pq.Array(&post.Tags),
		&post.ViewCount,
		&post.PublishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("post not found")
	}

	if err != nil {
		return nil, fmt.Errorf("error getting post: %w", err)
	}

	return post, nil
}

func (r *PostgresPostRepository) GetBySlug(ctx context.Context, slug string) (*models.Post, error) {
	query := `
		SELECT id, title, content, slug, author_id, category_id, status, tags, view_count, published_at, created_at, updated_at
		FROM posts
		WHERE slug = $1`

	post := &models.Post{}
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.Slug,
		&post.AuthorID,
		&post.CategoryID,
		&post.Status,
		pq.Array(&post.Tags),
		&post.ViewCount,
		&post.PublishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("post not found")
	}

	if err != nil {
		return nil, fmt.Errorf("error getting post: %w", err)
	}

	return post, nil
}

func (r *PostgresPostRepository) Update(ctx context.Context, post *models.Post) error {
	query := `
		UPDATE posts
		SET title = $1, content = $2, slug = $3, category_id = $4, status = $5, tags = $6, updated_at = $7
		WHERE id = $8`

	post.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.Slug,
		post.CategoryID,
		post.Status,
		pq.Array(post.Tags),
		post.UpdatedAt,
		post.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

func (r *PostgresPostRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM posts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

func (r *PostgresPostRepository) List(ctx context.Context, filter *models.PostFilter) ([]*models.Post, error) {
	var conditions []string
	var args []interface{}
	argPosition := 1

	query := `
		SELECT id, title, content, slug, author_id, category_id, status, tags, view_count, published_at, created_at, updated_at
		FROM posts`

	if filter.AuthorID != 0 {
		conditions = append(conditions, fmt.Sprintf("author_id = $%d", argPosition))
		args = append(args, filter.AuthorID)
		argPosition++
	}

	if filter.CategoryID != 0 {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argPosition))
		args = append(args, filter.CategoryID)
		argPosition++
	}

	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argPosition))
		args = append(args, filter.Status)
		argPosition++
	}

	if len(filter.Tags) > 0 {
		conditions = append(conditions, fmt.Sprintf("tags && $%d", argPosition))
		args = append(args, pq.Array(filter.Tags))
		argPosition++
	}

	if filter.SearchQuery != "" {
		conditions = append(conditions, fmt.Sprintf("(title ILIKE $%d OR content ILIKE $%d)", argPosition, argPosition))
		args = append(args, "%"+filter.SearchQuery+"%")
		argPosition++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if filter.PageSize > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argPosition, argPosition+1)
		args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listing posts: %w", err)
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.Slug,
			&post.AuthorID,
			&post.CategoryID,
			&post.Status,
			pq.Array(&post.Tags),
			&post.ViewCount,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning post: %w", err)
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating posts: %w", err)
	}

	return posts, nil
}

func (r *PostgresPostRepository) IncrementViewCount(ctx context.Context, id int64) error {
	query := `
		UPDATE posts
		SET view_count = view_count + 1
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error incrementing view count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}
