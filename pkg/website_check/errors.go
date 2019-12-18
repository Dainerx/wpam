package website_check

import "errors"

var (
	// ErrIdEmpty is returned when an instance does not have an ID.
	ErrIdEmpty = errors.New(`Id empty`)

	// ErrUrlNotValid is returned when an instance hsa a non valid url.
	ErrUrlNotValid = errors.New(`Url value is not valid.`)

	// ErrHttpMethodNotRecognized is returned when an instance has an unrecognizable HTTP method.
	ErrHttpMethodNotRecognized = errors.New("HTTP method not recognized.")

	// ErrHttpMethodNotSupported is returned when an instance has a not supported yet HTTP method.
	ErrHttpMethodNotSupported = errors.New("HTTP method not supported yet.")

	// ErrCheckIntervalNotInInterval is returned when the instance's check interval is not in the range.
	ErrCheckIntervalNotInInterval = errors.New("Check interval is not in the accepted range [5s,2m]")

	// ErrCheckIntervalNotInInterval is returned when the instance's timeout is not in the range.
	ErrTimeOutNowNotInInterval = errors.New("Timeout is not in the accepted range [1s,20s]")
)
