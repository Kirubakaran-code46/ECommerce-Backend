package services

import (
	"context"
	"log"
	"strings"
	"time"

	"checkout-backend/internal/errors"
	"checkout-backend/internal/models"
	"checkout-backend/internal/repositories"

	"github.com/google/uuid"
)

type CheckoutService interface {
	CalculateCheckout(ctx context.Context, req models.CheckoutRequest) (*models.CheckoutResponse, *errors.AppError)
}

type checkoutService struct {
	productRepo repositories.ProductRepository
	couponRepo  repositories.CouponRepository
}

func NewCheckoutService(productRepo repositories.ProductRepository, couponRepo repositories.CouponRepository) CheckoutService {
	return &checkoutService{
		productRepo: productRepo,
		couponRepo:  couponRepo,
	}
}

func (s *checkoutService) CalculateCheckout(ctx context.Context, req models.CheckoutRequest) (*models.CheckoutResponse, *errors.AppError) {
	// 1. Validate cart has items
	if len(req.Items) == 0 {
		return nil, errors.ErrEmptyCart()
	}

	// 2. Validate quantities and deduplicate items
	aggregatedQuantities := make(map[uuid.UUID]int)
	orderedIDs := make([]uuid.UUID, 0, len(req.Items))

	for _, item := range req.Items {
		if item.ProductID == uuid.Nil {
			return nil, errors.ErrInvalidRequest("Product ID cannot be nil or empty")
		}
		if item.Quantity <= 0 {
			return nil, errors.ErrInvalidQuantity(item.Quantity)
		}

		if _, exists := aggregatedQuantities[item.ProductID]; !exists {
			orderedIDs = append(orderedIDs, item.ProductID)
		}
		aggregatedQuantities[item.ProductID] += item.Quantity
	}

	// 3. Batch fetch products from database
	productsMap, err := s.productRepo.FindByIDs(ctx, orderedIDs)
	if err != nil {
		log.Printf("[CheckoutService] Database error fetching products: %v", err)
		return nil, errors.ErrInternal(err)
	}

	// 4. Verify product existence and stock inventory
	var subtotal int64 = 0
	currency := "INR"
	itemsResponse := make([]models.CheckoutItemResponse, 0, len(orderedIDs))

	for _, id := range orderedIDs {
		product, exists := productsMap[id]
		if !exists {
			return nil, errors.ErrProductNotFound(id.String())
		}

		qty := aggregatedQuantities[id]
		if qty > product.StockQuantity {
			return nil, errors.ErrInsufficientStock(product.Name, qty, product.StockQuantity)
		}

		// Safe 64-bit integer calculation in paise
		lineTotal := product.Price * int64(qty)
		subtotal += lineTotal
		currency = product.Currency

		itemsResponse = append(itemsResponse, models.CheckoutItemResponse{
			ProductID: product.ID,
			Name:      product.Name,
			Quantity:  qty,
			UnitPrice: product.Price,
			LineTotal: lineTotal,
		})
	}

	// 5. Coupon validation and discount computation
	var discount int64 = 0
	var couponSummary *models.CouponSummaryResponse

	if req.CouponCode != nil && strings.TrimSpace(*req.CouponCode) != "" {
		code := strings.TrimSpace(*req.CouponCode)
		coupon, err := s.couponRepo.FindByCode(ctx, code)
		if err != nil {
			log.Printf("[CheckoutService] Database error fetching coupon '%s': %v", code, err)
			return nil, errors.ErrInternal(err)
		}

		if coupon == nil {
			return nil, errors.ErrInvalidCoupon(code)
		}

		if !coupon.Active {
			return nil, errors.ErrInactiveCoupon(coupon.Code)
		}

		// Check expiration against current UTC time
		if coupon.ExpiresAt.Before(time.Now().UTC()) {
			return nil, errors.ErrExpiredCoupon(coupon.Code)
		}

		// Check minimum cart value requirement (in paise)
		if subtotal < coupon.MinimumCartValue {
			return nil, errors.ErrMinCartValueNotMet(coupon.Code, coupon.MinimumCartValue, subtotal)
		}

		// Compute discount based on coupon type
		switch coupon.DiscountType {
		case "percentage":
			// Safe integer percentage calculation in paise
			discount = (subtotal * coupon.DiscountValue) / 100
			if coupon.MaximumDiscount != nil && discount > *coupon.MaximumDiscount {
				discount = *coupon.MaximumDiscount
			}
		case "fixed_amount":
			discount = coupon.DiscountValue
		default:
			log.Printf("[CheckoutService] Unsupported discount type: %s", coupon.DiscountType)
			return nil, errors.ErrInvalidCoupon(coupon.Code)
		}

		// Discount must never exceed the subtotal
		if discount > subtotal {
			discount = subtotal
		}

		couponSummary = &models.CouponSummaryResponse{
			Code:           coupon.Code,
			DiscountType:   coupon.DiscountType,
			DiscountValue:  coupon.DiscountValue,
			DiscountAmount: discount,
		}
	}

	// 6. Final calculation
	total := max(subtotal-discount, 0)

	return &models.CheckoutResponse{
		Items:    itemsResponse,
		Subtotal: subtotal,
		Discount: discount,
		Total:    total,
		Currency: currency,
		Coupon:   couponSummary,
	}, nil
}
