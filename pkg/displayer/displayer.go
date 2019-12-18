// Package inspired by https://github.com/davecheney/httpstat
package displayer

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Dainerx/wpam/pkg/types"

	"github.com/Dainerx/wpam/pkg/stat"
	"github.com/fatih/color"
)

const (
	newLine    = "\n"
	sep        = "\n---------------------------\n"
	timeFormat = "02-Jan-2006 15:04:05"
)

// Omitted interface for the percent sign
func print(message string) (n int, err error) {
	return fmt.Fprintln(color.Output, message)
}

func grayscale(code color.Attribute) func(string, ...interface{}) string {
	return color.New(code + 232).SprintfFunc()
}

func colorize(s string) string {
	v := strings.Split(s, newLine)
	v[0] = grayscale(16)(v[0])
	return strings.Join(v, newLine)
}

func colorizeAlert(website string, time time.Time, availability float64) string {
	var colorizedAlertMessage string
	if availability < types.AvaiabilityThreshold {
		colorizedAlertMessage = color.RedString("Website " + website + " is down. Availability=" +
			fmt.Sprintf("%.2f%%", availability) + ", time=" + time.Format(timeFormat))
	} else {
		colorizedAlertMessage = color.GreenString("Website " + website + " has resumed. Availability=" +
			fmt.Sprintf("%.2f %%", availability) + ", time=" + time.Format(timeFormat))
	}
	return colorizedAlertMessage
}

func DisplayStatsAndAlerts(title string, time time.Time, mapAllAlerts map[string]types.Alerts, mapAllStats map[string]stat.Stat) {
	output := title + newLine
	output += color.CyanString("Metrics since " + time.Format(timeFormat))
	output += sep
	// Sort the map stat to assure the same display every time.
	var urls []string
	for url := range mapAllStats {
		urls = append(urls, url)
	}
	sort.Strings(urls)

	// Loop the sorted slice
	for _, url := range urls {
		stats := mapAllStats[url]
		urlColored := color.CyanString(url)
		var lastStatusColored string
		if stats.LastStatus == types.Up {
			lastStatusColored = color.GreenString(stats.LastStatus)
		} else if stats.LastStatus == types.Unkown {
			lastStatusColored = color.YellowString(stats.LastStatus)
		} else {
			lastStatusColored = color.RedString(stats.LastStatus)
		}
		var availabilityColored string
		if stats.Availability >= types.AvaiabilityThreshold {
			availabilityColored = color.GreenString(fmt.Sprintf("%.2f%%", stats.Availability))
		} else {
			availabilityColored = color.RedString(fmt.Sprintf("%.2f%%", stats.Availability))
		}
		var failuresCountColored string
		if stats.FailuresCount == 0 {
			failuresCountColored = color.GreenString(strconv.FormatInt(int64(stats.FailuresCount), 10))
		} else {
			failuresCountColored = color.RedString(strconv.FormatInt(int64(stats.FailuresCount), 10))
		}

		line := "[" + urlColored + "]"
		line += fmt.Sprintf("Last status=%s, Availability=%s, Failures count=%s, AvgRt=%.3fs, MaxRt=%.3fs, MinRt=%.3fs, Content Length=%d",
			lastStatusColored, availabilityColored, failuresCountColored, stats.AvgRt, stats.MaxRt, stats.MinRt, stats.ContentLength)

		// Alerts
		if mapAllAlerts[url].Display {
			for _, alert := range mapAllAlerts[url].Alerts {
				line += newLine + colorizeAlert(url, alert.Timestamp, alert.Availability)
			}
		}
		output += line + sep
	}
	output += newLine
	print(colorize(output))
}

func DisplaySuccessMessage(format string, a ...interface{}) {
	format = color.GreenString(format)
	fmt.Fprintf(color.Output, format, a...)
}

func DisplayWarning(format string, a ...interface{}) {
	format = color.YellowString(format)
	fmt.Fprintf(color.Output, format, a...)
}

func DisplayError(format string, a ...interface{}) {
	format = color.RedString(format)
	fmt.Fprintf(color.Output, format, a...)
}
