package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Price         int64     `json:"price"` // in paise (e.g., 5000000 = ₹50,000.00)
	Currency      string    `json:"currency"`
	StockQuantity int       `json:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Coupon struct {
	ID               uuid.UUID `json:"id"`
	Code             string    `json:"code"`
	DiscountType     string    `json:"discount_type"` // "percentage" or "fixed_amount"
	DiscountValue    int64     `json:"discount_value"`
	MinimumCartValue int64     `json:"minimum_cart_value"` // in paise
	MaximumDiscount  *int64    `json:"maximum_discount"`   // in paise, nil if uncapped
	ExpiresAt        time.Time `json:"expires_at"`
	Active           bool      `json:"active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Checkout request DTOs
type CartItemInput struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required"`
}

type CheckoutRequest struct {
	Items      []CartItemInput `json:"items" binding:"required"`
	CouponCode *string         `json:"coupon_code,omitempty"`
}

// Checkout response DTOs
type CheckoutItemResponse struct {
	ProductID uuid.UUID `json:"product_id"`
	Name      string    `json:"name"`
	Quantity  int       `json:"quantity"`
	UnitPrice int64     `json:"unit_price"` // in paise
	LineTotal int64     `json:"line_total"` // in paise
}

type CouponSummaryResponse struct {
	Code           string `json:"code"`
	DiscountType   string `json:"discount_type"`
	DiscountValue  int64  `json:"discount_value"`
	DiscountAmount int64  `json:"discount_amount"` // in paise
}

type CheckoutResponse struct {
	Items    []CheckoutItemResponse `json:"items"`
	Subtotal int64                  `json:"subtotal"` // in paise
	Discount int64                  `json:"discount"` // in paise
	Total    int64                  `json:"total"`    // in paise
	Currency string                 `json:"currency"`
	Coupon   *CouponSummaryResponse `json:"coupon,omitempty"`
}

// Envelope responses
type APIResponse struct {
	Data interface{} `json:"data"`
}

type APIErrorEnvelope struct {
	Error interface{} `json:"error"`
}
