ALTER TABLE schedule_shifts
    DROP CONSTRAINT IF EXISTS chk_shift_time;

ALTER TABLE schedule_shifts
    ALTER COLUMN time_start TYPE VARCHAR(5) USING TO_CHAR(time_start, 'HH24:MI'),
    ALTER COLUMN time_end   TYPE VARCHAR(5) USING TO_CHAR(time_end,   'HH24:MI');

ALTER TABLE schedule_shifts
    ADD CONSTRAINT chk_shift_time CHECK (time_end > time_start);
