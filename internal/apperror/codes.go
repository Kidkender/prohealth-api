package apperror

var ErrUserExists = Conflict("error.user.already_exists")
var ErrInvalidCredentials = Unauthorized("error.auth.invalid_credentials")
var ErrUserNotFound = NotFound("error.user.not_found")
