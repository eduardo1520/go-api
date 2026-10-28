package service

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go-api/model"
	"go-api/repository"
	"go-api/usecase"
)

func setupProductService(t *testing.T) (ProductService, sqlmock.Sqlmock) {
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
		_ = db.Close()
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet SQL expectations: %v", err)
		}
	})

	repo := repository.NewProductRepositoryWithTransaction(transaction)
	productUseCase := usecase.NewProductUseCase(*repo)
	return NewProductService(*productUseCase), mock
}

func TestProductServiceDelegatesProductOperations(t *testing.T) {
	t.Run("get products", func(t *testing.T) {
		productService, mock := setupProductService(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, product_name, price FROM product")).
			WillReturnRows(sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(1, "Mouse", 10))

		products, err := productService.GetProducts()
		if err != nil || len(products) != 1 {
			t.Fatalf("unexpected result: products=%+v err=%v", products, err)
		}
	})

	t.Run("create product", func(t *testing.T) {
		productService, mock := setupProductService(t)
		query := "INSERT INTO product (product_name,price) VALUES ($1,$2) RETURNING id"
		mock.ExpectPrepare(regexp.QuoteMeta(query))
		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("Mouse", 10.0).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

		product, err := productService.CreateProduct(model.Product{Name: "Mouse", Price: 10})
		if err != nil || product.ID != 2 {
			t.Fatalf("unexpected result: product=%+v err=%v", product, err)
		}
	})

	t.Run("get product by id", func(t *testing.T) {
		productService, mock := setupProductService(t)
		query := "SELECT id, product_name, price FROM product WHERE id = $1"
		mock.ExpectPrepare(regexp.QuoteMeta(query))
		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(2, "Mouse", 10))

		product, err := productService.GetProductById(2)
		if err != nil || product == nil || product.ID != 2 {
			t.Fatalf("unexpected result: product=%+v err=%v", product, err)
		}
	})

	t.Run("update product", func(t *testing.T) {
		productService, mock := setupProductService(t)
		mock.ExpectExec(regexp.QuoteMeta("UPDATE product SET product_name = $1, price = $2 WHERE id = $3")).
			WithArgs("Mouse", 10.0, 2).WillReturnResult(sqlmock.NewResult(0, 1))

		if err := productService.UpdateProduct(model.Product{ID: 2, Name: "Mouse", Price: 10}); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("delete product", func(t *testing.T) {
		productService, mock := setupProductService(t)
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM product WHERE id = $1")).
			WithArgs(2).WillReturnResult(sqlmock.NewResult(0, 1))

		if err := productService.DeleteProduct(2); err != nil {
			t.Fatal(err)
		}
	})
}

func TestProductServicePropagatesErrors(t *testing.T) {
	t.Run("get products", func(t *testing.T) {
		productService, mock := setupProductService(t)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, product_name, price FROM product")).
			WillReturnError(errors.New("query failed"))

		if _, err := productService.GetProducts(); err == nil {
			t.Fatal("expected error")
		}

	})

	t.Run("create product", func(t *testing.T) {
		productService, mock := setupProductService(t)
		query := "INSERT INTO product (product_name,price) VALUES ($1,$2) RETURNING id"
		mock.ExpectPrepare(regexp.QuoteMeta(query))
		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("Mouse", 10.0).
			WillReturnError(errors.New("insert failed"))

		if _, err := productService.CreateProduct(model.Product{Name: "Mouse", Price: 10}); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("get product by id", func(t *testing.T) {
		productService, mock := setupProductService(t)
		query := "SELECT id, product_name, price FROM product WHERE id = $1"
		mock.ExpectPrepare(regexp.QuoteMeta(query))
		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(2).
			WillReturnError(errors.New("query failed"))

		if _, err := productService.GetProductById(2); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("update product", func(t *testing.T) {
		productService, mock := setupProductService(t)
		mock.ExpectExec(regexp.QuoteMeta("UPDATE product SET product_name = $1, price = $2 WHERE id = $3")).
			WithArgs("Mouse", 10.0, 2).WillReturnError(errors.New("update failed"))

		if err := productService.UpdateProduct(model.Product{ID: 2, Name: "Mouse", Price: 10}); err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("delete product", func(t *testing.T) {
		productService, mock := setupProductService(t)
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM product WHERE id = $1")).
			WithArgs(2).WillReturnError(errors.New("delete failed"))

		if err := productService.DeleteProduct(2); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestProductServiceTranslatesNotFoundErrors(t *testing.T) {
	t.Run("get product by id", func(t *testing.T) {
		productService, mock := setupProductService(t)
		query := "SELECT id, product_name, price FROM product WHERE id = $1"
		mock.ExpectPrepare(regexp.QuoteMeta(query))
		mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(2).
			WillReturnRows(sqlmock.NewRows([]string{"id", "product_name", "price"}))
		_, err := productService.GetProductById(2)
		if !errors.Is(err, ErrProductNotFound) {
			t.Fatalf("expected service not found error, got %v", err)
		}
	})

	t.Run("update product", func(t *testing.T) {
		productService, mock := setupProductService(t)
		mock.ExpectExec(regexp.QuoteMeta("UPDATE product SET product_name = $1, price = $2 WHERE id = $3")).
			WithArgs("Mouse", 10.0, 2).WillReturnResult(sqlmock.NewResult(0, 0))
		err := productService.UpdateProduct(model.Product{ID: 2, Name: "Mouse", Price: 10})
		if !errors.Is(err, ErrProductNotFound) {
			t.Fatalf("expected service not found error, got %v", err)
		}
	})

	t.Run("delete product", func(t *testing.T) {
		productService, mock := setupProductService(t)
		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM product WHERE id = $1")).
			WithArgs(2).WillReturnResult(sqlmock.NewResult(0, 0))
		err := productService.DeleteProduct(2)
		if !errors.Is(err, ErrProductNotFound) {
			t.Fatalf("expected service not found error, got %v", err)
		}
	})
}
