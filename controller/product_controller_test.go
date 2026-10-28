package controller

import (
	"errors"
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
	controller := NewProductController(*usecase.NewProductUseCase(*repository.NewProductRepository(db)))
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/products", controller.GetProducts)
	router.GET("/products/:productId", controller.GetProductById)
	router.POST("/products", controller.CreateProduct)
	return router, mock, func() { _ = db.Close() }
}

func TestGetProducts(t *testing.T) {
	tests := []struct {
		name string
		err  error
		code int
	}{
		{"success", nil, http.StatusOK},
		{"error", errors.New("query failed"), http.StatusInternalServerError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, mock, cleanup := setupProductRouter(t)
			defer cleanup()
			query := "SELECT id, product_name, price FROM product"
			expectation := mock.ExpectQuery(regexp.QuoteMeta(query))
			if test.err != nil {
				expectation.WillReturnError(test.err)
			} else {
				expectation.WillReturnRows(sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(1, "Mouse", 10))
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/products", nil))
			if response.Code != test.code {
				t.Fatalf("expected %d, got %d", test.code, response.Code)
			}
		})
	}
}

func TestGetProductByID(t *testing.T) {
	tests := []struct {
		name string
		path string
		code int
		err  error
		rows bool
	}{
		{"invalid", "/products/nope", http.StatusBadRequest, nil, false},
		{"not found", "/products/9", http.StatusNotFound, nil, false},
		{"database error", "/products/9", http.StatusInternalServerError, errors.New("query failed"), false},
		{"success", "/products/1", http.StatusOK, nil, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, mock, cleanup := setupProductRouter(t)
			defer cleanup()
			if test.path != "/products/nope" {
				query := "SELECT id, product_name, price FROM product WHERE id = $1"
				mock.ExpectPrepare(regexp.QuoteMeta(query))
				id := 9
				if test.rows {
					id = 1
				}
				expectation := mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(id)
				if test.rows {
					expectation.WillReturnRows(sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(1, "Mouse", 10))
				} else if test.err != nil {
					expectation.WillReturnError(test.err)
				} else {
					expectation.WillReturnRows(sqlmock.NewRows([]string{"id", "product_name", "price"}))
				}
			}

			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))
			if response.Code != test.code {
				t.Fatalf("expected %d, got %d: %s", test.code, response.Code, response.Body.String())
			}
		})
	}
}

func TestGetProductByIDReturnsBadRequestForEmptyID(t *testing.T) {
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Params = gin.Params{{Key: "productId", Value: ""}}
	NewProductController(usecase.ProductUseCase{}).GetProductById(context)

	if context.Writer.Status() != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", context.Writer.Status())
	}
}

func TestCreateProduct(t *testing.T) {
	tests := []struct {
		name string
		body string
		code int
		err  error
	}{
		{"invalid JSON", `{"name":`, http.StatusBadRequest, nil},
		{"database error", `{"name":"Mouse","price":10}`, http.StatusInternalServerError, errors.New("insert failed")},
		{"success", `{"name":"Mouse","price":10}`, http.StatusCreated, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, mock, cleanup := setupProductRouter(t)
			defer cleanup()
			if test.err != nil || test.code == http.StatusCreated {
				query := "INSERT INTO product (product_name, price) VALUES($1, $2) RETURNING id"
				mock.ExpectPrepare(regexp.QuoteMeta(query))
				expectation := mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("Mouse", 10.0)
				if test.err != nil {
					expectation.WillReturnError(test.err)
				} else {
					expectation.WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(3))
				}
			}
			request := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != test.code {
				t.Fatalf("expected %d, got %d", test.code, response.Code)
			}
		})
	}
}
