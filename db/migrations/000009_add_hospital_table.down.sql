DROP INDEX IF EXISTS idx_appointments_hospital;

DROP INDEX IF EXISTS idx_schedule_shifts_hospital;

DROP INDEX IF EXISTS idx_schedule_periods_hospital;

DROP INDEX IF EXISTS idx_hospital_members_user;

ALTER TABLE appointments     DROP COLUMN IF EXISTS hospital_id;

ALTER TABLE schedule_shifts  DROP COLUMN IF EXISTS hospital_id;

ALTER TABLE schedule_periods DROP COLUMN IF EXISTS hospital_id;

ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'user';

DROP TABLE IF EXISTS hospital_members;

DROP TABLE IF EXISTS hospitals;
