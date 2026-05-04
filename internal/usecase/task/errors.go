package task

import "errors"

var ErrInvalidInput = errors.New("invalid task input")
var ErrInvalidSchedule = errors.New("invalid schedule input")
var ErrScheduleNotFound = errors.New("schedule not found")
var ErrDuplicateMaterialization = errors.New("task already materialized for schedule and date")
