INSERT INTO api_keys (key, user_id, description) VALUES (
    "TEST_API_KEY", "0", "test api key"
);

INSERT INTO api_rate_limits (api_key) VALUES ("TEST_API_KEY");