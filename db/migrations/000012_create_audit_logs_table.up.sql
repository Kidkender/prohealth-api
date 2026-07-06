CREATE TABLE audit_logs (
	id SERIAL PRIMARY KEY,
	hospital_id INT,
	actor_id INT,
	action VARCHAR(255) NOT NULL,
	entity_type VARCHAR(100) NOT NULL,
	entity_id INT,
	metadata JSONB,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_hospital_id ON audit_logs(hospital_id);
CREATE INDEX idx_audit_logs_actor_id ON audit_logs(actor_id);
