CREATE TABLE schedule_periods (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    start_date DATE NOT NULL,
    end_date   DATE NOT NULL,
    status     VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_date_range CHECK (end_date >= start_date)
);
