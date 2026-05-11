-- +migrate Up

CREATE UNIQUE INDEX IF NOT EXISTS idx_shorten_original_url
    ON shorten(original_url);