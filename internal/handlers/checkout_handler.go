package handlers

import (
	"net/http"

	"checkout-backend/internal/errors"
	"checkout-backend/internal/models"
	"checkout-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type CheckoutHandler struct {
	service services.CheckoutService
}

func NewCheckoutHandler(service services.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{service: service}
}

func (h *CheckoutHandler) Checkout(c *gin.Context) {
	var req models.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appErr := errors.ErrInvalidRequest("Invalid request body. Please verify that 'items' array is present and formatted properly.")
		c.JSON(appErr.HTTPStatus, models.APIErrorEnvelope{Error: appErr})
		return
	}

	result, appErr := h.service.CalculateCheckout(c.Request.Context(), req)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, models.APIErrorEnvelope{Error: appErr})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Data: result})
}
