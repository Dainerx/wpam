package website_check

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/Dainerx/wpam/pkg/safe_store"
	"github.com/Dainerx/wpam/pkg/types"
)

func TestFindString(t *testing.T) {

	var a = []string{"hello", "world", "random", "string"}
	s := "random"
	got := findString(a, s)
	if got != true {
		t.Errorf("FindString(%v,%s) = %s; want true", a, s, strconv.FormatBool(got))
	}

	s = "notfound"
	got = findString(a, s)
	if got != false {
		t.Errorf("FindString(%v,%s) = %s; want false", a, s, strconv.FormatBool(got))
	}

	s = "HELLO"
	got = findString(a, s)
	if got != false {
		t.Errorf("FindString(%v,%s) = %s; want false", a, s, strconv.FormatBool(got))
	}
}

func TestUrlValidation(t *testing.T) {
	instanceUrlNotValid := types.Instance{
		Id:                             "not valid",
		Url:                            "url",
		HttpMethod:                     types.HTTPGet,
		Timeout:                        time.Second * 4,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 10,
	}
	_, err := NewcheckRequestFromInstance(instanceUrlNotValid, &safe_store.SafeStore{})
	if err == nil {
		t.Errorf("Url validation failed got %v; want %v", err, ErrHttpMethodNotRecognized)
	}

	instanceUrlNotValid = types.Instance{
		Id:                             "google",
		Url:                            "http//google.com", //not valid missing :
		HttpMethod:                     types.HTTPGet,
		Timeout:                        time.Second * 4,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 10,
	}
	_, err = NewcheckRequestFromInstance(instanceUrlNotValid, &safe_store.SafeStore{})
	if err == nil {
		t.Errorf("Url validation failed got %v; want %v", err, ErrHttpMethodNotRecognized)
	}

	instanceUrlValid := types.Instance{
		Id:                             "google",
		Url:                            "http://google.com",
		HttpMethod:                     types.HTTPGet,
		Timeout:                        time.Second * 4,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 10,
	}
	_, err = NewcheckRequestFromInstance(instanceUrlValid, &safe_store.SafeStore{})
	if err != nil {
		t.Errorf("Url validation failed got %v; want %v", err, nil)
	}
}
func TestHttpMethodValidation(t *testing.T) {
	instanceHttpMethodNotValid := types.Instance{
		Id:                             "google",
		Url:                            "http://google.com",
		HttpMethod:                     "NOT_HTTP_METHOD",
		Timeout:                        time.Second * 4,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 10,
	}
	_, err := NewcheckRequestFromInstance(instanceHttpMethodNotValid, &safe_store.SafeStore{})
	if err == nil {
		t.Errorf("Http method validation failed got %v; want %v", err, ErrHttpMethodNotRecognized)
	}

	instanceHttpMethodNotSupported := types.Instance{
		Id:                             "google",
		Url:                            "http://google.com",
		HttpMethod:                     types.HTTPPut,
		Timeout:                        time.Second * 4,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 10,
	}
	_, err = NewcheckRequestFromInstance(instanceHttpMethodNotSupported, &safe_store.SafeStore{})
	if err == nil {
		t.Errorf("Http method validation failed got %v; want %v", err, ErrHttpMethodNotSupported)
	}

	instanceValidHttpMethod := types.Instance{
		Id:                             "google",
		Url:                            "http://google.com",
		HttpMethod:                     "get", //testing lower case
		Timeout:                        time.Second * 4,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 10,
	}
	_, err = NewcheckRequestFromInstance(instanceValidHttpMethod, &safe_store.SafeStore{})
	if err != nil {
		t.Errorf("Http method validation failed got %v; want %v", err, nil)
	}

}

func TestTimeOutValidation(t *testing.T) {
	instanceTimeOutNowNotInInterval := types.Instance{
		Id:                             "google",
		Url:                            "http://google.com",
		HttpMethod:                     types.HTTPGet,
		Timeout:                        time.Millisecond * 1,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 10,
	}
	_, err := NewcheckRequestFromInstance(instanceTimeOutNowNotInInterval, &safe_store.SafeStore{})
	if err == nil {
		t.Errorf("Timeout validation failed got %v; want %v", err, ErrTimeOutNowNotInInterval)
	}
}

func TestCheckIntervalValidation(t *testing.T) {
	instanceTimeOutNowNotInInterval := types.Instance{
		Id:                             "google",
		Url:                            "http://google.com",
		HttpMethod:                     types.HTTPGet,
		Timeout:                        time.Second * 6,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 2, // too small
	}
	_, err := NewcheckRequestFromInstance(instanceTimeOutNowNotInInterval, &safe_store.SafeStore{})
	if err == nil {
		t.Errorf("CheckInterval validation failed got %v; want %v", err, ErrCheckIntervalNotInInterval)
	}
}
