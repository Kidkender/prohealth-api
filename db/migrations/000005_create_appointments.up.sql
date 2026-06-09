CREATE TABLE appointments (
    id               BIGSERIAL PRIMARY KEY,
    shift_id         BIGINT NOT NULL REFERENCES schedule_shifts(id) ON DELETE RESTRICT,
    patient_name     VARCHAR(255) NOT NULL,
    patient_phone    VARCHAR(20),
    appointment_time TIME NOT NULL,
    reason           VARCHAR(50) NOT NULL DEFAULT 'newPatientVisit',
    status           VARCHAR(20) NOT NULL DEFAULT 'pending',
    notes            TEXT,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_appointments_shift  ON appointments(shift_id);

CREATE INDEX idx_appointments_status ON appointments(status);
