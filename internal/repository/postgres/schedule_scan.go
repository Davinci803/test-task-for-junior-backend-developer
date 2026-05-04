package postgres

import taskdomain "example.com/taskservice/internal/domain/task"

type scheduleScanner interface {
	Scan(dest ...any) error
}

func scanSchedule(scanner scheduleScanner) (*taskdomain.Schedule, error) {
	var (
		scheduleType   string
		statusTemplate string
		rawPayload     []byte
		schedule       taskdomain.Schedule
	)

	if err := scanner.Scan(
		&schedule.ID,
		&schedule.BaseTitle,
		&schedule.BaseDescription,
		&statusTemplate,
		&scheduleType,
		&rawPayload,
		&schedule.StartDate,
		&schedule.EndDate,
		&schedule.IsActive,
		&schedule.CreatedAt,
		&schedule.UpdatedAt,
	); err != nil {
		return nil, err
	}

	payload, err := decodeSchedulePayload(taskdomain.ScheduleType(scheduleType), rawPayload)
	if err != nil {
		return nil, err
	}

	schedule.StatusTemplate = taskdomain.Status(statusTemplate)
	schedule.Type = taskdomain.ScheduleType(scheduleType)
	schedule.Payload = payload

	return &schedule, nil
}
