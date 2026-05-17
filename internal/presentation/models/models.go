package models

type AuthRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	IsNewUser bool   `json:"-"`
}

type AuthResponse struct {
	Token string `json:"token"`
}
