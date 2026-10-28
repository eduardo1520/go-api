package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"go-api/model"
)

var ErrProductNotFound = errors.New("product not found")

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

	query := "SELECT id, product_name, price FROM product"
	rows, err := tx.Query(query)
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
	query, err := tx.Prepare("INSERT INTO product" +
		" (product_name, price)" +
		" VALUES($1, $2) RETURNING id")
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		fmt.Println(err)
		return 0, err
	}
	defer query.Close()

	err = query.QueryRow(product.Name, product.Price).Scan(&id)
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

	query, err := tx.Prepare("SELECT id, product_name, price FROM product WHERE id = $1")
	if err != nil {
		if ownsTransaction {
			_ = tx.Rollback()
		}
		fmt.Println(err)
		return nil, err
	}
	defer query.Close()

	var product model.Product
	err = query.QueryRow(id_product).Scan(&product.ID, &product.Name, &product.Price)
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

	result, err := tx.Exec("UPDATE product SET product_name = $1, price = $2 WHERE id = $3", product.Name, product.Price, product.ID)
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

	result, err := tx.Exec("DELETE FROM product WHERE id = $1", id)
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
