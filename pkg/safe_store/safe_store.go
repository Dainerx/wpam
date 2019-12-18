package safe_store

import (
	"sync"
	"time"

	"github.com/Dainerx/wpam/pkg/logger"
	"github.com/Dainerx/wpam/pkg/stat"
	"github.com/Dainerx/wpam/pkg/types"
)

type alerts map[string]types.Alerts
type store map[string][]types.Response
type statStore map[string]TupleStat

type TupleStat struct {
	twoMinutesAgoStats stat.Stat
	tenMinutesAgoStats stat.Stat
	oneHourAgoStats    stat.Stat
}

type SafeStat struct {
	sync.RWMutex //embedded field
	stats        statStore
}
type SafeStore struct {
	sync.RWMutex //embedded field
	data         store
	safeStat     *SafeStat
	alerts       alerts
}

// Creates a new SafeStat.
func newSafeStat() *SafeStat {
	ss := &SafeStat{
		stats: make(map[string]TupleStat),
	}
	return ss
}

// Fetches an url stat from the store.
// Locks then unlocks the safeStat on read.
// Returns an instance of TupleStat (two minutes, ten minutes and one hour ago).
func (safeStat *SafeStat) getUrlStat(url string) TupleStat {
	safeStat.RLock()
	defer safeStat.RUnlock()
	return safeStat.stats[url]
}

// Fetches an url stat from two minutes ago.
// Locks then unlocks the safeStat on read.
// Returns stat.Stat instance.
func (safeStat *SafeStat) getUrlStatTwoMinutesAgo(url string) stat.Stat {
	safeStat.RLock()
	defer safeStat.RUnlock()
	return safeStat.stats[url].twoMinutesAgoStats
}

// Fetches all stats of two minutes ago for all urls.
// Locks then unlocks the safeStat on read.
// Returns a map mapping each url with a Stat.
func (safeStat *SafeStat) getAllStatTwoMinutesAgo() map[string]stat.Stat {
	safeStat.RLock()
	defer safeStat.RUnlock()
	mapAllStatsTwoMinutesAgo := make(map[string]stat.Stat)
	for url, stat := range safeStat.stats {
		mapAllStatsTwoMinutesAgo[url] = stat.twoMinutesAgoStats
	}
	return mapAllStatsTwoMinutesAgo
}

// Fetches an url stat from ten minutes ago.
// Locks then unlocks the safeStat on read.
// Returns stat.Stat instance.
func (safeStat *SafeStat) getUrlStatTenMinutesAgo(url string) stat.Stat {
	safeStat.RLock()
	defer safeStat.RUnlock()
	return safeStat.stats[url].tenMinutesAgoStats
}

// Fetches all stats of ten minutes ago for all urls.
// Locks then unlocks the safeStat on read.
// Returns a map mapping each url with a Stat.
func (safeStat *SafeStat) getAllStatTenMinutesAgo() map[string]stat.Stat {
	safeStat.RLock()
	defer safeStat.RUnlock()
	mapAllStatsTenMinutesAgo := make(map[string]stat.Stat)
	for url, stat := range safeStat.stats {
		mapAllStatsTenMinutesAgo[url] = stat.tenMinutesAgoStats
	}
	return mapAllStatsTenMinutesAgo
}

// Fetches an url stat from one hour ago.
// Locks then unlocks the safeStat on read.
// Returns stat.Stat instance.
func (safeStat *SafeStat) getUrlStatOneHourAgo(url string) stat.Stat {
	safeStat.RLock()
	defer safeStat.RUnlock()
	return safeStat.stats[url].oneHourAgoStats
}

// Fetches all stats of one hour ago for all urls.
// Locks then unlocks the safeStat on read.
// Returns a map mapping each url with a Stat.
func (safeStat *SafeStat) getAllStatOneHourAgo() map[string]stat.Stat {
	safeStat.RLock()
	defer safeStat.RUnlock()
	mapAllStatsOneHourAgo := make(map[string]stat.Stat)
	for url, stat := range safeStat.stats {
		mapAllStatsOneHourAgo[url] = stat.oneHourAgoStats
	}
	return mapAllStatsOneHourAgo
}

// updateAlerts, takes an url and a time as param then proceeds to update alerts if the url changed the state.
// Locks and unlocks the safestore on Read and Write.
func (safeStore *SafeStore) updateAlerts(url string, time time.Time) {
	availabilityTwoMinutesAgo := safeStore.safeStat.getUrlStatTwoMinutesAgo(url).Availability
	safeStore.RLock()
	websiteAlerts := safeStore.alerts[url]
	data := safeStore.data[url]
	safeStore.RUnlock()

	if len(data) == 1 { // Is this the first check?
		if availabilityTwoMinutesAgo < types.AvaiabilityThreshold { // It went down
			websiteAlerts.Display = true // If a website goes down once always display its alerts
			websiteAlerts.Alerts = append(websiteAlerts.Alerts,
				types.AlertStatus{Timestamp: time,
					Availability: availabilityTwoMinutesAgo}) //Add the down alert
		} else {
			websiteAlerts.Display = false // If a website is up no need to display
			websiteAlerts.Alerts = append(websiteAlerts.Alerts,
				types.AlertStatus{Timestamp: time,
					Availability: availabilityTwoMinutesAgo}) //Add the up alert
		}
	} else {
		// If a website goes down once always display its alerts
		if availabilityTwoMinutesAgo < types.AvaiabilityThreshold { // It went down
			websiteAlerts.Display = true
			if websiteAlerts.Alerts[len(websiteAlerts.Alerts)-1].Availability >= types.AvaiabilityThreshold { // Was it up?
				//Add Alert Down
				websiteAlerts.Alerts = append(websiteAlerts.Alerts,
					types.AlertStatus{Timestamp: time,
						Availability: availabilityTwoMinutesAgo}) //Add the down alert
			}
		} else if websiteAlerts.Alerts[len(websiteAlerts.Alerts)-1].Availability < types.AvaiabilityThreshold && availabilityTwoMinutesAgo >= types.AvaiabilityThreshold { //website was down and resumed
			websiteAlerts.Alerts = append(websiteAlerts.Alerts,
				types.AlertStatus{Timestamp: time,
					Availability: availabilityTwoMinutesAgo}) //Add the resume alert
		}
	}
	safeStore.Lock()
	safeStore.alerts[url] = websiteAlerts // Resassign it
	safeStore.Unlock()
}

// updateStateStore locks the safeStat update entries and unlock it.
// This should be called after every put of data in the SafeStore.
func (safeStat *SafeStat) updateStatStore(url string, responsesTwoMinuteAgo, responsesTenMinuteAgo, responsesOneHourAgo []types.Response) {
	tupleStat := TupleStat{}
	twoMinutesAgoStats, err := stat.NewStat(responsesTwoMinuteAgo)
	if err != nil {
		logger.Logger.Warnf("Could not generate stats of two minutes ago: %v", err)
	} else {
		tupleStat.twoMinutesAgoStats = twoMinutesAgoStats
	}
	tenMinutesAgoStats, err := stat.NewStat(responsesTenMinuteAgo)
	if err != nil {
		logger.Logger.Warnf("Could not generate stats of ten minutes ago: %v", err)
	} else {
		tupleStat.tenMinutesAgoStats = tenMinutesAgoStats
	}
	oneHourAgoStats, err := stat.NewStat(responsesOneHourAgo)
	if err != nil {
		logger.Logger.Warnf("Could not generate stats of one hour ago: %v", err)
	} else {
		tupleStat.oneHourAgoStats = oneHourAgoStats
	}
	safeStat.Lock()
	safeStat.stats[url] = tupleStat
	safeStat.Unlock()
}

// New creates a new SafeStore.
func New() *SafeStore {
	return &SafeStore{
		data:     map[string][]types.Response{},
		alerts:   map[string]types.Alerts{},
		safeStat: newSafeStat(),
	}
}

// Get Responses from X minutes ago, where x of type time.Duration is passed in argument.
// Locks the SafeStore's read lock then unlock it
func getResponsesXMinutesAgo(responses []types.Response, x time.Duration) []types.Response {
	xMinutesAgo := time.Now().Add(-1 * x * time.Minute)
	sep := 0
	for _, response := range responses {
		if response.Timestamp() < xMinutesAgo.UnixNano() {
			sep++
		} else {
			// Since they are sorted by timestamp if the current element
			// is not from x minutes ago, the rest won't be
			break
		}
	}
	return responses[sep:]
}

// Get Responses from X hours ago, where x of type time.Duration is passed in argument.
// Locks the SafeStore's read lock then unlock it
// Used for data cleaning.
func getResponsesXHoursAgo(responses []types.Response, x time.Duration) []types.Response {
	xHoursAgo := time.Now().Add((-1 * x))
	sep := 0
	for _, response := range responses {
		if response.Timestamp() < xHoursAgo.UnixNano() {
			sep++
		} else {
			// Since they are sorted by timestamp if the current element
			// is not from x minutes ago, the rest won't be
			break
		}
	}
	return responses[sep:]
}

// Put will add a response to the url's data (responses) in the store.
// Not linear due to the update of statStore <- look more into this
// Locks the SafeStore's write lock then unlock it
func (s *SafeStore) Put(url string, response types.Response) {
	s.RLock()
	s.data[url] = append(s.data[url], response)
	currentResponses := s.data[url]
	s.RUnlock()
	//can be optimized
	s.safeStat.updateStatStore(url, getResponsesXMinutesAgo(currentResponses, 2), getResponsesXMinutesAgo(currentResponses, 10), getResponsesXMinutesAgo(currentResponses, 60))
	s.updateAlerts(url, time.Now())
}

// Remove data (responses) of an url from the store.
// Locks the SafeStore's write lock then unlock it
func (s *SafeStore) Remove(url string) {
	s.Lock()
	defer s.Unlock()
	// If key does not exist delete is no-op.
	delete(s.data, url)
	logger.Logger.Infof("%s 's data was removed from the store", url)
}

// Len will return the number of entries in the store's data.
// Locks the SafeStore's write lock then unlock it in order to keep the Len() integrity
func (s *SafeStore) Len() int {
	s.Lock()
	defer s.Unlock()
	return len(s.data)
}

// Get an array of responses from the store mapped by an url.
// O(1)
// Locks the SafeStore's read lock then unlock it
func (s *SafeStore) Get(url string) []types.Response {
	s.RLock()
	defer s.RUnlock()
	return s.data[url]
}

// Get an url's stats as Alerts.
// O(1)
// Locks the SafeStore read lock then unlock it.
func (s *SafeStore) GetUrlAlerts(url string) types.Alerts {
	s.RLock()
	defer s.RUnlock()
	return s.alerts[url]
}

// Get an all alerts as map mapping every url to its Alerts.
// Locks the SafeStore read lock then unlock it.
func (s *SafeStore) GetAllAlerts() map[string]types.Alerts {
	s.RLock()
	defer s.RUnlock()
	mapAllAlert := make(map[string]types.Alerts)
	for url, alerts := range s.alerts {
		mapAllAlert[url] = alerts
	}
	return mapAllAlert
}

// Get an url's stats as a TupleStat.
// O(1)
// Locks the SafeStore.safeStat's read lock then unlock it.
func (s *SafeStore) GetUrlStats(url string) TupleStat {
	return s.safeStat.getUrlStat(url)
}

// Get an url's stats from 10 minutes ago.
// O(1)
// Locks the SafeStore.safeStat's read lock then unlock it.
func (s *SafeStore) GetUrlStatsTwoMinutesAgo(url string) stat.Stat {
	return s.safeStat.getUrlStatTwoMinutesAgo(url)
}

func (s *SafeStore) GetAllStatsTwoMinutesAgo() map[string]stat.Stat {
	return s.safeStat.getAllStatTwoMinutesAgo()
}

// Get an url's stats from 10 minutes ago.
// O(1)
// Locks the SafeStore.safeStat's read lock then unlock it.
func (s *SafeStore) GetUrlStatsTenMinutesAgo(url string) stat.Stat {
	return s.safeStat.getUrlStatTenMinutesAgo(url)
}

func (s *SafeStore) GetAllStatsTenMinutesAgo() map[string]stat.Stat {
	return s.safeStat.getAllStatTenMinutesAgo()
}

// Get an url's stats from 1 hour ago.
// O(1)
// Locks the SafeStore.safeStat's read lock then unlock it.
func (s *SafeStore) GetUrlStatsOneHourAgo(url string) stat.Stat {
	return s.safeStat.getUrlStatOneHourAgo(url)
}

func (s *SafeStore) GetAllStatsOneHourAgo() map[string]stat.Stat {
	return s.safeStat.getAllStatOneHourAgo()
}

func (s *SafeStore) CleanDataFromXHoursAgo(x int) {
	s.Lock()
	defer s.Unlock()
	for url, responses := range s.data {
		s.data[url] = getResponsesXHoursAgo(responses, time.Duration(x)*time.Hour) //keep only data from one hour ago
	}
}

// CleanData cleans the store and remove responses dating more than one hour ago
func (s *SafeStore) CleanData() {
	logger.Logger.Info("Started data cleaning process. Dropping data dating more than one hour ago.")
	s.CleanDataFromXHoursAgo(1)
	logger.Logger.Info("Data cleaning process has finished.")
	// That's all cause stats are always updated and alerts are always kept for historical reasons
}
