package auth

import (
	"net/http"
	"strings"

	"dnd_schedule/internal/config"
	authenticateservice "dnd_schedule/internal/domain/service/authenticate-service"

	"github.com/gin-gonic/gin"
)

type BearerValidator struct {
	cfg         config.ConfigProvider
	authService authenticateservice.AuthService
}

func NewBearerValidator(cfg config.ConfigProvider, authService authenticateservice.AuthService) *BearerValidator {
	return &BearerValidator{cfg, authService}
}

func (b BearerValidator) VerifyJWT(c *gin.Context) {
	tokenParts := strings.Split(c.Request.Header.Get("Authorization"), " ")
	if len(tokenParts) != 2 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token not found"})
		return
	}
	token := tokenParts[1]

	claims, err := b.authService.VerifyJWT(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	username, _ := claims.GetSubject()
	issuer, _ := claims.GetIssuer()
	userID := claims.UserID

	c.Set("username", username)
	c.Set("issuer", issuer)
	c.Set("userID", userID)

	c.Next()
}
