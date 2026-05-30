-- +migrate Down

ALTER TABLE shorten DROP COLUMN IF EXISTS is_deleted;