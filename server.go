package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/byuoitav/central-monitor/monitors"
	"github.com/fatih/color"
)

type config struct {
	Interval     int      `json:"interval"`
	Designations []string `json:"designations"`
}

//non lambda main function
func main() {
	timer := time.NewTimer(1 * time.Second)

	for {
		<-timer.C
		log.Printf(color.HiGreenString("Starting run"))
		//get the config
		config, err := getConfig()
		if err != nil {
			log.Printf(color.HiRedString("Couldn't get Configuration: %v", err.Error()))
			timer.Reset(30 * time.Second)
			continue
		}

		log.Printf(color.HiGreenString("Configuration retrieved: %v", config))
		//we run
		monitors.RunMStatus(config.Designations)
		log.Printf(color.HiGreenString("Done."))

		timer.Reset(time.Duration(config.Interval) * time.Second)
	}
}

func getConfig() (config, error) {

	configLocation := os.Getenv("CONFIG_LOCATION")
	if len(configLocation) <= 0 {
		configLocation = "./config.json"
	}

	b, err := ioutil.ReadFile(configLocation)
	if err != nil {
		return config{}, err
	}

	toReturn := config{}

	err = json.Unmarshal(b, &toReturn)
	return toReturn, err
}

//we run you on a timer - get the interval from the config file
