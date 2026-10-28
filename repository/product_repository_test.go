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
	updateQuery   = "UPDATE product SET product_name = $1, price = $2 WHERE id = $3"
	deleteQuery   = "DELETE FROM product WHERE id = $1"
)

func setupRepository(t *testing.T) (*ProductRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	mock.ExpectBegin()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = tx.Rollback()
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet SQL expectations: %v", err)
		}
	})
	return NewProductRepositoryWithTransaction(tx), mock
}

func setupOwnedRepository(t *testing.T) (*ProductRepository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return NewProductRepository(db), mock
}

func TestRepositoryConstructorsAndBegin(t *testing.T) {
	t.Run("external transaction", func(t *testing.T) {
		repo, _ := setupRepository(t)
		tx, owns, err := repo.begin()
		if err != nil || tx == nil || owns {
			t.Fatalf("unexpected result: tx=%v owns=%t err=%v", tx, owns, err)
		}
	})

	t.Run("database transaction", func(t *testing.T) {
		repo, mock := setupOwnedRepository(t)
		mock.ExpectBegin()
		tx, owns, err := repo.begin()
		if err != nil || tx == nil || !owns {
			t.Fatalf("unexpected result: tx=%v owns=%t err=%v", tx, owns, err)
		}
		mock.ExpectRollback()
		if err := tx.Rollback(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("begin error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
		tx, owns, err := NewProductRepository(db).begin()
		if tx != nil || owns || err == nil {
			t.Fatalf("expected begin error, got tx=%v owns=%t err=%v", tx, owns, err)
		}
	})
}

func TestGetProductsExternalTransaction(t *testing.T) {
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

func TestGetProductsOwnedTransaction(t *testing.T) {
	tests := []struct {
		name      string
		setupRows func() *sqlmock.Rows
		queryErr  error
		beginErr  error
		commitErr error
		wantErr   bool
	}{
		{"success", func() *sqlmock.Rows {
			return sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(1, "Mouse", 10)
		}, nil, nil, nil, false},
		{"begin error", nil, nil, sql.ErrConnDone, nil, true},
		{"query error", nil, sql.ErrConnDone, nil, nil, true},
		{"scan error", func() *sqlmock.Rows {
			return sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow("bad", "Mouse", 10)
		}, nil, nil, nil, true},
		{"rows error", func() *sqlmock.Rows {
			return sqlmock.NewRows([]string{"id", "product_name", "price"}).
				AddRow(1, "Mouse", 10).
				AddRow(2, "Keyboard", 20).
				RowError(1, errors.New("row failed"))
		}, nil, nil, nil, true},
		{"commit error", func() *sqlmock.Rows {
			return sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(1, "Mouse", 10)
		}, nil, nil, sql.ErrTxDone, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := setupOwnedRepository(t)
			if test.beginErr != nil {
				mock.ExpectBegin().WillReturnError(test.beginErr)
			} else {
				mock.ExpectBegin()
				expectation := mock.ExpectQuery(regexp.QuoteMeta(productsQuery))
				if test.queryErr != nil {
					expectation.WillReturnError(test.queryErr)
					mock.ExpectRollback()
				} else {
					expectation.WillReturnRows(test.setupRows())
					if test.commitErr != nil {
						mock.ExpectCommit().WillReturnError(test.commitErr)
					} else if test.wantErr {
						mock.ExpectRollback()
					} else {
						mock.ExpectCommit()
					}
				}
			}
			products, err := repo.GetProducts()
			if test.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !test.wantErr && (err != nil || len(products) != 1) {
				t.Fatalf("unexpected result: %+v %v", products, err)
			}
		})
	}
}

func TestGetProductByIDExternalTransaction(t *testing.T) {
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

func TestGetProductByIDOwnedTransaction(t *testing.T) {
	tests := []struct {
		name       string
		prepareErr error
		queryErr   error
		rows       *sqlmock.Rows
		beginErr   error
		commitErr  error
		wantErr    bool
	}{
		{"success", nil, nil, sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(1, "Mouse", 10), nil, nil, false},
		{"begin error", nil, nil, nil, sql.ErrConnDone, nil, true},
		{"prepare error", sql.ErrConnDone, nil, nil, nil, nil, true},
		{"query error", nil, sql.ErrConnDone, nil, nil, nil, true},
		{"not found", nil, nil, sqlmock.NewRows([]string{"id", "product_name", "price"}), nil, nil, true},
		{"scan error", nil, nil, sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow("bad", "Mouse", 10), nil, nil, true},
		{"commit error", nil, nil, sqlmock.NewRows([]string{"id", "product_name", "price"}).AddRow(1, "Mouse", 10), nil, sql.ErrTxDone, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := setupOwnedRepository(t)
			if test.beginErr != nil {
				mock.ExpectBegin().WillReturnError(test.beginErr)
			} else {
				mock.ExpectBegin()
				prepare := mock.ExpectPrepare(regexp.QuoteMeta(productQuery))
				if test.prepareErr != nil {
					prepare.WillReturnError(test.prepareErr)
					mock.ExpectRollback()
				} else {
					expectation := mock.ExpectQuery(regexp.QuoteMeta(productQuery)).WithArgs(1)
					if test.queryErr != nil {
						expectation.WillReturnError(test.queryErr)
					} else {
						expectation.WillReturnRows(test.rows)
					}
					if !test.wantErr {
						if test.commitErr != nil {
							mock.ExpectCommit().WillReturnError(test.commitErr)
						} else {
							mock.ExpectCommit()
						}
					} else {
						mock.ExpectRollback()
					}
				}
			}
			product, err := repo.GetProductById(1)
			if test.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !test.wantErr && (err != nil || product == nil) {
				t.Fatalf("unexpected result: %+v %v", product, err)
			}
		})
	}
}

func TestCreateProductExternalTransaction(t *testing.T) {
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

func TestCreateProductOwnedTransaction(t *testing.T) {
	tests := []struct {
		name       string
		prepareErr error
		queryErr   error
		beginErr   error
		commitErr  error
		wantErr    bool
	}{
		{"success", nil, nil, nil, nil, false},
		{"begin error", nil, nil, sql.ErrConnDone, nil, true},
		{"prepare error", sql.ErrConnDone, nil, nil, nil, true},
		{"query error", nil, sql.ErrConnDone, nil, nil, true},
		{"commit error", nil, nil, nil, sql.ErrTxDone, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := setupOwnedRepository(t)
			if test.beginErr != nil {
				mock.ExpectBegin().WillReturnError(test.beginErr)
			} else {
				mock.ExpectBegin()
				prepare := mock.ExpectPrepare(regexp.QuoteMeta(insertQuery))
				if test.prepareErr != nil {
					prepare.WillReturnError(test.prepareErr)
					mock.ExpectRollback()
				} else {
					expectation := mock.ExpectQuery(regexp.QuoteMeta(insertQuery)).WithArgs("Mouse", 10.0)
					if test.queryErr != nil {
						expectation.WillReturnError(test.queryErr)
						mock.ExpectRollback()
					} else {
						expectation.WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
						if test.commitErr != nil {
							mock.ExpectCommit().WillReturnError(test.commitErr)
						} else {
							mock.ExpectCommit()
						}
					}
				}
			}
			id, err := repo.CreateProduct(model.Product{Name: "Mouse", Price: 10})
			if test.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !test.wantErr && (err != nil || id != 1) {
				t.Fatalf("unexpected result: id=%d err=%v", id, err)
			}
		})
	}
}

func TestUpdateProductExternalTransaction(t *testing.T) {
	tests := []struct {
		name string
		err  error
		rows int64
	}{
		{"success", nil, 1}, {"exec error", sql.ErrConnDone, 0}, {"not found", nil, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := setupRepository(t)
			expectation := mock.ExpectExec(regexp.QuoteMeta(updateQuery)).WithArgs("Mouse", 10.0, 1)
			if test.err != nil {
				expectation.WillReturnError(test.err)
			} else {
				expectation.WillReturnResult(sqlmock.NewResult(0, test.rows))
			}
			err := repo.UpdateProduct(model.Product{ID: 1, Name: "Mouse", Price: 10})
			if test.name == "success" && err != nil {
				t.Fatal(err)
			}
			if test.name != "success" && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestDeleteProductExternalTransaction(t *testing.T) {
	tests := []struct {
		name string
		err  error
		rows int64
	}{
		{"success", nil, 1}, {"exec error", sql.ErrConnDone, 0}, {"not found", nil, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := setupRepository(t)
			expectation := mock.ExpectExec(regexp.QuoteMeta(deleteQuery)).WithArgs(1)
			if test.err != nil {
				expectation.WillReturnError(test.err)
			} else {
				expectation.WillReturnResult(sqlmock.NewResult(0, test.rows))
			}
			err := repo.DeleteProduct(1)
			if test.name == "success" && err != nil {
				t.Fatal(err)
			}
			if test.name != "success" && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestUpdateProductOwnedTransaction(t *testing.T) {
	tests := []struct {
		name      string
		beginErr  error
		execErr   error
		resultErr error
		rows      int64
		commitErr error
		wantErr   bool
	}{
		{"success", nil, nil, nil, 1, nil, false},
		{"begin error", sql.ErrConnDone, nil, nil, 0, nil, true},
		{"exec error", nil, sql.ErrConnDone, nil, 0, nil, true},
		{"rows affected error", nil, nil, sql.ErrConnDone, 0, nil, true},
		{"not found", nil, nil, nil, 0, nil, true},
		{"commit error", nil, nil, nil, 1, sql.ErrTxDone, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := setupOwnedRepository(t)
			if test.beginErr != nil {
				mock.ExpectBegin().WillReturnError(test.beginErr)
			} else {
				mock.ExpectBegin()
				expectation := mock.ExpectExec(regexp.QuoteMeta(updateQuery)).WithArgs("Mouse", 10.0, 1)
				switch {
				case test.execErr != nil:
					expectation.WillReturnError(test.execErr)
					mock.ExpectRollback()
				case test.resultErr != nil:
					expectation.WillReturnResult(sqlmock.NewErrorResult(test.resultErr))
					mock.ExpectRollback()
				default:
					expectation.WillReturnResult(sqlmock.NewResult(0, test.rows))
					if test.commitErr != nil {
						mock.ExpectCommit().WillReturnError(test.commitErr)
					} else if test.wantErr {
						mock.ExpectRollback()
					} else {
						mock.ExpectCommit()
					}
				}
			}
			err := repo.UpdateProduct(model.Product{ID: 1, Name: "Mouse", Price: 10})
			if test.wantErr != (err != nil) {
				t.Fatalf("expected error=%t, got %v", test.wantErr, err)
			}
		})
	}
}

func TestDeleteProductOwnedTransaction(t *testing.T) {
	tests := []struct {
		name      string
		beginErr  error
		execErr   error
		resultErr error
		rows      int64
		commitErr error
		wantErr   bool
	}{
		{"success", nil, nil, nil, 1, nil, false},
		{"begin error", sql.ErrConnDone, nil, nil, 0, nil, true},
		{"exec error", nil, sql.ErrConnDone, nil, 0, nil, true},
		{"rows affected error", nil, nil, sql.ErrConnDone, 0, nil, true},
		{"not found", nil, nil, nil, 0, nil, true},
		{"commit error", nil, nil, nil, 1, sql.ErrTxDone, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo, mock := setupOwnedRepository(t)
			if test.beginErr != nil {
				mock.ExpectBegin().WillReturnError(test.beginErr)
			} else {
				mock.ExpectBegin()
				expectation := mock.ExpectExec(regexp.QuoteMeta(deleteQuery)).WithArgs(1)
				switch {
				case test.execErr != nil:
					expectation.WillReturnError(test.execErr)
					mock.ExpectRollback()
				case test.resultErr != nil:
					expectation.WillReturnResult(sqlmock.NewErrorResult(test.resultErr))
					mock.ExpectRollback()
				default:
					expectation.WillReturnResult(sqlmock.NewResult(0, test.rows))
					if test.commitErr != nil {
						mock.ExpectCommit().WillReturnError(test.commitErr)
					} else if test.wantErr {
						mock.ExpectRollback()
					} else {
						mock.ExpectCommit()
					}
				}
			}
			err := repo.DeleteProduct(1)
			if test.wantErr != (err != nil) {
				t.Fatalf("expected error=%t, got %v", test.wantErr, err)
			}
		})
	}
}
