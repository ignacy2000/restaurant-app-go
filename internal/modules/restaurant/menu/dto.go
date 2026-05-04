package menu

type CreateReq struct {
	Name        string `json:"name"        binding:"required"`
	Description string `json:"description"`
}

type Response struct {
	ID           string `json:"id"`
	RestaurantID string `json:"restaurant_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	CreatedAt    string `json:"created_at"`
}

type CreateItemReq struct {
	Name        string  `json:"name"     binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Position    int     `json:"position"`
}

type ItemResponse struct {
	ID          string  `json:"id"`
	MenuID      string  `json:"menu_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Position    int     `json:"position"`
}
