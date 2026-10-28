package controller

import (
	"errors"
	"go-api/model"
	"go-api/repository"
	"go-api/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	productUseCase usecase.ProductUseCase
}

func NewProductController(useCase usecase.ProductUseCase) *ProductController {
	return &ProductController{
		productUseCase: useCase,
	}
}

func (p *ProductController) GetProducts(ctx *gin.Context) {
	products, err := p.productUseCase.GetProducts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
	}
	ctx.JSON(http.StatusOK, products)
}

func (p *ProductController) GetProductById(ctx *gin.Context) {
	id := ctx.Param("productId")
	if id == "" {
		response := model.Response{Message: "Id do produto não pode ser nulo."}
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	productId, err := strconv.Atoi(id)
	if err != nil {
		response := model.Response{Message: "Id do produto inválido."}
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	product, err := p.productUseCase.GetProductById(productId)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			response := model.Response{Message: "Produto não encontrado na base de dados."}
			ctx.JSON(http.StatusNotFound, response)
			return
		}

		response := model.Response{Message: "Erro ao buscar produto."}
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}

	ctx.JSON(http.StatusOK, product)

}

func (p *ProductController) CreateProduct(ctx *gin.Context) {
	var product model.Product
	err := ctx.BindJSON(&product)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	insertedProduct, err := p.productUseCase.CreateProduct(product)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, insertedProduct)
}
