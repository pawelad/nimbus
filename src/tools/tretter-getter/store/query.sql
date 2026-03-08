-- name: UpsertEpisode :one
INSERT INTO episodes (
    episode_number,
    programme_id,
    title,
    description,
    web_url,
    image_url,
    since,
    till,
    recording_started,
    duration_seconds,
    year,
    status
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
ON CONFLICT(episode_number) DO UPDATE SET
    programme_id = excluded.programme_id,
    title = excluded.title,
    description = excluded.description,
    web_url = excluded.web_url,
    image_url = excluded.image_url,
    since = excluded.since,
    till = excluded.till,
    duration_seconds = excluded.duration_seconds,
    year = excluded.year,
    status = CASE WHEN episodes.status = 'recorded' THEN 'recorded' ELSE excluded.status END
RETURNING *;

-- name: UpdateEpisodeStatus :exec
UPDATE episodes
SET status = ?
WHERE episode_number = ?;

-- name: GetEpisodes :many
SELECT * FROM episodes
ORDER BY episode_number DESC
LIMIT ? OFFSET ?;

-- name: GetEpisode :one
SELECT * FROM episodes
WHERE episode_number = ? LIMIT 1;

-- name: GetRecordedCount :one
SELECT COUNT(*) FROM episodes;

-- name: GetActiveRecordings :many
SELECT * FROM episodes
WHERE status = 'recording'
ORDER BY episode_number ASC;
