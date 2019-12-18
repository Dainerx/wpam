package stat_test

import (
	"math"
	"net/http"
	"testing"
	"time"

	"github.com/Dainerx/wpam/pkg/stat"
	"github.com/Dainerx/wpam/pkg/types"
	"github.com/Dainerx/wpam/pkg/website_check"
)

var httpAcceptedStatusCodes = []int{http.StatusOK}

const (
	float64EqualityThreshold         = 1e-6
	expectedAvailability     float64 = 75.0
	expectedFailuresCount    int     = 1
	expectedMaxRt            float64 = 3.9
	expectedMinRt            float64 = 0.3
	expectedAvgRt            float64 = 1.7
)

func feedGenericResponses() []types.Response {
	var responses []types.Response
	cr1 := website_check.NewCheckResponseWithStatus(httpAcceptedStatusCodes, http.StatusOK, time.Duration(1*time.Second), 0)
	responses = append(responses, *cr1)
	cr2 := website_check.NewCheckResponseWithStatus(httpAcceptedStatusCodes, http.StatusNotFound, 300*time.Millisecond, 0)
	responses = append(responses, *cr2)
	cr3 := website_check.NewCheckResponseWithStatus(httpAcceptedStatusCodes, http.StatusOK, ((3 * time.Second) + (900 * time.Millisecond)), 0)
	responses = append(responses, *cr3)
	cr4 := website_check.NewCheckResponseWithStatus(httpAcceptedStatusCodes, http.StatusOK, time.Duration((1*time.Second)+(600*time.Millisecond)), 0)
	responses = append(responses, *cr4)
	return responses
}
func TestStatAvailability(t *testing.T) {
	responses := feedGenericResponses()
	stat, err := stat.NewStat(responses)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if gotAvailability := stat.Availability; gotAvailability != expectedAvailability {
		t.Errorf("Failed StatAvailability()=%f, want %f", gotAvailability, expectedAvailability)
	}
	if gotFailuresCount := stat.FailuresCount; gotFailuresCount != expectedFailuresCount {
		t.Errorf("Failed StatAvailability()=%d, want %d", gotFailuresCount, expectedFailuresCount)
	}
}

func TestStatMaxMinAvgRT(t *testing.T) {
	responses := feedGenericResponses()
	stat, err := stat.NewStat(responses)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if gotmax := stat.MaxRt; gotmax != expectedMaxRt {
		t.Errorf("Failed StatMaxMinAvgResponsesTime() max=%f, want %f", gotmax, expectedMaxRt)
	}
	if gotmin := stat.MinRt; gotmin != expectedMinRt {
		t.Errorf("Failed StatMaxMinAvgResponsesTime() min=%f, want %f", gotmin, expectedMinRt)
	}
	if gotavg := stat.AvgRt; math.Abs(gotavg-expectedAvgRt) > float64EqualityThreshold {
		t.Errorf("Failed StatMaxMinAvgResponsesTime() avg=%f, want %f", gotavg, expectedAvgRt)
	}
}

func TestStatWithInvalidDataSize(t *testing.T) {
	_, err := stat.NewStat([]types.Response{})
	if err != stat.ErrDataSizeInvalid {
		t.Errorf("Invalid size of data is unhandled")
	}
}
