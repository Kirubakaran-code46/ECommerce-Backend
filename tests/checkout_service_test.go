package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"checkout-backend/internal/errors"
	"checkout-backend/internal/models"
	"checkout-backend/internal/services"

	"github.com/google/uuid"
)

// Mock Product Repository
type mockProductRepo struct {
	products map[uuid.UUID]models.Product
}

func (m *mockProductRepo) FindAll(ctx context.Context) ([]models.Product, error) {
	result := make([]models.Product, 0, len(m.products))
	for _, p := range m.products {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockProductRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	p, exists := m.products[id]
	if !exists {
		return nil, nil
	}
	return &p, nil
}

func (m *mockProductRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]models.Product, error) {
	result := make(map[uuid.UUID]models.Product)
	for _, id := range ids {
		if p, exists := m.products[id]; exists {
			result[id] = p
		}
	}
	return result, nil
}

// Mock Coupon Repository
type mockCouponRepo struct {
	coupons map[string]models.Coupon
}

func (m *mockCouponRepo) FindByCode(ctx context.Context, code string) (*models.Coupon, error) {
	c, exists := m.coupons[strings.ToLower(strings.TrimSpace(code))]
	if !exists {
		return nil, nil
	}
	return &c, nil
}

func setupTestService() (services.CheckoutService, uuid.UUID, uuid.UUID) {
	laptopID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	mouseID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	maxCap := int64(300000) // ₹3,000 cap

	productRepo := &mockProductRepo{
		products: map[uuid.UUID]models.Product{
			laptopID: {
				ID:            laptopID,
				Name:          "Laptop Pro",
				Price:         5000000, // ₹50,000 (5,000,000 paise)
				Currency:      "INR",
				StockQuantity: 10,
			},
			mouseID: {
				ID:            mouseID,
				Name:          "Wireless Mouse",
				Price:         200000, // ₹2,000 (200,000 paise)
				Currency:      "INR",
				StockQuantity: 5,
			},
		},
	}

	couponRepo := &mockCouponRepo{
		coupons: map[string]models.Coupon{
			"save10": {
				ID:               uuid.New(),
				Code:             "SAVE10",
				DiscountType:     "percentage",
				DiscountValue:    10,
				MinimumCartValue: 100000, // ₹1,000
				MaximumDiscount:  nil,
				ExpiresAt:        time.Now().Add(24 * time.Hour),
				Active:           true,
			},
			"save20cap": {
				ID:               uuid.New(),
				Code:             "SAVE20CAP",
				DiscountType:     "percentage",
				DiscountValue:    20,
				MinimumCartValue: 100000,
				MaximumDiscount:  &maxCap, // Capped at ₹3,000
				ExpiresAt:        time.Now().Add(24 * time.Hour),
				Active:           true,
			},
			"mincart50k": {
				ID:               uuid.New(),
				Code:             "MINCART50K",
				DiscountType:     "percentage",
				DiscountValue:    15,
				MinimumCartValue: 5000000, // ₹50,000 minimum
				ExpiresAt:        time.Now().Add(24 * time.Hour),
				Active:           true,
			},
			"flat500": {
				ID:               uuid.New(),
				Code:             "FLAT500",
				DiscountType:     "fixed_amount",
				DiscountValue:    50000, // ₹500
				MinimumCartValue: 100000,
				ExpiresAt:        time.Now().Add(24 * time.Hour),
				Active:           true,
			},
			"expired10": {
				ID:               uuid.New(),
				Code:             "EXPIRED10",
				DiscountType:     "percentage",
				DiscountValue:    10,
				MinimumCartValue: 0,
				ExpiresAt:        time.Now().Add(-24 * time.Hour), // Expired yesterday
				Active:           true,
			},
			"inactive15": {
				ID:               uuid.New(),
				Code:             "INACTIVE15",
				DiscountType:     "percentage",
				DiscountValue:    15,
				MinimumCartValue: 0,
				ExpiresAt:        time.Now().Add(24 * time.Hour),
				Active:           false, // Disabled
			},
		},
	}

	return services.NewCheckoutService(productRepo, couponRepo), laptopID, mouseID
}

func TestSuccessfulCheckoutWithoutCoupon(t *testing.T) {
	service, laptopID, mouseID := setupTestService()

	req := models.CheckoutRequest{
		Items: []models.CartItemInput{
			{ProductID: laptopID, Quantity: 1}, // 50,000
			{ProductID: mouseID, Quantity: 2},  // 4,000
		},
	}

	resp, appErr := service.CalculateCheckout(context.Background(), req)
	if appErr != nil {
		t.Fatalf("Expected success, got error: %v", appErr)
	}

	expectedSubtotal := int64(5400000) // ₹54,000 in paise
	if resp.Subtotal != expectedSubtotal {
		t.Errorf("Expected subtotal %d, got %d", expectedSubtotal, resp.Subtotal)
	}
	if resp.Discount != 0 {
		t.Errorf("Expected discount 0, got %d", resp.Discount)
	}
	if resp.Total != expectedSubtotal {
		t.Errorf("Expected total %d, got %d", expectedSubtotal, resp.Total)
	}
	if resp.Coupon != nil {
		t.Errorf("Expected nil coupon summary, got %+v", resp.Coupon)
	}
}

func TestEmptyCart(t *testing.T) {
	service, _, _ := setupTestService()

	req := models.CheckoutRequest{
		Items: []models.CartItemInput{},
	}

	_, appErr := service.CalculateCheckout(context.Background(), req)
	if appErr == nil {
		t.Fatal("Expected error for empty cart, got nil")
	}
	if appErr.Code != errors.CodeEmptyCart {
		t.Errorf("Expected code %s, got %s", errors.CodeEmptyCart, appErr.Code)
	}
}

func TestInvalidQuantity(t *testing.T) {
	service, laptopID, _ := setupTestService()

	testCases := []int{0, -1, -10}
	for _, qty := range testCases {
		req := models.CheckoutRequest{
			Items: []models.CartItemInput{
				{ProductID: laptopID, Quantity: qty},
			},
		}

		_, appErr := service.CalculateCheckout(context.Background(), req)
		if appErr == nil {
			t.Fatalf("Expected error for quantity %d, got nil", qty)
		}
		if appErr.Code != errors.CodeInvalidQuantity {
			t.Errorf("Expected code %s, got %s", errors.CodeInvalidQuantity, appErr.Code)
		}
	}
}

func TestProductNotFound(t *testing.T) {
	service, _, _ := setupTestService()
	randomID := uuid.New()

	req := models.CheckoutRequest{
		Items: []models.CartItemInput{
			{ProductID: randomID, Quantity: 1},
		},
	}

	_, appErr := service.CalculateCheckout(context.Background(), req)
	if appErr == nil {
		t.Fatal("Expected error for non-existent product, got nil")
	}
	if appErr.Code != errors.CodeProductNotFound {
		t.Errorf("Expected code %s, got %s", errors.CodeProductNotFound, appErr.Code)
	}
}

func TestInsufficientStock(t *testing.T) {
	service, laptopID, _ := setupTestService() // Stock is 10

	req := models.CheckoutRequest{
		Items: []models.CartItemInput{
			{ProductID: laptopID, Quantity: 11},
		},
	}

	_, appErr := service.CalculateCheckout(context.Background(), req)
	if appErr == nil {
		t.Fatal("Expected error for stock excess, got nil")
	}
	if appErr.Code != errors.CodeInsufficientStock {
		t.Errorf("Expected code %s, got %s", errors.CodeInsufficientStock, appErr.Code)
	}
}

func TestValidPercentageCoupon(t *testing.T) {
	service, laptopID, _ := setupTestService()
	coupon := "SAVE10"

	req := models.CheckoutRequest{
		Items: []models.CartItemInput{
			{ProductID: laptopID, Quantity: 1}, // ₹50,000 = 5,000,000 paise
		},
		CouponCode: &coupon,
	}

	resp, appErr := service.CalculateCheckout(context.Background(), req)
	if appErr != nil {
		t.Fatalf("Expected success, got error: %v", appErr)
	}

	expectedSubtotal := int64(5000000)
	expectedDiscount := int64(500000) // 10% of 50,000 = ₹5,000
	expectedTotal := int64(4500000)    // ₹45,000

	if resp.Subtotal != expectedSubtotal {
		t.Errorf("Expected subtotal %d, got %d", expectedSubtotal, resp.Subtotal)
	}
	if resp.Discount != expectedDiscount {
		t.Errorf("Expected discount %d, got %d", expectedDiscount, resp.Discount)
	}
	if resp.Total != expectedTotal {
		t.Errorf("Expected total %d, got %d", expectedTotal, resp.Total)
	}
	if resp.Coupon == nil || resp.Coupon.DiscountAmount != expectedDiscount {
		t.Errorf("Expected coupon summary with discount %d", expectedDiscount)
	}
}

func TestCouponCaseInsensitivity(t *testing.T) {
	service, laptopID, _ := setupTestService()
	variations := []string{"save10", "SAVE10", "Save10", "  save10  "}

	for _, code := range variations {
		req := models.CheckoutRequest{
			Items: []models.CartItemInput{
				{ProductID: laptopID, Quantity: 1},
			},
			CouponCode: &code,
		}

		resp, appErr := service.CalculateCheckout(context.Background(), req)
		if appErr != nil {
			t.Fatalf("Code '%s' failed with error: %v", code, appErr)
		}
		if resp.Discount != 500000 {
			t.Errorf("Code '%s' expected discount 500000, got %d", code, resp.Discount)
		}
	}
}

func TestMaximumDiscountCap(t *testing.T) {
	service, laptopID, _ := setupTestService()
	coupon := "SAVE20CAP" // 20% on ₹50,000 is ₹10,000, but cap is ₹3,000 (300,000 paise)

	req := models.CheckoutRequest{
		Items: []models.CartItemInput{
			{ProductID: laptopID, Quantity: 1},
		},
		CouponCode: &coupon,
	}

	resp, appErr := service.CalculateCheckout(context.Background(), req)
	if appErr != nil {
		t.Fatalf("Expected success, got error: %v", appErr)
	}

	expectedCap := int64(300000) // ₹3,000
	if resp.Discount != expectedCap {
		t.Errorf("Expected discount capped at %d, got %d", expectedCap, resp.Discount)
	}
	expectedTotal := int64(5000000) - expectedCap
	if resp.Total != expectedTotal {
		t.Errorf("Expected total %d, got %d", expectedTotal, resp.Total)
	}
}

func TestMinimumCartValueNotSatisfied(t *testing.T) {
	service, _, mouseID := setupTestService()
	coupon := "MINCART50K" // Requires ₹50,000

	req := models.CheckoutRequest{
		Items: []models.CartItemInput{
			{ProductID: mouseID, Quantity: 1}, // Only ₹2,000
		},
		CouponCode: &coupon,
	}

	_, appErr := service.CalculateCheckout(context.Background(), req)
	if appErr == nil {
		t.Fatal("Expected error for minimum cart not met, got nil")
	}
	if appErr.Code != errors.CodeMinCartValueNotMet {
		t.Errorf("Expected code %s, got %s", errors.CodeMinCartValueNotMet, appErr.Code)
	}
}

func TestExpiredCoupon(t *testing.T) {
	service, laptopID, _ := setupTestService()
	coupon := "EXPIRED10"

	req := models.CheckoutRequest{
		Items: []models.CartItemInput{
			{ProductID: laptopID, Quantity: 1},
		},
		CouponCode: &coupon,
	}

	_, appErr := service.CalculateCheckout(context.Background(), req)
	if appErr == nil {
		t.Fatal("Expected error for expired coupon, got nil")
	}
	if appErr.Code != errors.CodeExpiredCoupon {
		t.Errorf("Expected code %s, got %s", errors.CodeExpiredCoupon, appErr.Code)
	}
}

func TestInactiveCoupon(t *testing.T) {
	service, laptopID, _ := setupTestService()
	coupon := "INACTIVE15"

	req := models.CheckoutRequest{
		Items: []models.CartItemInput{
			{ProductID: laptopID, Quantity: 1},
		},
		CouponCode: &coupon,
	}

	_, appErr := service.CalculateCheckout(context.Background(), req)
	if appErr == nil {
		t.Fatal("Expected error for inactive coupon, got nil")
	}
	if appErr.Code != errors.CodeInactiveCoupon {
		t.Errorf("Expected code %s, got %s", errors.CodeInactiveCoupon, appErr.Code)
	}
}

func TestNonExistentCoupon(t *testing.T) {
	service, laptopID, _ := setupTestService()
	coupon := "DOESNOTEXIST"

	req := models.CheckoutRequest{
		Items: []models.CartItemInput{
			{ProductID: laptopID, Quantity: 1},
		},
		CouponCode: &coupon,
	}

	_, appErr := service.CalculateCheckout(context.Background(), req)
	if appErr == nil {
		t.Fatal("Expected error for invalid coupon, got nil")
	}
	if appErr.Code != errors.CodeInvalidCoupon {
		t.Errorf("Expected code %s, got %s", errors.CodeInvalidCoupon, appErr.Code)
	}
}

func TestFixedAmountCoupon(t *testing.T) {
	service, _, mouseID := setupTestService()
	coupon := "FLAT500" // ₹500 off

	req := models.CheckoutRequest{
		Items: []models.CartItemInput{
			{ProductID: mouseID, Quantity: 1}, // ₹2,000 (200,000 paise)
		},
		CouponCode: &coupon,
	}

	resp, appErr := service.CalculateCheckout(context.Background(), req)
	if appErr != nil {
		t.Fatalf("Expected success, got error: %v", appErr)
	}

	expectedDiscount := int64(50000)  // ₹500
	expectedTotal := int64(150000)     // ₹1,500
	if resp.Discount != expectedDiscount {
		t.Errorf("Expected discount %d, got %d", expectedDiscount, resp.Discount)
	}
	if resp.Total != expectedTotal {
		t.Errorf("Expected total %d, got %d", expectedTotal, resp.Total)
	}
}
