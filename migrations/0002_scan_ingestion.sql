ALTER TABLE tracks
    ADD COLUMN artist_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN genre TEXT NOT NULL DEFAULT '',
    ADD COLUMN release_year INTEGER;

ALTER TABLE track_files
    ADD COLUMN last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE UNIQUE INDEX artists_lower_name_unique_idx ON artists ((lower(name)));
CREATE INDEX albums_library_title_idx ON albums (library_id, title);
CREATE INDEX track_files_last_seen_at_idx ON track_files (last_seen_at);
