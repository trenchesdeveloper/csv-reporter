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
    output_file_path = $2,  -- VARCHAR
    download_url = $3,  -- VARCHAR
    download_expires_at = $4,  -- TIMESTAMPTZ
    error_message = $5,  -- VARCHAR
    started_at = $6,  -- TIMESTAMPTZ
    failed_at = $7,  -- TIMESTAMPTZ
    completed_at = $8   -- TIMESTAMPTZ
WHERE
    user_id = $1  -- UUID
  AND id      = $9 -- UUID
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