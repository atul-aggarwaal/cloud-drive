package domain

import "errors"

var(
	ErrorPermissionDenied = errors.New("permission denied")
	ErrorMethodNotAllowed = errors.New("method not allowed")
	ErrorUnauthorizedUser = errors.New("unauthorized user")
	ErrorInvalidRequest = errors.New("invalid request")
)