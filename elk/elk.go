package elk

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/byuoitav/event-translator-microservice/elkreporting"
	"github.com/byuoitav/state-parsing/alerts/base"
	"github.com/fatih/color"
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

func GetNonSuppressedControlProcessors() (map[string]string, error) {
	addr := fmt.Sprintf("%s/%s/_search", os.Getenv("ELK_ADDR"), "oit-static-av-devices,oit-static-av-rooms")

	respCode, body, err := base.MakeELKRequest(addr, "POST", []byte(ControlProcsAndSuppressedRooms), 1)

	if err != nil {
		//there's an error
		log.Printf(color.HiRedString("[ELK-Query] There was an error with the initial query: %v", err.Error()))
		return nil, err
	}

	if respCode/100 != 2 {
		msg := fmt.Sprintf("[ELK-Query] Non 200 response received from the initial query: %v, %s", respCode, body)
		log.Printf(color.HiRedString(msg))
		return nil, errors.New(msg)
	}

	resp := BaseResponse{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		msg := fmt.Sprintf("[ELK-Query] Unable to unmarshal response: %v, %s", err.Error(), body)
		log.Printf(color.HiRedString(msg))
		return nil, errors.New(msg)
	}

	toReturn := make(map[string]string)

	for _, h := range resp.Hits.Hits {
		//we go through and build our map
		toReturn[h.ID] = h.Type
	}
	return toReturn, nil
}
