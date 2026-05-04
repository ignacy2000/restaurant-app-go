package restaurant

type CreateReq struct {
	Name        string `json:"name"        binding:"required"`
	Description string `json:"description"`
	Address     string `json:"address"`
}

type Response struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	CreatedAt   string `json:"created_at"`
}
