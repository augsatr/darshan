CREATE TABLE IF NOT EXISTS temples (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    slug        TEXT NOT NULL UNIQUE,
    state       TEXT NOT NULL,
    city        TEXT NOT NULL DEFAULT '',
    deity       TEXT NOT NULL DEFAULT '',
    arch_style  TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    history     TEXT NOT NULL DEFAULT '',
    image_url   TEXT NOT NULL DEFAULT '',
    model_url   TEXT NOT NULL DEFAULT '',
    ambient_audio TEXT NOT NULL DEFAULT '',
    latitude    DOUBLE PRECISION NOT NULL DEFAULT 0,
    longitude   DOUBLE PRECISION NOT NULL DEFAULT 0,
    visit_duration INT NOT NULL DEFAULT 0,
    is_published BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS images (
    id          BIGSERIAL PRIMARY KEY,
    temple_id   BIGINT NOT NULL REFERENCES temples(id) ON DELETE CASCADE,
    url         TEXT NOT NULL,
    alt         TEXT NOT NULL DEFAULT '',
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_temples_state ON temples(state);
CREATE INDEX IF NOT EXISTS idx_temples_slug ON temples(slug);
CREATE INDEX IF NOT EXISTS idx_images_temple_id ON images(temple_id);
