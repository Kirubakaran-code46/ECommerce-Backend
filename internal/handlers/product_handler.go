package handlers

import (
	"net/http"

	"checkout-backend/internal/errors"
	"checkout-backend/internal/models"
	"checkout-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProductHandler struct {
	service services.ProductService
}

func NewProductHandler(service services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	products, appErr := h.service.GetAllProducts(c.Request.Context())
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, models.APIErrorEnvelope{Error: appErr})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Data: products})
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		appErr := errors.ErrInvalidRequest("Invalid product ID format. Must be a valid UUID.")
		c.JSON(appErr.HTTPStatus, models.APIErrorEnvelope{Error: appErr})
		return
	}

	product, appErr := h.service.GetProductByID(c.Request.Context(), id)
	if appErr != nil {
		c.JSON(appErr.HTTPStatus, models.APIErrorEnvelope{Error: appErr})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Data: product})
}
