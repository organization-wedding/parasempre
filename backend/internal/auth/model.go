package auth

type SendOTPInput struct {
	Phone string `json:"phone" validate:"required,brphone"`
}

type VerifyOTPInput struct {
	Phone string `json:"phone" validate:"required,brphone"`
	Code  string `json:"code"  validate:"required,len=6"`
}

type TokenResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
	URACF string `json:"uracf"`
}
