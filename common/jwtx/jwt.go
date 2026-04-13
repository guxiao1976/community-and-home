package jwtx

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID             int64  `json:"user_id"`
	Username           string `json:"username"`
	AdministrativeCode string `json:"administrative_code"`
	jwt.RegisteredClaims
}

func GenerateToken(secret string, userID int64, username, adminCode string, expireDuration time.Duration) (string, error) {
	claims := Claims{
		UserID:             userID,
		Username:           username,
		AdministrativeCode: adminCode,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrSignatureInvalid
}
