package product

type Product struct {
	ID     int64   `json:"id,omitempty"`
	Name   string  `json:"name"`
	Width  float32 `json:"width"`
	Length float32 `json:"length"`
	Height float32 `json:"height"`
	Weight float32 `json:"weight"`
	Place  string  `json:"place"`
}

type Storage interface {
	FindByID(id int64) (*Product, error)
}
