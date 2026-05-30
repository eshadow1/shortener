-- +migrate Down

ALTER TABLE shorten DROP COLUMN IF EXISTS user_id;