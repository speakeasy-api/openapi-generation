package ecommerce

import "time"

type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ProductList struct {
	NextCursor *string    `json:"nextCursor"`
	Products   []*Product `json:"products"`
}

type NewProductForm struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type ProductUpdateForm struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
}

type ProductInventoryUpdateForm struct {
	QuantityDelta int `json:"quantityDelta"`
}

type ProductInventoryStatus struct {
	ProductID string    `json:"productId"`
	Quantity  int       `json:"quantity"`
	UpdatedAt time.Time `json:"updatedAt"`
}
