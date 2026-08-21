BEGIN;

CREATE TABLE users (
    id UUID PRIMARY KEY,
    display_name TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('user', 'admin')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE libraries (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    root_path TEXT NOT NULL UNIQUE,
    visibility TEXT NOT NULL CHECK (visibility IN ('private', 'shared')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE library_memberships (
    library_id UUID NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    can_read BOOLEAN NOT NULL DEFAULT true,
    can_manage BOOLEAN NOT NULL DEFAULT false,
    PRIMARY KEY (library_id, user_id)
);

CREATE TABLE artists (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    sort_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE albums (
    id UUID PRIMARY KEY,
    library_id UUID NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    album_artist_id UUID REFERENCES artists(id) ON DELETE SET NULL,
    release_year INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE tracks (
    id UUID PRIMARY KEY,
    library_id UUID NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
    album_id UUID REFERENCES albums(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    track_number INTEGER NOT NULL DEFAULT 0,
    disc_number INTEGER NOT NULL DEFAULT 0,
    duration_ms BIGINT NOT NULL DEFAULT 0 CHECK (duration_ms >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE track_files (
    id UUID PRIMARY KEY,
    track_id UUID NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    path TEXT NOT NULL UNIQUE,
    codec TEXT NOT NULL DEFAULT '',
    bitrate INTEGER NOT NULL DEFAULT 0 CHECK (bitrate >= 0),
    bit_depth INTEGER NOT NULL DEFAULT 0 CHECK (bit_depth >= 0),
    sample_rate INTEGER NOT NULL DEFAULT 0 CHECK (sample_rate >= 0),
    channels INTEGER NOT NULL DEFAULT 0 CHECK (channels >= 0),
    size_bytes BIGINT NOT NULL DEFAULT 0 CHECK (size_bytes >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX albums_library_id_idx ON albums(library_id);
CREATE INDEX tracks_library_id_idx ON tracks(library_id);
CREATE INDEX tracks_album_id_idx ON tracks(album_id);
CREATE INDEX track_files_track_id_idx ON track_files(track_id);

COMMIT;
