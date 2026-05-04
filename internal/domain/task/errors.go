package task

import "errors"

var ErrNotFound = errors.New("task not found")
var ErrInvalidSchedule = errors.New("invalid schedule")
var ErrScheduleNotFound = errors.New("schedule not found")
