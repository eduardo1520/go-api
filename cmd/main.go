package main

import (
	"fmt"
	"go-api/controller"
	"go-api/db"
	"go-api/repository"
	"go-api/service"
	"go-api/usecase"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "go-api/docs"
)

// @title Go API
// @version 1.0
// @description API REST de produtos desenvolvida em Go com Gin e PostgreSQL.
// @host localhost:8000
// @BasePath /

var connectDB = db.ConnectDB
var runServer = func(server *gin.Engine) error {
	return server.Run(":8000")
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}

func run() error {
	server, err := newServer()
	if err != nil {
		return err
	}
	return runServer(server)
}

func newServer() (*gin.Engine, error) {
	server := gin.Default()

	dbConnection, err := connectDB()
	if err != nil {
		return nil, err
	}

	ProductRepository := repository.NewProductRepository(dbConnection)
	ProductUseCase := usecase.NewProductUseCase(*ProductRepository)
	ProductService := service.NewProductService(*ProductUseCase)
	ProductController := controller.NewProductController(ProductService)

	server.GET("/ping", ping)
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	server.GET("/products", ProductController.GetProducts)
	server.GET("/products/:productId", ProductController.GetProductById)
	server.POST("/products", ProductController.CreateProduct)
	server.PUT("/products/:productId", ProductController.UpdateProduct)
	server.PATCH("/products/:productId", ProductController.PatchProduct)
	server.DELETE("/products/:productId", ProductController.DeleteProduct)

	return server, nil
}

// ping godoc
// @Summary Verifica o serviço
// @Tags health
// @Produce json
// @Success 200 {object} response.Message
// @Router /ping [get]
func ping(c *gin.Context) {
	c.JSON(200, gin.H{"message": "pong"})
}
