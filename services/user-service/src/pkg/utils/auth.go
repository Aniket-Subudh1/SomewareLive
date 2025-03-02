package utils

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/your-username/slido-clone/user-service/config"
)

var (
	// ErrNoAuthHeader is returned when no authorization header is present
	ErrNoAuthHeader = errors.New("no authorization header present")
	// ErrInvalidAuthHeader is returned when the authorization header is invalid
	ErrInvalidAuthHeader = errors.New("invalid authorization header format")
	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired is returned when the token is expired
	ErrTokenExpired = errors.New("token expired")
	// ErrInvalidTokenIssuer is returned when the token issuer is invalid
	ErrInvalidTokenIssuer = errors.New("invalid token issuer")
	// ErrInvalidTokenType is returned when the token type is invalid
	ErrInvalidTokenType = errors.New("invalid token type")
)

// TokenClaims represents the JWT claims
type TokenClaims struct {
	jwt.RegisteredClaims
	Email string   `json:"email,omitempty"`
	Role  string   `json:"role,omitempty"`
	Type  string   `json:"type,omitempty"`
	Roles []string `json:"roles,omitempty"`
}

// ExtractToken extracts the token from the authorization header
func ExtractToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrNoAuthHeader
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", ErrInvalidAuthHeader
	}

	return parts[1], nil
}

// ValidateToken validates a JWT token
func ValidateToken(tokenString string, cfg *config.JWTConfig) (*TokenClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.Secret), nil
	})

	// Handle parsing errors
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Extract claims
	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate token type
	if claims.Type != "access" {
		return nil, ErrInvalidTokenType
	}

	// Validate issuer if specified
	if cfg.Issuer != "" && claims.Issuer != cfg.Issuer {
		return nil, ErrInvalidTokenIssuer
	}

	// Check expiration
	if claims.ExpiresAt != nil {
		expiresAt, err := claims.ExpiresAt.Time()
		if err != nil || expiresAt.Before(time.Now()) {
			return nil, ErrTokenExpired
		}
	}

	return claims, nil
}
