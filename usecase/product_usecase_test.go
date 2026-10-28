package usecase

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go-api/model"
	"go-api/repository"
)

func setupUseCase(t *testing.T) (*ProductUseCase, sqlmock.Sqlmock) {
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
	return NewProductUseCase(*repo), mock
}

func TestGetProductsDelegatesToRepository(t *testing.T) {
	useCase, mock := setupUseCase(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, product_name, price FROM product")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(1, "Mouse", 10))

	products, err := useCase.GetProducts()
	if err != nil || len(products) != 1 {
		t.Fatalf("unexpected result: products=%+v err=%v", products, err)
	}
}

func TestGetProductsReturnsRepositoryError(t *testing.T) {
	useCase, mock := setupUseCase(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, product_name, price FROM product")).
		WillReturnError(errors.New("query failed"))

	products, err := useCase.GetProducts()
	if err == nil || len(products) != 0 {
		t.Fatalf("expected repository error, got products=%+v err=%v", products, err)
	}
}

func TestCreateProductAssignsGeneratedID(t *testing.T) {
	useCase, mock := setupUseCase(t)

	query := "INSERT INTO product (product_name, price) VALUES($1, $2) RETURNING id"
	mock.ExpectPrepare(regexp.QuoteMeta(query))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("Mouse", 10.0).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(4))

	product, err := useCase.CreateProduct(model.Product{Name: "Mouse", Price: 10})
	if err != nil || product.ID != 4 {
		t.Fatalf("unexpected result: product=%+v err=%v", product, err)
	}
}

func TestCreateProductReturnsRepositoryError(t *testing.T) {
	useCase, mock := setupUseCase(t)

	query := "INSERT INTO product (product_name, price) VALUES($1, $2) RETURNING id"
	mock.ExpectPrepare(regexp.QuoteMeta(query))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("Mouse", 10.0).
		WillReturnError(errors.New("insert failed"))

	product, err := useCase.CreateProduct(model.Product{Name: "Mouse", Price: 10})
	if err == nil || product.ID != 0 {
		t.Fatalf("expected repository error, got product=%+v err=%v", product, err)
	}
}

func TestGetProductByIDDelegatesToRepository(t *testing.T) {
	useCase, mock := setupUseCase(t)

	query := "SELECT id, product_name, price FROM product WHERE id = $1"
	mock.ExpectPrepare(regexp.QuoteMeta(query))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(4).
		WillReturnRows(sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(4, "Mouse", 10))

	product, err := useCase.GetProductById(4)
	if err != nil || product == nil || product.ID != 4 {
		t.Fatalf("unexpected result: product=%+v err=%v", product, err)
	}
}

func TestGetProductByIDReturnsRepositoryError(t *testing.T) {
	useCase, mock := setupUseCase(t)

	query := "SELECT id, product_name, price FROM product WHERE id = $1"
	mock.ExpectPrepare(regexp.QuoteMeta(query))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(4).
		WillReturnError(errors.New("query failed"))

	product, err := useCase.GetProductById(4)
	if err == nil || product != nil {
		t.Fatalf("expected repository error, got product=%+v err=%v", product, err)
	}
}

func TestUpdateProductDelegatesToRepository(t *testing.T) {
	useCase, mock := setupUseCase(t)
	mock.ExpectExec(regexp.QuoteMeta("UPDATE product SET product_name = $1, price = $2 WHERE id = $3")).
		WithArgs("Mouse", 10.0, 4).WillReturnResult(sqlmock.NewResult(0, 1))

	if err := useCase.UpdateProduct(model.Product{ID: 4, Name: "Mouse", Price: 10}); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteProductReturnsRepositoryError(t *testing.T) {
	useCase, mock := setupUseCase(t)
	mock.ExpectExec(regexp.QuoteMeta("DELETE FROM product WHERE id = $1")).
		WithArgs(4).WillReturnError(errors.New("delete failed"))

	if err := useCase.DeleteProduct(4); err == nil {
		t.Fatal("expected repository error")
	}
}
