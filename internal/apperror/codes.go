package apperror

// General
var ErrUnauthorized = Unauthorized("error.unauthorized")
var ErrForbidden = Forbidden("error.forbidden")
var ErrRateLimitExceeded = RateLimitExceeded("error.rate_limit.exceeded")
var ErrInternalServerError = InternalServerError("error.internal_server_error")

// Auth
var ErrInvalidCredentials = Unauthorized("error.auth.invalid_credentials")
var ErrInvalidToken = Unauthorized("error.auth.invalid_token")

// User
var ErrUserExists = Conflict("error.user.already_exists")
var ErrUserNotFound = NotFound("error.user.not_found")
var ErrUserDetailNotFound = NotFound("error.user_detail.not_found")

// Hospital
var ErrHospitalMemberNotFound = NotFound("error.hospital_member.not_found")
var ErrHospitalSlugAlreadyExists = Conflict("error.hospital.slug.already_exists")

// Department
var ErrDepartmentNotFound = NotFound("error.department.not_found")

// Schedule
var ErrInvalidTimeRange = BadRequest("error.schedule.invalid_time_range")
var ErrSchedulePeriodNotFound = NotFound("error.schedule.period.not_found")
var ErrSchedulePeriodAlreadyExists = Conflict("error.schedule.period.already_exists")
var ErrSchedulePeriodHasShifts = Conflict("error.schedule.period.has_shifts")
var ErrScheduleShiftNotFound = NotFound("error.schedule.shift.not_found")
var ErrDoctorAlreadyAssigned = Conflict("error.schedule.doctor.already_assigned")

// Appointment
var ErrAppointmentNotFound = NotFound("error.appointment.not_found")
