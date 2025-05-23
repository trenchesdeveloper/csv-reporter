package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/trenchesdeveloper/csv-reporter/helpers"
)

// createRandomUser is a helper function to create a random user for token tests
func createRandomUser(t *testing.T) User {
	hashedPassword, err := helpers.HashPasswordBase64(helpers.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		HashedPassword: hashedPassword,
		Email:          helpers.RandomEmail(),
	}

	user, err := testStore.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	return user
}

// createRandomRefreshToken is a helper function to create a random refresh token
func createRandomRefreshToken(t *testing.T, user User) RefreshToken {
	// Generate random token and hash it
	hashedToken := helpers.RandomString(32)

	// Set expiry time to 24 hours in the future
	expiryTime := time.Now().Add(24 * time.Hour)

	arg := CreateRefreshTokenParams{
		HashedToken: hashedToken,
		UserID:      user.ID,
		ExpiresAt:   expiryTime,
	}

	token, err := testStore.CreateRefreshToken(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	return token
}

func TestCreateRefreshToken(t *testing.T) {
	user := createRandomUser(t)

	hashedToken := helpers.RandomString(32)
	expiryTime := time.Now().Add(24 * time.Hour)

	arg := CreateRefreshTokenParams{
		HashedToken: hashedToken,
		UserID:      user.ID,
		ExpiresAt:   expiryTime,
	}

	token, err := testStore.CreateRefreshToken(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Verify the token properties
	require.Equal(t, arg.HashedToken, token.HashedToken)
	require.Equal(t, arg.UserID, token.UserID)
	require.WithinDuration(t, arg.ExpiresAt, token.ExpiresAt, time.Second)
	require.NotZero(t, token.CreatedAt)
}

func TestGetRefreshToken(t *testing.T) {
	user := createRandomUser(t)
	token1 := createRandomRefreshToken(t, user)

	// Get the token
	token2, err := testStore.GetRefreshToken(context.Background(), token1.HashedToken)
	require.NoError(t, err)
	require.NotEmpty(t, token2)

	// Verify token properties
	require.Equal(t, token1.HashedToken, token2.HashedToken)
	require.Equal(t, token1.UserID, token2.UserID)
	require.WithinDuration(t, token1.ExpiresAt, token2.ExpiresAt, time.Second)
	require.WithinDuration(t, token1.CreatedAt, token2.CreatedAt, time.Second)

	// Test getting a non-existent token
	nonExistentToken := helpers.RandomString(32)
	_, err = testStore.GetRefreshToken(context.Background(), nonExistentToken)
	require.Error(t, err)
}

func TestDeleteRefreshToken(t *testing.T) {
	user := createRandomUser(t)
	token := createRandomRefreshToken(t, user)

	// Delete the token
	err := testStore.DeleteRefreshToken(context.Background(), token.HashedToken)
	require.NoError(t, err)

	// Try to get the deleted token - should fail
	_, err = testStore.GetRefreshToken(context.Background(), token.HashedToken)
	require.Error(t, err)
}

func TestDeleteUserRefreshTokens(t *testing.T) {
	user := createRandomUser(t)

	// Create multiple tokens for the same user
	token1 := createRandomRefreshToken(t, user)
	token2 := createRandomRefreshToken(t, user)
	token3 := createRandomRefreshToken(t, user)

	// Create a token for a different user to make sure it's not deleted
	otherUser := createRandomUser(t)
	otherToken := createRandomRefreshToken(t, otherUser)

	// Delete all tokens for the first user using the new exec query
	err := testStore.DeleteAllUserRefreshTokens(context.Background(), user.ID)
	require.NoError(t, err)

	// Try to get the deleted tokens - should all error
	_, err = testStore.GetRefreshToken(context.Background(), token1.HashedToken)
	require.Error(t, err)
	_, err = testStore.GetRefreshToken(context.Background(), token2.HashedToken)
	require.Error(t, err)
	_, err = testStore.GetRefreshToken(context.Background(), token3.HashedToken)
	require.Error(t, err)

	// The other user's token should still exist
	tokenCheck, err := testStore.GetRefreshToken(context.Background(), otherToken.HashedToken)
	require.NoError(t, err)
	require.Equal(t, otherToken.HashedToken, tokenCheck.HashedToken)
}

func TestUpdateRefreshTokenExpiry(t *testing.T) {
	user := createRandomUser(t)
	token := createRandomRefreshToken(t, user)

	// Update the token's expiry time
	newExpiryTime := time.Now().Add(48 * time.Hour) // 2 days in the future

	arg := UpdateRefreshTokenExpiryParams{
		HashedToken: token.HashedToken,
		ExpiresAt:   newExpiryTime,
	}

	updatedToken, err := testStore.UpdateRefreshTokenExpiry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedToken)

	// Verify the token was updated
	require.Equal(t, token.HashedToken, updatedToken.HashedToken)
	require.Equal(t, token.UserID, updatedToken.UserID)
	require.WithinDuration(t, newExpiryTime, updatedToken.ExpiresAt, time.Second)

	// Fetch the token again to confirm changes persisted
	fetchedToken, err := testStore.GetRefreshToken(context.Background(), token.HashedToken)
	require.NoError(t, err)
	require.WithinDuration(t, newExpiryTime, fetchedToken.ExpiresAt, time.Second)
}

func TestDeleteExpiredRefreshTokens(t *testing.T) {
	user := createRandomUser(t)

	// Create an expired token
	hashedToken := helpers.RandomString(32)
	expiredTime := time.Now().Add(-24 * time.Hour) // 1 day in the past

	expiredArg := CreateRefreshTokenParams{
		HashedToken: hashedToken,
		UserID:      user.ID,
		ExpiresAt:   expiredTime,
	}

	expiredToken, err := testStore.CreateRefreshToken(context.Background(), expiredArg)
	require.NoError(t, err)

	// Create a valid token
	validToken := createRandomRefreshToken(t, user)

	// Delete expired tokens
	err = testStore.DeleteExpiredRefreshTokens(context.Background())
	require.NoError(t, err)

	// Try to get the expired token - should fail
	_, err = testStore.GetRefreshToken(context.Background(), expiredToken.HashedToken)
	require.Error(t, err)

	// The valid token should still exist - we'll verify by retrieving it
	validCheck, err := testStore.GetRefreshToken(context.Background(), validToken.HashedToken)
	require.NoError(t, err)
	require.Equal(t, validToken.HashedToken, validCheck.HashedToken)
}
