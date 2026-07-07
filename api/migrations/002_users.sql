CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    name          TEXT NOT NULL DEFAULT '',
    role          TEXT NOT NULL DEFAULT 'user',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS favorites (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    temple_id  BIGINT NOT NULL REFERENCES temples(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, temple_id)
);

CREATE TABLE IF NOT EXISTS reviews (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    temple_id  BIGINT NOT NULL REFERENCES temples(id) ON DELETE CASCADE,
    rating     INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    text       TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS page_views (
    id        BIGSERIAL PRIMARY KEY,
    temple_id BIGINT NOT NULL REFERENCES temples(id) ON DELETE CASCADE,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_favorites_user ON favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_favorites_temple ON favorites(temple_id);
CREATE INDEX IF NOT EXISTS idx_reviews_temple ON reviews(temple_id);
CREATE INDEX IF NOT EXISTS idx_page_views_temple ON page_views(temple_id);
CREATE INDEX IF NOT EXISTS idx_page_views_date ON page_views(viewed_at);
