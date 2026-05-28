package authenticate_service

import (
	"context"
	"errors"
	"fmt"

	_ "dnd_schedule/internal/config"
	"dnd_schedule/internal/domain/hash_manager"
	"dnd_schedule/internal/domain/models"
	"dnd_schedule/internal/domain/tokenmanager"
	"dnd_schedule/internal/presentation/features/authenticate/authenticate-requests"
	"dnd_schedule/internal/repository/datasources/authenticate"
)

var InvalidPassword = errors.New("invalid password")
var UserNotFound = errors.New("user not found")

type AuthService interface {
	VerifyJWT(string) (*tokenmanager.Claims, error)
	RegisterUser(ctx context.Context, object authenticate_requests.AuthRequest) (string, error)
	GetToken(ctx context.Context, object authenticate_requests.AuthRequest) (string, error)
	GetUser(ctx context.Context, object authenticate_requests.AuthRequest) (*models.User, error)
	GenerateJWT(ctx context.Context, object authenticate_requests.AuthRequest) (string, error)
}

type AuthServiceImpl struct {
	manager tokenmanager.Manager
	db      authenticate.IDatasource
	hash    hashmanager.HashValidator
}

func NewAuthService(db authenticate.IDatasource) AuthService {
	manager := tokenmanager.New()
	return &AuthServiceImpl{manager: manager, db: db, hash: hashmanager.NewHashValidator()}
}

func (a *AuthServiceImpl) GetUser(ctx context.Context, object authenticate_requests.AuthRequest) (*models.User, error) {
	user, err := a.db.GetUser(ctx, object.User.Username)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (a *AuthServiceImpl) GenerateJWT(ctx context.Context, object authenticate_requests.AuthRequest) (string, error) {
	token, err := a.manager.IssueJWT(&object.User)
	if err != nil {
		return "", fmt.Errorf("generate jwt: %w", err)
	}
	return token, nil
}

func (a *AuthServiceImpl) GetToken(ctx context.Context, object authenticate_requests.AuthRequest) (string, error) {
	user, err := a.GetUser(ctx, object)
	if err != nil {
		return "", fmt.Errorf("get user: %w", err)
	}
	if user == nil {
		return "", UserNotFound
	}
	if user.Hash != a.hash.CalculateHash(object.User.Password) {
		return "", InvalidPassword
	}
	token, err := a.GenerateJWT(ctx, object)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return token, nil
}

func (a *AuthServiceImpl) RegisterUser(ctx context.Context, object authenticate_requests.AuthRequest) (string, error) {
	user := &models.User{Username: object.User.Username, Hash: a.hash.CalculateHash(object.User.Password)}
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

func (a *AuthServiceImpl) VerifyJWT(token string) (*tokenmanager.Claims, error) {
	return a.manager.VerifyJWT(token)
}
