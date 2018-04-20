package monitors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/byuoitav/authmiddleware/bearertoken"
	"github.com/byuoitav/central-monitor/elk"
	"github.com/byuoitav/central-monitor/monitors/common"
	"github.com/byuoitav/central-monitor/monitors/mstatus"
	"github.com/byuoitav/configuration-database-microservice/structs"
	"github.com/fatih/color"
)

func RunMStatus(branches []string) ([]common.Report, error) {
	//We assume that we want to run mstatus on all cp's - go get the list from the DB
	allDevs := []structs.Device{}

	for _, b := range branches {
		dev, err := GetDevicesByBranch(b, "ControlProcessor", "pi")
		if err != nil {
			log.Printf(color.HiRedString("Couldn't get devices for branch %v: %v", b, err.Error()))
			continue
		}
		allDevs = append(allDevs, dev...)
	}

	if len(allDevs) <= 0 {
		log.Printf(color.HiYellowString("No devices found. Aborting"))
		return []common.Report{}, nil
	}

	//We have the list of all possible devices: get the list of suppressed alerts from ELK.
	suppressedDevs, err := elk.GetNonSuppressedControlProcessors()
	if err != nil {
		log.Printf(color.HiRedString("[Mstatus] Couldn't get suppressed notification devices: error: %v.", err.Error()))
		return []common.Report{}, errors.New(fmt.Sprintf("[Mstatus] Couldn't get suppressed notification devices: error: %v.", err.Error()))

	}

	outChannel := make(chan common.Report, len(allDevs))
	wg := sync.WaitGroup{}
	//now we validate that the CP isn't suppressed - then start it running
	for _, dev := range allDevs {

		if _, ok := suppressedDevs[dev.GetFullName()]; !ok {
			log.Printf(color.HiCyanString("[resolution-%v] Device not in monitoring. Skipping.", dev.GetFullName()))
			continue
		}
		if _, ok := suppressedDevs[dev.GetFullName()[:strings.LastIndex(dev.GetFullName(), "-")]]; ok {
			log.Printf(color.HiCyanString("[resolution-%v] Device suppressed at room level. Skipping.", dev.GetFullName()))
			continue
		}

		wg.Add(1)

		//for now this will be fine, eventually we may want to limit the number of concurrent routines
		go func(dev structs.Device) {
			rep := mstatus.RunCheck(dev.Address, dev.GetFullName())
			outChannel <- rep
			wg.Done()
		}(dev)
	}
	wg.Wait()

	toReturn := []common.Report{}

	close(outChannel)

	for value := range outChannel {
		toReturn = append(toReturn, value)
	}

	return toReturn, nil
}

func GetDevicesByBranch(branch, deviceRole, deviceType string) ([]structs.Device, error) {

	client := &http.Client{}
	url := fmt.Sprintf("%v/deployment/devices/roles/%v/types/%v/%v", os.Getenv("CONFIGURATION_DATABASE_MICROSERVICE_ADDRESS"), deviceRole, deviceType, branch)

	log.Printf("Making request for all devices to: %v", url)
	req, _ := http.NewRequest("GET", url, nil)

	//get the bearer token
	token, err := bearertoken.GetToken()
	if err != nil {
		return []structs.Device{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)

	//make the request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error getting devices 1: %v", err.Error())
		return []structs.Device{}, err
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error getting devices 2: %v", err.Error())
		return []structs.Device{}, err
	}

	allDevices := []structs.Device{}
	err = json.Unmarshal(b, &allDevices)
	if err != nil {
		log.Printf("Error getting devices 3: %v", err.Error())
		return []structs.Device{}, err
	}

	return allDevices, nil
}
