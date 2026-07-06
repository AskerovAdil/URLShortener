CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT users_email_unique UNIQUE (email)
);

CREATE TABLE urls (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alias        VARCHAR(32) NOT NULL,
    original_url TEXT NOT NULL,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT urls_alias_unique UNIQUE (alias)
);

CREATE INDEX urls_user_id_idx ON urls (user_id);
