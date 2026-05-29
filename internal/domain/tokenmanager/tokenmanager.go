package tokenmanager

import (
	"fmt"
	"time"

	"dnd_schedule/internal/domain/models"

	"github.com/golang-jwt/jwt/v5"
)

type Manager interface {
	IssueJWT(user *models.User) (string, error)
	VerifyJWT(string) (*Claims, error)
}

type Claims struct {
	UserID models.UserID `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTManager struct {
}

func New() *JWTManager {
	return &JWTManager{}
}

func (manager *JWTManager) IssueJWT(user *models.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "gophkeeper",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// TODO change to production key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("error signing token: %v", err)
	}
	return tokenString, nil
}

func (manager *JWTManager) VerifyJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Method.Alg())
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

const secretKey = "SECRET_KEY"
