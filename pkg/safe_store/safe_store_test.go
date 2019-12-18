package safe_store_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Dainerx/wpam/pkg/safe_store"
	"github.com/Dainerx/wpam/pkg/types"
	"github.com/Dainerx/wpam/pkg/website_check"
)

const (
	keyFirst     = "first"
	KeySecond    = "second"
	TenResponses = 10
)

// This file tests the store creation and different methods.
// Contains benchmarking for read/write, read only and write only.

// Get Handler
func getHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestStoreCreation(t *testing.T) {
	s := safe_store.New()
	if s.Len() != 0 {
		t.Errorf("Failed, invalid store length.")
	}
}

func TestStorePut(t *testing.T) {
	s := safe_store.New()
	s.Put(keyFirst, website_check.CheckResponse{})
	if got := s.Len(); got != 1 {
		t.Errorf("s.Len() = %d, want 1.", got)
	}
	s.Put(keyFirst, website_check.CheckResponse{})
	if got := s.Len(); got != 1 {
		t.Errorf("s.Len() = %d, want 1.", got)
	}
	if got := s.Get(keyFirst); len(got) != 2 {
		t.Errorf("len(s.Get(%s)) = %d, want 2.", keyFirst, len(got))
	}

	s.Put("second", website_check.CheckResponse{})
	if got := s.Len(); got != 2 {
		t.Errorf("s.Len() = %d, want 2.", got)
	}
}

func TestStoreRemove(t *testing.T) {
	s := safe_store.New()
	s.Put(keyFirst, website_check.CheckResponse{})
	if got := s.Len(); got != 1 {
		t.Errorf("s.Len() = %d, want 1.", got)
	}
	s.Remove(keyFirst)
	if got := s.Len(); got != 0 {
		t.Errorf("s.Len() = %d, want 0.", got)
	}
	// Testing the no-op
	s.Remove(KeySecond)
	if got := s.Len(); got != 0 {
		t.Errorf("s.Len() = %d, want 0.", got)
	}
}

func TestStoreGet(t *testing.T) {
	ts := httptest.NewServer(
		http.HandlerFunc(getHandler),
	)
	defer ts.Close()
	url := ts.URL
	s := safe_store.New()
	instance := types.Instance{
		Id:                             "TestStoreGet",
		Url:                            url,
		HttpMethod:                     types.HTTPGet,
		Timeout:                        time.Second * 1,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 10,
	}
	checkRequest, _ := website_check.NewcheckRequestFromInstance(instance, s)
	checkResponse, _ := checkRequest.Response()
	s.Put(keyFirst, checkResponse)
	if got := s.Get(keyFirst); got[0].Status() != types.Up {
		t.Errorf("s.Get(%s).Status = %s, want %s.", keyFirst, got[0].Status(), types.Up)
	}

}

func TestCleanUpData(t *testing.T) {
	s := safe_store.New()
	ts := httptest.NewServer(
		http.HandlerFunc(getHandler),
	)
	defer ts.Close()
	url := ts.URL
	instance := types.Instance{
		Id:                             "TestCleanUpData",
		Url:                            url,
		HttpMethod:                     types.HTTPGet,
		Timeout:                        time.Second * 1,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 5,
	}
	checkRequest, err := website_check.NewcheckRequestFromInstance(instance, s)
	if err != nil {
		t.Fatalf("%v", err)
	}
	checkRequest.RunForXSeconds(time.Second * 3)
	if len(s.Get(url)) == 0 {
		t.Error("No request were performed.")
	}
	s.CleanDataFromXHoursAgo(-1) // Go to the future and clean the data
	if len(s.Get(url)) != 0 {
		t.Error("Data cleaning process failed.")
	}
}

func BenchmarkRead(b *testing.B) {
	s := safe_store.New()
	nbr := b.N
	for i := 0; i < nbr; i++ {
		s.Put(strconv.FormatInt(int64(i), 10), website_check.CheckResponse{})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Get(strconv.FormatInt(int64(i-1), 10))
	}
}

func BenchmarkWrite(b *testing.B) {
	s := safe_store.New()
	for i := 0; i < b.N; i++ {
		s.Put(strconv.FormatInt(int64(i), 10), website_check.CheckResponse{})
	}
}

func BenchmarkReadWrite(b *testing.B) {
	s := safe_store.New()
	r := false
	for i := 0; i < b.N; i++ {
		if r == false {
			s.Put(strconv.FormatInt(int64(i), 10), website_check.CheckResponse{})
			r = true
		} else {
			_ = s.Get(strconv.FormatInt(int64(i), 10))
			r = false
		}
	}
}
