CREATE TABLE hospitals (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    slug        VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    address     TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE hospital_members (
    hospital_id BIGINT NOT NULL REFERENCES hospitals(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        VARCHAR(50) NOT NULL DEFAULT 'doctor',
    joined_at   TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (hospital_id, user_id)
);

ALTER TABLE schedule_periods
    ADD COLUMN hospital_id BIGINT REFERENCES hospitals(id);

ALTER TABLE schedule_shifts
    ADD COLUMN hospital_id BIGINT REFERENCES hospitals(id);

ALTER TABLE appointments
    ADD COLUMN hospital_id BIGINT REFERENCES hospitals(id);

ALTER TABLE users DROP COLUMN role;

CREATE INDEX idx_hospital_members_user     ON hospital_members(user_id);

CREATE INDEX idx_schedule_periods_hospital ON schedule_periods(hospital_id);

CREATE INDEX idx_schedule_shifts_hospital  ON schedule_shifts(hospital_id);

CREATE INDEX idx_appointments_hospital     ON appointments(hospital_id);
