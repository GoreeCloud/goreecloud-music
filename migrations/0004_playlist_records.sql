CREATE TABLE IF NOT EXISTS playlist_records (
    user_id TEXT NOT NULL,
    playlist_id TEXT NOT NULL,
    revision NUMERIC(20, 0) NOT NULL CHECK (revision >= 0 AND revision <= 18446744073709551615),
    payload BYTEA NOT NULL CHECK (octet_length(payload) BETWEEN 1 AND 65536),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, playlist_id)
);

CREATE INDEX IF NOT EXISTS idx_playlist_records_user_id
    ON playlist_records (user_id, playlist_id);
