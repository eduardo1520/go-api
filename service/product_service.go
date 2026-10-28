package service

import (
	"go-api/model"
	"go-api/usecase"
)

type ProductService interface {
	GetProducts() ([]model.Product, error)
	CreateProduct(product model.Product) (model.Product, error)
	GetProductById(id int) (*model.Product, error)
	UpdateProduct(product model.Product) error
	DeleteProduct(id int) error
}

type productService struct {
	productUseCase usecase.ProductUseCase
}

func NewProductService(productUseCase usecase.ProductUseCase) ProductService {
	return &productService{productUseCase: productUseCase}
}

func (ps *productService) GetProducts() ([]model.Product, error) {
	return ps.productUseCase.GetProducts()
}

func (ps *productService) CreateProduct(product model.Product) (model.Product, error) {
	return ps.productUseCase.CreateProduct(product)
}

func (ps *productService) GetProductById(id int) (*model.Product, error) {
	return ps.productUseCase.GetProductById(id)
}

func (ps *productService) UpdateProduct(product model.Product) error {
	return ps.productUseCase.UpdateProduct(product)
}

func (ps *productService) DeleteProduct(id int) error {
	return ps.productUseCase.DeleteProduct(id)
}
