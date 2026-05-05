package order

type ItemReq struct {
	Name     string `json:"name"     binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
	Notes    string `json:"notes"`
}

type CreateReq struct {
	Items      []ItemReq `json:"items"       binding:"required,min=1,dive"`
	Notes      string    `json:"notes"`
	GuestEmail string    `json:"guest_email" binding:"required,email"`
}

type ConfirmResponse struct {
	OrderID      string `json:"order_id"`
	RestaurantID string `json:"restaurant_id"`
	TableID      string `json:"table_id"`
}

type UpdateStatusReq struct {
	Status OrderStatus `json:"status" binding:"required"`
}

type ItemResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Notes    string `json:"notes,omitempty"`
}

type Response struct {
	ID           string        `json:"id"`
	RestaurantID string        `json:"restaurant_id"`
	TableID      string        `json:"table_id"`
	Status       OrderStatus   `json:"status"`
	Notes        string        `json:"notes,omitempty"`
	Items        []ItemResponse `json:"items"`
	CreatedAt    string        `json:"created_at"`
}
