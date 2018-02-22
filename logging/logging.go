package logging

import (
	"github.com/byuoitav/central-monitor/elk"
	ei "github.com/byuoitav/event-router-microservice/eventinfrastructure"
)

//this is an entry point that could log to mulitple places
func Log(info ei.EventInfo, device string) {
	elk.SendELKEvent(info, device)
}
