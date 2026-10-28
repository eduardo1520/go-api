package controller

import (
	"errors"
	"go-api/dto"
	"go-api/model"
	"go-api/response"
	"go-api/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	productService service.ProductService
}

func NewProductController(productService service.ProductService) *ProductController {
	return &ProductController{
		productService: productService,
	}
}

func (p *ProductController) GetProducts(ctx *gin.Context) {
	products, err := p.productService.GetProducts()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Message{Message: "Erro ao listar produtos."})
		return
	}
	ctx.JSON(http.StatusOK, productResponses(products))
}

func (p *ProductController) GetProductById(ctx *gin.Context) {
	id := ctx.Param("productId")
	if id == "" {
		response := response.Message{Message: "Id do produto não pode ser nulo."}
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	productId, err := strconv.Atoi(id)
	if err != nil {
		response := response.Message{Message: "Id do produto inválido."}
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	product, err := p.productService.GetProductById(productId)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			response := response.Message{Message: "Produto não encontrado na base de dados."}
			ctx.JSON(http.StatusNotFound, response)
			return
		}

		response := response.Message{Message: "Erro ao buscar produto."}
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}

	ctx.JSON(http.StatusOK, productResponse(*product))

}

func (p *ProductController) CreateProduct(ctx *gin.Context) {
	var request dto.ProductRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Message{Message: "Dados do produto inválidos."})
		return
	}

	insertedProduct, err := p.productService.CreateProduct(productFromRequest(request))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Message{Message: "Erro ao criar produto."})
		return
	}

	ctx.JSON(http.StatusCreated, productResponse(insertedProduct))
}

func (p *ProductController) UpdateProduct(ctx *gin.Context) {
	productID, valid := productIDFromContext(ctx)
	if !valid {
		return
	}

	var request dto.ProductRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Message{Message: "Dados do produto inválidos."})
		return
	}
	product := productFromRequest(request)
	product.ID = productID

	if err := p.productService.UpdateProduct(product); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			ctx.JSON(http.StatusNotFound, response.Message{Message: "Produto não encontrado na base de dados."})
			return
		}
		ctx.JSON(http.StatusInternalServerError, response.Message{Message: "Erro ao atualizar produto."})
		return
	}

	ctx.JSON(http.StatusOK, productResponse(product))
}

func (p *ProductController) DeleteProduct(ctx *gin.Context) {
	productID, valid := productIDFromContext(ctx)
	if !valid {
		return
	}

	if err := p.productService.DeleteProduct(productID); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			ctx.JSON(http.StatusNotFound, response.Message{Message: "Produto não encontrado na base de dados."})
			return
		}
		ctx.JSON(http.StatusInternalServerError, response.Message{Message: "Erro ao excluir produto."})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func productIDFromContext(ctx *gin.Context) (int, bool) {
	id := ctx.Param("productId")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, response.Message{Message: "Id do produto não pode ser nulo."})
		return 0, false
	}

	productID, err := strconv.Atoi(id)
	if err != nil || productID <= 0 {
		ctx.JSON(http.StatusBadRequest, response.Message{Message: "Id do produto inválido."})
		return 0, false
	}
	return productID, true
}

func productFromRequest(request dto.ProductRequest) model.Product {
	return model.Product{Name: request.Name, Price: request.Price}
}

func productResponse(product model.Product) response.Product {
	return response.Product{ID: product.ID, Name: product.Name, Price: product.Price}
}

func productResponses(products []model.Product) []response.Product {
	responses := make([]response.Product, 0, len(products))
	for _, product := range products {
		responses = append(responses, productResponse(product))
	}
	return responses
}
