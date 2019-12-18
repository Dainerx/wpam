package stat

import (
	"math"

	"github.com/Dainerx/wpam/pkg/types"
)

type Stat struct {
	LastStatus    string
	Availability  float64
	FailuresCount int
	MaxRt         float64
	MinRt         float64
	SumRt         float64
	AvgRt         float64
	ContentLength int64
}

// Create a new stat and returns it.
// Returns error not nil if failed to compute one of the stats.
func NewStat(responses []types.Response) (Stat, error) {
	s := Stat{}
	s.LastStatus = types.Unkown
	err := s.statMaxMinAvgResponsesTime(responses)
	if err != nil {
		return s, err
	}
	err = s.statAvailability(responses)
	if err != nil {
		return s, err
	}
	err = s.statContentLength(responses)
	if err != nil {
		return s, err
	}

	s.LastStatus = responses[len(responses)-1].Status()
	return s, err
}

// Computes the website's max, min and average response time.
// Returns error if failed to compute.
func (s *Stat) statMaxMinAvgResponsesTime(responses []types.Response) error {
	if len(responses) > 0 {
		var sum float64 = 0
		var min float64 = math.MaxFloat64
		var max float64 = -1
		for _, response := range responses {
			rt := response.ResponseTime()
			sum += rt.Seconds()
			max = math.Max(max, rt.Seconds())
			min = math.Min(min, rt.Seconds())
		}

		avg := (sum / float64(len(responses)))
		s.MaxRt, s.MinRt, s.AvgRt = max, min, avg
		return nil
	} else {
		s.MaxRt, s.MinRt, s.AvgRt = -1, -1, -1
		return ErrDataSizeInvalid
	}
}

// Computes the website's availability as percentage and failures count
// Returns error if failed to compute.
func (s *Stat) statAvailability(responses []types.Response) error {
	if len(responses) > 0 {
		var upCount int = 0
		for _, response := range responses {
			if response.Status() == types.Up {
				upCount++
			}
		}
		s.Availability, s.FailuresCount = ((float64(upCount) / float64(len(responses))) * 100), (len(responses) - upCount)
		return nil
	} else {
		s.Availability, s.FailuresCount = -1, -1
		return ErrDataSizeInvalid
	}
}

// Compute content length of the response (given in header)
// Returns error if failed to compute.
func (s *Stat) statContentLength(responses []types.Response) error {
	sizeResponses := len(responses)
	if sizeResponses > 0 {
		s.ContentLength = responses[sizeResponses-1].ContentLength()
		return nil
	} else {
		s.ContentLength = -1
		return ErrDataSizeInvalid
	}
}
