package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserId uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

type Interface interface {
	GenerateToken(userId uuid.UUID, role string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
}

type JsonWebToken struct {
	SecretKey   []byte
	ExpiredTime time.Duration
}

func NewJWTToken(secretKey string, expiredTime time.Duration) Interface {
  return &JsonWebToken{
    SecretKey: []byte(secretKey),
    ExpiredTime: expiredTime,
  }
}

func (j *JsonWebToken) GenerateToken(userId uuid.UUID, role string) (string, error) {
  claims := Claims{
    UserId: userId,
    Role: role,
    RegisteredClaims: jwt.RegisteredClaims{
      ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.ExpiredTime)),
      IssuedAt: jwt.NewNumericDate(time.Now()),
    },
  }

  token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
  return token.SignedString(j.SecretKey)
}

func (j *JsonWebToken) ValidateToken(tokenString string) (*Claims, error) {
  token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func (token *jwt.Token) (interface{}, error) {
    return j.SecretKey, nil
  })
  
  if token == nil {
    return nil, errors.New("Token is nil")
  }
  
  if err != nil {
    return nil, err
  }

  claims, ok := token.Claims.(*Claims)
  if !ok || !token.Valid {
    return nil, errors.New("Invalid token")
  }

  return claims, nil
}
