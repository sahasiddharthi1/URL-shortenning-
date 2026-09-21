CREATE TABLE IF NOT EXISTS urls (
    id          SERIAL PRIMARY KEY,
    short_code  VARCHAR(10) UNIQUE NOT NULL,
    long_url    TEXT NOT NULL,
    created_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS analytics (
    id          SERIAL PRIMARY KEY,
    short_code  VARCHAR(10) REFERENCES urls(short_code),
    referrer    TEXT,
    user_agent  TEXT,
    ip_address  TEXT,
    clicked_at  TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_analytics_code ON analytics(short_code);
