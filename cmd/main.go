package main

import (
	"fmt"
	"go-api/controller"
	"go-api/db"
	"go-api/repository"
	"go-api/service"
	"go-api/usecase"

	"github.com/gin-gonic/gin"
)

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

	server.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	server.GET("/products", ProductController.GetProducts)
	server.GET("/products/:productId", ProductController.GetProductById)
	server.POST("/products", ProductController.CreateProduct)
	server.PUT("/products/:productId", ProductController.UpdateProduct)
	server.PATCH("/products/:productId", ProductController.UpdateProduct)
	server.DELETE("/products/:productId", ProductController.DeleteProduct)

	return server, nil
}
