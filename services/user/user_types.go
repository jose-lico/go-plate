package user

type RegisterUserPayload struct {
	Email    string `json:"email" validate:"required,min=6,max=254,email"`
	Password string `json:"password" validate:"required,min=7,max=16"`
}
