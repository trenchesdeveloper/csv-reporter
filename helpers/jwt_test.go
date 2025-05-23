package helpers

import (
	"github.com/google/uuid"
	"github.com/trenchesdeveloper/csv-reporter/config"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	conf, err := config.LoadConfig("../..")
	require.NoError(t, err)
	jwtManager := NewJwtManager(conf)
	assert.NotNil(t, jwtManager)

	userID := uuid.New()
	tokenPairs, err := jwtManager.GenerateTokenPairs(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenPairs)
	assert.NotEmpty(t, tokenPairs.AccessToken)
	assert.NotEmpty(t, tokenPairs.RefreshToken)
	assert.NotEqual(t, tokenPairs.AccessToken, tokenPairs.RefreshToken)

	// Validate the access token
	accessToken, err := jwt.ParseWithClaims(tokenPairs.AccessToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(conf.JWT_SECRET), nil
	})
	require.NoError(t, err)
	assert.NotNil(t, accessToken)
	claims, ok := accessToken.Claims.(*CustomClaims)
	require.True(t, ok)
	assert.True(t, accessToken.Valid)
	assert.Equal(t, claims.TokenType, "access")
	assert.Equal(t, claims.Issuer, conf.AppName)
	assert.Equal(t, claims.Subject, userID.String())
	assert.NotZero(t, claims.IssuedAt)
	assert.NotZero(t, claims.ExpiresAt)
	assert.WithinDuration(t, claims.ExpiresAt.Time, time.Now().Add(time.Minute*15), time.Second)
	assert.WithinDuration(t, claims.IssuedAt.Time, time.Now(), time.Second)
	// Validate the refresh token
	refreshToken, err := jwt.ParseWithClaims(tokenPairs.RefreshToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(conf.JWT_SECRET), nil

	})
	require.NoError(t, err)
	assert.NotNil(t, refreshToken)
	claims, ok = refreshToken.Claims.(*CustomClaims)
	require.True(t, ok)
	assert.True(t, refreshToken.Valid)
	assert.Equal(t, claims.TokenType, "refresh")
	assert.Equal(t, claims.Issuer, conf.AppName)
	assert.Equal(t, claims.Subject, userID.String())
	assert.NotZero(t, claims.IssuedAt)
	assert.NotZero(t, claims.ExpiresAt)
	assert.WithinDuration(t, claims.ExpiresAt.Time, time.Now().Add(time.Hour*24*7), time.Second)
	assert.WithinDuration(t, claims.IssuedAt.Time, time.Now(), time.Second)
	// Check if the access token is not expired
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	// Check if the refresh token is not expired
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	// Check if the access token is not equal to the refresh token
	assert.NotEqual(t, tokenPairs.AccessToken, tokenPairs.RefreshToken)
	// Check if the access token is not expired
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
	// Check if the refresh token is not expired
	assert.True(t, claims.ExpiresAt.Time.After(time.Now()))
}
func TestValidateToken(t *testing.T) {
	conf, err := config.LoadConfig("../..")
	require.NoError(t, err)
	jwtManager := NewJwtManager(conf)
	assert.NotNil(t, jwtManager)

	userID := uuid.New()
	tokenPairs, err := jwtManager.GenerateTokenPairs(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenPairs)
	assert.NotEmpty(t, tokenPairs.AccessToken)
	assert.NotEmpty(t, tokenPairs.RefreshToken)
	assert.NotEqual(t, tokenPairs.AccessToken, tokenPairs.RefreshToken)

	// Validate the access token
	accessToken, err := jwtManager.ValidateToken(tokenPairs.AccessToken)
	require.NoError(t, err)
	assert.NotNil(t, accessToken)
	claims, ok := accessToken.Claims.(*CustomClaims)
	require.True(t, ok)
	assert.True(t, accessToken.Valid)
	assert.Equal(t, claims.TokenType, "access")
}
func TestValidateTokenExpired(t *testing.T) {
	conf, err := config.LoadConfig("../..")
	require.NoError(t, err)
	jwtManager := NewJwtManager(conf)
	assert.NotNil(t, jwtManager)

	userID := uuid.New()
	tokenPairs, err := jwtManager.GenerateTokenPairs(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenPairs)
	assert.NotEmpty(t, tokenPairs.AccessToken)
	assert.NotEmpty(t, tokenPairs.RefreshToken)
	assert.NotEqual(t, tokenPairs.AccessToken, tokenPairs.RefreshToken)

	// Validate the access token
	accessToken, err := jwtManager.ValidateToken(tokenPairs.AccessToken)
	require.NoError(t, err)
	assert.NotNil(t, accessToken)
	claims, ok := accessToken.Claims.(*CustomClaims)
	require.True(t, ok)
	assert.True(t, accessToken.Valid)
	assert.Equal(t, claims.TokenType, "access")
}
func TestValidateTokenInvalid(t *testing.T) {
	conf, err := config.LoadConfig("../..")
	require.NoError(t, err)
	jwtManager := NewJwtManager(conf)
	assert.NotNil(t, jwtManager)

	// Validate the access token
	accessToken, err := jwtManager.ValidateToken("invalid_token")
	require.Error(t, err)
	assert.Nil(t, accessToken)
}
