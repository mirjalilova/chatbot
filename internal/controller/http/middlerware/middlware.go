package middleware

import (
	"net/http"
	"time"

	"chatbot/internal/controller/http/token"
	"chatbot/internal/usecase"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

func Identity(userRepo usecase.UserRepoI) gin.HandlerFunc {
	return func(c *gin.Context) {

		cookie, err := c.Request.Cookie("access_token")
		if err == nil && cookie.Value != "" {
			claims, err := token.ExtractClaim(cookie.Value)
			if err == nil {
				id, _ := claims["id"].(string)
				role, _ := claims["role"].(string)

				c.Set("id", id)
				c.Set("role", role)
				c.Next()
				return
			}
		}

		ip := c.ClientIP()
		ua := c.Request.UserAgent()

		guestID, err := userRepo.GetGuestByIPAndUA(c.Request.Context(), ip, ua)
		if err != nil {
			guestID, err = userRepo.CreateGuest(c.Request.Context(), ip, ua)
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": "failed to create guest"})
				return
			}
		}

		tokens := token.GenerateJWTToken(guestID, "guest")
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "access_token",
			Value:    tokens.AccessToken,
			Path:     "/",
			Domain:   "ccenter.uz",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			Expires:  time.Now().Add(365 * 24 * time.Hour),
		})

		c.Set("id", guestID)
		c.Set("role", "guest")
		c.Next()
	}
}

func Authorize(enforcer *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		role := c.GetString("role")
		path := c.FullPath()
		method := c.Request.Method

		allowed, err := enforcer.Enforce(role, path, method)
		if err != nil || !allowed {
			c.AbortWithStatusJSON(403, gin.H{"error": "permission denied"})
			return
		}

		c.Next()
	}
}

