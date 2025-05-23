package helpers

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/trenchesdeveloper/csv-reporter/config"
	"time"
)

type JwtManager struct {
	config *config.AppConfig
}

func NewJwtManager(config *config.AppConfig) *JwtManager {
	return &JwtManager{
		config: config,
	}
}

type TokenPairs struct {
	AccessToken  string
	RefreshToken string
}

type CustomClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func (j JwtManager) GenerateTokenPairs(userID uuid.UUID) (*TokenPairs, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.AppName,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
	})

	key := []byte(j.config.JWT_SECRET)
	accessToken, err := jwtToken.SignedString(key)

	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.config.AppName,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	})
	key = []byte(j.config.JWT_SECRET)
	refreshTokenString, err := refreshToken.SignedString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return &TokenPairs{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
	}, nil
}
func (j JwtManager) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.config.JWT_SECRET), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	return token, nil
}

func (j JwtManager) IsAccessToken(token *jwt.Token) bool {
	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return false
	}
	return claims.TokenType == "access"
}
