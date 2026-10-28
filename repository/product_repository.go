package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"go-api/model"

	sq "github.com/Masterminds/squirrel"
)

var ErrProductNotFound = errors.New("product not found")
var queryBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type ProductRepository struct {
	connection  *sql.DB
	transaction *sql.Tx
}

func NewProductRepository(connection *sql.DB) *ProductRepository {
	return &ProductRepository{connection: connection}
}

func NewProductRepositoryWithTransaction(transaction *sql.Tx) *ProductRepository {
	return &ProductRepository{transaction: transaction}
}

func (pr *ProductRepository) begin() (*sql.Tx, bool, error) {
	if pr.transaction != nil {
		return pr.transaction, false, nil
	}

	transaction, err := pr.connection.Begin()
	if err != nil {
		return nil, false, err
	}
	return transaction, true, nil
}

func (pr *ProductRepository) GetProducts() ([]model.Product, error) {
	tx, ownsTransaction, err := pr.begin()
	if err != nil {
		return []model.Product{}, err
	}

	query, args, err := queryBuilder.
		Select("id", "product_name", "price").
		From("product").
		ToSql()
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		return []model.Product{}, err
	}

	rows, err := tx.Query(query, args...)
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		fmt.Println(err)
		return []model.Product{}, err
	}

	var productList []model.Product
	var productObj model.Product

	for rows.Next() {
		err := rows.Scan(&productObj.ID, &productObj.Name, &productObj.Price)
		if err != nil {
			_ = rows.Close()
			if ownsTransaction {
				_ = tx.Rollback()
			}
			fmt.Println(err)
			return []model.Product{}, err
		}
		productList = append(productList, productObj)
	}

	_ = rows.Close()
	if ownsTransaction {
		if err := tx.Commit(); err != nil {
			return []model.Product{}, err
		}
	}

	return productList, nil
}

func (pr *ProductRepository) CreateProduct(product model.Product) (int, error) {
	tx, ownsTransaction, err := pr.begin()
	if err != nil {
		return 0, err
	}

	var id int
	query, args, err := queryBuilder.
		Insert("product").
		Columns("product_name", "price").
		Values(product.Name, product.Price).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		fmt.Println(err)
		return 0, err
	}
	statement, err := tx.Prepare(query)
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		return 0, err
	}
	defer statement.Close()

	err = statement.QueryRow(args...).Scan(&id)
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		fmt.Println(err)
		return 0, err
	}

	if ownsTransaction {
		if err := tx.Commit(); err != nil {
			return 0, err
		}
	}
	return id, nil
}

func (pr *ProductRepository) GetProductById(id_product int) (*model.Product, error) {
	tx, ownsTransaction, err := pr.begin()
	if err != nil {
		return nil, err
	}

	query, args, err := queryBuilder.
		Select("id", "product_name", "price").
		From("product").
		Where(sq.Eq{"id": id_product}).
		ToSql()
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		fmt.Println(err)
		return nil, err
	}

	var product model.Product
	statement, err := tx.Prepare(query)
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		return nil, err
	}
	defer statement.Close()

	err = statement.QueryRow(args...).Scan(&product.ID, &product.Name, &product.Price)
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	if ownsTransaction {
		if err := tx.Commit(); err != nil {
			return nil, err
		}
	}
	return &product, nil
}

func (pr *ProductRepository) UpdateProduct(product model.Product) error {
	tx, ownsTransaction, err := pr.begin()
	if err != nil {
		return err
	}

	query, args, err := queryBuilder.
		Update("product").
		Set("product_name", product.Name).
		Set("price", product.Price).
		Where(sq.Eq{"id": product.ID}).
		ToSql()
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		return err
	}

	result, err := tx.Exec(query, args...)
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		return err
	}
	if rowsAffected == 0 {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		return ErrProductNotFound
	}

	if ownsTransaction {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func (pr *ProductRepository) DeleteProduct(id int) error {
	tx, ownsTransaction, err := pr.begin()
	if err != nil {
		return err
	}

	query, args, err := queryBuilder.
		Delete("product").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		return err
	}

	result, err := tx.Exec(query, args...)
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		return err
	}
	if rowsAffected == 0 {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		return ErrProductNotFound
	}

	if ownsTransaction {
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
