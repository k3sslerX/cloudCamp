CREATE TABLE api_keys (
    key VARCHAR(64) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    active BOOLEAN DEFAULT TRUE
);

CREATE TABLE api_rate_limits (
    api_key VARCHAR(64) PRIMARY KEY REFERENCES api_keys(key),
    capacity INTEGER NOT NULL DEFAULT 100,
    rate INTEGER NOT NULL DEFAULT 10,
    frequency BIGINT NOT NULL DEFAULT 600
);