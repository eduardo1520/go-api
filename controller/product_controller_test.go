package controller

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"go-api/repository"
	"go-api/usecase"
)

func setupProductRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewProductRepository(db)
	productUseCase := usecase.NewProductUseCase(*repo)
	controller := NewProductController(*productUseCase)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/products/:productId", controller.GetProductById)
	router.POST("/products", controller.CreateProduct)
	return router, mock, func() { db.Close() }
}

func TestGetProductByIdReturnsBadRequestForInvalidID(t *testing.T) {
	router, mock, cleanup := setupProductRouter(t)
	defer cleanup()

	request := httptest.NewRequest(http.MethodGet, "/products/not-a-number", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "inválido") {
		t.Fatalf("expected invalid ID message, got %s", response.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProductByIdReturnsNotFound(t *testing.T) {
	router, mock, cleanup := setupProductRouter(t)
	defer cleanup()

	query := "SELECT id, product_name, price FROM product WHERE id = $1"
	mock.ExpectPrepare(regexp.QuoteMeta(query))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(999).WillReturnRows(sqlmock.NewRows([]string{"id", "product_name", "price"}))

	request := httptest.NewRequest(http.MethodGet, "/products/999", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "não encontrado") {
		t.Fatalf("expected not found message, got %s", response.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateProductReturnsCreatedProduct(t *testing.T) {
	router, mock, cleanup := setupProductRouter(t)
	defer cleanup()

	query := "INSERT INTO product (product_name, price) VALUES($1, $2) RETURNING id"
	mock.ExpectPrepare(regexp.QuoteMeta(query))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("Keyboard", 149.90).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(12))

	request := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(`{"name":"Keyboard","price":149.90}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"id_product":12`) {
		t.Fatalf("expected created product ID, got %s", response.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
