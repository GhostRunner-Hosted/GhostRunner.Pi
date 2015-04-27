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
    commandTimeout = 20
)

type GitRepository struct {
    Location string `json:"location"`
    Username string `json:"username"`
    Password string `json:"password"`
}

func RunCommandLineScript(config *utils.Configuration, taskId, script string, scriptId int) (string, string) {
    logging.Debug("task.commandLine", "RunCommandLineScript", "Starting to process commandLine task");
    
    if strings.ToLower(runtime.GOOS) != "linux" {
        return "Errored", "Incorrect operating system"
    }

    status := "Completed"

    var shellScriptName string
    var shellFileLocation string

    logging.Debug("task.commandLine", "RunCommandLineScript", "System detected as Linux")

    shellScriptName = taskId + "_" + strconv.Itoa(scriptId) + ".sh"

    script = strings.TrimSpace(script)

    if !strings.HasPrefix(strings.ToLower(script), "#!/bin/bash") {
        script = "#!/bin/bash\n" + script
    }

    taskFolderLocation := config.ProcessingLocation + "/" + taskId

    if _, err := os.Stat(taskFolderLocation); os.IsNotExist(err) {
        logging.Debug("task.commandLine", "RunCommandLineScript", "Creating task processing location")
        err := os.Mkdir(taskFolderLocation, 0777)

        if err != nil {
            logging.Error("task.commandLine", "RunCommandLineScript", "Error creating task processing location " + taskFolderLocation, err)
        }
    }

    shellFileLocation = config.ProcessingLocation + "/" + taskId + "/" + shellScriptName

    if _, err := os.Stat(shellFileLocation); err == nil {
        err := os.Remove(shellFileLocation)

        if err != nil {
            logging.Error("task.commandLine", "RunCommandLineScript", "Error deleting previous shell script at " + shellFileLocation, err)

            return "Errored", "Error deleting previous shell script: " + err.Error()
        }
    }

    shellFile, err := os.Create(shellFileLocation)

    if err != nil {
        logging.Error("task.commandLine", "RunCommandLineScript", "Error creating shell script at " + shellFileLocation, err)
        
        return "Errored", "Error creating shell script: " + err.Error()
    }
    
    defer shellFile.Close()
    defer os.Remove(shellFileLocation)

    _, err = shellFile.Write([]byte(script))

    if err != nil {
        logging.Error("task.commandLine", "RunCommandLineScript", "Error writing shell script to " + shellFileLocation, err)
        
        return "Errored", "Error writing shell script: " + err.Error()
    }

    cmd := exec.Command("sh", shellFileLocation)
    cmd.Dir = taskFolderLocation

    go func() {
        time.Sleep(commandTimeout * time.Minute)
        cmd.Process.Kill()

        logging.Error("task.commandLine", "RunCommandLineScript", "Command line script timed out", nil)
        
        status = "Errored"
    }()

	scriptLog, err := cmd.Output()

    if err != nil {
        logging.Error("task.commandLine", "RunCommandLineScript", "Error running shell script", err)
        
        return "Errored", "Error running shell script + " + err.Error()
    }

    return status, string(scriptLog)
}
