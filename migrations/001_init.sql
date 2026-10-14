-- +goose Up
CREATE TABLE users(
    id BIGSERIAL PRIMARY KEY,
    tg_id BIGINT NOT NULL UNIQUE,
    username TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE items(
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    url TEXT NOT NULL,
    title TEXT,
    target_price NUMERIC(12,2) NOT NULL,
    last_price NUMERIC(12,2),
    last_currency TEXT,
    last_checked_at TIMESTAMPTZ,
    last_notified_at TIMESTAMPTZ,
    last_notified_price NUMERIC(12,2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user_id, url),
    CONSTRAINT items_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX items_user_last_checked_at_idx ON items (last_checked_at);

-- +goose Down
DROP TABLE items;
DROP TABLE users;