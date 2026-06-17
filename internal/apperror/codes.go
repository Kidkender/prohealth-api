package apperror

var ErrUserExists = Conflict("error.user.already_exists")
var ErrInvalidCredentials = Unauthorized("error.auth.invalid_credentials")
var ErrUserNotFound = NotFound("error.user.not_found")
var ErrUserDetailNotFound = NotFound("error.user_detail.not_found")
var ErrInvalidTimeRange = BadRequest("error.schedule.invalid_time_range")
var ErrSchedulePeriodNotFound = NotFound("error.schedule.period.not_found")
var ErrDepartmentNotFound = NotFound("error.department.not_found")
var ErrScheduleShiftNotFound = NotFound("error.schedule.shift.not_found")
var ErrInvalidToken = Unauthorized("error.auth.invalid_token")
var ErrDoctorAlreadyAssigned = Conflict("error.schedule.doctor.already_assigned")
var ErrSchedulePeriodHasShifts = Conflict("error.schedule.period.has_shifts")
