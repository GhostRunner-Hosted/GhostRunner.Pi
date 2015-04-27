package utils

import (
    "os"
    "io/ioutil"
    "encoding/json"
    "ghostrunner/logging"
    "errors"
)

type Configuration struct {
	HostUrl string
	ProcessingLocation string
    NodeLocation string
    NpmLocation	string
    RunnerId string
    ApiKey string
    ApiSecret string
	SessionId string
}

func LoadConfiguration() (Configuration, error) {
	if _, err := os.Stat("/etc/ghostrunner.conf"); os.IsNotExist(err) {
        	logging.Error("utils.configuration", "LoadConfiguration", "Unable to fing configuration file '/etc//ghostrunner.conf'", err)

		return Configuration{}, errors.New("Unable to fing configuration file '/etc/ghostrunner.conf'")
	}
	
	file, _ := os.Open("/etc/ghostrunner.conf")

	decoder := json.NewDecoder(file)

	configuration := Configuration{}

	err := decoder.Decode(&configuration)

	if err != nil {
		logging.Error("utils.configuration", "LoadConfiguration", "Error decoding configuration file", err)
	}

	if len(configuration.RunnerId) == 0 {
		runnerId, _ := GenerateUUID()
        	configuration.RunnerId = runnerId 

		UpdateConfiguration(&configuration) 
	}

	return configuration, nil
}

func UpdateConfiguration(config *Configuration) {
	stringifiedConfig, _ := json.Marshal(config)

    err := ioutil.WriteFile("/etc/ghostrunner.conf", []byte(stringifiedConfig), 0644)

    if err != nil {
        logging.Error("utils.configuration", "UpdateConfiguration", "Error writing configuration", err)

        return
    }
}
