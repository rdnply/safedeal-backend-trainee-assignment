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
	getAllStmt *sql.Stmt
}

func NewOrderStorage(db *DB) (*OrderStorage, error) {
	s := &OrderStorage{statementStorage: newStatementsStorage(db)}

	stmts := []stmt{
		{Query: createOrderQuery, Dst: &s.createStmt},
		{Query: getAllOrderQuery, Dst: &s.getAllStmt},
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

const getAllOrderQuery = "SELECT id, " + orderFields + " FROM orders"

func (s *OrderStorage) GetAll() ([]*order.Order, error) {
	rows, err := s.getAllStmt.Query()
	if err != nil {
		return nil, errors.Wrap(err, "can't exec query to get all orders")
	}

	defer rows.Close()

	orders := make([]*order.Order, 0)

	for rows.Next() {
		var o order.Order

		err = scanOrder(rows, &o)
		if err != nil {
			return nil, errors.Wrap(err, "can't scan row with order")
		}

		orders = append(orders, &o)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows contain error")
	}

	return orders, nil
}
