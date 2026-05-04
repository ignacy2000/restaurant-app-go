package call

type UpdateStatusReq struct {
	Status CallStatus `json:"status" binding:"required"`
}

type Response struct {
	ID           string     `json:"id"`
	RestaurantID string     `json:"restaurant_id"`
	TableID      string     `json:"table_id"`
	Status       CallStatus `json:"status"`
	CreatedAt    string     `json:"created_at"`
}
