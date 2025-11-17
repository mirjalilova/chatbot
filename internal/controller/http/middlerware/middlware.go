package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"chatbot/internal/controller/http/token"
	"chatbot/internal/usecase"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/exp/slog"
)

const (
	key          = "vctr"
	unauthorized = "unauthorized"
)


func NewAuth(enforcer *casbin.Enforcer, userRepo usecase.UserRepoI) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		allow, err := CheckPermission(ctx.Writer, ctx.FullPath(), ctx.Request, enforcer, userRepo)

		if err != nil {
			slog.Error("Error checking permission: %v", err)
			if ve, ok := err.(*jwt.ValidationError); ok && ve.Errors == jwt.ValidationErrorExpired {
				RequireRefresh(ctx)
			} else {
				RequirePermission(ctx)
			}
			return
		}

		if !allow {
			RequirePermission(ctx)
			return
		}

		claims, err := ExtractToken(ctx.Writer, ctx.Request, userRepo)
		if err != nil {
			slog.Error("Error extracting token: %v", err)
			InvalidToken(ctx)
			return
		}

		var id string
		var ok bool
		id, ok = claims["id"].(string)
		if !ok {
			slog.Warn("id not found in claims")
		}
		ctx.Set("id", id)
		ctx.Set("claims", claims)

		ctx.Next()
	}
}

func OptionalAuth(userRepo usecase.UserRepoI) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		jwtToken := ctx.Request.Header.Get("Authorization")
		if jwtToken != "" {
			claims, err := ExtractToken(ctx.Writer, ctx.Request, userRepo)
			if err != nil {
				if strings.Contains(err.Error(), "token expired") {
					slog.Warn("Token expired")
					ctx.AbortWithStatusJSON(401, gin.H{"error": "Token expired"})
					return
				}

				slog.Error("Invalid token: %v", err)
				ctx.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
				return
			}

			slog.Info("Extracted optional claims: %v", claims)
			ctx.Set("claims", claims)
		}

		ctx.Next()
	}
}

func ExtractToken(w http.ResponseWriter, r *http.Request, userRepo usecase.UserRepoI) (jwt.MapClaims, error) {
	// jwtToken := r.Header.Get("Authorization")
	// if jwtToken != "" {
	// 	if strings.Contains(jwtToken, "Basic") {
	// 		return nil, fmt.Errorf("invalid token format")
	// 	}
	// 	tokenString := strings.TrimSpace(strings.TrimPrefix(jwtToken, "Bearer "))
	// 	return token.ExtractClaim(tokenString)
	// }

    cookie, err := r.Cookie("access_token")
    if err == nil && cookie.Value != "" {
        return token.ExtractClaim(cookie.Value)
    }

	guestID, err := userRepo.CreateGuest(r.Context())
    if err != nil {
        return nil, fmt.Errorf("guest user yaratishda xatolik: %w", err)
    }
	slog.Info("Created new guest user with ID: %s", guestID)

	tokens := token.GenerateJWTToken(guestID, "guest")

	access := &http.Cookie{
        Name:     "access_token",
        Value:    tokens.AccessToken,
        Path:     "/",
		Domain:   "ccenter.uz",
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteNoneMode,
        Expires:  time.Now().Add(1000 * 24 * time.Hour),
    }
	http.SetCookie(w, access)

	return token.ExtractClaim(cookie.Value)
}


func GetRole(w http.ResponseWriter, r *http.Request, userrepo usecase.UserRepoI) (string, error) {
	claims, err := ExtractToken(w, r, userrepo)
	if err != nil {
		return unauthorized, err
	}

	role, ok := claims["role"].(string)
	if !ok {
		return unauthorized, errors.New("role claim not found")
	}
	return role, nil
}

func CheckPermission(w http.ResponseWriter, path string, r *http.Request, enforcer *casbin.Enforcer, userrepo usecase.UserRepoI) (bool, error) {
	role, err := GetRole(w, r, userrepo)
	if err != nil {
		slog.Error("Error getting role from token", err)
		return false, err
	}

	allowed, err := enforcer.Enforce(role, path, r.Method)
	if err != nil {
		slog.Error("Error during Casbin enforce", err)
		return false, err
	}

	return allowed, nil
}

func InvalidToken(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error": "Invalid token!",
	})
}

func RequirePermission(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error": "Permission denied",
	})
}

func RequireRefresh(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"error": "Access token expired",
	})
}
