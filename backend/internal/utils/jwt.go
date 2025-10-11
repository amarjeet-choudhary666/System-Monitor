package utils

import (
	"errors"
	"time"

	"github.com/amarjeet-choudhary666/CodeXray/backend/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

var cfg *config.Config

func InitConfig(c *config.Config) {
	cfg = c
}

func GenerateToken(userId uint, username string) (string, string, error) {
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":   userId,
		"username": username,
		"exp":      time.Now().Add(35 * time.Minute).Unix(),
	}).SignedString([]byte(cfg.Auth.JWTSecret))

	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":   userId,
		"username": username,
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(),
	}).SignedString([]byte(cfg.Auth.JWTSecret))

	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(cfg.Auth.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func RefreshToken(refreshTokenString string) (string, error) {
	claims, err := ValidateToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	userId, ok := claims["userId"].(float64)
	if !ok {
		return "", errors.New("invalid user ID in token")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New("invalid username in token")
	}

	accessToken, _, err := GenerateToken(uint(userId), username)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func GetUserIDFromToken(tokenString string) (uint, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	userId, ok := claims["userId"].(float64)
	if !ok {
		return 0, errors.New("invalid user ID in token")
	}

	return uint(userId), nil
}

func GetUsernameFromToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New("invalid username in token")
	}

	return username, nil
}
