package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const accessTokenTTL = 15 * time.Minute

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

var secret []byte

func init() {
	s := os.Getenv("JWT_SECRET")
	if s == "" {
		s = "darshan-dev-secret-change-in-production"
	}
	secret = []byte(s)
}

func GenerateToken(userID int64, email, role string) (string, error) {
	return GenerateAccessToken(userID, email, role)
}

func GenerateAccessToken(userID int64, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}
