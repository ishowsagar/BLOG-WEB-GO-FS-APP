package controller


// register payload req type
type RegisterRequest struct {
	Name string `json:"name" binding:"required,min=3"`
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required,min=4"`
	Bio string `json:"bio"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

//  login payload req type
type LoginRequest struct {
	Email string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}