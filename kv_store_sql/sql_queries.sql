-- first i created 3 instances of postgres on local, then created below table in all of them.

CREATE TABLE key_value_store (
    key TEXT PRIMARY KEY,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    deleted_at TIMESTAMP
);