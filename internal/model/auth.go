package model

type AuthResponse struct {
	Auth interface{} `json:"auth"`
} //	@name	auth

type Auth struct {
	OwnerID      int64  `db:"owner_id" redis:"owner_id"`
	Jti          string `db:"jti" redis:"jti"`
	Valid        bool   `redis:"valid"`
	SID          string `db:"sid"`
	Email        string `db:"email" redis:"email"`
	AccessToken  string // access token do not saved at db
	RefreshToken string `db:"refresh_token"`
	ExpiredAt    string `db:"expired_at"` // refresh token expired at
	CreatedAt    string `db:"created_at"`
	UpdatedAt    string `db:"updated_at"`
}

type AuthRenewAccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiredAt   string `json:"expired_at"`
}

type AuthLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,alphanum,min=8,omitempty"`
}

type AuthLogoutRequest struct {
	SID      string `json:"-"`
	Password string `json:"password" validate:"required,alphanum,min=8,omitempty"`
}

type AuthSessionResponse struct {
	SID     string `json:"-"`
	OwnerID int64  `json:"owner_id"`
	Valid   bool   `json:"valid"`
	Jti     string `json:"jti"`
}

type AuthLoginResponse struct {
	SID          string `json:"-"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}
