CREATE TABLE IF NOT EXISTS task_schedules (
	id BIGSERIAL PRIMARY KEY,
	base_title TEXT NOT NULL,
	base_description TEXT NOT NULL DEFAULT '',
	status_template TEXT NOT NULL,
	schedule_type TEXT NOT NULL,
	schedule_payload JSONB NOT NULL,
	start_date DATE NOT NULL,
	end_date DATE NULL,
	is_active BOOLEAN NOT NULL DEFAULT TRUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	CONSTRAINT chk_task_schedules_status_template
		CHECK (status_template IN ('new', 'in_progress', 'done')),
	CONSTRAINT chk_task_schedules_type
		CHECK (schedule_type IN ('daily', 'monthly_day', 'specific_dates', 'odd_even_days')),
	CONSTRAINT chk_task_schedules_date_range
		CHECK (end_date IS NULL OR end_date >= start_date),
	CONSTRAINT chk_task_schedules_payload_is_object
		CHECK (jsonb_typeof(schedule_payload) = 'object'),
	CONSTRAINT chk_task_schedules_payload_shape
		CHECK (
			(schedule_type = 'daily'
				AND schedule_payload ? 'interval'
				AND jsonb_typeof(schedule_payload->'interval') = 'number')
			OR
			(schedule_type = 'monthly_day'
				AND schedule_payload ? 'day'
				AND jsonb_typeof(schedule_payload->'day') = 'number')
			OR
			(schedule_type = 'specific_dates'
				AND schedule_payload ? 'dates'
				AND jsonb_typeof(schedule_payload->'dates') = 'array')
			OR
			(schedule_type = 'odd_even_days'
				AND schedule_payload ? 'mode'
				AND jsonb_typeof(schedule_payload->'mode') = 'string'
				AND schedule_payload->>'mode' IN ('odd', 'even'))
		)
	);

CREATE INDEX IF NOT EXISTS idx_task_schedules_active ON task_schedules (is_active);
CREATE INDEX IF NOT EXISTS idx_task_schedules_type ON task_schedules (schedule_type);

ALTER TABLE tasks
	ADD COLUMN IF NOT EXISTS schedule_id BIGINT NULL,
	ADD COLUMN IF NOT EXISTS planned_for DATE NULL;

DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = 'fk_tasks_schedule'
	) THEN
		ALTER TABLE tasks
			ADD CONSTRAINT fk_tasks_schedule
			FOREIGN KEY (schedule_id)
			REFERENCES task_schedules (id)
			ON DELETE SET NULL;
	END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_tasks_schedule_id ON tasks (schedule_id);
CREATE INDEX IF NOT EXISTS idx_tasks_planned_for ON tasks (planned_for);

CREATE UNIQUE INDEX IF NOT EXISTS ux_tasks_schedule_planned_for
	ON tasks (schedule_id, planned_for)
	WHERE schedule_id IS NOT NULL AND planned_for IS NOT NULL;
