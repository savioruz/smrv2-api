package model

type UsersResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Major    string `json:"major"`
	Semester string `json:"semester"`
}

type UsersRegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UsersLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UsersRegisterResponse struct {
	Email string `json:"email"`
}

type UsersLoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UsersVerifyEmailRequest struct {
	Token string `param:"token" validate:"required"`
}

type UsersRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,jwt"`
}

type UserRefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type PortalVerificationMessage struct {
	UserID   string `json:"user_id"`
	NIM      string `json:"nim"`
	Password string `json:"password"`
}

type UserResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type UserResetPassword struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required"`
}
