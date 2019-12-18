package website_check

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Dainerx/wpam/pkg/logger"
	"github.com/Dainerx/wpam/pkg/safe_store"
	"github.com/Dainerx/wpam/pkg/types"
)

var (
	httpMethods          = []string{types.HTTPGet, types.HTTPHead, types.HTTPPost, types.HTTPPut, types.HTTPDelete, types.HTTPOptions, types.HTTPTrace}
	httpMethodsSupported = []string{types.HTTPGet, types.HTTPPost}
)

const (
	PKG              = "website_check"
	maxTimeOut       = (20 * time.Second)
	minTimeOut       = (1 * time.Second)
	maxCheckInterval = (2 * time.Minute)
	minCheckInterval = (5 * time.Second)
)

//add default values for mashling
type CheckRequest struct {
	id                             string
	url                            string
	httpMethod                     string
	timeout                        time.Duration
	httpAcceptedResponseStatusCode []int //if it is not here then it is down
	checkInterval                  time.Duration
	data                           map[string]interface{}
	netClient                      *http.Client
	store                          *safe_store.SafeStore
	firstRequest                   bool
	stop                           chan bool
}

func findString(strings []string, element string) bool {
	for _, s := range strings {
		if s == element {
			return true
		}
	}
	return false
}

func NewCheckRequest(id, url string) CheckRequest {
	checkRequest := CheckRequest{}
	checkRequest.id = id
	checkRequest.url = url
	checkRequest.httpMethod = types.HTTPGet
	checkRequest.timeout = time.Duration(10 * time.Second)
	checkRequest.httpAcceptedResponseStatusCode = append(checkRequest.httpAcceptedResponseStatusCode, http.StatusOK)
	checkRequest.checkInterval = time.Duration(10 * time.Second)
	checkRequest.data = make(map[string]interface{})
	checkRequest.netClient = &http.Client{
		Timeout: checkRequest.timeout,
	}
	checkRequest.firstRequest = true
	checkRequest.stop = make(chan bool)
	return checkRequest
}

func NewcheckRequestFromInstance(instance types.Instance, store *safe_store.SafeStore) (*CheckRequest, error) {
	checkRequest := &CheckRequest{}
	if instance.Id == "" {
		return checkRequest, ErrIdEmpty
	}
	checkRequest.id = instance.Id

	if _, err := url.ParseRequestURI(instance.Url); err != nil {
		return checkRequest, ErrUrlNotValid
	}
	checkRequest.url = instance.Url

	if instance.HttpMethod == "" {
		checkRequest.httpMethod = strings.ToUpper(types.HTTPGet)
	} else {
		checkRequest.httpMethod = strings.ToUpper(instance.HttpMethod)
	}

	if !findString(httpMethods, checkRequest.httpMethod) {
		return checkRequest, ErrHttpMethodNotRecognized
	}
	if !findString(httpMethodsSupported, checkRequest.httpMethod) {
		return checkRequest, ErrHttpMethodNotSupported
	}
	if instance.Timeout == time.Duration(0*time.Second) {
		checkRequest.timeout = time.Duration(10 * time.Second)
	} else {
		checkRequest.timeout = instance.Timeout
		if checkRequest.timeout > maxTimeOut || checkRequest.timeout < minTimeOut {
			return checkRequest, ErrTimeOutNowNotInInterval
		}
	}
	if len(instance.HttpAcceptedResponseStatusCode) == 0 {
		checkRequest.httpAcceptedResponseStatusCode = []int{http.StatusOK}
	} else {
		checkRequest.httpAcceptedResponseStatusCode = instance.HttpAcceptedResponseStatusCode
	}
	if instance.CheckInterval == time.Duration(0*time.Second) {
		checkRequest.checkInterval = time.Duration(10 * time.Second)
	} else {
		checkRequest.checkInterval = instance.CheckInterval //converts to seconds
		if checkRequest.checkInterval > maxCheckInterval || checkRequest.checkInterval < minCheckInterval {
			return checkRequest, ErrCheckIntervalNotInInterval
		}

	}
	checkRequest.data = instance.Data
	checkRequest.netClient = &http.Client{
		Timeout: checkRequest.timeout,
	}
	checkRequest.store = store
	checkRequest.firstRequest = true
	checkRequest.stop = make(chan bool)
	return checkRequest, nil
}

func (checkRequest CheckRequest) Id() string {
	return checkRequest.id
}

func (checkRequest CheckRequest) Url() string {
	return checkRequest.url
}

func (checkRequest CheckRequest) doRequest() (*http.Response, error) {
	switch checkRequest.httpMethod {
	case types.HTTPPost:
		requestBody, err := json.Marshal(checkRequest.data)
		if err != nil {
			return nil, err
		}
		return checkRequest.netClient.Post(checkRequest.url, "application/json", bytes.NewBuffer(requestBody))
	case types.HTTPGet:
		return checkRequest.netClient.Get(checkRequest.url)
	default: // Will never be reached on runtime, since the HttpMethod check happens on configuration's parsing.
		return nil, ErrHttpMethodNotRecognized
	}
}

// Returns Response with Status DOWN when any of the following occur:
// - The request to url times out.
// - The response code is not in the httpAcceptedResponseStatusCode slice
// Otherwise returns Response with Status UP.
func (checkRequest *CheckRequest) Response() (CheckResponse, error) {
	start := time.Now()
	res, err := checkRequest.doRequest()
	if err != nil {
		checkResponse := NewCheckResponse(-1, 0, -1)
		checkResponse.status = types.Down
		logger.Logger.Warnf("Website is %s, reason: %v", checkResponse.status, err)
		return *checkResponse, err
	}
	responseTime := time.Since(start)
	httpResponseStatusCode := res.StatusCode
	checkResponse := NewCheckResponse(httpResponseStatusCode, responseTime, res.ContentLength)
	//Add an if statement for when the http method is not reconigzed
	if checkResponse.matchesAcceptedCodes(checkRequest.httpAcceptedResponseStatusCode) {
		checkResponse.status = types.Up
		logger.Logger.Infof("Website %s is %s, with http_response_status_code=%d, took %v s to respond", checkRequest.url, checkResponse.status, checkResponse.httpStatusCode, checkResponse.responseTime.Seconds())
	} else {
		checkResponse.status = types.Down
		logger.Logger.Infof("Website %s is %s with http_code = %d", checkRequest.url, checkResponse.status, checkResponse.httpStatusCode)
	}
	// Update Last check
	return *checkResponse, nil
}

// Does a firstRequest if checkRequest.firstRequest has false as value
// Otherwise enter an infinite loop, fetch a response put in the safe store.
// Stops when the channel when the checkRequest.stop channel is unlocked through the Stop() method.
func (checkRequest *CheckRequest) Run() {
	// If it is the first request do it instantly to feed data to the store
	if checkRequest.firstRequest {
		checkResponse, _ := checkRequest.Response()
		checkRequest.store.Put(checkRequest.url, checkResponse)
		checkRequest.firstRequest = false
	}
	ticker := time.NewTicker(checkRequest.checkInterval)
	for {
		select {
		case <-ticker.C:
			checkResponse, _ := checkRequest.Response()
			checkRequest.store.Put(checkRequest.url, checkResponse)
		case <-checkRequest.stop: // If stop is called this chan is unlocked and the infinite loop is exited
			logger.Logger.Errorf("check request stopped!")
			return
		}
	}
}

// Behaves exactly like Run(), only it stops after x seconds
func (checkRequest *CheckRequest) RunForXSeconds(x time.Duration) {
	// If it is the first request do it instantly to feed data to the store
	if checkRequest.firstRequest {
		checkResponse, err := checkRequest.Response()
		if err != nil {
			logger.Logger.Errorf("%v", err)
		}
		checkRequest.store.Put(checkRequest.url, checkResponse)
		checkRequest.firstRequest = false
	}
	ticker := time.NewTicker(checkRequest.checkInterval)
	stopTime := time.NewTicker(x)
	for {
		select {
		case <-ticker.C:
			checkResponse, err := checkRequest.Response()
			if err != nil {
				logger.Logger.Errorf("%v", err)
			}
			checkRequest.store.Put(checkRequest.url, checkResponse)
		case <-stopTime.C: // Stop after X seconds
			return
		}
	}
}

// Unlocks the stop channel.
func (checkRequest *CheckRequest) Stop() {
	checkRequest.stop <- true
}
