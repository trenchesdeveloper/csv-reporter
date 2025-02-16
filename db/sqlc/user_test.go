package db

import (
	"context"
	"testing"

	"github.com/trenchesdeveloper/csv-reporter/helpers"

	"github.com/stretchr/testify/require"
)

func TestCreateRandomUser(t *testing.T) {
	hashedPassword, err := helpers.HashPasswordBase64(helpers.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		HashedPassword: hashedPassword,
		Email:          helpers.RandomEmail(),
	}

	user, err := testStore.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.Email, user.Email)
	require.NotZero(t, user.CreatedAt)

}
