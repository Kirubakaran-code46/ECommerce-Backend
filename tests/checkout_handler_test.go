package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"checkout-backend/internal/handlers"
	"checkout-backend/internal/models"
	"checkout-backend/internal/routes"
	"checkout-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setupTestRouter() (*gin.Engine, uuid.UUID, uuid.UUID) {
	gin.SetMode(gin.TestMode)
	service, laptopID, mouseID := setupTestService()

	// Wrap mock product repo in product service
	laptopProduct := models.Product{
		ID:            laptopID,
		Name:          "Laptop Pro",
		Price:         5000000,
		Currency:      "INR",
		StockQuantity: 10,
	}
	mouseProduct := models.Product{
		ID:            mouseID,
		Name:          "Wireless Mouse",
		Price:         200000,
		Currency:      "INR",
		StockQuantity: 5,
	}
	prodRepo := &mockProductRepo{
		products: map[uuid.UUID]models.Product{
			laptopID: laptopProduct,
			mouseID:  mouseProduct,
		},
	}

	prodService := services.NewProductService(prodRepo)
	prodHandler := handlers.NewProductHandler(prodService)
	checkoutHandler := handlers.NewCheckoutHandler(service)

	router := routes.SetupRouter("http://localhost:5173", prodHandler, checkoutHandler)
	return router, laptopID, mouseID
}

func TestHTTPGetProducts(t *testing.T) {
	router, _, _ := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	data, exists := response["data"]
	if !exists {
		t.Fatal("Expected 'data' field in response")
	}

	items := data.([]interface{})
	if len(items) != 2 {
		t.Errorf("Expected 2 products, got %d", len(items))
	}
}

func TestHTTPCheckoutSuccess(t *testing.T) {
	router, laptopID, _ := setupTestRouter()

	coupon := "SAVE10"
	payload := models.CheckoutRequest{
		Items: []models.CartItemInput{
			{ProductID: laptopID, Quantity: 1},
		},
		CouponCode: &coupon,
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/checkout", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response struct {
		Data models.CheckoutResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Data.Subtotal != 5000000 {
		t.Errorf("Expected subtotal 5000000, got %d", response.Data.Subtotal)
	}
	if response.Data.Discount != 500000 {
		t.Errorf("Expected discount 500000, got %d", response.Data.Discount)
	}
	if response.Data.Total != 4500000 {
		t.Errorf("Expected total 4500000, got %d", response.Data.Total)
	}
}

func TestHTTPCheckoutInvalidCoupon(t *testing.T) {
	router, laptopID, _ := setupTestRouter()

	coupon := "NONEXISTENT"
	payload := models.CheckoutRequest{
		Items: []models.CartItemInput{
			{ProductID: laptopID, Quantity: 1},
		},
		CouponCode: &coupon,
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/checkout", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("Expected status 422, got %d. Body: %s", w.Code, w.Body.String())
	}

	var errResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	errObj := errResp["error"].(map[string]interface{})
	if errObj["code"] != "INVALID_COUPON" {
		t.Errorf("Expected error code INVALID_COUPON, got %v", errObj["code"])
	}
}

func TestHTTPCheckoutStockExcess(t *testing.T) {
	router, laptopID, _ := setupTestRouter() // Stock is 10

	payload := models.CheckoutRequest{
		Items: []models.CartItemInput{
			{ProductID: laptopID, Quantity: 20},
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/checkout", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("Expected status 409 Conflict, got %d. Body: %s", w.Code, w.Body.String())
	}
}
