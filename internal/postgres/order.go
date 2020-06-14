package postgres

import (
	"database/sql"
	"github.com/pkg/errors"
	"safedeal-backend-trainee/internal/order"
)

var _ order.Storage = &OrderStorage{}

type OrderStorage struct {
	statementStorage

	createStmt *sql.Stmt
}

func NewOrderStorage(db *DB) (*OrderStorage, error) {
	s := &OrderStorage{statementStorage: newStatementsStorage(db)}

	stmts := []stmt{
		{Query: createOrderQuery, Dst: &s.createStmt},
	}

	if err := s.initStatements(stmts); err != nil {
		return nil, errors.Wrap(err, "can't init statements")
	}

	return s, nil
}

func scanOrder(scanner sqlScanner, o *order.Order) error {
	return scanner.Scan(&o.ID, &o.ProductID, &o.Name, &o.From, &o.Destination, &o.Time)
}

const orderFields = "product_id, name, from_place, destination, time"
const createOrderQuery = "INSERT INTO orders(" + orderFields + ") VALUES ($1, $2, $3, $4, $5) RETURNING id"

func (s *OrderStorage) Create(o *order.Order) error {
	if err := s.createStmt.QueryRow(o.ProductID, o.Name, o.From, o.Destination, o.Time).Scan(&o.ID); err != nil {
		return errors.Wrap(err, "can't exec query")
	}

	return nil
}
