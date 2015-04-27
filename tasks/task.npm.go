package tasks

import (
    "strings"
    "strconv"
    "runtime"
    "time"
    "os"
    "os/exec"
    "ghostrunner/utils"
    "ghostrunner/logging"
)

const (
    npmTimeout = 15
)
    
func RunNpmScript(config *utils.Configuration, taskId string, scriptId int, packageNames []string) (string) {
    logging.Debug("task.npm", "RunNpmScript", "Starting to process spm task");
    
    logging.Debug("task.npm", "RunNpmScript", "Checking operating system");

    if strings.ToLower(runtime.GOOS) != "linux" {
        return "Incorrect operating system"
    }

    logging.Debug("task.npm", "RunNpmScript", "System detected as Linux");

    npmLocations := strings.Split(config.NpmLocation, "|")
    npmFound := false

    for _, npmLocation := range npmLocations {
        if _, err := os.Stat(npmLocation); err == nil {
            npmFound = true
        }
    }

    if npmFound == false {
        return "npm is not installed"
    }

    taskFolderLocation := config.ProcessingLocation + "/" + taskId
    npmFileLocation := config.ProcessingLocation + "/" + taskId + "/" + taskId + "_" + strconv.Itoa(scriptId) + "_npm.sh"

    logging.Debug("task.npm", "RunNpmScript", "Checking for task folder location");

    if _, err := os.Stat(taskFolderLocation); os.IsNotExist(err) {
        logging.Debug("task.npm", "RunNpmScript", "Creating task folder  location");
        os.Mkdir(taskFolderLocation, 0777)
    }

    npmFile, err := os.Create(npmFileLocation)

    if err != nil {
        logging.Error("task.npm", "RunNpmScript", "Error creating npm script at " + npmFileLocation, err)
        return "Error creating npm script: " + err.Error()
    }
    
    defer npmFile.Close()
    //defer os.Remove(npmFileLocation)

    script := "#!/bin/bash\n"
    script += "npm config set registry http://registry.npmjs.org/\n"

    for _, npmPackage := range packageNames {
        script += "npm install " + npmPackage + "\n"
    }

    _, err = npmFile.Write([]byte(script))

    if err != nil {
        logging.Error("task.node", "RunNpmScript", "Error writing npm script to " + npmFileLocation, err)
        return "Error writing npm script: " + err.Error()
    }

    logging.Debug("task.npm", "RunNpmScript", "New script successfully created at " + npmFileLocation)

    scriptLog := ""

    cmd := exec.Command("sh", npmFileLocation)
    cmd.Dir = taskFolderLocation

    go func() {
        time.Sleep(npmTimeout * time.Minute)
        cmd.Process.Kill()

        scriptLog += "npm install timed out\n"
    }()

    output, err := cmd.Output()

    if err != nil {
        logging.Error("task.npm", "RunNpmScript", "Error running npm install", err)
        
        scriptLog += "Error running npm install: " + err.Error() + "\n"
    }

    scriptLog += string(output)

    return scriptLog
}