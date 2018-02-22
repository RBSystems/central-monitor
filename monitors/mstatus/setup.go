package mstatus

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"

	"golang.org/x/crypto/ssh"
)

//putting this here in case in the future we want to do more robust checking of the response
//we can include a function in here to check the response and return a good/bad response - potentially the specific
//steps to resolution for that error
type msCheck struct {
	Name          string
	Endpoint      string
	GetResolution func(ErrorStr string, address string, ms msCheck) func() (string, error)
}

//returns the list of microservices to check
func getMSToCheck() []msCheck {
	return []msCheck{
		msCheck{"touchpanel-ui", "http://$ADDRESS:8888/mstatus", BaseResolution},
		msCheck{"control-api", "http://$ADDRESS:8000/mstatus", BaseResolution},
		msCheck{"event-router", "http://$ADDRESS:7000/mstatus", BaseResolution},
		msCheck{"translator", "http://$ADDRESS:6998/mstatus", BaseResolution},
		msCheck{"config-db", "http://$ADDRESS:8006/mstatus", BaseResolution},
	}
}

func BaseResolution(ErrorStr string, address string, ms msCheck) func() (string, error) {
	if ErrorStr != "timeout" && ErrorStr != "conn_refused" {
		//we don't know the reror, go ahead and report it
		return func() (string, error) {
			msg := fmt.Sprintf("Unknown Error String from MStatus: %v. Need to report", ErrorStr)
			return msg, errors.New(msg)
		}
	}

	//it's a timeout, so you'll need to ssh in and restart the container in question
	sshconfig := GetSSHConfig(address)
	return func() (string, error) {

		//we just need to restart the docker container
		log.Printf(color.HiGreenString("[resolution-%v] Starting Resolution function for issue %v, microservice %v, address %v", address, ErrorStr, ms.Name, address))
		conn, err := ssh.Dial("tcp", address+":22", sshconfig)
		if err != nil {
			log.Printf(color.HiRedString("[resolution-%v] Could not ssh into device, error: %v", address, err.Error()))
			return "error", errors.New(fmt.Sprintf("Error connecting to device %v over ssh. Error: %v", address, err.Error()))
		}

		session, err := conn.NewSession()
		if err != nil {
			log.Printf(color.HiRedString("[resolution-%v] Could not create ssh session, error: %v", address, err.Error()))
			return "error", errors.New(fmt.Sprintf("Error creating ssh session on device %v. Error: %v", address, err.Error()))
		}
		out, err := session.CombinedOutput(getRestartDockerCommandFromMS(ms.Name))
		if err != nil {
			cmd := getRestartDockerCommandFromMS(ms.Name)
			log.Printf(color.HiRedString("[resolution-%v] Error running command %v on host %v. Error: %v", cmd, address, err.Error()))
			return "error", errors.New(fmt.Sprintf("Error running command %v on host %v. Error: %v", cmd, address, err.Error()))
		}

		// check the output
		if strings.TrimSpace(string(out)) == getContainerNameFromMS(ms.Name) {
			log.Printf(color.HiGreenString("[resolution-%v] Container restarted.", address))
			return "resolved", nil
		}

		log.Printf(color.HiRedString("[resolution-%v] Resolution failed. Unexpected output: '%v'", address, string(out)))
		return "failure", errors.New(fmt.Sprintf("Unexpected response received: %s", out))
	}
}

func getContainerNameFromMS(name string) string {
	nametocontainer := map[string]string{}
	nametocontainer["touchpanel-ui"] = "tmp_touchpanel-ui-microservice_1"
	nametocontainer["control-api"] = "tmp_av-api_1"
	nametocontainer["event-router"] = "tmp_event-router-microservice_1"
	nametocontainer["translator"] = "tmp_event-translator-microservice_1"
	nametocontainer["config-db"] = "tmp_configuration-database-microservice_1"

	return nametocontainer[name]
}

func getRestartDockerCommandFromMS(name string) string {
	return fmt.Sprintf("docker restart %v", getContainerNameFromMS(name))
}

func GetSSHConfig(address string) *ssh.ClientConfig {
	return &ssh.ClientConfig{
		User: os.Getenv("PI_SSH_USERNAME"),
		Auth: []ssh.AuthMethod{
			ssh.Password(os.Getenv("PI_SSH_PASSWORD")),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         3 * time.Second,
	}
}
