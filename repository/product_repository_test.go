package repository

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go-api/model"
)

func newRepositoryTest(t *testing.T) (*ProductRepository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	return NewProductRepository(db), mock, func() { _ = db.Close() }
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
			repo, mock, cleanup := newRepositoryTest(t)
			defer cleanup()
			expectation := mock.ExpectQuery(regexp.QuoteMeta("SELECT id, product_name, price FROM product"))
			if test.err != nil {
				expectation.WillReturnError(test.err)
			} else {
				expectation.WillReturnRows(test.rows)
			}
			products, err := repo.GetProducts()
			if test.err == nil && test.name == "success" && (err != nil || len(products) != 1) {
				t.Fatalf("unexpected result: %+v %v", products, err)
			}
			if test.err != nil && err == nil {
				t.Fatal("expected query error")
			}
			if test.name == "scan error" && err == nil {
				t.Fatal("expected scan error")
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
			repo, mock, cleanup := newRepositoryTest(t)
			defer cleanup()
			query := "SELECT id, product_name, price FROM product WHERE id = $1"
			mock.ExpectPrepare(regexp.QuoteMeta(query))
			expectation := mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(1)
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

func TestGetProductByIDReturnsScanError(t *testing.T) {
	repo, mock, cleanup := newRepositoryTest(t)
	defer cleanup()

	query := "SELECT id, product_name, price FROM product WHERE id = $1"
	mock.ExpectPrepare(regexp.QuoteMeta(query))
	mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "product_name", "price"}).
			AddRow("bad", "Mouse", 10))

	product, err := repo.GetProductById(1)
	if product != nil || err == nil {
		t.Fatalf("expected scan error, got product=%+v err=%v", product, err)
	}

}

func TestGetProductByIDReturnsPrepareError(t *testing.T) {
	repo, mock, cleanup := newRepositoryTest(t)
	defer cleanup()

	query := "SELECT id, product_name, price FROM product WHERE id = $1"
	mock.ExpectPrepare(regexp.QuoteMeta(query)).WillReturnError(sql.ErrConnDone)

	product, err := repo.GetProductById(1)
	if product != nil || err == nil {
		t.Fatalf("expected prepare error, got product=%+v err=%v", product, err)
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
			repo, mock, cleanup := newRepositoryTest(t)
			defer cleanup()
			query := "INSERT INTO product (product_name, price) VALUES($1, $2) RETURNING id"
			prepare := mock.ExpectPrepare(regexp.QuoteMeta(query))
			if test.prepareErr != nil {
				prepare.WillReturnError(test.prepareErr)
			} else {
				expectation := mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs("Mouse", 10.0)
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
