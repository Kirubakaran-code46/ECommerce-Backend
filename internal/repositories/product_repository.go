package repositories

import (
	"context"
	"database/sql"
	"errors"

	"checkout-backend/internal/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ProductRepository interface {
	FindAll(ctx context.Context) ([]models.Product, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]models.Product, error)
}

type sqlProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
	return &sqlProductRepository{db: db}
}

func (r *sqlProductRepository) FindAll(ctx context.Context) ([]models.Product, error) {
	query := `
		SELECT id, name, description, price, currency, stock_quantity, created_at, updated_at
		FROM products
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Currency,
			&p.StockQuantity,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *sqlProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	query := `
		SELECT id, name, description, price, currency, stock_quantity, created_at, updated_at
		FROM products
		WHERE id = $1
	`
	var p models.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.Price,
		&p.Currency,
		&p.StockQuantity,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &p, nil
}

func (r *sqlProductRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]models.Product, error) {
	if len(ids) == 0 {
		return make(map[uuid.UUID]models.Product), nil
	}

	query := `
		SELECT id, name, description, price, currency, stock_quantity, created_at, updated_at
		FROM products
		WHERE id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]models.Product, len(ids))
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Price,
			&p.Currency,
			&p.StockQuantity,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result[p.ID] = p
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
