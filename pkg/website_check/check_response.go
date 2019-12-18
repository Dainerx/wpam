package website_check

import (
	"time"

	"github.com/Dainerx/wpam/pkg/types"
)

type CheckResponse struct {
	timestamp      int64
	httpStatusCode int
	responseTime   time.Duration
	contentLength  int64
	status         string
}

func NewCheckResponse(httpStatusCode int, responseTime time.Duration, contentLength int64) *CheckResponse {
	checkResponse := &CheckResponse{}
	checkResponse.timestamp = time.Now().UnixNano()
	checkResponse.httpStatusCode = httpStatusCode
	checkResponse.responseTime = responseTime
	checkResponse.contentLength = contentLength
	return checkResponse
}

func NewCheckResponseWithStatus(httpAcceptedResponseStatusCode []int, httpStatusCode int, responseTime time.Duration, contentLength int64) *CheckResponse {
	checkResponse := &CheckResponse{}
	checkResponse.timestamp = time.Now().UnixNano()
	checkResponse.httpStatusCode = httpStatusCode
	checkResponse.responseTime = responseTime
	if checkResponse.matchesAcceptedCodes(httpAcceptedResponseStatusCode) {
		checkResponse.status = types.Up
	} else {
		checkResponse.status = types.Down
	}
	checkResponse.contentLength = contentLength
	return checkResponse
}

func (checkResponse CheckResponse) Timestamp() int64 {
	return checkResponse.timestamp
}

func (checkResponse CheckResponse) HttpStatusCode() int {
	return checkResponse.httpStatusCode
}

func (checkResponse CheckResponse) ResponseTime() time.Duration {
	return checkResponse.responseTime
}

func (checkResponse CheckResponse) ContentLength() int64 {
	return checkResponse.contentLength
}

func (checkResponse CheckResponse) Status() string {
	return checkResponse.status
}

// matchesAcceptedCodes tells wether an url is up depending on the checkRequest httpAcceptedResponseStatusCodes
// And checkResponseesponse Http_status_code
// It returns true if checkResponseesponse.Http_status_code is in checkRequest.httpAcceptedResponseStatusCode.
func (checkResponse CheckResponse) matchesAcceptedCodes(httpAcceptedResponseStatusCodes []int) bool {
	for _, httpAcceptedResponseStatusCode := range httpAcceptedResponseStatusCodes {
		if checkResponse.httpStatusCode == httpAcceptedResponseStatusCode {
			return true
		}
	}
	return false
}
