CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    bucket TEXT NOT NULL,
    object_key TEXT NOT NULL,

    original_filename TEXT,
    content_type TEXT,

    size_bytes BIGINT NOT NULL,
    checksum TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NULL,
    downloaded_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_files_expires_at ON files (expires_at);
CREATE UNIQUE INDEX idx_files_object_key ON files (object_key);
