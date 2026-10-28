package usecase

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go-api/model"
	"go-api/repository"
)

func newUseCaseTest(t *testing.T) (*ProductUseCase, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewProductRepository(db)
	return NewProductUseCase(*repo), mock, func() { _ = db.Close() }
}

func TestGetProductsDelegatesToRepository(t *testing.T) {
	useCase, mock, cleanup := newUseCaseTest(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, product_name, price FROM product")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(1, "Mouse", 10))

	products, err := useCase.GetProducts()
	if err != nil || len(products) != 1 {
		t.Fatalf("unexpected result: products=%+v err=%v", products, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProductsReturnsRepositoryError(t *testing.T) {
	useCase, mock, cleanup := newUseCaseTest(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, product_name, price FROM product")).
		WillReturnError(errors.New("query failed"))

	products, err := useCase.GetProducts()
	if err == nil || len(products) != 0 {
		t.Fatalf("expected repository error, got products=%+v err=%v", products, err)
	}
}

func TestCreateProductAssignsGeneratedID(t *testing.T) {
	useCase, mock, cleanup := newUseCaseTest(t)
	defer cleanup()

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
	useCase, mock, cleanup := newUseCaseTest(t)
	defer cleanup()

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
	useCase, mock, cleanup := newUseCaseTest(t)
	defer cleanup()

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
	useCase, mock, cleanup := newUseCaseTest(t)
	defer cleanup()

	query := "SELECT id, product_name, price FROM product WHERE id = $1"
	mock.ExpectPrepare(regexp.QuoteMeta(query))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(4).
		WillReturnError(errors.New("query failed"))

	product, err := useCase.GetProductById(4)
	if err == nil || product != nil {
		t.Fatalf("expected repository error, got product=%+v err=%v", product, err)
	}
}
