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
	"go-api/service"
	"go-api/usecase"
)

func setupProductRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	transaction, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = transaction.Rollback()
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet SQL expectations: %v", err)
		}
	})
	controller := NewProductController(service.NewProductService(*usecase.NewProductUseCase(*repository.NewProductRepositoryWithTransaction(transaction))))
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/products", controller.GetProducts)
	router.GET("/products/:productId", controller.GetProductById)
	router.POST("/products", controller.CreateProduct)
	router.PUT("/products/:productId", controller.UpdateProduct)
	router.PATCH("/products/:productId", controller.UpdateProduct)
	router.DELETE("/products/:productId", controller.DeleteProduct)
	t.Cleanup(func() { _ = db.Close() })
	return router, mock
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
			router, mock := setupProductRouter(t)
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
			router, mock := setupProductRouter(t)
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
	NewProductController(service.NewProductService(usecase.ProductUseCase{})).GetProductById(context)

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
		{"missing required fields", `{}`, http.StatusBadRequest, nil},
		{"invalid price", `{"name":"Mouse","price":0}`, http.StatusBadRequest, nil},
		{"database error", `{"name":"Mouse","price":10}`, http.StatusInternalServerError, errors.New("insert failed")},
		{"success", `{"name":"Mouse","price":10}`, http.StatusCreated, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, mock := setupProductRouter(t)
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

func TestUpdateProduct(t *testing.T) {
	tests := []struct {
		name, path, body string
		code             int
		err              error
		rows             int64
	}{
		{"invalid ID", "/products/nope", `{"name":"Mouse","price":10}`, http.StatusBadRequest, nil, 0},
		{"zero ID", "/products/0", `{"name":"Mouse","price":10}`, http.StatusBadRequest, nil, 0},
		{"invalid JSON", "/products/1", `{"name":`, http.StatusBadRequest, nil, 0},
		{"invalid fields", "/products/1", `{}`, http.StatusBadRequest, nil, 0},
		{"not found", "/products/1", `{"name":"Mouse","price":10}`, http.StatusNotFound, nil, 0},
		{"database error", "/products/1", `{"name":"Mouse","price":10}`, http.StatusInternalServerError, errors.New("update failed"), 0},
		{"success", "/products/1", `{"name":"Mouse","price":10}`, http.StatusOK, nil, 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, mock := setupProductRouter(t)
			if test.code != http.StatusBadRequest {
				expectation := mock.ExpectExec(regexp.QuoteMeta("UPDATE product SET product_name = $1, price = $2 WHERE id = $3")).WithArgs("Mouse", 10.0, 1)
				if test.err != nil {
					expectation.WillReturnError(test.err)
				} else {
					expectation.WillReturnResult(sqlmock.NewResult(0, test.rows))
				}
			}
			request := httptest.NewRequest(http.MethodPut, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != test.code {
				t.Fatalf("expected %d, got %d: %s", test.code, response.Code, response.Body.String())
			}
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	tests := []struct {
		name, path string
		code       int
		err        error
		rows       int64
	}{
		{"invalid ID", "/products/nope", http.StatusBadRequest, nil, 0}, {"not found", "/products/1", http.StatusNotFound, nil, 0},
		{"database error", "/products/1", http.StatusInternalServerError, errors.New("delete failed"), 0}, {"success", "/products/1", http.StatusNoContent, nil, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			router, mock := setupProductRouter(t)
			if test.path != "/products/nope" {
				expectation := mock.ExpectExec(regexp.QuoteMeta("DELETE FROM product WHERE id = $1")).WithArgs(1)
				if test.err != nil {
					expectation.WillReturnError(test.err)
				} else {
					expectation.WillReturnResult(sqlmock.NewResult(0, test.rows))
				}
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete, test.path, nil))
			if response.Code != test.code {
				t.Fatalf("expected %d, got %d: %s", test.code, response.Code, response.Body.String())
			}
		})
	}
}

func TestUpdateProductRejectsEmptyID(t *testing.T) {
	context, _ := gin.CreateTestContext(httptest.NewRecorder())
	context.Params = gin.Params{{Key: "productId", Value: ""}}
	NewProductController(service.NewProductService(usecase.ProductUseCase{})).UpdateProduct(context)

	if context.Writer.Status() != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", context.Writer.Status())
	}
}
