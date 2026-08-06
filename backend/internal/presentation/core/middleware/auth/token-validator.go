package auth

import (
	"crypto/hmac"
	"net/http"

	"dnd_schedule/internal/config"
	"github.com/gin-gonic/gin"
)

type ClientAuthVerifier struct {
	expectedKey string
}

func NewClientAuthVerifier(cfg config.ConfigProvider) *ClientAuthVerifier {
	return &ClientAuthVerifier{expectedKey: cfg.GetAPIKey()}
}

func (v *ClientAuthVerifier) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Api-Key")
		if !hmac.Equal([]byte(key), []byte(v.expectedKey)) { // constant-time сравнение
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
