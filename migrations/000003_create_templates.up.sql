CREATE TABLE templates (
    id          UUID PRIMARY KEY,
    name        VARCHAR(255)         NOT NULL UNIQUE,
    body        TEXT                 NOT NULL,
    created_at  TIMESTAMPTZ          NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ          NOT NULL DEFAULT NOW()
);
