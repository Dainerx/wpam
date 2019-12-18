package cmd

import (
	"bytes"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dainerx/wpam/pkg/displayer"
	"github.com/Dainerx/wpam/pkg/logger"
	"github.com/Dainerx/wpam/pkg/safe_store"
	"github.com/Dainerx/wpam/pkg/types"
	"github.com/Dainerx/wpam/pkg/website_check"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Author                  = "Oussama Ben Ghorbel"
	PKG                     = "cmd"
	titleStatsTenMinutesAgo = "Perodic 10s metrics for the past 10 minutes."
	titleStatsOneHourAgo    = "Perodic 1m metrics in the past 1 hour."
	tenMinutes              = 10
	config                  = "config"
)

var rootCmd = &cobra.Command{
	Use:   "wpam",
	Short: "wpam: Website Availability & Performance Monitoring Tool made by " + Author,
	Run: func(cmd *cobra.Command, args []string) {
		viper.SetConfigType("yaml")
		configFilePath := viper.GetString(config)
		if configFilePath != "" {
			viper.SetConfigFile(configFilePath)
			if err := viper.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); ok {
					// Config file not found
					displayer.DisplayError("Config file not found.\n")
					logger.Logger.Fatal("Config file not found.")

				} else {
					// Config file was found but another error was produced
					displayer.DisplayError("Failed to read config file: %v.\n", err)
					logger.Logger.Fatalf("Failed to read config file: %v.", err)
				}
			}
			// Config file found and successfully parsed
			logger.Logger.Infof("Log file found and successfully parsed.")
			displayer.DisplaySuccessMessage("Read config from file: %s\n", configFilePath)
		} else {
			// Run with default config
			var yamlExample = []byte(`
input:
  - id: datadog
    url: https://www.datadoghq.com/
`)
			err := viper.ReadConfig(bytes.NewBuffer(yamlExample))
			if err != nil {
				logger.Logger.Fatalf("Failed to read default config")
			}
			displayer.DisplayWarning("No config file given, launching one instance: https://www.datadoghq.com/.\n")
		}

		// Unmarshal configuration using Viper
		var config types.Configuration
		if err := viper.Unmarshal(&config); err != nil {
			displayer.DisplayError("Failed to unmarshal configuration: %v.\n", err)
			logger.Logger.Fatalf("Failed to unmarshal configuration: %v", err)
		}
		// Configuration unmarshalled
		logger.Logger.Infof("Configuration successfully unmarshalled.")
		// Create the safe store
		safeStore := safe_store.New()

		// Run valid instances on different Go routine
		var instances []website_check.CheckRequest
		seenIds, seenUrls := make(map[string]string), make(map[string]string)
		for _, instance := range config.Input {
			instance.CheckInterval *= 1e9 // Defaults nano seconds, converts before moving on.
			instance.Timeout *= 1e9       // Defaults nano seconds, converts before moving on.
			checkRequest, err := website_check.NewcheckRequestFromInstance(instance, safeStore)
			if err != nil {
				displayer.DisplayWarning("Instance with Id {%s} will not be considered: %v\n", checkRequest.Id(), err)
				logger.Logger.Warnf("%s", err.Error())
				continue
			}
			if id, seen := seenIds[checkRequest.Id()]; seen {
				displayer.DisplayWarning("Instance with Id {%s} already seen, thus duplicated instance will not be considered.\n", id)
				logger.Logger.Warnf("Instance with Id {%s} already seen, thus duplicated instance will not be considered.", id)
				continue
			}
			if url, seen := seenUrls[checkRequest.Url()]; seen {
				displayer.DisplayWarning("Instance with Url {%s} already seen, thus duplicated instance will not be considered.\n", url)
				logger.Logger.Warnf("Instance with Url {%s} already seen, thus duplicated instance will not be considered.", url)
				continue
			}

			// Consider this instance
			instances = append(instances, *checkRequest)
			// Add it in seen ids and urls
			seenIds[checkRequest.Id()] = checkRequest.Id()
			seenUrls[checkRequest.Url()] = checkRequest.Url()
			// Run checkRequest on different go routines (for each instance a goroutine)
			go checkRequest.Run()
		}

		// Keep going with the main thread to display
		tickerTenSeconds := time.NewTicker(time.Duration(10 * time.Second))
		tickerOneMinute := time.NewTicker(time.Duration(1 * time.Minute))
		tickerOneHour := time.NewTicker(time.Duration(1 * time.Hour))
		// Channel to capture System exit
		c := make(chan os.Signal, 1) //buffer size 1 or risk missing the signal
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		for {
			select {
			case <-c:
				// Stop all the instances
				displayer.DisplayWarning("\nCaptured shutdown signal.\n")
				displayer.DisplayWarning("Stopping all instances.\n")
				for _, instance := range instances {
					instance.Stop()
				}
				displayer.DisplaySuccessMessage("All Instances have stopped.\n")
				displayer.DisplaySuccessMessage("Bye!\n")
				// Safe Store will be garbage collected
				os.Exit(0)
			case <-tickerOneHour.C:
				// Clean data after one hour
				logger.Logger.Info("Started data cleaning process. Dropping data dating more than one hour ago.\n")
				safeStore.CleanData()
				logger.Logger.Info("Data cleaning process has finished.\n")
			case <-tickerOneMinute.C:
				mapAllStats := safeStore.GetAllStatsOneHourAgo()
				mapAllAlerts := safeStore.GetAllAlerts()
				displayer.DisplayStatsAndAlerts(titleStatsOneHourAgo, time.Now().Add(-1*tenMinutes*time.Minute), mapAllAlerts, mapAllStats)

			case <-tickerTenSeconds.C:
				// map used for display metrics
				mapAllStats := safeStore.GetAllStatsTenMinutesAgo()
				// map used for alerts needed to be displayed
				mapAllAlerts := safeStore.GetAllAlerts()
				displayer.DisplayStatsAndAlerts(titleStatsTenMinutesAgo, time.Now().Add(-1*tenMinutes*time.Minute), mapAllAlerts, mapAllStats)
			}
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "--config path/to/configfile.yaml")
	err := viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	if err != nil {
		logger.Logger.Fatalf("Failed to bind flag: %v", err)
	}
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
