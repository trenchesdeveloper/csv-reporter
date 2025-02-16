-- name: CreateUser :one
INSERT INTO users (email, hashed_password) VALUES ($1, $2) RETURNING *;

-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: FindUserById :one
SELECT * FROM users WHERE id = $1;

