package tasks

import (
    "strings"
    "runtime"
    "strconv"
    "time"
    "regexp"
    "os"
    "os/exec"
    "ghostrunner/utils"
    "ghostrunner/logging"
)

const (
    nodeTimeout = 10
)
    
func RunNodeScript(config *utils.Configuration, taskId, script string, scriptId int) (string, string) {
    logging.Debug("task.node", "RunNodeScript", "Starting to process node task");

    if strings.ToLower(runtime.GOOS) != "linux" { 
        return "Errored", "Incorrect operating system"
    }

    scriptLog := ""
    status := "Completed"
    
    logging.Debug("task.node", "RunNodeScript", "Checking operating system as this allows us to check for node");

    logging.Debug("task.node", "RunNodeScript", "System detected as Linux");

    nodeLocations := strings.Split(config.NodeLocation, "|")
    nodeFound := false

    for _, nodeLocation := range nodeLocations {
        if _, err := os.Stat(nodeLocation); err == nil {
            nodeFound = true
        }
    }

    if nodeFound == false {
        return "Errored", "node is not installed"
    }

    taskFolderLocation := config.ProcessingLocation + "/" + taskId
    nodeFileLocation := config.ProcessingLocation + "/" + taskId + "/" + taskId + "_" + strconv.Itoa(scriptId) + ".js"

    logging.Debug("task.node", "RunNodeScript", "Attempting to write the node file out to the processing location " + nodeFileLocation)

    if _, err := os.Stat(nodeFileLocation); err == nil {
        err := os.Remove(nodeFileLocation)

        if err != nil {
            logging.Error("task.node", "RunNodeScript", "Error deleting previous node script at " + nodeFileLocation, err)

            return "Errored", "Error deleting previous node script"
        }
    }

    logging.Debug("task.node", "RunNodeScript", "Checking for processing location");

    if _, err := os.Stat(config.ProcessingLocation); os.IsNotExist(err) {
        logging.Debug("task.node", "RunNodeScript", "Creating processing location");
        os.Mkdir(config.ProcessingLocation, 0777)
    }

    logging.Debug("task.node", "RunNodeScript", "Checking for task processing location");

    if _, err := os.Stat(taskFolderLocation); os.IsNotExist(err) {
        logging.Debug("task.node", "RunNodeScript", "Creating task processing location");
        os.Mkdir(taskFolderLocation, 0777)
    }

    nodeFile, err := os.Create(nodeFileLocation)

    if err != nil {
        logging.Error("task.node", "RunNodeScript", "Error creating node script at " + nodeFileLocation, err)
        return "Errored", "Error creating node script"
    }
    
    defer nodeFile.Close()
    defer os.Remove(nodeFileLocation)

    _, err = nodeFile.Write([]byte(script))

    if err != nil {
        logging.Error("task.node", "RunNodeScript", "Error writing node script to " + nodeFileLocation, err)
        return "Errored", "Error writing node script"
    }

    logging.Debug("task.node", "RunNodeScript", "New script successfully craeted at " + nodeFileLocation)

    scriptLog += loadNodeNpmPackages(config, taskId, script, scriptId)

    cmd := exec.Command("node", nodeFileLocation)
    cmd.Dir = taskFolderLocation

    go func() {
        time.Sleep(nodeTimeout * time.Minute)
        cmd.Process.Kill()

        logging.Error("task.node", "RunNodeScript", scriptLog + "Node script timed out " + nodeFileLocation, err)
        
        status = "Errored"
    }()

    output, err := cmd.Output()

    if err != nil {
        logging.Error("task.node", "RunNodeScript", "Error running node script", err)
        
        return "Errored", scriptLog + "Error running node script: " + err.Error()
    }

    scriptLog += string(output)

    return status, scriptLog
}

func loadNodeNpmPackages(config *utils.Configuration, taskId, scriptContent string, scriptId int) (string){
    requireRegExp, _ := regexp.Compile("require\\(\".*?\"\\)")
    npmRegExp, _ := regexp.Compile("\".*?\"")

    var npmPackages []string

    for _, require := range requireRegExp.FindAllString(scriptContent, -1) {
        for _, rawNpmPackage := range npmRegExp.FindAllString(require, -1) {
            npmPackage := strings.Trim(rawNpmPackage, "\"")

            npmPackages = append(npmPackages, npmPackage)
        }
    }

    npmInstallLog := RunNpmScript(config, taskId, scriptId, npmPackages)
    npmInstallLog += "\n"

    return npmInstallLog
}
