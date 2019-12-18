package website_check

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dainerx/wpam/pkg/safe_store"

	"github.com/Dainerx/wpam/pkg/types"
)

// Post handler
func postHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	var req map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Get Handler
func getHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Test simple HTTP get.
func TestResponseWithGet(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(postHandler),
	)
	defer ts.Close()
	url := ts.URL
	checkRequest := NewCheckRequest("TestResponseWithGet", url)
	got, err := checkRequest.Response()
	if err == ErrHttpMethodNotRecognized || err == ErrHttpMethodNotSupported {
		t.Fatalf("checkRequest.Response() failed: %v", err)
	}

	if got.status != types.Up {
		t.Errorf("checkRequest.Response() = %s; want %s", got.status, types.Up)
	}

	checkRequest.id = "BadCheckX"
	checkRequest.url = "https://www.facebook.com/BadCheckX"
	got, err = checkRequest.Response()
	if err != nil {
		t.Fatalf("checkRequest.Response() failed: %v", err)
	}
	if got.Status() != types.Down {
		t.Errorf("checkRequest.Response() = %s; want %s", got.Status(), types.Down)
	}

}

// Test http Post
func TestResponseWithPost(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(postHandler),
	)
	defer ts.Close()
	url := ts.URL
	data := make(map[string]string)
	data["k1"] = "v1"
	data["k2"] = "v2"
	instance := types.Instance{
		Id:                             "TestResponseWithPost",
		Url:                            url,
		HttpMethod:                     types.HTTPPost,
		Timeout:                        time.Second * 4,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK, http.StatusBadRequest}, //Facebook answers BadRequest on Post
		CheckInterval:                  time.Second * 10,
	}
	checkRequest, err := NewcheckRequestFromInstance(instance, &safe_store.SafeStore{})
	if err != nil {
		t.Fatalf("%v", err)
	}
	got, err := checkRequest.Response() // It should be up since http.StatusBadRequest is accepted
	if err != nil {
		t.Fatalf("checkRequest.Response() failed: %v", err)
	}
	if got.Status() == types.Down {
		t.Errorf("checkRequest.Response() = %s; want %s", got.Status(), types.Down)
	}
}

// Test wrong http method
func TestResponseAgainstWrongHttpMethod(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(getHandler),
	)
	defer ts.Close()
	url := ts.URL
	instance := types.Instance{
		Id:                             "TestResponseAgainstWrongHttpMethod",
		Url:                            url,
		HttpMethod:                     "NOTVALIDMETHOD",
		Timeout:                        time.Nanosecond * 1,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 10,
	}
	_, err := NewcheckRequestFromInstance(instance, &safe_store.SafeStore{})
	if err != ErrHttpMethodNotRecognized {
		t.Fatal("checkRequest.Response() failed: HttpMethodNotRecognized not handled")
		return
	}
}

// Test if timeout is being handled.
func TestResponseAgainstTimeOut(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(getHandler),
	)
	defer ts.Close()
	url := ts.URL
	instance := types.Instance{
		Id:                             "TestResponseAgainstTimeOut",
		Url:                            url,
		HttpMethod:                     types.HTTPGet,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 10,
	}
	checkRequest, _ := NewcheckRequestFromInstance(instance, &safe_store.SafeStore{})
	checkRequest.netClient.Timeout = time.Nanosecond * 1 // Bad way to access it
	_, err := checkRequest.Response()
	if err == nil {
		t.Fatal("checkRequest.Response() failed: timeout not handled")
		return
	}

}
