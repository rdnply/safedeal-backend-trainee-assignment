package order

import (
	"safedeal-backend-trainee/internal/ftime"
)

type Order struct {
	ID          int64             `json:"id"`
	ProductID   int64             `json:"product_id"`
	Name        string            `json:"name"`
	From        string            `json:"from,omitempty"`
	Destination string            `json:"destination,omitempty"`
	Time        *ftime.FormatTime `json:"time,omitempty"`
}

type Storage interface {
	Create(o *Order) error
	GetAll() ([]*Order, error)
	FindByID(id int64) (*Order, error)
}
