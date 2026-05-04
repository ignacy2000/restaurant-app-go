package table

type CreateReq struct {
	Number   int `json:"number"   binding:"required,min=1"`
	Capacity int `json:"capacity" binding:"required,min=1"`
}

type UpdateReq struct {
	Capacity int `json:"capacity" binding:"required,min=1"`
}

type Response struct {
	ID           string `json:"id"`
	RestaurantID string `json:"restaurant_id"`
	Number       int    `json:"number"`
	Capacity     int    `json:"capacity"`
	CreatedAt    string `json:"created_at"`
}
