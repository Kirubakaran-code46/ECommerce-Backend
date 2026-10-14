package services

import (
	"context"
	"log"

	"checkout-backend/internal/errors"
	"checkout-backend/internal/models"
	"checkout-backend/internal/repositories"

	"github.com/google/uuid"
)

type ProductService interface {
	GetAllProducts(ctx context.Context) ([]models.Product, *errors.AppError)
	GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, *errors.AppError)
}

type productService struct {
	repo repositories.ProductRepository
}

func NewProductService(repo repositories.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) GetAllProducts(ctx context.Context) ([]models.Product, *errors.AppError) {
	products, err := s.repo.FindAll(ctx)
	if err != nil {
		log.Printf("[ProductService] Error fetching products: %v", err)
		return nil, errors.ErrInternal(err)
	}

	if products == nil {
		products = []models.Product{}
	}

	return products, nil
}

func (s *productService) GetProductByID(ctx context.Context, id uuid.UUID) (*models.Product, *errors.AppError) {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		log.Printf("[ProductService] Error fetching product %s: %v", id, err)
		return nil, errors.ErrInternal(err)
	}

	if product == nil {
		return nil, errors.ErrProductNotFound(id.String())
	}

	return product, nil
}
