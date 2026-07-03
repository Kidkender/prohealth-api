CREATE TABLE sessions (
	id         BIGSERIAL PRIMARY KEY,
	user_id     BIGINT NOT NULL,
	token_hash  TEXT NOT NULL UNIQUE,
	expires_at  TIMESTAMPTZ NOT NULL,
	ip         TEXT,
	user_agent  TEXT,
	revoked    BOOLEAN NOT NULL DEFAULT FALSE,
	revoked_at  TIMESTAMPTZ,
	last_used_at TIMESTAMPTZ,
	created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_sessions_user ON sessions (user_id);
