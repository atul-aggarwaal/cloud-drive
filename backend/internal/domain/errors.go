package domain

import "errors"

var(
	ErrorPermissionDenied = errors.New("permission denied")
	ErrorMethodNotAllowed = errors.New("method not allowed")
	ErrorUnauthorizedUser = errors.New("unauthorized user")
	ErrorInvalidRequest = errors.New("invalid request")
	ErrorFileNotFound = errors.New("file not found")
	ErorFileCorrupted = errors.New("file has no versions")
)