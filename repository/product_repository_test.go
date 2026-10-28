package repository

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go-api/model"
)

const (
	productsQuery = "SELECT id, product_name, price FROM product"
	productQuery  = "SELECT id, product_name, price FROM product WHERE id = $1"
	insertQuery   = "INSERT INTO product (product_name, price) VALUES($1, $2) RETURNING id"
)

func setupRepository(t *testing.T) (*ProductRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
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
	return NewProductRepositoryWithTransaction(transaction), mock
}

func TestNewProductRepositoryUsesDatabaseConnection(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	repo := NewProductRepository(db)
	if repo.connection != db || repo.transaction != nil {
		t.Fatalf("unexpected repository configuration: %+v", repo)
	}
	mock.ExpectBegin()
	transaction, _, err := repo.begin()
	if err != nil {
		t.Fatalf("expected transaction to begin, got %v", err)
	}
	mock.ExpectRollback()
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryBeginReturnsDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	repo := NewProductRepository(db)
	transaction, ownsTransaction, err := repo.begin()
	if transaction != nil || ownsTransaction || err == nil {
		t.Fatalf("expected begin error, got transaction=%v owns=%t err=%v", transaction, ownsTransaction, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetProducts(t *testing.T) {
	tests := []struct {
		name string
		rows *sqlmock.Rows
		err  error
	}{
		{"success", sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(1, "Mouse", 10), nil},
		{"query error", nil, sql.ErrConnDone},
		{"scan error", sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow("bad", "Mouse", 10), nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := setupRepository(t)
			expectation := mock.ExpectQuery(regexp.QuoteMeta(productsQuery))
			if test.err != nil {
				expectation.WillReturnError(test.err)
			} else {
				expectation.WillReturnRows(test.rows)
			}

			products, err := repo.GetProducts()
			if test.name == "success" && (err != nil || len(products) != 1) {
				t.Fatalf("unexpected result: %+v %v", products, err)
			}
			if test.name != "success" && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestGetProductByID(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		rows    *sqlmock.Rows
		wantErr error
	}{
		{"success", nil, sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(1, "Mouse", 10), nil},
		{"not found", nil, sqlmock.NewRows([]string{"id", "product_name", "price"}), ErrProductNotFound},
		{"query error", sql.ErrConnDone, nil, sql.ErrConnDone},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := setupRepository(t)
			mock.ExpectPrepare(regexp.QuoteMeta(productQuery))
			expectation := mock.ExpectQuery(regexp.QuoteMeta(productQuery)).WithArgs(1)
			if test.err != nil {
				expectation.WillReturnError(test.err)
			} else {
				expectation.WillReturnRows(test.rows)
			}
			product, err := repo.GetProductById(1)
			if test.wantErr != nil && !errors.Is(err, test.wantErr) {
				t.Fatalf("expected %v, got %v", test.wantErr, err)
			}
			if test.wantErr == nil && (err != nil || product == nil) {
				t.Fatalf("unexpected result: %+v %v", product, err)
			}
		})
	}
}

func TestCreateProduct(t *testing.T) {
	tests := []struct {
		name       string
		prepareErr error
		queryErr   error
	}{
		{"success", nil, nil},
		{"prepare error", sql.ErrConnDone, nil},
		{"query error", nil, sql.ErrConnDone},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := setupRepository(t)
			prepare := mock.ExpectPrepare(regexp.QuoteMeta(insertQuery))
			if test.prepareErr != nil {
				prepare.WillReturnError(test.prepareErr)
			} else {
				expectation := mock.ExpectQuery(regexp.QuoteMeta(insertQuery)).WithArgs("Mouse", 10.0)
				if test.queryErr != nil {
					expectation.WillReturnError(test.queryErr)
				} else {
					expectation.WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				}
			}

			id, err := repo.CreateProduct(model.Product{Name: "Mouse", Price: 10})
			if test.name == "success" && (err != nil || id != 1) {
				t.Fatalf("unexpected result: %d %v", id, err)
			}
			if test.name != "success" && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
