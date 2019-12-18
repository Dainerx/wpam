package cmd

import (
	"testing"

	"github.com/Dainerx/wpam/pkg/safe_store"
	"github.com/Dainerx/wpam/pkg/types"
	"github.com/Dainerx/wpam/pkg/website_check"
	"github.com/spf13/viper"
)

func TestUnmarshall(t *testing.T) {
	viper.SetConfigType("yaml")
	// Payload with three entries, two valid and one not
	viper.SetConfigName("test_unmarshall")
	viper.AddConfigPath("test_payloads")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found
			t.Fatal("Config file not found.")
		} else {
			// Config file was found but another error was produced
			t.Fatalf("Failed to read config file: %v.", err)
		}
	}

	var config types.Configuration
	if err := viper.Unmarshal(&config); err != nil {
		t.Errorf("Failed to unmarshal configuration: %v", err)
	}

	safeStore := safe_store.New()
	var instances []website_check.CheckRequest
	for _, instance := range config.Input {
		checkRequest, err := website_check.NewcheckRequestFromInstance(instance, safeStore)
		t.Log(checkRequest)
		if err != nil {
			t.Logf("%s", err.Error())
		} else {
			instances = append(instances, *checkRequest)
		}
	}

	const wanted = 2
	got := len(instances)
	if got < wanted {
		t.Errorf("Unmarshall did not work len(instances)= %d; want %d", got, wanted)
	} else if got > wanted {
		t.Errorf("Unmarshall did not detect wrong entry len(instances)= %d; want %d", got, wanted)
	}

}
