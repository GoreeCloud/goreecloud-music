CREATE TABLE album_artwork (
    album_id uuid PRIMARY KEY REFERENCES albums(id) ON DELETE CASCADE,
    source_type text NOT NULL CHECK (source_type IN ('sidecar', 'embedded')),
    source_path text NOT NULL,
    mime_type text NOT NULL DEFAULT '',
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX album_artwork_source_path_idx ON album_artwork (source_path);
