CREATE TABLE outbox_jobs (
	id SERIAL PRIMARY KEY,
	type VARCHAR(255) NOT NULL,
	status VARCHAR(20) NOT NULL DEFAULT 'pending',
	payload BYTEA NOT NULL,
	retry_count INT NOT NULL DEFAULT 0,
	last_error TEXT,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_outbox_jobs_status ON outbox_jobs (status);
