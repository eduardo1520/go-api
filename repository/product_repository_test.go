package repository

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go-api/model"
)

func TestGetProductByIdReturnsProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectPrepare(regexp.QuoteMeta("SELECT id, product_name, price FROM product WHERE id = $1"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, product_name, price FROM product WHERE id = $1")).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(7, "Mouse", 89.90))

	product, err := NewProductRepository(db).GetProductById(7)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if product == nil || product.ID != 7 || product.Name != "Mouse" || product.Price != 89.90 {
		t.Fatalf("unexpected product: %+v", product)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProductByIdReturnsNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	query := "SELECT id, product_name, price FROM product WHERE id = $1"
	mock.ExpectPrepare(regexp.QuoteMeta(query))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(999).WillReturnRows(sqlmock.NewRows([]string{"id", "product_name", "price"}))

	product, err := NewProductRepository(db).GetProductById(999)
	if product != nil {
		t.Fatalf("expected nil product, got %+v", product)
	}
	if !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestCreateProductReturnsGeneratedID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	query := "INSERT INTO product (product_name, price) VALUES($1, $2) RETURNING id"
	mock.ExpectPrepare(regexp.QuoteMeta(query))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("Keyboard", 149.90).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(12))

	id, err := NewProductRepository(db).CreateProduct(model.Product{Name: "Keyboard", Price: 149.90})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 12 {
		t.Fatalf("expected ID 12, got %d", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
