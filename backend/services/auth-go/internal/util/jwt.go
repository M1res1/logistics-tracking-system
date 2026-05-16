package util

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtUtil struct {
	Secret            []byte
	AccessExpiration  time.Duration
	RefreshExpiration time.Duration
}

func NewJwtUtil(secret string, accessExp, refreshExp int64) *JwtUtil {
	return &JwtUtil{
		Secret:            []byte(secret),
		AccessExpiration:  time.Duration(accessExp) * time.Millisecond,
		RefreshExpiration: time.Duration(refreshExp) * time.Millisecond,
	}
}

func (j *JwtUtil) GenerateToken(email string, userType string, userId uint) (string, error) {
	claims := jwt.MapClaims{
		"sub":       email,
		"user_type": userType,
		"user_id":   userId,
		"iat":       time.Now().Unix(),
		"exp":       time.Now().Add(j.AccessExpiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.Secret)
}

func (j *JwtUtil) GenerateRefreshToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"sub": email,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(j.RefreshExpiration).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.Secret)
}

func (j *JwtUtil) ExtractUsername(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return j.Secret, nil
	})

	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", err
	}

	return claims["sub"].(string), nil
}

func (j *JwtUtil) ExtractExpiration(tokenString string) (time.Time, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return j.Secret, nil
	})

	if err != nil {
		return time.Time{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return time.Time{}, err
	}

	exp := int64(claims["exp"].(float64))
	return time.Unix(exp, 0), nil
}
