package authenticate_service

import (
	"context"
	"fmt"
	"time"

	_ "dnd_schedule/internal/config"
	"dnd_schedule/internal/domain/models"
	requests "dnd_schedule/internal/presentation/models"
	"dnd_schedule/internal/repository/datasources/authenticate"
	"github.com/golang-jwt/jwt/v5"

	"github.com/xianghuzhao/kdfcrypt"
)

const secretKey = "SECRET_KEY"

type AuthService interface {
	VerifyJWT(string) (*Claims, error)
	RegisterUser(ctx context.Context, object requests.AuthRequest) (string, error)
}

type AuthServiceImpl struct {
	manager Manager
	db      authenticate.IDatasource
	hash    HashValidator
}

func NewAuthService(db authenticate.IDatasource) AuthService {
	manager := New()
	return &AuthServiceImpl{manager: manager, db: db, hash: NewHashValidator()}
}

func (a *AuthServiceImpl) RegisterUser(ctx context.Context, object requests.AuthRequest) (string, error) {
	user := &models.User{Username: object.Username, Hash: a.hash.CalculateHash(object.Password)}
	userAdded, err := a.db.AddUser(ctx, user)
	if err != nil {
		return "", fmt.Errorf("add user error: %w", err)
	}

	token, err := a.manager.IssueJWT(userAdded)
	if err != nil {
		return "", fmt.Errorf("issue token: %w", err)
	}
	return token, nil

}

func (a *AuthServiceImpl) VerifyJWT(token string) (*Claims, error) {
	return a.manager.VerifyJWT(token)
}

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

type HashValidator interface {
	CalculateHash(password string) string
	Validate(password string, encoded string) bool
}

type HashValidatorImpl struct {
}

func NewHashValidator() HashValidator {
	return &HashValidatorImpl{}
}

func (h *HashValidatorImpl) CalculateHash(password string) string {
	encoded, _ := kdfcrypt.Encode(password, &kdfcrypt.Option{
		Algorithm:        "argon2id",
		Param:            "m=65536,t=1,p=4",
		RandomSaltLength: 16,
		HashLength:       32,
	})
	return encoded
}

func (h *HashValidatorImpl) Validate(password string, encoded string) bool {
	ok, err := kdfcrypt.Verify(password, encoded)
	if err != nil || !ok {
		return false
	}
	return true
}
