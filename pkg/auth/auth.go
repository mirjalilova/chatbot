package auth

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
)

type GooglePayload struct {
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
	Sub           string
}

func VerifyGoogleIDToken(ctx context.Context, idToken, clientID string) (*GooglePayload, error) {
	payload, err := idtoken.Validate(ctx, idToken, clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid google id token: %w", err)
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)

	return &GooglePayload{
		Email:         email,
		EmailVerified: emailVerified,
		Name:          name,
		Picture:       picture,
		Sub:           payload.Subject,
	}, nil
}
