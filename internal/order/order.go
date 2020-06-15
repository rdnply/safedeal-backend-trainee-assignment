package order

import "time"

type Order struct {
	ID          int64     `json:"id"`
	ProductID   int64     `json:"product_id"`
	Name        string    `json:"name"`
	From        string    `json:"from,omitempty"`
	Destination string    `json:"destination,omitempty"`
	Time        *time.Time `json:"time,omitempty"`
}

type Storage interface {
	Create(o *Order) error
	GetAll() ([]*Order, error)
}
