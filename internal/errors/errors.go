package errors

import (
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	CodeInvalidRequest       ErrorCode = "INVALID_REQUEST"
	CodeProductNotFound      ErrorCode = "PRODUCT_NOT_FOUND"
	CodeInsufficientStock    ErrorCode = "INSUFFICIENT_STOCK"
	CodeEmptyCart            ErrorCode = "EMPTY_CART"
	CodeInvalidQuantity      ErrorCode = "INVALID_QUANTITY"
	CodeInvalidCoupon        ErrorCode = "INVALID_COUPON"
	CodeExpiredCoupon        ErrorCode = "EXPIRED_COUPON"
	CodeInactiveCoupon       ErrorCode = "INACTIVE_COUPON"
	CodeMinCartValueNotMet   ErrorCode = "MIN_CART_VALUE_NOT_MET"
	CodeInternalServerError  ErrorCode = "INTERNAL_SERVER_ERROR"
)

type AppError struct {
	Code       ErrorCode   `json:"code"`
	Message    string      `json:"message"`
	HTTPStatus int         `json:"-"`
	Details    interface{} `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func New(code ErrorCode, message string, httpStatus int, details interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Details:    details,
	}
}

// Common error constructors
func ErrInvalidRequest(message string) *AppError {
	return New(CodeInvalidRequest, message, http.StatusBadRequest, nil)
}

func ErrProductNotFound(id string) *AppError {
	return New(CodeProductNotFound, fmt.Sprintf("Product with ID '%s' was not found", id), http.StatusNotFound, map[string]string{"product_id": id})
}

func ErrInsufficientStock(productName string, requested, available int) *AppError {
	return New(
		CodeInsufficientStock,
		fmt.Sprintf("Insufficient stock for '%s'. Requested: %d, Available: %d", productName, requested, available),
		http.StatusConflict,
		map[string]interface{}{
			"product_name": productName,
			"requested":    requested,
			"available":    available,
		},
	)
}

func ErrEmptyCart() *AppError {
	return New(CodeEmptyCart, "Cart cannot be empty for checkout", http.StatusUnprocessableEntity, nil)
}

func ErrInvalidQuantity(quantity int) *AppError {
	return New(CodeInvalidQuantity, fmt.Sprintf("Quantity must be greater than 0, got %d", quantity), http.StatusBadRequest, map[string]int{"quantity": quantity})
}

func ErrInvalidCoupon(code string) *AppError {
	return New(CodeInvalidCoupon, fmt.Sprintf("Coupon code '%s' is invalid", code), http.StatusUnprocessableEntity, map[string]string{"coupon_code": code})
}

func ErrExpiredCoupon(code string) *AppError {
	return New(CodeExpiredCoupon, fmt.Sprintf("Coupon code '%s' has expired", code), http.StatusUnprocessableEntity, map[string]string{"coupon_code": code})
}

func ErrInactiveCoupon(code string) *AppError {
	return New(CodeInactiveCoupon, fmt.Sprintf("Coupon code '%s' is not currently active", code), http.StatusUnprocessableEntity, map[string]string{"coupon_code": code})
}

func ErrMinCartValueNotMet(code string, minVal, currentVal int64) *AppError {
	return New(
		CodeMinCartValueNotMet,
		fmt.Sprintf("Coupon '%s' requires a minimum cart value of %d paise, but current cart is %d paise", code, minVal, currentVal),
		http.StatusUnprocessableEntity,
		map[string]interface{}{
			"coupon_code":        code,
			"minimum_cart_value": minVal,
			"current_cart_value": currentVal,
		},
	)
}

func ErrInternal(err error) *AppError {
	return New(CodeInternalServerError, "An unexpected error occurred. Please try again later.", http.StatusInternalServerError, nil)
}
