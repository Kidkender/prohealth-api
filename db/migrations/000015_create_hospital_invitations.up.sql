CREATE TABLE hospital_invitations (
    id          BIGSERIAL PRIMARY KEY,
    hospital_id BIGINT NOT NULL REFERENCES hospitals(id) ON DELETE CASCADE,
    email       VARCHAR(255) NOT NULL,
    role        VARCHAR(50) NOT NULL DEFAULT 'doctor',
    token_hash  VARCHAR(64) NOT NULL UNIQUE,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    invited_by  BIGINT NOT NULL REFERENCES users(id),
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_hospital_invitations_hospital ON hospital_invitations(hospital_id);
CREATE INDEX idx_hospital_invitations_email    ON hospital_invitations(email);

CREATE UNIQUE INDEX uq_hospital_invitations_pending
    ON hospital_invitations(hospital_id, email)
    WHERE status = 'pending';
