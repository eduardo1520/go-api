package service

import (
	"errors"
	"go-api/model"
	"go-api/repository"
	"go-api/usecase"
)

var ErrProductNotFound = errors.New("product not found")

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
	product, err := ps.productUseCase.GetProductById(id)
	if err != nil {
		return nil, translateError(err)
	}
	return product, nil
}

func (ps *productService) UpdateProduct(product model.Product) error {
	return translateError(ps.productUseCase.UpdateProduct(product))
}

func (ps *productService) DeleteProduct(id int) error {
	return translateError(ps.productUseCase.DeleteProduct(id))
}

func translateError(err error) error {
	if errors.Is(err, repository.ErrProductNotFound) {
		return ErrProductNotFound
	}
	return err
}
