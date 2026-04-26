-- +migrate Up

CREATE TABLE IF NOT EXISTS shorten (
    id BIGSERIAL PRIMARY KEY,
    shorten_url VARCHAR(255) NOT NULL,
    original_url TEXT NOT NULL UNIQUE
    );