CREATE TABLE IF NOT EXISTS episodes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    episode_number INTEGER NOT NULL UNIQUE,
    programme_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    web_url TEXT NOT NULL,
    image_url TEXT NOT NULL DEFAULT '',
    since DATETIME NOT NULL,
    till DATETIME NOT NULL,
    recording_started DATETIME NOT NULL,
    duration_seconds INTEGER NOT NULL,
    year INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
