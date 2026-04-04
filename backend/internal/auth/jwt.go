package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

type Claims struct {
	UserID int64  `json:"user_id"`
	URACF  string `json:"uracf"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret []byte
	expiry time.Duration
}

func NewJWTService(secret string, expiry time.Duration) *JWTService {
	return &JWTService{secret: []byte(secret), expiry: expiry}
}

func (s *JWTService) Generate(userID int64, uracf, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		URACF:  uracf,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperror.Unauthorized("invalid signing method")
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, apperror.Unauthorized("invalid or expired token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, apperror.Unauthorized("invalid token claims")
	}

	return claims, nil
}
