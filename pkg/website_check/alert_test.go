package website_check

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Dainerx/wpam/pkg/safe_store"

	"github.com/Dainerx/wpam/pkg/types"
)

const (
	alwaysDown            = "/always_down"
	alwaysDownAlertsCount = 1
	downUp                = "/down_up"
	downUpAlertsCount     = 4
	alwaysUp              = "/always_up"
	alwaysUpAlertsCount   = 1
)

var requestCount int32 = 0

// Fails at 0
// Resumes at 4 with avaibility of 80%
// Fails again at 5 with avaibility of 66.66%
// Resumes back again at 7 with avaibility of 85.71%
func downUpUrlHandler(w http.ResponseWriter, r *http.Request) {
	if requestCount < 1 || (requestCount == 5) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	atomic.AddInt32(&requestCount, 1) // Very important to make sure requestCount is synchronized
}

// Always failing
func alwaysDownUrlHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}

// Always up
func alwaysUpUrlHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Test Alerts for an url that's always down.
func TestAlertsAlwaysDown(t *testing.T) {
	safeStore := safe_store.New()
	// Start the mock HTTP Server
	t.Log("Starting mock HTTP server")
	ts := httptest.NewServer(
		http.HandlerFunc(alwaysDownUrlHandler),
	)
	defer ts.Close()

	url := ts.URL
	instanceAlwaysDown := types.Instance{
		Id:                             alwaysDown,
		Url:                            url,
		HttpMethod:                     types.HTTPPost,
		Timeout:                        time.Second * 4,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 6, //check every 2s
	}
	checkRequestAlwaysDown, err := NewcheckRequestFromInstance(instanceAlwaysDown, safeStore)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Run the instance for 10 seconds
	checkRequestAlwaysDown.RunForXSeconds(10 * time.Second)
	alerts := safeStore.GetUrlAlerts(url)
	// This instance is always down it should have only one alert that it was down.
	if len(alerts.Alerts) != alwaysDownAlertsCount {
		t.Errorf("Failed to find the right number of alerts got %d; want %d", len(alerts.Alerts), alwaysDownAlertsCount)
	}
	// Website went down should have an alerts.Display value of true.
	if !alerts.Display {
		t.Errorf("Failed to have the right display value for alerts got %t; want true", alerts.Display)
	}
	// The only existing alert should have an avaibility of 0%.
	if alerts.Alerts[0].Availability != 0 {
		t.Errorf("Failed to have the right avaibility value for alerts.Alerts[0] got %f; want 0", alerts.Alerts[0].Availability)
	}
}

// Test Alerts for an url that's always up.
func TestAlertsAlwaysUp(t *testing.T) {
	safeStore := safe_store.New()
	// Start the mock HTTP Server
	t.Log("Starting mock HTTP server")
	ts := httptest.NewServer(
		http.HandlerFunc(alwaysUpUrlHandler),
	)
	defer ts.Close()

	url := ts.URL
	instanceDownUp := types.Instance{
		Id:                             alwaysUp,
		Url:                            url,
		HttpMethod:                     types.HTTPGet,
		Timeout:                        time.Second * 2,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 5,
	}

	checkRequestAlwaysUp, err := NewcheckRequestFromInstance(instanceDownUp, safeStore)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Run the instance for 10 seconds
	checkRequestAlwaysUp.RunForXSeconds(10 * time.Second)
	alerts := safeStore.GetUrlAlerts(url)
	// This instance should have 1 alerts saying it up and not changing.
	if len(alerts.Alerts) != alwaysUpAlertsCount {
		t.Errorf("Failed to find the right number of alerts got %d; want %d", len(alerts.Alerts), alwaysUpAlertsCount)
	}
	// Website never went down should have an alerts.Display value of false.
	if alerts.Display {
		t.Errorf("Failed to have the right display value for alerts got %t; want false", alerts.Display)
	}
}

// Test Alerts for an url that fails twice and resumes twice.
func TestAlertsDownUp(t *testing.T) {
	safeStore := safe_store.New()
	// Start the mock HTTP Server
	t.Log("Starting mock HTTP server")
	ts := httptest.NewServer(
		http.HandlerFunc(downUpUrlHandler),
	)
	defer ts.Close()

	url := ts.URL
	instanceDownUp := types.Instance{
		Id:                             downUp,
		Url:                            url,
		HttpMethod:                     types.HTTPGet,
		Timeout:                        time.Second * 2,
		HttpAcceptedResponseStatusCode: []int{http.StatusOK},
		CheckInterval:                  time.Second * 5,
	}

	checkRequestDownUp, err := NewcheckRequestFromInstance(instanceDownUp, safeStore)
	if err != nil {
		t.Fatalf("%v", err)
	}
	// Run the instance for 60 seconds
	checkRequestDownUp.RunForXSeconds(60 * time.Second)
	alerts := safeStore.GetUrlAlerts(url)
	// This instance should 4 alerts since it fails and resumes twice.
	if len(alerts.Alerts) != downUpAlertsCount {
		t.Errorf("Failed to find the right number of alerts got %d; want %d", len(alerts.Alerts), downUpAlertsCount)
	}
	// Website went down once should have an alerts.Display value of true.
	if !alerts.Display {
		t.Errorf("Failed to have the right display value for alerts got %t; want true", alerts.Display)
	}
}
