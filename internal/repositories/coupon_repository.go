package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"checkout-backend/internal/models"
)

type CouponRepository interface {
	FindByCode(ctx context.Context, code string) (*models.Coupon, error)
}

type sqlCouponRepository struct {
	db *sql.DB
}

func NewCouponRepository(db *sql.DB) CouponRepository {
	return &sqlCouponRepository{db: db}
}

func (r *sqlCouponRepository) FindByCode(ctx context.Context, code string) (*models.Coupon, error) {
	sanitizedCode := strings.TrimSpace(code)
	if sanitizedCode == "" {
		return nil, nil
	}

	query := `
		SELECT id, code, discount_type, discount_value, minimum_cart_value, maximum_discount, expires_at, active, created_at, updated_at
		FROM coupons
		WHERE LOWER(code) = LOWER($1)
	`

	var c models.Coupon
	var maxDiscount sql.NullInt64

	err := r.db.QueryRowContext(ctx, query, sanitizedCode).Scan(
		&c.ID,
		&c.Code,
		&c.DiscountType,
		&c.DiscountValue,
		&c.MinimumCartValue,
		&maxDiscount,
		&c.ExpiresAt,
		&c.Active,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if maxDiscount.Valid {
		val := maxDiscount.Int64
		c.MaximumDiscount = &val
	}

	return &c, nil
}
