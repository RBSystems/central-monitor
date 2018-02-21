package logging

import (
	"os"
	"strings"
	"time"

	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
)

func SendELKEvent(info ei.EventInfo, device string) {
	//from the info we can fill out the event information
	split := strings.Split(device, "-")

	event := ei.Event{
		Hostname:         device,
		Timestamp:        time.Now().Format(time.RFC3339),
		LocalEnvironment: false,
		Event:            info,
		Building:         split[0],
		Room:             split[1],
	}

	if len(os.Getenv("ELASTIC_API_EVENTS")) != 0 {
		elkreporting.SendElkEvent(os.Getenv("ELASTIC_API_EVENTS"), event)
	}
	if len(os.Getenv("ELASTIC_API_EVENTS_DEV")) != 0 {
		elkreporting.SendElkEvent(os.Getenv("ELASTIC_API_EVENTS_DEV"), event)
	}
}

//this is an entry point that could log to mulitple places
func Log(info ei.EventInfo, device string) {
	SendELKEvent(info, device)
}
