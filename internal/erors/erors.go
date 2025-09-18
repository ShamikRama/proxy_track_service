package erors

import "errors"

var (
	ErrTrackCodeRequired = errors.New("track code is required")
	ErrTrackCodeInvalid  = errors.New("track code is invalid")
	ErrTrackCodeNotFound = errors.New("track code not found")
)