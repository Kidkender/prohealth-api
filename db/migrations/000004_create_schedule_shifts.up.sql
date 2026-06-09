CREATE TABLE schedule_shifts (
    id            BIGSERIAL PRIMARY KEY,
    period_id     BIGINT NOT NULL REFERENCES schedule_periods(id) ON DELETE CASCADE,
    department_id BIGINT NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,
    date          DATE NOT NULL,
    time_start    TIME NOT NULL,
    time_end      TIME NOT NULL,
    room_number   VARCHAR(100),
    status        VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    note          TEXT,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_shift_time CHECK (time_end > time_start)
);

CREATE TABLE schedule_shift_doctors (
    shift_id BIGINT NOT NULL REFERENCES schedule_shifts(id) ON DELETE CASCADE,
    user_id  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (shift_id, user_id)
);

CREATE INDEX idx_shifts_date          ON schedule_shifts(date);

CREATE INDEX idx_shifts_period        ON schedule_shifts(period_id);

CREATE INDEX idx_shifts_department    ON schedule_shifts(department_id);
