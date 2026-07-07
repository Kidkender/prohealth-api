package apperror

// General
var ErrUnauthorized = Unauthorized("error.unauthorized")
var ErrForbidden = Forbidden("error.forbidden")
var ErrRateLimitExceeded = RateLimitExceeded("error.rate_limit.exceeded")
var ErrInternalServerError = InternalServerError("error.internal_server_error")

// Auth
var ErrInvalidCredentials = Unauthorized("error.auth.invalid_credentials")
var ErrInvalidToken = Unauthorized("error.auth.invalid_token")
var ErrTokenRevoked = Unauthorized("error.auth.token_revoked")
var ErrTokenReused = Unauthorized("error.auth.token_reused")
var ErrTokenExpired = Unauthorized("error.auth.token_expired")

// User
var ErrUserExists = Conflict("error.user.already_exists")
var ErrUserNotFound = NotFound("error.user.not_found")
var ErrUserDetailNotFound = NotFound("error.user_detail.not_found")
var ErrUserBanned = Forbidden("error.user.banned")
var ErrUserInactive = Forbidden("error.user.inactive")
var ErrCannotBanSelf = BadRequest("error.user.cannot_ban_self")
var ErrInvalidUserStatus = BadRequest("error.user.invalid_status")

// Hospital
var ErrHospitalMemberNotFound = NotFound("error.hospital_member.not_found")
var ErrHospitalSlugAlreadyExists = Conflict("error.hospital.slug.already_exists")
var ErrHospitalNotFound = NotFound("error.hospital.not_found")
var ErrHospitalHasDependents = Conflict("error.hospital.has_dependents")

// Department
var ErrDepartmentNotFound = NotFound("error.department.not_found")
var ErrDepartmentAlreadyExists = Conflict("error.department.already_exists")
var ErrDepartmentInUse = Conflict("error.department.in_use")

// Schedule
var ErrInvalidTimeRange = BadRequest("error.schedule.invalid_time_range")
var ErrSchedulePeriodNotFound = NotFound("error.schedule.period.not_found")
var ErrSchedulePeriodAlreadyExists = Conflict("error.schedule.period.already_exists")
var ErrSchedulePeriodHasShifts = Conflict("error.schedule.period.has_shifts")
var ErrScheduleShiftNotFound = NotFound("error.schedule.shift.not_found")
var ErrDoctorAlreadyAssigned = Conflict("error.schedule.doctor.already_assigned")

// Appointment
var ErrAppointmentNotFound = NotFound("error.appointment.not_found")

var ErrSchedulePeriodGenerating = Conflict("error.schedule.period.generating")
