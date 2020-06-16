package postgres

import (
	"database/sql"
	"safedeal-backend-trainee/internal/product"

	"github.com/pkg/errors"
)

var _ product.Storage = &ProductStorage{}

type ProductStorage struct {
	statementStorage

	findByIDStmt *sql.Stmt
}

func NewProductStorage(db *DB) (*ProductStorage, error) {
	s := &ProductStorage{statementStorage: newStatementsStorage(db)}

	stmts := []stmt{
		{Query: findProductByIDQuery, Dst: &s.findByIDStmt},
	}

	if err := s.initStatements(stmts); err != nil {
		return nil, errors.Wrap(err, "can't init statements")
	}

	return s, nil
}

func scanProduct(scanner sqlScanner, p *product.Product) error {
	return scanner.Scan(&p.ID, &p.Name, &p.Width, &p.Length, &p.Height, &p.Weight, &p.Place)
}

const productFields = "name, width, length, height, weight, place"
const findProductByIDQuery = "SELECT id, " + productFields + " FROM products WHERE id=$1"

func (s *ProductStorage) FindByID(id int64) (*product.Product, error) {
	var p product.Product

	row := s.findByIDStmt.QueryRow(id)
	if err := scanProduct(row, &p); err != nil {
		if err == sql.ErrNoRows {
			return &p, nil
		}

		return &p, errors.Wrap(err, "can't scan product")
	}

	return &p, nil
}
