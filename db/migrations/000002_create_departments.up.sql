CREATE TABLE departments (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

ALTER TABLE user_details
    ADD COLUMN department_id BIGINT REFERENCES departments(id) ON DELETE SET NULL;
