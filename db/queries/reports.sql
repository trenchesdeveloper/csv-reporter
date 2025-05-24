-- name: CreateReport :one
INSERT INTO reports (
    user_id,
    report_type,
    output_file_path,
    download_url,
    download_expires_at,
    error_message,
    started_at,
    failed_at,
    completed_at
) VALUES (
             $1,  -- user_id
             $2,  -- report_type
             $3,  -- output_file_path
             $4,  -- download_url
             $5,  -- download_expires_at
             $6,  -- error_message
             $7,  -- started_at
             $8,  -- failed_at
             $9   -- completed_at
         )
RETURNING *;

-- name: GetReport :one
SELECT
    user_id,
    id,
    report_type,
    output_file_path,
    download_url,
    download_expires_at,
    error_message,
    created_at,
    started_at,
    failed_at,
    completed_at
FROM reports
WHERE
    user_id = $1  -- UUID
  AND id      = $2; -- UUID

-- name: DeleteReport :exec
DELETE FROM reports
WHERE
    user_id = $1  -- UUID
  AND id      = $2; -- UUID

-- name: UpdateReport :one
UPDATE reports
SET
    output_file_path = COALESCE(sqlc.narg('output_file_path'), output_file_path),
    download_url = COALESCE(sqlc.narg('download_url'), download_url),
    download_expires_at = COALESCE(sqlc.narg('download_expires_at'), download_expires_at),
    error_message = COALESCE(sqlc.narg('error_message'), error_message),
    started_at = COALESCE(sqlc.narg('started_at'), started_at),
    failed_at = COALESCE(sqlc.narg('failed_at'), failed_at),
    completed_at = COALESCE(sqlc.narg('completed_at'), completed_at)
WHERE
    user_id = sqlc.arg('user_id')  -- UUID
  AND id = sqlc.arg('id') -- UUID
RETURNING
    user_id,
    id,
    report_type,
    output_file_path,
    download_url,
    download_expires_at,
    error_message,
    created_at,
    started_at,
    failed_at,
    completed_at;