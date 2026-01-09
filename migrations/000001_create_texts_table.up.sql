CREATE TABLE IF NOT EXISTS texts (
    id bigserial PRIMARY KEY,
    name TEXT NULL,
    text_details TEXT,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) with time zone NOT NULL DEFAULT NOW()
);