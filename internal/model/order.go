package model

import "time"

type Order struct {
	ID        int       `json:"id"`
	Item      string    `json:"item"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
