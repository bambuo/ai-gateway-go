package admin

import (
	"fmt"
	"time"

	"ai/gateway/internal/database"

	"github.com/golang-jwt/jwt/v5"
)

const tokenExpiry = 24 * time.Hour

type JWTClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func generateJWT(username string) (string, error) {
	secret, err := database.GetJWTSecret()
	if err != nil {
		return "", fmt.Errorf("获取 JWT 密钥: %w", err)
	}
	claims := JWTClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ai-gateway",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("JWT 签名: %w", err)
	}
	return tokenString, nil
}

func validateJWT(tokenString string) (*JWTClaims, error) {
	secret, err := database.GetJWTSecret()
	if err != nil {
		return nil, fmt.Errorf("获取 JWT 密钥: %w", err)
	}
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名算法: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("JWT 解析: %w", err)
	}
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("JWT 无效")
	}
	return claims, nil
}
