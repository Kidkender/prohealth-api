ALTER TABLE schedule_shifts
    DROP CONSTRAINT IF EXISTS chk_shift_time;

ALTER TABLE schedule_shifts
    ALTER COLUMN time_start TYPE TIME USING time_start::TIME,
    ALTER COLUMN time_end   TYPE TIME USING time_end::TIME;

ALTER TABLE schedule_shifts
    ADD CONSTRAINT chk_shift_time CHECK (time_end > time_start);
