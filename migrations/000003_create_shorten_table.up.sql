-- +migrate Up

ALTER TABLE shorten ADD COLUMN IF NOT EXISTS user_id UUID;