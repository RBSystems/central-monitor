package mstatus

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/byuoitav/central-monitor/logging"
	"github.com/byuoitav/central-monitor/monitors/common"
	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
	"github.com/fatih/color"
)

//RunCheck runs the check (imagine that)
func RunCheck(addr, hostname string) common.Report {
	//go through and check each endpoint
	ms := getMSToCheck()
	for _, curMS := range ms {
		log.Printf(color.HiBlueString("[Mstatus] check against device %v, microservice %v", hostname, curMS.Name))

		address := strings.Replace(curMS.Endpoint, "$ADDRESS", addr, 1)

		//set a timeout
		timeout := time.Duration(3 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}

		_, err := client.Get(address)
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {

				//it was a timeout error log it, run the resolution, log result
				LogEvent(ei.HEALTH, "System Error", "[MSTATUS] Timeout. Resolution Steps pending..", hostname)

				msg, err := curMS.GetResolution("timeout", address, curMS)()
				if err != nil && msg == "error" {

					//something happend where we couldn't carry out the resolution
					LogEvent(ei.HEALTH, "System Error", fmt.Sprintf("[MSTATUS] Problem running resolution: %v", err.Error()), hostname)
					LogEvent(ei.ERROR, "[MSTATUS]", fmt.Sprintf("Mstatus timed out for microservice: %v.", curMS.Name), hostname)

				} else if err != nil && msg == "failure" {

					//resolution failed
					LogEvent(ei.HEALTH, "System Error", fmt.Sprintf("[MSTATUS] Resolution Failed: %v", err.Error()), hostname)
					LogEvent(ei.ERROR, "[MSTATUS]", fmt.Sprintf("Mstatus timed out for microservice: %v.", curMS.Name), hostname)
				}

			} else {
				log.Printf("There was a problem with the request: %v", err.Error())

				//need to log the error - probably throw an alert?
				LogEvent(ei.ERROR, "[MSTATUS]", fmt.Sprintf("Problem running mstatus for microservice %v: %v", curMS.Name, err.Error()), hostname)
			}
		}

		log.Printf(color.HiBlueString("[Mstatus] check against device %v, microservice %v Done.", hostname, curMS.Name))
	}

	return common.Report{}
}

func LogEvent(etype ei.EventType, key, value, device string) error {
	info := ei.EventInfo{
		Type:           etype,
		Requestor:      os.Getenv("HOSTNAME"),
		EventCause:     ei.AWS,
		Device:         device,
		EventInfoKey:   key,
		EventInfoValue: value,
	}

	logging.Log(info, device)
	return nil
}
