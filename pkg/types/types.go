package types

import (
	"time"
)

// Instance is a struct that holds an instance of input from the user configuration file.
type Instance struct {
	Id                             string
	Url                            string
	HttpMethod                     string
	Timeout                        time.Duration
	HttpAcceptedResponseStatusCode []int //if it is not here then it is down
	CheckInterval                  time.Duration
	Data                           map[string]interface{}
}

// Configuration is struct holding an array of instances.
type Configuration struct {
	Input []Instance
}

// Response is an interface having five methods, CheckResponse for instance implements this interface.
type Response interface {
	Timestamp() int64
	HttpStatusCode() int
	ResponseTime() time.Duration
	ContentLength() int64
	Status() string
}

// Alerts status is a struct pairing every availability and timestamp.
type AlertStatus struct {
	Timestamp    time.Time
	Availability float64
}

// Alerts is a truct holding an array of Alert Status and bool display (true needs to display, false no).
type Alerts struct {
	Alerts  []AlertStatus
	Display bool
}
