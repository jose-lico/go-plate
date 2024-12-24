package user

type RegisterUserPayload struct {
	Email    string `json:"email" validate:"required,min=6,max=254,email" example:"example@email.com"`
	Name     string `json:"name" validate:"required,min=2,max=32" example:"John"`
	Password string `json:"password" validate:"required,min=7,max=16" example:"password"`
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email" example:"example@email.com"`
	Password string `json:"password" validate:"required" example:"password"`
}
